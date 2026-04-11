package dbx

import (
	"context"
	"strconv"
	"time"

	"github.com/DaiYuANg/arcgo/dbx"
	"github.com/DaiYuANg/arcgo/dbx/repository"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/ports"
)

type sessionEventRow struct {
	ID        int64     `dbx:"id"`
	SessionID int64     `dbx:"session_id"`
	EventType string    `dbx:"event_type"`
	Payload   *string   `dbx:"payload"`
	CreatedAt time.Time `dbx:"created_at"`
}

type sessionEventSchema struct {
	dbx.Schema[sessionEventRow]
	ID        dbx.IDColumn[sessionEventRow, int64, dbx.IDSnowflake] `dbx:"id,pk"`
	SessionID dbx.Column[sessionEventRow, int64]                    `dbx:"session_id"`
	EventType dbx.Column[sessionEventRow, string]                   `dbx:"event_type"`
	Payload   dbx.Column[sessionEventRow, *string]                  `dbx:"payload"`
	CreatedAt dbx.Column[sessionEventRow, time.Time]                `dbx:"created_at"`
}

type sessionEventRepo struct {
	repo *repository.Base[sessionEventRow, sessionEventSchema]
}

func NewSessionEventRepository(db *dbx.DB) ports.SessionEventRepository {
	ss := dbx.MustSchema("bastion_session_events", sessionEventSchema{})
	return &sessionEventRepo{repo: repository.New[sessionEventRow](db, ss)}
}

func (r *sessionEventRepo) CreateSessionEvent(ctx context.Context, in ports.CreateSessionEventInput) error {
	sessionID, err := strconv.ParseInt(in.SessionID, 10, 64)
	if err != nil {
		return err
	}
	return r.repo.Create(ctx, &sessionEventRow{
		SessionID: sessionID,
		EventType: in.EventType,
		Payload:   in.Payload,
		CreatedAt: in.CreatedAt,
	})
}
