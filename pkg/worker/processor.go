package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"table-service.pl/pkg/mailer"
)

type EmailProcessor struct {
	mailer *mailer.Mailer
}

func NewEmailProcessor(m *mailer.Mailer) *EmailProcessor {
	return &EmailProcessor{mailer: m}
}

func (p *EmailProcessor) ProcessTask(_ context.Context, t *asynq.Task) error {
	switch t.Type() {
	case TypePasswordReset:
		var payload PasswordResetPayload
		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return fmt.Errorf("unmarshal payload: %w", err)
		}
		return p.mailer.SendPasswordReset(payload.To, payload.ResetURL)
	default:
		return fmt.Errorf("unknown task type: %s", t.Type())
	}
}
