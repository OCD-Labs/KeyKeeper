package api

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

var sessionCleanupJobTicker *time.Ticker

func (app *KeyKeeper) startSessionCleanupJob() {
	sessionCleanupJobTicker = time.NewTicker(25 * time.Minute)

	// Launch the background goroutine.
	go func() {
		defer func() {
			if err := recover(); err != nil {
				app.logger.Error().Err(fmt.Errorf("%s", err))
			}
		}()

		for range sessionCleanupJobTicker.C {
			err := app.store.DeleteExpiredSession(context.Background())
			if err != nil {
				log.Error().Err(err).Msg("error occurred")
			}
		}
	}()
}
