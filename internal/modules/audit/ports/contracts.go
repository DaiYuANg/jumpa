package ports

import (
	"context"
	"time"
)

type CreateSessionEventInput struct {
	SessionID string
	EventType string
	Payload   *string
	CreatedAt time.Time
}

type SessionEventRepository interface {
	CreateSessionEvent(ctx context.Context, in CreateSessionEventInput) error
}
