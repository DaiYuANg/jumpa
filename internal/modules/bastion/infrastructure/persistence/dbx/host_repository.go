package dbx

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/DaiYuANg/arcgo/collectionx"
	"github.com/DaiYuANg/arcgo/dbx"
	"github.com/DaiYuANg/arcgo/dbx/repository"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/ports"
	"github.com/samber/mo"
)

type hostRow struct {
	ID                 int64     `dbx:"id"`
	Name               string    `dbx:"name"`
	Address            string    `dbx:"address"`
	Port               int       `dbx:"port"`
	Protocol           string    `dbx:"protocol"`
	Environment        *string   `dbx:"environment"`
	Platform           *string   `dbx:"platform"`
	AuthenticationType string    `dbx:"authentication_type"`
	CredentialRef      *string   `dbx:"credential_ref"`
	JumpEnabled        bool      `dbx:"jump_enabled"`
	RecordingPolicy    string    `dbx:"recording_policy"`
	CreatedAt          time.Time `dbx:"created_at"`
}

type hostSchema struct {
	dbx.Schema[hostRow]
	ID                 dbx.IDColumn[hostRow, int64, dbx.IDSnowflake] `dbx:"id,pk"`
	Name               dbx.Column[hostRow, string]                   `dbx:"name"`
	Address            dbx.Column[hostRow, string]                   `dbx:"address"`
	Port               dbx.Column[hostRow, int]                      `dbx:"port"`
	Protocol           dbx.Column[hostRow, string]                   `dbx:"protocol"`
	Environment        dbx.Column[hostRow, *string]                  `dbx:"environment"`
	Platform           dbx.Column[hostRow, *string]                  `dbx:"platform"`
	AuthenticationType dbx.Column[hostRow, string]                   `dbx:"authentication_type"`
	CredentialRef      dbx.Column[hostRow, *string]                  `dbx:"credential_ref"`
	JumpEnabled        dbx.Column[hostRow, bool]                     `dbx:"jump_enabled"`
	RecordingPolicy    dbx.Column[hostRow, string]                   `dbx:"recording_policy"`
	CreatedAt          dbx.Column[hostRow, time.Time]                `dbx:"created_at"`
}

type hostRepo struct {
	hs   hostSchema
	repo *repository.Base[hostRow, hostSchema]
}

func NewHostRepository(db *dbx.DB) ports.HostRepository {
	hs := dbx.MustSchema("bastion_hosts", hostSchema{})
	return &hostRepo{hs: hs, repo: repository.New[hostRow](db, hs)}
}

func (r *hostRepo) ListHosts(ctx context.Context) ([]ports.HostRecord, error) {
	rows, err := r.repo.ListSpec(ctx, repository.OrderBy(r.hs.Name.Asc()))
	if err != nil {
		return nil, err
	}
	return collectionx.MapList(rows, func(_ int, row hostRow) ports.HostRecord {
		return toHostRecord(row)
	}).Values(), nil
}

func (r *hostRepo) GetHostByID(ctx context.Context, id string) (mo.Option[ports.HostRecord], error) {
	hostID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return mo.None[ports.HostRecord](), err
	}
	row, err := r.repo.FirstSpec(ctx, repository.Where(r.hs.ID.Eq(hostID)))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return mo.None[ports.HostRecord](), nil
		}
		return mo.None[ports.HostRecord](), err
	}
	return mo.Some(toHostRecord(row)), nil
}

func (r *hostRepo) GetHostByName(ctx context.Context, name string) (mo.Option[ports.HostRecord], error) {
	row, err := r.repo.FirstSpec(ctx, repository.Where(r.hs.Name.Eq(name)))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return mo.None[ports.HostRecord](), nil
		}
		return mo.None[ports.HostRecord](), err
	}
	return mo.Some(toHostRecord(row)), nil
}

func (r *hostRepo) CreateHost(ctx context.Context, in ports.CreateHostRecordInput) (ports.HostRecord, error) {
	row := &hostRow{
		Name:               in.Name,
		Address:            in.Address,
		Port:               in.Port,
		Protocol:           in.Protocol,
		Environment:        in.Environment,
		Platform:           in.Platform,
		AuthenticationType: in.AuthenticationType,
		CredentialRef:      in.CredentialRef,
		JumpEnabled:        in.JumpEnabled,
		RecordingPolicy:    in.RecordingPolicy,
		CreatedAt:          in.CreatedAt,
	}
	if err := r.repo.Create(ctx, row); err != nil {
		return ports.HostRecord{}, err
	}
	return toHostRecord(*row), nil
}

func (r *hostRepo) UpdateHost(ctx context.Context, id string, in ports.PatchHostRecordInput) (mo.Option[ports.HostRecord], error) {
	hostID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return mo.None[ports.HostRecord](), err
	}
	assignments := make([]dbx.Assignment, 0, 10)
	if in.Name != nil {
		assignments = append(assignments, r.hs.Name.Set(*in.Name))
	}
	if in.Address != nil {
		assignments = append(assignments, r.hs.Address.Set(*in.Address))
	}
	if in.Port != nil {
		assignments = append(assignments, r.hs.Port.Set(*in.Port))
	}
	if in.Protocol != nil {
		assignments = append(assignments, r.hs.Protocol.Set(*in.Protocol))
	}
	if in.Environment != nil {
		assignments = append(assignments, r.hs.Environment.Set(in.Environment))
	}
	if in.Platform != nil {
		assignments = append(assignments, r.hs.Platform.Set(in.Platform))
	}
	if in.AuthenticationType != nil {
		assignments = append(assignments, r.hs.AuthenticationType.Set(*in.AuthenticationType))
	}
	if in.CredentialRef != nil {
		assignments = append(assignments, r.hs.CredentialRef.Set(in.CredentialRef))
	}
	if in.JumpEnabled != nil {
		assignments = append(assignments, r.hs.JumpEnabled.Set(*in.JumpEnabled))
	}
	if in.RecordingPolicy != nil {
		assignments = append(assignments, r.hs.RecordingPolicy.Set(*in.RecordingPolicy))
	}
	if len(assignments) == 0 {
		return r.GetHostByID(ctx, id)
	}
	res, err := r.repo.UpdateByID(ctx, hostID, assignments...)
	if err != nil {
		return mo.None[ports.HostRecord](), err
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return mo.None[ports.HostRecord](), nil
	}
	return r.GetHostByID(ctx, id)
}

func (r *hostRepo) DeleteHost(ctx context.Context, id string) (bool, error) {
	hostID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return false, err
	}
	res, err := r.repo.DeleteByID(ctx, hostID)
	if err != nil {
		return false, err
	}
	ra, _ := res.RowsAffected()
	return ra > 0, nil
}

func toHostRecord(row hostRow) ports.HostRecord {
	return ports.HostRecord{
		ID:                 strconv.FormatInt(row.ID, 10),
		Name:               row.Name,
		Address:            row.Address,
		Port:               row.Port,
		Protocol:           row.Protocol,
		Environment:        row.Environment,
		Platform:           row.Platform,
		AuthenticationType: row.AuthenticationType,
		CredentialRef:      row.CredentialRef,
		JumpEnabled:        row.JumpEnabled,
		RecordingPolicy:    row.RecordingPolicy,
		CreatedAt:          row.CreatedAt,
	}
}
