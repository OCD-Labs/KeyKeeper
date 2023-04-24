package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	db "github.com/OCD-Labs/KeyKeeper/db/sqlc"
	"github.com/OCD-Labs/KeyKeeper/internal/utils"
	"github.com/OCD-Labs/KeyKeeper/internal/worker"
	"github.com/go-playground/validator/v10"
	"github.com/hibiken/asynq"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

type createUserRequest struct {
	FirstName       string  `json:"first_name" validate:"required,min=1"`
	LastName        string  `json:"last_name" validate:"required,min=1"`
	Password        string  `json:"password" validate:"required,min=8"`
	Email           string  `json:"email" validate:"required,email"`
	ProfileImageUrl *string `json:"profile_image_url"`
}

type userResponse struct {
	ID int64 `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
	Email string `json:"email"`
	ProfileImageurl string `json:"profil_image_url"`
	CreatedAt         time.Time `json:"created_at"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	IsActive bool `json:"is_active"`
	IsEmailVerified bool `json:"is_email_verified"`
}

func newUserResponse(user db.User) userResponse {
	fields := strings.Fields(user.FullName)

	return userResponse{
		ID:          user.ID,
		FirstName:          fields[0],
		LastName:          fields[1],
		Email:             user.Email,
		ProfileImageurl: user.ProfileImageUrl.String,
		CreatedAt:         user.CreatedAt,
		PasswordChangedAt: user.PasswordChangedAt,
		IsActive: user.IsActive,
		IsEmailVerified: user.IsEmailVerified,
	}
}

func (app *KeyKeeper) createUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := app.readJSON(w, r, &req); err != nil {
		werr := fmt.Errorf("failed to parse request body: %w", err)
		app.errorResponse(w, r, http.StatusBadRequest, werr.Error())
		log.Error().Err(werr)
		return
	}

	if err := app.bindJSONWithValidation(w, r, req, validator.New()); err != nil {
		return
	}

	hashedPassword, err := utils.HashedPassword(req.Password)
	if err != nil {
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to hash password")
		log.Error().Err(err).Msg("failed to hash password")
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

	resp, err := app.store.CreateUserTx(r.Context(), arg)
	if err != nil {
		if pqError, ok := err.(*pq.Error); ok {
			switch pqError.Code.Name() {
			case "unique_violation":
				app.errorResponse(w, r, http.StatusForbidden, "user already exist")
				log.Error().Err(err).Msg("username already exist")
				return
			}
		}
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to create user")
		log.Error().Err(err).Msg("failed to create user")
		return
	}

	app.writeJSON(
		w, 
		http.StatusCreated, 
		envelop{"data": newUserResponse(resp.User)}, 
		nil,
	)
}
