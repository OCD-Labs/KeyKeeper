// Package worker (task_send_reset_password_email) provides functions to distribute
// and process reset password email tasks.
package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const (
	// TaskSendResetPasswordEmail represents the name of the task that sends the reset password email.
	TaskSendResetPasswordEmail = "task:send_reset_password_email"
)

// PayloadResetPasswordEmail provides the userEmail.
type PayloadSendResetPasswordEmail struct {
	SessionID uuid.UUID `json:"payload_id"`
	UserEmail string    `json:"user_email"`
}

// DistributeTaskSendResetPasswordEmail enqueues the given task to be processed by a worker. It returns an error if the task could
// not be enqueued.
func (distributor *RedisTaskDistributor) DistributeTaskSendResetPasswordEmail(
	ctx context.Context,
	payload *PayloadSendResetPasswordEmail,
	opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	task := asynq.NewTask(TaskSendResetPasswordEmail, jsonPayload, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	log.Info().Str("type", task.Type()).
		Bytes("payload", task.Payload()).
		Str("queue", info.Queue).
		Int("max_retry", info.MaxRetry).
		Msg("enqueued task")

	return nil
}

// ProcessTaskSendResetPasswordEmail processes a 'TaskSendResetPasswordEmail' task.
func (processor *RedisTaskProcessor) ProcessTaskSendResetPasswordEmail(
	ctx context.Context,
	task *asynq.Task,
) error {
	var payload PayloadSendResetPasswordEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	resetPassword, err := processor.store.GetSession(ctx, payload.SessionID)
	if err != nil {
		return fmt.Errorf("failed to get reset password record: %w", err)
	}

	resetPasswordURL := fmt.Sprintf(
		"https://%s/api/v1/reset_password?reset_token=%s",
		processor.configs.APIHost,
		resetPassword.Token,
	)
	subject := "Reset password instructions"
	content := fmt.Sprintf(`Hi, <br/>
	Someone has requested a link to change your password. You can do this through the link below! <br/>
	Please <a href="%s">Click here</a> to reset your password.<br/>
	`, resetPasswordURL)
	to := []string{payload.UserEmail}

	err = processor.mailer.SendEmail(subject, content, to, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to send reset password email: %w", err)
	}

	log.Info().Str("type", task.Type()).
		Bytes("payload", task.Payload()).
		Str("email", payload.UserEmail).
		Msg("processed task")

	return nil
}
