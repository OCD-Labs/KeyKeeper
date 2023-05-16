package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/OCD-Labs/KeyKeeper/internal/token"
	"github.com/rs/zerolog/log"
)

const (
	authorizationHeaderKey  = "Authorization"
	authorizationTypeBearer = "Bearer"
)

var (
	authorizationPayloadKey = &struct{}{}
)

// contextSetToken registers an authentication token per connection
func (app *KeyKeeper) contextSetToken(r *http.Request, payload *token.Payload) *http.Request {
	ctx := context.WithValue(r.Context(), authorizationPayloadKey, payload)
	return r.WithContext(ctx)
}

// contextGetToken retrieves n authentication token.
func (app *KeyKeeper) contextGetToken(r *http.Request) *token.Payload {
	user, ok := r.Context().Value(authorizationPayloadKey).(*token.Payload)
	if !ok {
		panic("missing user value in request context")
	}
	return user
}

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

// authenticate helps know who the user is through their 'Bearer <token>'.
func (app *KeyKeeper) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This indicates to any caches that the response may
		// vary based on the value of Authorization.
		w.Header().Set("Vary", authorizationHeaderKey)

		authHeader := r.Header.Get(authorizationHeaderKey)
		if authHeader == "" {
			app.errorResponse(w, r, http.StatusUnauthorized, "authorization header is not provided")
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != authorizationTypeBearer {
			app.errorResponse(w, r, http.StatusUnauthorized, "invalid authorization header format")
			return
		}

		accessToken := tokenParts[1]
		payload, err := app.tokenMaker.VerifyToken(accessToken)
		if err != nil {
			switch {
			case errors.Is(err, token.ErrExpiredToken):
				app.errorResponse(w, r, http.StatusBadRequest, token.ErrExpiredToken.Error())
			case errors.Is(err, token.ErrInvalidToken):
				fmt.Println(authHeader)
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

		r = app.contextSetToken(r, payload)

		next.ServeHTTP(w, r)
	})
}

// enableCORS enables cross-site requests for web user-agents.
func (app *KeyKeeper) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Vary", "Access-Control-Request-Method")

		origin := r.Header.Get("Origin")

		// Preflight
		if origin != "" && len(app.configs.CorsTrustedOrigins) != 0 {
			for _, v := range app.configs.CorsTrustedOrigins {
				if origin == v {
					w.Header().Set("Access-Control-Allow-Origin", origin)

					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
						w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
						w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

						w.WriteHeader(http.StatusOK)
						return
					}
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

// recoverPanic graciouly recovers any panic within the goroutine handling the request
func (app *KeyKeeper) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")

				werr := fmt.Errorf("%s", err)
				app.errorResponse(w, r, http.StatusInternalServerError, werr.Error())
			}
		}()

		next.ServeHTTP(w, r)
	})
}
