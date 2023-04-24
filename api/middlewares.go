package api

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// ResponseRecorder wraps http.ResponseWriter to provide extra
// custome functions.
type ResponseRecorder struct {
	http.ResponseWriter
	StatusCode int
	Body       []byte
}

// Write capture the response status code as it's being
// written by the next handler.
func (rec *ResponseRecorder) WriteHeader(statusCode int) {
	rec.StatusCode = statusCode
	rec.ResponseWriter.WriteHeader(statusCode)
}

// Write capture the response body as it's being written
// by the next handler
func (rec *ResponseRecorder) Write(body []byte) (int, error) {
	rec.Body = body
	return rec.ResponseWriter.Write(body)
}

func (app *KeyKeeper) httpLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		rec := &ResponseRecorder{
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
		}

		next.ServeHTTP(rec, r)

		duration := time.Since(startTime)

		logger := log.Info()
		if rec.StatusCode < http.StatusOK || rec.StatusCode >= http.StatusBadRequest {
			logger = log.Error().Bytes("body", rec.Body)
		} else if rec.StatusCode >= http.StatusMultipleChoices && rec.StatusCode < http.StatusBadRequest {
			logger = log.Warn().Bytes("body", rec.Body)
		}

		logger.Str("protocol", "HTTP").
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status_code", rec.StatusCode).
			Str("status_text", http.StatusText(rec.StatusCode)).
			Dur("duration", duration).
			Msg("received an HTTP request")
	})
}
