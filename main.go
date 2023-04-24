package main

import (
	"database/sql"
	"embed"
	"io/fs"
	"os"
	"time"

	migrate "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"

	"github.com/OCD-Labs/KeyKeeper/api"
	db "github.com/OCD-Labs/KeyKeeper/db/sqlc"
	"github.com/OCD-Labs/KeyKeeper/internal/mailer"
	"github.com/OCD-Labs/KeyKeeper/internal/token"
	"github.com/OCD-Labs/KeyKeeper/internal/utils"
	"github.com/OCD-Labs/KeyKeeper/internal/worker"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
)

//go:embed docs/swagger
var swaggerDocs embed.FS

func main() {
	configs, err := utils.ParseConfigs("./")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse configurations")
	}

	if configs.Env == "development" {
		log.Logger = log.Output(
			zerolog.ConsoleWriter{
				Out:        os.Stderr,
				TimeFormat: time.RFC3339,
			},
		).With().Caller().Logger()
	}

	log.Info().Msg("connecting to DB")
	dbConn, err := sql.Open(configs.DBDriver, configs.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to DB")
	}

	runDBMigrations(configs.MigrationURL, configs.DBSource)
	store := db.NewStore(dbConn)

	// Retrieve the swagger-ui files.
	swaggerFiles, err := fs.Sub(swaggerDocs, "docs/swagger")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get subcontent from swaggerDocs")
	}

	tokenMaker, err := token.NewPasetoMaker(configs.SymmetricKey)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to setup tokeMaker")
	}

	redisOpt := asynq.RedisClientOpt{
		Addr: configs.RedisAddress,
	}
	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)

	log.Info().Msg("setting up application")
	app := api.NewServer(configs, store, swaggerFiles, log.Logger, tokenMaker, taskDistributor)

	log.Info().Msg("starting redis server")
	go runTaskProcessor(configs, redisOpt, store, tokenMaker)

	log.Info().Msg("starting/stopping service")
	err = app.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("failed starting/stopping server")
	}
}

func runTaskProcessor(config utils.Configs, redisOpt asynq.RedisClientOpt, store db.Store, tokenMaker token.TokenMaker) {
	mailer := mailer.NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)
	taskProcessor := worker.NewRedisTaskProcessor(redisOpt, store, mailer, config, tokenMaker)
	log.Info().Msg("starting task processor")
	err := taskProcessor.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start task processor")
	}
}

func runDBMigrations(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create a new migrate instance")
	}

	if err := migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().Err(err).Msg("failed to run migrateup")
	}

	log.Info().Msg("db migrated successfully")
}
