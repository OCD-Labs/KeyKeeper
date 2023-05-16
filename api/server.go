package api

import (
	"context"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	db "github.com/OCD-Labs/KeyKeeper/db/sqlc"
	"github.com/OCD-Labs/KeyKeeper/internal/cache"
	"github.com/OCD-Labs/KeyKeeper/internal/token"
	"github.com/OCD-Labs/KeyKeeper/internal/utils"
	"github.com/OCD-Labs/KeyKeeper/internal/worker"
	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/julienschmidt/httprouter"
)

// KeyKeeper is the application, holds the necessary dependencies.
type KeyKeeper struct {
	configs         utils.Configs
	store           db.Store
	tokenMaker      token.TokenMaker
	swaggerUI       fs.FS
	router          http.Handler
	logger          zerolog.Logger
	wg              sync.WaitGroup
	taskDistributor worker.TaskDistributor
	googleConfig    *oauth2.Config
	cache           cache.Cache
}

type envelop map[string]interface{}

// NewServer setups a KeyKeeper object, and the application
// routes.
func NewServer(
	configs utils.Configs,
	store db.Store,
	swaggerFiles fs.FS,
	logger zerolog.Logger,
	tokenMaker token.TokenMaker,
	taskDistributor worker.TaskDistributor,
	cache cache.Cache,
) *KeyKeeper {

	googleConfig := &oauth2.Config{
		Endpoint:     google.Endpoint,
		RedirectURL:  configs.GoogleRedirectURL,
		ClientID:     configs.GoogleClientID,
		ClientSecret: configs.GoogleClientSecret,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
	}

	server := &KeyKeeper{
		configs:         configs,
		store:           store,
		tokenMaker:      tokenMaker,
		swaggerUI:       swaggerFiles,
		logger:          logger,
		taskDistributor: taskDistributor,
		cache:           cache,
		googleConfig:    googleConfig,
	}
	server.setupRoutes()

	return server
}

func (app *KeyKeeper) setupRoutes() {
	mux := httprouter.New()

	fsysHandler := http.FileServer(http.FS(app.swaggerUI))
	mux.Handler(http.MethodGet, "/api/v1/swagger/*any", http.StripPrefix("/api/v1/swagger/", fsysHandler))

	mux.HandlerFunc(http.MethodGet, "/api/v1/healthcheck", app.ping)
	mux.HandlerFunc(http.MethodPost, "/api/v1/users", app.createUser)
	mux.HandlerFunc(http.MethodPatch, "/api/v1/verify_email", app.verifyEmail)
	mux.HandlerFunc(http.MethodGet, "/api/v1/auth/login", app.login)
	mux.HandlerFunc(http.MethodGet, "/api/v1/auth/google/login", app.googleLogin)
	mux.HandlerFunc(http.MethodGet, "/api/v1/auth/google/callback", app.googleLoginCallback)
	mux.HandlerFunc(http.MethodPost, "/api/v1/resend_email_verification", app.resendVerifyEmail)
	mux.HandlerFunc(http.MethodPost, "/api/v1/forgot_password", app.forgotPassword)
	mux.HandlerFunc(http.MethodPatch, "/api/v1/reset_password", app.resetUserPassword)

	mux.Handler(http.MethodPost, "/api/v1/auth/logout", app.authenticate(http.HandlerFunc(app.logout)))
	mux.Handler(http.MethodPatch, "/api/v1/users/:id/deactivate", app.authenticate(http.HandlerFunc(app.deactivateUser)))
	mux.Handler(http.MethodPatch, "/api/v1/users/:id/change_password", app.authenticate(http.HandlerFunc(app.changeUserPassword)))
	mux.Handler(http.MethodGet, "/api/v1/users/:id", app.authenticate(http.HandlerFunc(app.getUser)))

	mux.Handler(http.MethodPost, "/api/v1/reminders", app.authenticate(http.HandlerFunc(app.createReminder)))
	mux.Handler(http.MethodGet, "/api/v1/reminders/:id", app.authenticate(http.HandlerFunc(app.getReminder)))
	mux.Handler(http.MethodGet, "/api/v1/reminders", app.authenticate(http.HandlerFunc(app.listReminders)))

	app.router = app.recoverPanic(app.enableCORS(app.httpLogger(mux)))
}

// Start setup amd starts a server.
func (app *KeyKeeper) Start() error {
	server := &http.Server{
		Addr:         app.configs.ServerAddress,
		Handler:      app.router,
		ErrorLog:     log.New(app.logger, "", 0),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	shutdownErr := make(chan error)

	// Background job to listen for any shutdown signal
	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		app.logger.Info().
			Str("signal", s.String()).
			Msg("shutting down server")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := server.Shutdown(ctx)
		if err != nil {
			shutdownErr <- err
		}

		app.logger.Info().
			Str("addr", server.Addr).
			Msg("completing background tasks")

		app.wg.Wait()
		shutdownErr <- nil
	}()

	app.startSessionCleanupJob()

	app.logger.Info().
		Str("environment", app.configs.Env).
		Str("addr", server.Addr).
		Msg("starting server")

	err := server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownErr
	if err != nil {
		return err
	}

	app.logger.Info().
		Str("addr", server.Addr).
		Msg("server stopped")

	return nil
}
