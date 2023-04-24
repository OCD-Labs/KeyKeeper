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
	"github.com/OCD-Labs/KeyKeeper/internal/token"
	"github.com/OCD-Labs/KeyKeeper/internal/utils"
	"github.com/OCD-Labs/KeyKeeper/internal/worker"
	"github.com/rs/zerolog"

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
) *KeyKeeper {
	server := &KeyKeeper{
		configs:         configs,
		store:           store,
		tokenMaker:      tokenMaker,
		swaggerUI:       swaggerFiles,
		logger:          logger,
		taskDistributor: taskDistributor,
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

	app.router = app.httpLogger(mux)
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
