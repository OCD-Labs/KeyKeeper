package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	db "github.com/OCD-Labs/KeyKeeper/db/sqlc"
	"github.com/OCD-Labs/KeyKeeper/internal/pagination"
	"github.com/OCD-Labs/KeyKeeper/internal/types"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
	"github.com/tabbed/pqtype"
)

type createReminderRequest struct {
	WebsiteUrl string         `json:"website_url" validate:"required"`
	Interval   types.Interval `json:"interval" vslidate:"required"`
}

type createReminderResponse struct {
	db.Reminder
	Extension string `json:"extension"`
}

func newReminderResponse(reminder db.Reminder) (*createReminderResponse, error) {
	val, err := reminder.Extension.Value()
	if err != nil {
		return nil, err
	}

	var buf []byte
	var ok bool

	buf, ok = val.([]byte)
	if !ok {
		buf, ok = val.([]uint8)
		if !ok {
			return nil, fmt.Errorf("mismatched type")
		}
	}

	return &createReminderResponse{
		Reminder:  reminder,
		Extension: string(buf),
	}, nil
}

func (app *KeyKeeper) createReminder(w http.ResponseWriter, r *http.Request) {
	var req createReminderRequest
	if err := app.readJSON(w, r, &req); err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, "failed to parse request")
		log.Error().Err(err).Msg("error occurred")
		return
	}

	if err := app.bindJSONWithValidation(w, r, &req, validator.New()); err != nil {
		return
	}

	authPayload := app.contextGetToken(r)

	reminder, err := app.store.CreateReminder(r.Context(), db.CreateReminderParams{
		UserID:     authPayload.UserID,
		WebsiteUrl: req.WebsiteUrl,
		Interval:   int64(req.Interval),
		Extension: pqtype.NullRawMessage{
			RawMessage: []byte(fmt.Sprintf(`{"get_email_notifications": %t}`, false)),
			Valid:      true,
		},
	})
	if err != nil { // TODO: Handle error due to Postgres constraints
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to create new reminder")
		log.Error().Err(err).Msg("error occurred")
		return
	}

	resp, err := newReminderResponse(reminder)
	if err != nil {
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to restructure reminder")
		return
	}

	app.writeJSON(w, http.StatusOK, envelop{"data": envelop{"reminder": resp}}, nil)
}

type getReminderPathVariable struct {
	ID int64 `json:"id" validate:"required,min=1"`
}

func (app *KeyKeeper) getReminder(w http.ResponseWriter, r *http.Request) {
	var pathVar getReminderPathVariable
	var err error

	pathVar.ID, err = app.retrieveIDParam(r)
	if err != nil || pathVar.ID == 0 {
		app.errorResponse(w, r, http.StatusBadRequest, "invalid reminder id")
		return
	}

	reminder, err := app.store.GetReminder(r.Context(), pathVar.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.errorResponse(w, r, http.StatusNotFound, "reminder not found")
		default:
			app.errorResponse(w, r, http.StatusInternalServerError, "failed to fetch reminder")
		}
		log.Error().Err(err).Msg("error occurred")
		return
	}

	resp, err := newReminderResponse(reminder)
	if err != nil {
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to restructure reminder")
		return
	}

	app.writeJSON(w, http.StatusOK, envelop{"data": resp}, nil)
}

type listRemindersQueryStr struct {
	WebsiteURL string `json:"website_url"`
	Page       int    `json:"page" validate:"min=1,max=10000000"`
	PageSize   int    `json:"page_size" validate:"min=1,max=20"`
	Sort       string `json:"sort"`
}

func (app *KeyKeeper) listReminders(w http.ResponseWriter, r *http.Request) {
	queryStr := r.URL.Query()
	var reqQueryStr listRemindersQueryStr

	reqQueryStr.WebsiteURL = app.readStr(queryStr, "website_url", "")
	reqQueryStr.Sort = app.readStr(queryStr, "sort", "")

	reqQueryStr.Page, _ = app.readInt(queryStr, "page", 1)
	reqQueryStr.PageSize, _ = app.readInt(queryStr, "page_size", 15)

	if err := app.bindJSONWithValidation(w, r, &reqQueryStr, validator.New()); err != nil {
		return
	}

	arg := db.ListRemindersParamsX{
		WebsiteURL: reqQueryStr.WebsiteURL,
		Filters: pagination.Filters{
			Page:         reqQueryStr.Page,
			PageSize:     reqQueryStr.PageSize,
			Sort:         reqQueryStr.Sort,
			SortSafelist: []string{"id", "website_url", "interval", "updated_at", "-id", "-website_url", "-interval", "-updated_at"},
		},
	}

	reminders, metadata, err := app.store.ListRemindersX(r.Context(), arg)
	if err != nil {
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to retrieve reminders")
		log.Error().Err(err).Msg("error occurred")
		return
	}

	resp := []*createReminderResponse{}
	for _, v := range reminders {
		var r *createReminderResponse
		r, err = newReminderResponse(v)
		if err != nil {
			continue
		}
		resp = append(resp, r)
	}

	if err != nil {
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to restructure reminder")
		log.Error().Err(err).Msg("error occurred")
		return
	}

	app.writeJSON(w, http.StatusOK, envelop{"data": envelop{"reminders": resp, "metadata": metadata}}, nil)
}
