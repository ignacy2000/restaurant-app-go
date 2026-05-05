package worker

import (
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

const TypePasswordReset = "email:password_reset"

type PasswordResetPayload struct {
	To       string `json:"to"`
	ResetURL string `json:"reset_url"`
}

func NewPasswordResetTask(to, resetURL string) (*asynq.Task, error) {
	payload, err := json.Marshal(PasswordResetPayload{To: to, ResetURL: resetURL})
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}
	return asynq.NewTask(TypePasswordReset, payload), nil
}
