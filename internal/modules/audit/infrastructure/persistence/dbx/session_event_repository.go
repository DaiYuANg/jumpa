package dbx

import (
	"context"
	"strconv"
	"time"

	"github.com/arcgolabs/dbx"
	columnx "github.com/arcgolabs/dbx/column"
	"github.com/arcgolabs/dbx/idgen"
	"github.com/arcgolabs/dbx/repository"
	schemax "github.com/arcgolabs/dbx/schema"
	"github.com/DaiYuANg/jumpa/internal/modules/audit/ports"
)

type sessionEventRow struct {
	ID        int64     `dbx:"id"`
	SessionID int64     `dbx:"session_id"`
	EventType string    `dbx:"event_type"`
	Payload   *string   `dbx:"payload"`
	CreatedAt time.Time `dbx:"created_at"`
}

type sessionEventSchema struct {
	schemax.Schema[sessionEventRow]
	ID        columnx.IDColumn[sessionEventRow, int64, idgen.IDSnowflake] `dbx:"id,pk"`
	SessionID columnx.Column[sessionEventRow, int64]                      `dbx:"session_id"`
	EventType columnx.Column[sessionEventRow, string]                     `dbx:"event_type"`
	Payload   columnx.Column[sessionEventRow, *string]                    `dbx:"payload"`
	CreatedAt columnx.Column[sessionEventRow, time.Time]                  `dbx:"created_at"`
}

type sessionEventRepo struct {
	repo *repository.Base[sessionEventRow, sessionEventSchema]
}

func NewSessionEventRepository(db *dbx.DB) ports.SessionEventRepository {
	ss := schemax.MustSchema("bastion_session_events", sessionEventSchema{})
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
