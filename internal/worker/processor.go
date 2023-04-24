package worker

import (
	"context"

	db "github.com/OCD-Labs/KeyKeeper/db/sqlc"
	"github.com/OCD-Labs/KeyKeeper/internal/logger"
	"github.com/OCD-Labs/KeyKeeper/internal/mailer"
	"github.com/OCD-Labs/KeyKeeper/internal/token"
	"github.com/OCD-Labs/KeyKeeper/internal/utils"
	"github.com/go-redis/redis/v8"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const (
	// QueueCritical is the name of the critical queue.
	QueueCritical = "critical"
	// QueueDefault is the name of the default queue.
	QueueDefault = "default"
)

// TaskProcessor is an interface for a worker that processes tasks.
type TaskProcessor interface {
	// Start starts the RedisTaskProcessor.
	Start() error

	// ProcessTaskSendVerifyEmail processes a TaskSendVerifyEmail task.
	ProcessTaskSendVerifyEmail(context.Context, *asynq.Task) error
}

type RedisTaskProcessor struct {
	server     *asynq.Server
	store      db.Store
	configs    utils.Configs
	mailer     mailer.EmailSender
	tokenMaker token.TokenMaker
}

// NewRedisTaskProcessor creates a new RedisTaskProcessor.
func NewRedisTaskProcessor(
	redisOpt asynq.RedisClientOpt,
	store db.Store,
	mailer mailer.EmailSender,
	configs utils.Configs,
	tokenMaker token.TokenMaker,
) TaskProcessor {
	logger := logger.New()
	redis.SetLogger(logger)

	server := asynq.NewServer(redisOpt, asynq.Config{
		Queues: map[string]int{
			QueueCritical: 10,
			QueueDefault:  5,
		},
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			log.Error().Err(err).
				Str("type", task.Type()).
				Bytes("payload", task.Payload()).
				Msg("process task failed")
		}),
		Logger: logger,
	})

	return &RedisTaskProcessor{
		server:     server,
		store:      store,
		configs:    configs,
		mailer:     mailer,
		tokenMaker: tokenMaker,
	}
}

// Start starts the RedisTaskProcessor
func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskSendVerifyEmail, processor.ProcessTaskSendVerifyEmail)
	mux.HandleFunc(TaskSendResetPasswordEmail, processor.ProcessTaskSendResetPasswordEmail)
	return processor.server.Start(mux)
}
