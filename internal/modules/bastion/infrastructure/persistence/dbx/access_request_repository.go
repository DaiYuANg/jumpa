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

type accessRequestRow struct {
	ID             int64      `dbx:"id"`
	PolicyID       int64      `dbx:"policy_id"`
	PrincipalName  string     `dbx:"principal_name"`
	PrincipalEmail *string    `dbx:"principal_email"`
	HostName       string     `dbx:"host_name"`
	HostAccount    string     `dbx:"host_account"`
	Protocol       string     `dbx:"protocol"`
	Status         string     `dbx:"status"`
	RequestedAt    time.Time  `dbx:"requested_at"`
	ReviewedAt     *time.Time `dbx:"reviewed_at"`
	ReviewedBy     *string    `dbx:"reviewed_by"`
	ReviewComment  *string    `dbx:"review_comment"`
}

type accessRequestSchema struct {
	dbx.Schema[accessRequestRow]
	ID             dbx.IDColumn[accessRequestRow, int64, dbx.IDSnowflake] `dbx:"id,pk"`
	PolicyID       dbx.Column[accessRequestRow, int64]                    `dbx:"policy_id"`
	PrincipalName  dbx.Column[accessRequestRow, string]                   `dbx:"principal_name"`
	PrincipalEmail dbx.Column[accessRequestRow, *string]                  `dbx:"principal_email"`
	HostName       dbx.Column[accessRequestRow, string]                   `dbx:"host_name"`
	HostAccount    dbx.Column[accessRequestRow, string]                   `dbx:"host_account"`
	Protocol       dbx.Column[accessRequestRow, string]                   `dbx:"protocol"`
	Status         dbx.Column[accessRequestRow, string]                   `dbx:"status"`
	RequestedAt    dbx.Column[accessRequestRow, time.Time]                `dbx:"requested_at"`
	ReviewedAt     dbx.Column[accessRequestRow, *time.Time]               `dbx:"reviewed_at"`
	ReviewedBy     dbx.Column[accessRequestRow, *string]                  `dbx:"reviewed_by"`
	ReviewComment  dbx.Column[accessRequestRow, *string]                  `dbx:"review_comment"`
}

type accessRequestRepo struct {
	ars  accessRequestSchema
	repo *repository.Base[accessRequestRow, accessRequestSchema]
}

func NewAccessRequestRepository(db *dbx.DB) ports.AccessRequestRepository {
	ars := dbx.MustSchema("bastion_access_requests", accessRequestSchema{})
	return &accessRequestRepo{ars: ars, repo: repository.New[accessRequestRow](db, ars)}
}

func (r *accessRequestRepo) ListRequests(ctx context.Context) ([]ports.AccessRequestRecord, error) {
	rows, err := r.repo.ListSpec(ctx, repository.OrderBy(r.ars.RequestedAt.Desc()))
	if err != nil {
		return nil, err
	}
	return collectionx.MapList(rows, func(_ int, row accessRequestRow) ports.AccessRequestRecord {
		return toAccessRequestRecord(row)
	}).Values(), nil
}

func (r *accessRequestRepo) GetRequestByID(ctx context.Context, id string) (mo.Option[ports.AccessRequestRecord], error) {
	requestID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return mo.None[ports.AccessRequestRecord](), err
	}
	row, err := r.repo.FirstSpec(ctx, repository.Where(r.ars.ID.Eq(requestID)))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return mo.None[ports.AccessRequestRecord](), nil
		}
		return mo.None[ports.AccessRequestRecord](), err
	}
	return mo.Some(toAccessRequestRecord(row)), nil
}

func (r *accessRequestRepo) FindLatestRequest(ctx context.Context, in ports.FindAccessRequestInput) (mo.Option[ports.AccessRequestRecord], error) {
	policyID, err := strconv.ParseInt(in.PolicyID, 10, 64)
	if err != nil {
		return mo.None[ports.AccessRequestRecord](), err
	}
	specs := []repository.Spec{
		repository.Where(r.ars.PolicyID.Eq(policyID)),
		repository.Where(r.ars.PrincipalName.Eq(in.PrincipalName)),
		repository.Where(r.ars.HostName.Eq(in.HostName)),
		repository.Where(r.ars.HostAccount.Eq(in.HostAccount)),
		repository.Where(r.ars.Protocol.Eq(in.Protocol)),
		repository.OrderBy(r.ars.RequestedAt.Desc()),
	}
	if in.PrincipalEmail != nil {
		specs = append(specs, repository.Where(r.ars.PrincipalEmail.Eq(in.PrincipalEmail)))
	}
	row, err := r.repo.FirstSpec(ctx, specs...)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return mo.None[ports.AccessRequestRecord](), nil
		}
		return mo.None[ports.AccessRequestRecord](), err
	}
	return mo.Some(toAccessRequestRecord(row)), nil
}

func (r *accessRequestRepo) CreateRequest(ctx context.Context, in ports.CreateAccessRequestInput) (ports.AccessRequestRecord, error) {
	policyID, err := strconv.ParseInt(in.PolicyID, 10, 64)
	if err != nil {
		return ports.AccessRequestRecord{}, err
	}
	row := &accessRequestRow{
		PolicyID:       policyID,
		PrincipalName:  in.PrincipalName,
		PrincipalEmail: in.PrincipalEmail,
		HostName:       in.HostName,
		HostAccount:    in.HostAccount,
		Protocol:       in.Protocol,
		Status:         "pending",
		RequestedAt:    in.RequestedAt,
	}
	if err := r.repo.Create(ctx, row); err != nil {
		return ports.AccessRequestRecord{}, err
	}
	return toAccessRequestRecord(*row), nil
}

func (r *accessRequestRepo) UpdateRequestStatus(ctx context.Context, id, status string, reviewedAt time.Time, reviewedBy string, reviewComment *string) (mo.Option[ports.AccessRequestRecord], error) {
	requestID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return mo.None[ports.AccessRequestRecord](), err
	}
	assignments := []dbx.Assignment{
		r.ars.Status.Set(status),
		r.ars.ReviewedAt.Set(&reviewedAt),
		r.ars.ReviewedBy.Set(&reviewedBy),
		r.ars.ReviewComment.Set(reviewComment),
	}
	res, err := r.repo.UpdateByID(ctx, requestID, assignments...)
	if err != nil {
		return mo.None[ports.AccessRequestRecord](), err
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return mo.None[ports.AccessRequestRecord](), nil
	}
	return r.GetRequestByID(ctx, id)
}

func toAccessRequestRecord(row accessRequestRow) ports.AccessRequestRecord {
	return ports.AccessRequestRecord{
		ID:             strconv.FormatInt(row.ID, 10),
		PolicyID:       strconv.FormatInt(row.PolicyID, 10),
		PrincipalName:  row.PrincipalName,
		PrincipalEmail: row.PrincipalEmail,
		HostName:       row.HostName,
		HostAccount:    row.HostAccount,
		Protocol:       row.Protocol,
		Status:         row.Status,
		RequestedAt:    row.RequestedAt,
		ReviewedAt:     row.ReviewedAt,
		ReviewedBy:     row.ReviewedBy,
		ReviewComment:  row.ReviewComment,
	}
}
