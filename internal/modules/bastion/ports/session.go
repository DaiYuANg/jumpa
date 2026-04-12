package ports

import (
	"context"
	"time"

	"github.com/samber/mo"
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

type SessionRepository interface {
	ListSessions(ctx context.Context) ([]SessionRecord, error)
	GetSessionByID(ctx context.Context, id string) (mo.Option[SessionRecord], error)
	CreateSession(ctx context.Context, in CreateSessionInput) (string, error)
	UpdateSessionStatus(ctx context.Context, id, status string, endedAt *time.Time) error
}
