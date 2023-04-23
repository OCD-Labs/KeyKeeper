package main

import (
	"database/sql"
	"embed"
	"io/fs"
	"os"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/OCD-Labs/KeyKeeper/api"
	db "github.com/OCD-Labs/KeyKeeper/db/sqlc"
	"github.com/OCD-Labs/KeyKeeper/internal/util"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
)

//go:embed docs/swagger
var swaggerDocs embed.FS

func main() {
	configs, err := util.ParseConfigs("./")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse configurations")
	}

	if configs.Env == "development" {
		log.Logger = log.Output(
			zerolog.ConsoleWriter{
				Out: os.Stderr, 
				TimeFormat: time.RFC3339,
			},
		).With().Caller().Logger()
	}

	log.Info().Msg("connecting to DB")
	dbConn, err := sql.Open(configs.DBDriver, configs.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to DB")
	}
	store := db.New(dbConn)

	// Retrieve the swagger-ui files.
	swaggerFiles, err := fs.Sub(swaggerDocs, "docs/swagger")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get subcontent from swaggerDocs")
	}

	log.Info().Msg("setting up application")
	app, err := api.NewServer(configs, store, swaggerFiles, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("failed setting up application")
	}

	log.Info().Msg("starting/stopping server")
	err = app.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("failed starting/stopping server")
	}
}
