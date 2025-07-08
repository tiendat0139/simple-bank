package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const (
	TaskTypeSendVerifyEmail = "task:send_verify_email"
)

type PayloadSendVerifyEmail struct {
	Username string `json:"username"`
}

// Method to create new send verify email task and put it on the queue
func (distributor *RedisTaskDistributor) DistributeTaskSendVerifyEmail(
	ctx context.Context,
	payload *PayloadSendVerifyEmail,
	opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TaskTypeSendVerifyEmail, jsonPayload, opts...)
	taskInfo, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	log.Info().Str("type", task.Type()).Bytes("payload", jsonPayload).Str("queue", taskInfo.Queue).
		Int("max_retry", taskInfo.MaxRetry).Msg("task enqueued")
	return nil
}

// Process send verify email task
func(processor *RedisTaskProcessor)	ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error {
	var payload *PayloadSendVerifyEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	user, err := processor.store.GetUser(ctx, payload.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user does not exist: %s", asynq.SkipRetry)
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// TODO: Implement the logic to send verification email
	log.Info().Str("email", user.Email).Msg("processed task")
	return nil

}
