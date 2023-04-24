package worker

import (
	"context"

	"github.com/hibiken/asynq"
)

// TaskDistributor defines the inteface required to
// distribute asynchronous tasks.
type TaskDistributor interface {
	DistributeTaskSendVerifyEmail(
		ctx context.Context,
		payload *PayloadSendVerifyEmail,
		opts ...asynq.Option,
	) error

	DistributeTaskSendResetPasswordEmail(
		ctx context.Context,
		payload *PayloadSendResetPasswordEmail,
		opts ...asynq.Option,
	) error
}

// RedisTaskDistributor defines and wrap a asynq client
// to distribute task to redis.
type RedisTaskDistributor struct {
	client *asynq.Client
}

func NewRedisTaskDistributor(redisOpts asynq.RedisClientOpt) TaskDistributor {
	client := asynq.NewClient(redisOpts)
	return &RedisTaskDistributor{
		client: client,
	}
}
