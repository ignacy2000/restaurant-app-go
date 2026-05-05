package worker

import (
	"fmt"

	"github.com/hibiken/asynq"
)

func NewServer(redisAddr string) *asynq.Server {
	return asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{Concurrency: 5},
	)
}

func Start(srv *asynq.Server, proc *EmailProcessor) error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TypePasswordReset, proc.ProcessTask)
	if err := srv.Start(mux); err != nil {
		return fmt.Errorf("start worker: %w", err)
	}
	return nil
}
