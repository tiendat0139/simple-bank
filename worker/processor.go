package worker

import (
	"context"

	"github.com/hibiken/asynq"
	db "github.com/tiendat0139/simple-bank/db/sqlc"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
)
type Processor interface {
	Start() error
	ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error
}

type RedisTaskProcessor struct {
	server *asynq.Server
	store  db.Store
}

func NewRedisTaskProcessor(redisOpts asynq.RedisClientOpt, store db.Store) Processor {
	server := asynq.NewServer(
		redisOpts,
		asynq.Config{
			Queues: map[string]int{
				QueueCritical: 10,
				QueueDefault:  5,
			},
		},
	)

	return &RedisTaskProcessor{
		server: server,
		store:  store,
	}
}

func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskTypeSendVerifyEmail, processor.ProcessTaskSendVerifyEmail)
	return processor.server.Start(mux)
}
