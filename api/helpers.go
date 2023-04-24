package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

// readJSON reads/parses request body.
func (app *KeyKeeper) readJSON(w http.ResponseWriter, r *http.Request, input interface{}) error {
	// Restrict r.Body to 1MB
	maxBytes := 1_048_578
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	decoder := json.NewDecoder(r.Body)

	// Disallow unknown fields.
	decoder.DisallowUnknownFields()
	err := decoder.Decode(input)

	if err != nil {
		// expected error types
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("request body contains badly-formatted JSON (at character %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return fmt.Errorf("request body contains badly-formatted JSON")
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("request body contains badly-formatted JSON for the field: %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("request body contains badly-formatted JSON ((at character %d))", unmarshalTypeError.Offset)
		case errors.Is(err, io.EOF):
			return fmt.Errorf("empty request body")
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			field := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("request body contains unknow field: %s", field)
		case err.Error() == "http: request body too large":
			return fmt.Errorf("request body must not be larger than %d bytes", maxBytes)
		default:
			return err
		}
	}

	// ensure r.Body is one json object
	err = decoder.Decode(&struct{}{})
	if err != io.EOF {
		return fmt.Errorf("request body must contain only a single JSON")
	}

	return nil
}

// writeJSON writes and sends JSON response.
func (app *KeyKeeper) writeJSON(w http.ResponseWriter, statusCode int, data envelop, header http.Header) error {
	resp, err := json.Marshal(data)
	if err != nil {
		return err
	}
	resp = append(resp, '\n')

	for key, value := range header {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(resp)

	return nil
}

// errorResponse writes error response.
func (app *KeyKeeper) errorResponse(
	w http.ResponseWriter,
	r *http.Request,
	statusCode int,
	message interface{},
) {
	resp := envelop{"error": message}
	err := app.writeJSON(w, statusCode, resp, nil)
	if err != nil {
		log.Error().Err(err).
			Str("request_method", r.Method).
			Str("request_url", r.URL.String()).
			Msg("failed to write response body")
		w.WriteHeader(500)
	}
}
