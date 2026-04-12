package dbx

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/DaiYuANg/arcgo/dbx"
	"github.com/DaiYuANg/arcgo/dbx/repository"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/ports"
	"github.com/samber/mo"
)

type sessionRow struct {
	ID            int64      `dbx:"id"`
	HostID        int64      `dbx:"host_id"`
	HostAccountID *int64     `dbx:"host_account_id"`
	PrincipalID   string     `dbx:"principal_id"`
	Protocol      string     `dbx:"protocol"`
	Status        string     `dbx:"status"`
	SourceAddr    *string    `dbx:"source_addr"`
	StartedAt     time.Time  `dbx:"started_at"`
	EndedAt       *time.Time `dbx:"ended_at"`
}

type sessionSchema struct {
	dbx.Schema[sessionRow]
	ID            dbx.IDColumn[sessionRow, int64, dbx.IDSnowflake] `dbx:"id,pk"`
	HostID        dbx.Column[sessionRow, int64]                    `dbx:"host_id"`
	HostAccountID dbx.Column[sessionRow, *int64]                   `dbx:"host_account_id"`
	PrincipalID   dbx.Column[sessionRow, string]                   `dbx:"principal_id"`
	Protocol      dbx.Column[sessionRow, string]                   `dbx:"protocol"`
	Status        dbx.Column[sessionRow, string]                   `dbx:"status"`
	SourceAddr    dbx.Column[sessionRow, *string]                  `dbx:"source_addr"`
	StartedAt     dbx.Column[sessionRow, time.Time]                `dbx:"started_at"`
	EndedAt       dbx.Column[sessionRow, *time.Time]               `dbx:"ended_at"`
}

type sessionRepo struct {
	db   *dbx.DB
	ss   sessionSchema
	repo *repository.Base[sessionRow, sessionSchema]
}

func NewSessionRepository(db *dbx.DB) ports.SessionRepository {
	ss := dbx.MustSchema("bastion_sessions", sessionSchema{})
	return &sessionRepo{db: db, ss: ss, repo: repository.New[sessionRow](db, ss)}
}

func (r *sessionRepo) CreateSession(ctx context.Context, in ports.CreateSessionInput) (string, error) {
	hostID, err := strconv.ParseInt(in.HostID, 10, 64)
	if err != nil {
		return "", err
	}
	row := &sessionRow{
		HostID:      hostID,
		PrincipalID: in.PrincipalID,
		Protocol:    in.Protocol,
		Status:      in.Status,
		SourceAddr:  in.SourceAddr,
		StartedAt:   in.StartedAt,
	}
	if in.HostAccountID != nil && *in.HostAccountID != "" {
		accountID, parseErr := strconv.ParseInt(*in.HostAccountID, 10, 64)
		if parseErr != nil {
			return "", parseErr
		}
		row.HostAccountID = &accountID
	}
	if err := r.repo.Create(ctx, row); err != nil {
		return "", err
	}
	return strconv.FormatInt(row.ID, 10), nil
}

func (r *sessionRepo) UpdateSessionStatus(ctx context.Context, id, status string, endedAt *time.Time) error {
	sessionID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}
	assignments := []dbx.Assignment{
		r.ss.Status.Set(status),
		r.ss.EndedAt.Set(endedAt),
	}
	_, err = r.repo.UpdateByID(ctx, sessionID, assignments...)
	return err
}

func (r *sessionRepo) GetSessionByID(ctx context.Context, id string) (mo.Option[ports.SessionRecord], error) {
	sessionID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return mo.None[ports.SessionRecord](), err
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT
  s.id,
  s.host_id,
  h.name AS host_name,
  s.host_account_id,
  COALESCE(a.account_name, '') AS host_account,
  s.principal_id,
  s.principal_id AS principal_name,
  s.protocol,
  s.status,
  s.source_addr,
  s.started_at,
  s.ended_at
FROM bastion_sessions s
JOIN bastion_hosts h ON h.id = s.host_id
LEFT JOIN bastion_host_accounts a ON a.id = s.host_account_id
WHERE s.id = ?
LIMIT 1
`, sessionID)
	if err != nil {
		return mo.None[ports.SessionRecord](), err
	}
	defer func() { _ = rows.Close() }()

	if !rows.Next() {
		return mo.None[ports.SessionRecord](), rows.Err()
	}

	record, err := scanSessionRecord(rows)
	if err != nil {
		return mo.None[ports.SessionRecord](), err
	}
	return mo.Some(record), rows.Err()
}

func (r *sessionRepo) ListSessions(ctx context.Context) ([]ports.SessionRecord, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT
  s.id,
  s.host_id,
  h.name AS host_name,
  s.host_account_id,
  COALESCE(a.account_name, '') AS host_account,
  s.principal_id,
  s.principal_id AS principal_name,
  s.protocol,
  s.status,
  s.source_addr,
  s.started_at,
  s.ended_at
FROM bastion_sessions s
JOIN bastion_hosts h ON h.id = s.host_id
LEFT JOIN bastion_host_accounts a ON a.id = s.host_account_id
ORDER BY s.started_at DESC
`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]ports.SessionRecord, 0, 16)
	for rows.Next() {
		record, err := scanSessionRecord(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, record)
	}

	return items, rows.Err()
}

func scanSessionRecord(rows *sql.Rows) (ports.SessionRecord, error) {
	var record ports.SessionRecord
	var id int64
	var hostID int64
	var hostAccountID sql.NullInt64
	var hostAccount sql.NullString
	if err := rows.Scan(
		&id,
		&hostID,
		&record.HostName,
		&hostAccountID,
		&hostAccount,
		&record.PrincipalID,
		&record.PrincipalName,
		&record.Protocol,
		&record.Status,
		&record.SourceAddr,
		&record.StartedAt,
		&record.EndedAt,
	); err != nil {
		return ports.SessionRecord{}, err
	}
	record.ID = strconv.FormatInt(id, 10)
	record.HostID = strconv.FormatInt(hostID, 10)
	if hostAccountID.Valid {
		value := strconv.FormatInt(hostAccountID.Int64, 10)
		record.HostAccountID = &value
	}
	if hostAccount.Valid {
		record.HostAccount = hostAccount.String
	}
	return record, nil
}
