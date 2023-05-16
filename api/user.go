package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	db "github.com/OCD-Labs/KeyKeeper/db/sqlc"
	"github.com/OCD-Labs/KeyKeeper/internal/token"
	"github.com/OCD-Labs/KeyKeeper/internal/utils"
	"github.com/OCD-Labs/KeyKeeper/internal/worker"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

type createUserRequest struct {
	FirstName       string  `json:"first_name" validate:"required,min=1"`
	LastName        string  `json:"last_name" validate:"required,min=1"`
	Password        string  `json:"password" validate:"required,min=8"`
	Email           string  `json:"email" validate:"required,email"`
	ProfileImageUrl *string `json:"profile_image_url"`
}

type userResponse struct {
	ID                int64     `json:"user_id"`
	FirstName         string    `json:"first_name"`
	LastName          string    `json:"last_name"`
	Email             string    `json:"email"`
	ProfileImageurl   string    `json:"profil_image_url"`
	CreatedAt         time.Time `json:"created_at"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	IsActive          bool      `json:"is_active"`
	IsEmailVerified   bool      `json:"is_email_verified"`
}

func newUserResponse(user db.User) userResponse {
	fields := strings.Fields(user.FullName)

	return userResponse{
		ID:                user.ID,
		FirstName:         fields[0],
		LastName:          fields[1],
		Email:             user.Email,
		ProfileImageurl:   user.ProfileImageUrl.String,
		CreatedAt:         user.CreatedAt,
		PasswordChangedAt: user.PasswordChangedAt,
		IsActive:          user.IsActive,
		IsEmailVerified:   user.IsEmailVerified,
	}
}

func (app *KeyKeeper) createUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := app.readJSON(w, r, &req); err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, "failed to parse request body")
		log.Error().Err(err).Msg("error occurred")
		return
	}

	if err := app.bindJSONWithValidation(w, r, req, validator.New()); err != nil {
		return
	}

	hashedPassword, err := utils.HashedPassword(req.Password)
	if err != nil {
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to hash password")
		log.Error().Err(err).Msg("error occurred")
		return
	}

	createUserArg := db.CreateUserParams{
		FullName:       fmt.Sprintf("%s %s", req.FirstName, req.LastName),
		HashedPassword: hashedPassword,
		Email:          req.Email,
	}

	if req.ProfileImageUrl != nil {
		createUserArg.ProfileImageUrl.String = *req.ProfileImageUrl
		createUserArg.ProfileImageUrl.Valid = true
	}

	arg := db.CreateUserTxParams{
		CreateUserParams: createUserArg,
		AfterCreate: func(user db.User) error {
			taskPayload := &worker.PayloadSendVerifyEmail{
				UserID:    user.ID,
				ClientIp:  r.RemoteAddr,
				UserAgent: r.UserAgent(),
			}

			opts := []asynq.Option{
				asynq.MaxRetry(10),
				asynq.ProcessIn(10 * time.Second),
				asynq.Queue(worker.QueueCritical),
			}

			err := app.taskDistributor.DistributeTaskSendVerifyEmail(r.Context(), taskPayload, opts...)
			return err
		},
	}

	_, err = app.store.CreateUserTx(r.Context(), arg)
	if err != nil {
		if pqError, ok := err.(*pq.Error); ok {
			switch pqError.Code.Name() {
			case "unique_violation":
				app.errorResponse(w, r, http.StatusForbidden, "user already exist")
				log.Error().Err(err).Msg("error occurred")
				return
			}
		}
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to create user")
		log.Error().Err(err).Msg("error occurred")
		return
	}

	app.writeJSON(
		w,
		http.StatusCreated,
		envelop{"result": "new user created successfully"},
		nil,
	)
}

type verifyEmailRequest struct {
	Email      string `json:"email" validate:"required,email"`
	SecretCode string `json:"secret_code" validate:"required"`
}

func (app *KeyKeeper) verifyEmail(w http.ResponseWriter, r *http.Request) {
	queryMap := r.URL.Query()
	var req verifyEmailRequest
	req.Email = app.readStr(queryMap, "email", "")
	req.SecretCode = app.readStr(queryMap, "secret_code", "")

	if err := app.bindJSONWithValidation(w, r, &req, validator.New()); err != nil {
		return
	}

	tokenExists, err := app.store.CheckSessionExistence(r.Context(), req.SecretCode)
	if err != nil {
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to check secret code")
		log.Error().Err(err).Msg("error occurred")
		return
	}

	if !tokenExists {
		app.errorResponse(w, r, http.StatusInternalServerError, "invalid secret code")
		return
	}

	payload, err := app.tokenMaker.VerifyToken(utils.Concat(req.SecretCode))
	if err != nil {
		switch {
		case errors.Is(err, token.ErrExpiredToken):
			app.errorResponse(w, r, http.StatusBadRequest, token.ErrExpiredToken.Error())
		case errors.Is(err, token.ErrInvalidToken):
			app.errorResponse(w, r, http.StatusBadRequest, token.ErrInvalidToken.Error())
		default:
			app.errorResponse(w, r, http.StatusInternalServerError, "failed to verify secret code")
		}

		log.Error().Err(err).Msg("error occurred")
		return
	}

	exists, err := app.cache.IsSessionBlacklisted(r.Context(), payload.ID.String())
	if err != nil || exists {
		app.errorResponse(w, r, http.StatusUnauthorized, "invalid token")
		return
	}

	arg := db.UpdateUserParams{
		IsEmailVerified: sql.NullBool{
			Bool:  true,
			Valid: true,
		},
		IsActive: sql.NullBool{
			Bool:  true,
			Valid: true,
		},
		Email: sql.NullString{
			String: req.Email,
			Valid:  true,
		},
	}

	_, err = app.store.UpdateUser(r.Context(), arg)
	if err != nil {
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to verify user's email")
		log.Error().Err(err).Msg("error occurred")
		return
	}

	expiredAt := payload.ExpiredAt
	duration := time.Until(expiredAt)

	err = app.cache.BlacklistSession(r.Context(), payload.ID.String(), duration)
	if err != nil {
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to blacklist access token")
		log.Error().Err(err).Msg("error occurred")
		return
	}

	app.writeJSON(
		w,
		http.StatusOK,
		envelop{"result": "user verified successfully"},
		nil,
	)
}

type resendVerifyEmailRequest struct {
	UserID int `json:"user_id" validate:"required,min=1"`
}

func (app *KeyKeeper) resendVerifyEmail(w http.ResponseWriter, r *http.Request) {
	queryMap := r.URL.Query()

	var req resendVerifyEmailRequest
	var err error

	req.UserID, err = app.readInt(queryMap, "user_id", 0)
	if err != nil || req.UserID == 0 {
		app.errorResponse(w, r, http.StatusBadRequest, "user id is not valid")
		return
	}

	if err := app.bindJSONWithValidation(w, r, &req, validator.New()); err != nil {
		return
	}

	taskPayload := &worker.PayloadSendVerifyEmail{
		UserID:    int64(req.UserID),
		ClientIp:  r.RemoteAddr,
		UserAgent: r.UserAgent(),
	}

	opts := []asynq.Option{
		asynq.MaxRetry(10),
		asynq.ProcessIn(10 * time.Second),
		asynq.Queue(worker.QueueCritical),
	}

	err = app.taskDistributor.DistributeTaskSendVerifyEmail(r.Context(), taskPayload, opts...)
	if err != nil {
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to resend verify email")
		log.Error().Err(err).Msg("error occurred")
		return
	}

	app.writeJSON(w, http.StatusOK, envelop{"result": "new verify email sent"}, nil)
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

func (app *KeyKeeper) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := app.readJSON(w, r, &req); err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, "failed to parse request body")
		log.Error().Err(err).Msg("error occurred")
		return
	}

	if err := app.bindJSONWithValidation(w, r, &req, validator.New()); err != nil {
		return
	}

	user, err := app.store.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.errorResponse(w, r, http.StatusNotFound, "user not found")
		default:
			app.errorResponse(w, r, http.StatusInternalServerError, "failed to fetch user's profile")
		}
		log.Error().Err(err).Msg("error occurred")
		return
	}

	if !user.IsEmailVerified {
		newReq, err := http.NewRequest(http.MethodPost, "/api/v1/resend_email_verification", r.Body)
		if err != nil {
			app.errorResponse(w, r, http.StatusInternalServerError, "failed to resend email verification mail")
			log.Error().Err(err).Msg("error occurred")
			return
		}

		for key, value := range r.Header {
			newReq.Header.Set(key, value[0])
		}

		http.Redirect(w, r, fmt.Sprintf("/api/v1/resend_email_verification?user_id=%d", user.ID), http.StatusTemporaryRedirect)
	}

	if !user.IsActive {
		app.errorResponse(w, r, http.StatusNoContent, "user is not activated")
		return
	}

	err = utils.VerifyPassword(user.HashedPassword, req.Password)
	if err != nil {
		app.errorResponse(w, r, http.StatusUnauthorized, "Invalid login credentials")
		log.Error().Err(err).Msg("error occurred")
		return
	}

	token, _, err := app.tokenMaker.CreateToken(24*time.Hour, user.ID)
	if err != nil {
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to generate access token")
		log.Error().Err(err).Msg("error occurred")
		return
	}

	app.writeJSON(
		w,
		http.StatusOK,
		envelop{"data": envelop{
			"user":         newUserResponse(user),
			"access_token": token,
		}},
		nil,
	)
}

func (app *KeyKeeper) logout(w http.ResponseWriter, r *http.Request) {
	authPayload := app.contextGetToken(r)

	expiredAt := authPayload.ExpiredAt
	duration := time.Until(expiredAt)

	err := app.cache.BlacklistSession(r.Context(), authPayload.ID.String(), duration)
	if err != nil {
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to blacklist access token")
		log.Error().Err(err).Msg("error occurred")
		return
	}

	app.writeJSON(w, http.StatusOK, envelop{"result": "Logged out user successfully"}, nil)
}

// Define the Google Sign-in route handler
func (app *KeyKeeper) googleLogin(w http.ResponseWriter, r *http.Request) {
	url := app.googleConfig.AuthCodeURL(app.configs.GoogleRandomString, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

type profile struct {
	Email         string `json:"email"`
	Name          string `json:"name"`
	FirstName     string `json:"given_name"`
	LastName      string `json:"family_name"`
	EmailVerified bool   `json:"email_verified"`
}

// Define the Google Sign-in callback route handler
func (app *KeyKeeper) googleLoginCallback(w http.ResponseWriter, r *http.Request) {
	queryMap := r.URL.Query()

	// Check state is valid.
	state := queryMap.Get("state")
	if state != app.configs.GoogleRandomString {
		app.errorResponse(w, r, http.StatusUnauthorized, "invalid state value")
		return
	}

	// Exchange the authorization code for an access token and ID token
	code := queryMap.Get("code")
	token, err := app.googleConfig.Exchange(r.Context(), code)
	if err != nil {
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to exchange code")
		log.Error().Err(err).Msg("error occurred")
		return
	}

	// Get the user's Google profile using the access token
	client := app.googleConfig.Client(r.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		app.errorResponse(w, r, http.StatusNotFound, "failed to get user info")
		log.Error().Err(err).Msg("error occurred")
		return
	}
	defer resp.Body.Close()

	// Parse the user's profile JSON
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to read Google response")
		log.Error().Err(err).Msg("error occurred")
		return
	}
	userProfile := &profile{}
	if err := json.Unmarshal(body, userProfile); err != nil {
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to parse user profile")
		log.Error().Err(err).Msg("error occurred")
		return
	}

	// Retrieve user by email
	user, err := app.store.GetUserByEmail(r.Context(), userProfile.Email)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.errorResponse(w, r, http.StatusNotFound, "user not found")
		default:
			app.errorResponse(w, r, http.StatusInternalServerError, "failed to fetch user's profile")
		}
		log.Error().Err(err).Msg("error occurred")
		return
	}

	pasetoToken, _, err := app.tokenMaker.CreateToken(24*time.Hour, user.ID)
	if err != nil {
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to generate access token")
		log.Error().Err(err).Msg("error occurred")
		return
	}

	app.writeJSON(
		w,
		http.StatusOK,
		envelop{"data": envelop{
			"user":         newUserResponse(user),
			"access_token": pasetoToken,
		}},
		nil,
	)
}

type deactivateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type deactivateUserPathVariable struct {
	ID int64 `json:"id" validate:"required,min=1"`
}

func (app *KeyKeeper) deactivateUser(w http.ResponseWriter, r *http.Request) {
	authPayload := app.contextGetToken(r)

	var pathVar deactivateUserPathVariable
	var err error

	pathVar.ID, err = app.retrieveIDParam(r)
	if err != nil || pathVar.ID == 0 {
		app.errorResponse(w, r, http.StatusBadRequest, "invalid user id")
		return
	}

	if pathVar.ID != authPayload.UserID {
		app.errorResponse(w, r, http.StatusUnauthorized, "mismatched user")
		return
	}

	var req deactivateUserRequest
	if err := app.readJSON(w, r, &req); err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, "failed to parse request")
		log.Error().Err(err).Msg("error occurred")
		return
	}

	if err := app.bindJSONWithValidation(w, r, &req, validator.New()); err != nil {
		return
	}

	user, err := app.store.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.errorResponse(w, r, http.StatusNotFound, "user not found")
		default:
			app.errorResponse(w, r, http.StatusInternalServerError, "failed to fetch user's profile")
		}
		log.Error().Err(err).Msg("error occurred")
		return
	}

	err = utils.VerifyPassword(user.HashedPassword, req.Password)
	if err != nil {
		app.errorResponse(w, r, http.StatusUnauthorized, "Invalid login credentials")
		log.Error().Err(err).Msg("error occurred")
		return
	}

	arg := db.UpdateUserParams{
		ID: sql.NullInt64{
			Int64: authPayload.UserID,
			Valid: true,
		},
		Email: sql.NullString{
			String: req.Email,
			Valid:  true,
		},
		IsActive: sql.NullBool{
			Bool:  false,
			Valid: true,
		},
	}
	_, err = app.store.UpdateUser(r.Context(), arg)
	if err != nil {
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to deactivate user")
		return
	}

	expiredAt := authPayload.ExpiredAt
	duration := time.Until(expiredAt)

	err = app.cache.BlacklistSession(r.Context(), authPayload.ID.String(), duration)
	if err != nil {
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to blacklist access token")
		log.Error().Err(err).Msg("error occurred")
		return
	}

	app.writeJSON(w, http.StatusOK, envelop{"result": "user successfully deleted"}, nil)
}

type changeUserPasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required,min=8"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_new_password" validate:"required,min=8,eqfield=NewPassword"`
}

type changeUserPasswordPathVariable struct {
	ID int64 `json:"id" validate:"required,min=1"`
}

func (app *KeyKeeper) changeUserPassword(w http.ResponseWriter, r *http.Request) {
	var pathVar changeUserPasswordPathVariable
	var err error
	pathVar.ID, err = app.retrieveIDParam(r)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, "mismatched user")
		return
	}
	if err := app.bindJSONWithValidation(w, r, &pathVar, validator.New()); err != nil {
		return
	}

	authPayload := app.contextGetToken(r)
	if pathVar.ID != authPayload.UserID {
		app.errorResponse(w, r, http.StatusUnauthorized, "mismatched user")
		return
	}

	var req changeUserPasswordRequest
	if err := app.bindJSONWithValidation(w, r, &req, validator.New()); err != nil {
		return
	}

	user, err := app.store.GetUser(r.Context(), authPayload.UserID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.errorResponse(w, r, http.StatusNotFound, "user not found")
		default:
			app.errorResponse(w, r, http.StatusInternalServerError, "failed to fetch user's profile")
		}
		log.Error().Err(err).Msg("error occurred")
		return
	}

	err = utils.VerifyPassword(req.CurrentPassword, user.HashedPassword)
	if err != nil {
		app.errorResponse(w, r, http.StatusUnauthorized, "invalid login credentials")
		log.Error().Err(err).Msg("error occurred")
		return
	}

	hashedPassword, err := utils.HashedPassword(req.NewPassword)
	if err != nil {
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to hash password")
		log.Error().Err(err).Msg("error occurred")
		return
	}

	updateUserParams := db.UpdateUserParams{
		HashedPassword: sql.NullString{
			String: hashedPassword,
			Valid:  true,
		},
		PasswordChangedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		ID: sql.NullInt64{
			Int64: pathVar.ID,
			Valid: true,
		},
	}

	_, err = app.store.UpdateUser(r.Context(), updateUserParams)
	if err != nil {
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to update user's password")
		log.Error().Err(err).Msg("error occurred")
		return
	}

	app.writeJSON(w, http.StatusOK, envelop{"result": "password updated successfully"}, nil)
}

type forgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

func (app *KeyKeeper) forgotPassword(w http.ResponseWriter, r *http.Request) {
	var req forgotPasswordRequest
	if err := app.readJSON(w, r, &req); err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, "failed to parse request")
		log.Error().Err(err).Msg("error occurred")
		return
	}

	if err := app.bindJSONWithValidation(w, r, &req, validator.New()); err != nil {
		return
	}

	user, err := app.store.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.errorResponse(w, r, http.StatusNotFound, "user not found")
		default:
			app.errorResponse(w, r, http.StatusInternalServerError, "failed to fetch user's profile")
		}
		log.Error().Err(err).Msg("error occurred")
		return
	}

	uuid, err := uuid.NewRandom()
	if err != nil {
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to create reset password session")
		log.Error().Err(err).Msg("error occurred")
		return
	}

	token, payload, err := app.tokenMaker.CreateToken(25*time.Minute, user.ID)
	if err != nil {
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to create reset password session")
		log.Error().Err(err).Msg("error occurred")
		return
	}

	_, err = app.store.CreateSession(r.Context(), db.CreateSessionParams{
		ID:        uuid,
		UserID:    user.ID,
		Token:     utils.Extract(token),
		Scope:     "reset_password",
		ClientIp:  r.RemoteAddr,
		UserAgent: r.UserAgent(),
		IsBlocked: false,
		ExpiresAt: payload.ExpiredAt,
	})
	if err != nil {
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to create reset password session")
		log.Error().Err(err).Msg("error occurred")
		return
	}

	task := &worker.PayloadSendResetPasswordEmail{
		SessionID: uuid,
		UserEmail: user.Email,
	}
	opts := []asynq.Option{
		asynq.MaxRetry(10),
		asynq.ProcessIn(5 * time.Second),
		asynq.Queue(worker.QueueCritical),
	}
	err = app.taskDistributor.DistributeTaskSendResetPasswordEmail(r.Context(), task, opts...)
	if err != nil {
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to create reset password task")
		log.Error().Err(err).Msg("error occurred")
		return
	}

	app.writeJSON(w, http.StatusOK, envelop{"result": "reset password email sent"}, nil)
}

type resetUserPassword struct {
	NewPassword     string `json:"new_password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_new_password" validate:"required,min=8,eqfield=NewPassword"`
}

type resetUserPasswordQueryStr struct {
	ResetToken string `json:"reset_token" validate:"required"`
}

func (app *KeyKeeper) resetUserPassword(w http.ResponseWriter, r *http.Request) {
	queryMap := r.URL.Query()
	var reqQueryStr resetUserPasswordQueryStr
	var err error

	reqQueryStr.ResetToken = app.readStr(queryMap, "id", "")
	if reqQueryStr.ResetToken == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "invalid reset token")
		return
	}

	payload, err := app.tokenMaker.VerifyToken(utils.Concat(reqQueryStr.ResetToken))
	if err != nil {
		switch {
		case errors.Is(err, token.ErrExpiredToken):
			app.errorResponse(w, r, http.StatusBadRequest, token.ErrExpiredToken.Error())
		case errors.Is(err, token.ErrInvalidToken):
			app.errorResponse(w, r, http.StatusBadRequest, token.ErrInvalidToken.Error())
		default:
			app.errorResponse(w, r, http.StatusInternalServerError, "failed to verify secret code")
		}

		log.Error().Err(err).Msg("error occurred")
		return
	}

	var reqBody resetUserPassword
	if err := app.readJSON(w, r, &reqBody); err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, "failed to parse request body")
		log.Error().Err(err).Msg("error occurred")
		return
	}

	if err := app.bindJSONWithValidation(w, r, &reqBody, validator.New()); err != nil {
		return
	}

	hashedPassword, err := utils.HashedPassword(reqBody.NewPassword)
	if err != nil {
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to hash password")
		log.Error().Err(err).Msg("error occurred")
		return
	}

	arg := db.UpdateUserParams{
		HashedPassword: sql.NullString{
			String: hashedPassword,
			Valid:  true,
		},
		PasswordChangedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		ID: sql.NullInt64{
			Int64: payload.UserID,
			Valid: true,
		},
	}

	_, err = app.store.UpdateUser(r.Context(), arg)
	if err != nil {
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to reset user's password")
		log.Error().Err(err).Msg("error occurred")
		return
	}

	app.writeJSON(w, http.StatusOK, envelop{"result": "password updated successfully"}, nil)
}

type getUserPathVariable struct {
	ID int64 `json:"id" validate:"required,min=1"`
}

func (app *KeyKeeper) getUser(w http.ResponseWriter, r *http.Request) {
	authPayload := app.contextGetToken(r)
	var pathVar getUserPathVariable
	var err error

	pathVar.ID, err = app.retrieveIDParam(r)
	if err != nil || pathVar.ID == 0 {
		app.errorResponse(w, r, http.StatusBadRequest, "invalid user id")
		return
	}

	if pathVar.ID != authPayload.UserID {
		app.errorResponse(w, r, http.StatusUnauthorized, "mismatched user")
		return
	}

	user, err := app.store.GetUser(r.Context(), authPayload.UserID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.errorResponse(w, r, http.StatusNotFound, "user not found")
		default:
			app.errorResponse(w, r, http.StatusInternalServerError, "failed to fetch user's profile")
		}
		log.Error().Err(err).Msg("error occurred")
		return
	}

	app.writeJSON(w, http.StatusOK, envelop{"data": envelop{"user": newUserResponse(user)}}, nil)
}
