package ports

import (
	"context"
	"time"
)

type SessionRecord struct {
	ID            string
	HostID        string
	HostName      string
	HostAccountID *string
	HostAccount   string
	PrincipalID   string
	PrincipalName string
	Protocol      string
	Status        string
	SourceAddr    *string
	StartedAt     time.Time
	EndedAt       *time.Time
}

type CreateSessionInput struct {
	HostID        string
	HostAccountID *string
	PrincipalID   string
	Protocol      string
	Status        string
	SourceAddr    *string
	StartedAt     time.Time
}

type CreateSessionEventInput struct {
	SessionID string
	EventType string
	Payload   *string
	CreatedAt time.Time
}

type SessionRepository interface {
	ListSessions(ctx context.Context) ([]SessionRecord, error)
	CreateSession(ctx context.Context, in CreateSessionInput) (string, error)
	UpdateSessionStatus(ctx context.Context, id, status string, endedAt *time.Time) error
}

type SessionEventRepository interface {
	CreateSessionEvent(ctx context.Context, in CreateSessionEventInput) error
}
