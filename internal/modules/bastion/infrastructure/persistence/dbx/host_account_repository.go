package dbx

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/arcgolabs/collectionx"
	"github.com/arcgolabs/dbx"
	columnx "github.com/arcgolabs/dbx/column"
	"github.com/arcgolabs/dbx/idgen"
	"github.com/arcgolabs/dbx/querydsl"
	"github.com/arcgolabs/dbx/repository"
	schemax "github.com/arcgolabs/dbx/schema"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/ports"
	"github.com/samber/mo"
)

type hostAccountRow struct {
	ID                 int64     `dbx:"id"`
	HostID             int64     `dbx:"host_id"`
	AccountName        string    `dbx:"account_name"`
	AuthenticationType string    `dbx:"authentication_type"`
	CredentialRef      *string   `dbx:"credential_ref"`
	CreatedAt          time.Time `dbx:"created_at"`
}

type hostAccountSchema struct {
	schemax.Schema[hostAccountRow]
	ID                 columnx.IDColumn[hostAccountRow, int64, idgen.IDSnowflake] `dbx:"id,pk"`
	HostID             columnx.Column[hostAccountRow, int64]                      `dbx:"host_id"`
	AccountName        columnx.Column[hostAccountRow, string]                     `dbx:"account_name"`
	AuthenticationType columnx.Column[hostAccountRow, string]                     `dbx:"authentication_type"`
	CredentialRef      columnx.Column[hostAccountRow, *string]                    `dbx:"credential_ref"`
	CreatedAt          columnx.Column[hostAccountRow, time.Time]                  `dbx:"created_at"`
}

type hostAccountRepo struct {
	hs   hostAccountSchema
	repo *repository.Base[hostAccountRow, hostAccountSchema]
}

func NewHostAccountRepository(db *dbx.DB) ports.HostAccountRepository {
	hs := schemax.MustSchema("bastion_host_accounts", hostAccountSchema{})
	return &hostAccountRepo{hs: hs, repo: repository.New[hostAccountRow](db, hs)}
}

func (r *hostAccountRepo) GetHostAccountByName(ctx context.Context, hostID, accountName string) (mo.Option[ports.HostAccountRecord], error) {
	hostKey, err := strconv.ParseInt(hostID, 10, 64)
	if err != nil {
		return mo.None[ports.HostAccountRecord](), err
	}
	row, err := r.repo.FirstSpec(ctx,
		repository.Where(r.hs.HostID.Eq(hostKey)),
		repository.Where(r.hs.AccountName.Eq(accountName)),
	)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return mo.None[ports.HostAccountRecord](), nil
		}
		return mo.None[ports.HostAccountRecord](), err
	}
	return mo.Some(toHostAccountRecord(row)), nil
}

func (r *hostAccountRepo) GetHostAccountByID(ctx context.Context, hostID, accountID string) (mo.Option[ports.HostAccountRecord], error) {
	hostKey, err := strconv.ParseInt(hostID, 10, 64)
	if err != nil {
		return mo.None[ports.HostAccountRecord](), err
	}
	accountKey, err := strconv.ParseInt(accountID, 10, 64)
	if err != nil {
		return mo.None[ports.HostAccountRecord](), err
	}
	row, err := r.repo.FirstSpec(ctx,
		repository.Where(r.hs.HostID.Eq(hostKey)),
		repository.Where(r.hs.ID.Eq(accountKey)),
	)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return mo.None[ports.HostAccountRecord](), nil
		}
		return mo.None[ports.HostAccountRecord](), err
	}
	return mo.Some(toHostAccountRecord(row)), nil
}

func (r *hostAccountRepo) ListHostAccountsByHostID(ctx context.Context, hostID string) ([]ports.HostAccountRecord, error) {
	hostKey, err := strconv.ParseInt(hostID, 10, 64)
	if err != nil {
		return nil, err
	}
	rows, err := r.repo.ListSpec(ctx,
		repository.Where(r.hs.HostID.Eq(hostKey)),
		repository.OrderBy(r.hs.AccountName.Asc()),
	)
	if err != nil {
		return nil, err
	}
	return collectionx.MapList(rows, func(_ int, row hostAccountRow) ports.HostAccountRecord {
		return toHostAccountRecord(row)
	}).Values(), nil
}

func (r *hostAccountRepo) CreateHostAccount(ctx context.Context, in ports.CreateHostAccountRecordInput) (ports.HostAccountRecord, error) {
	hostKey, err := strconv.ParseInt(in.HostID, 10, 64)
	if err != nil {
		return ports.HostAccountRecord{}, err
	}
	row := &hostAccountRow{
		HostID:             hostKey,
		AccountName:        in.AccountName,
		AuthenticationType: in.AuthenticationType,
		CredentialRef:      in.CredentialRef,
		CreatedAt:          in.CreatedAt,
	}
	if err := r.repo.Create(ctx, row); err != nil {
		return ports.HostAccountRecord{}, err
	}
	return toHostAccountRecord(*row), nil
}

func (r *hostAccountRepo) UpdateHostAccount(ctx context.Context, hostID, accountID string, in ports.PatchHostAccountRecordInput) (mo.Option[ports.HostAccountRecord], error) {
	hostKey, err := strconv.ParseInt(hostID, 10, 64)
	if err != nil {
		return mo.None[ports.HostAccountRecord](), err
	}
	accountKey, err := strconv.ParseInt(accountID, 10, 64)
	if err != nil {
		return mo.None[ports.HostAccountRecord](), err
	}
	assignments := make([]querydsl.Assignment, 0, 3)
	if in.AccountName != nil {
		assignments = append(assignments, r.hs.AccountName.Set(*in.AccountName))
	}
	if in.AuthenticationType != nil {
		assignments = append(assignments, r.hs.AuthenticationType.Set(*in.AuthenticationType))
	}
	if in.CredentialRef != nil {
		assignments = append(assignments, r.hs.CredentialRef.Set(in.CredentialRef))
	}
	if len(assignments) == 0 {
		return r.GetHostAccountByID(ctx, hostID, accountID)
	}
	res, err := r.repo.Update(ctx, querydsl.Update(r.hs).
		Set(assignments...).
		Where(r.hs.HostID.Eq(hostKey)).
		Where(r.hs.ID.Eq(accountKey)))
	if err != nil {
		return mo.None[ports.HostAccountRecord](), err
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return mo.None[ports.HostAccountRecord](), nil
	}
	return r.GetHostAccountByID(ctx, hostID, accountID)
}

func (r *hostAccountRepo) DeleteHostAccount(ctx context.Context, hostID, accountID string) (bool, error) {
	hostKey, err := strconv.ParseInt(hostID, 10, 64)
	if err != nil {
		return false, err
	}
	accountKey, err := strconv.ParseInt(accountID, 10, 64)
	if err != nil {
		return false, err
	}
	res, err := r.repo.Delete(ctx, querydsl.DeleteFrom(r.hs).Where(r.hs.HostID.Eq(hostKey)).Where(r.hs.ID.Eq(accountKey)))
	if err != nil {
		return false, err
	}
	ra, _ := res.RowsAffected()
	return ra > 0, nil
}

func toHostAccountRecord(row hostAccountRow) ports.HostAccountRecord {
	return ports.HostAccountRecord{
		ID:                 strconv.FormatInt(row.ID, 10),
		HostID:             strconv.FormatInt(row.HostID, 10),
		AccountName:        row.AccountName,
		AuthenticationType: row.AuthenticationType,
		CredentialRef:      row.CredentialRef,
		CreatedAt:          row.CreatedAt,
	}
}
