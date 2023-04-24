package api

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
)

// ValidationError represents a custom validation error that
// contains information about the violated fields and their messages.
type ValidationError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

// BindJSONWithValidation is a helper function that binds the JSON request
// body to the given interface and validates it with the specified validator.
func (app *KeyKeeper) bindJSONWithValidation(
	w http.ResponseWriter,
	r *http.Request,
	data interface{},
	validate *validator.Validate,
) error {
	if err := validate.Struct(data); err != nil {
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			return err
		}

		validationErrors := make([]ValidationError, 0, len(errs))

		for _, err := range errs {
			validationErrors = append(validationErrors, ValidationError{
				Field: err.Field(),
				Error: fmt.Sprintf("%s validation failed on '%s'", err.Tag(), err.Param()),
			})
		}

		app.errorResponse(w, r, http.StatusBadRequest, validationErrors)

		return err
	}

	return nil
}
