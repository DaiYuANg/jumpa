package dbx

import (
	"context"
	"errors"
	"strconv"
	"strings"
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

type accessRequestRow struct {
	ID                int64      `dbx:"id"`
	PolicyID          int64      `dbx:"policy_id"`
	PrincipalName     string     `dbx:"principal_name"`
	PrincipalEmail    *string    `dbx:"principal_email"`
	HostName          string     `dbx:"host_name"`
	HostAccount       string     `dbx:"host_account"`
	Protocol          string     `dbx:"protocol"`
	Status            string     `dbx:"status"`
	RequestedAt       time.Time  `dbx:"requested_at"`
	ReviewedAt        *time.Time `dbx:"reviewed_at"`
	ReviewedBy        *string    `dbx:"reviewed_by"`
	ReviewComment     *string    `dbx:"review_comment"`
	ApprovedUntil     *time.Time `dbx:"approved_until"`
	ConsumedAt        *time.Time `dbx:"consumed_at"`
	ConsumedSessionID *int64     `dbx:"consumed_session_id"`
}

type accessRequestSchema struct {
	schemax.Schema[accessRequestRow]
	ID                columnx.IDColumn[accessRequestRow, int64, idgen.IDSnowflake] `dbx:"id,pk"`
	PolicyID          columnx.Column[accessRequestRow, int64]                      `dbx:"policy_id"`
	PrincipalName     columnx.Column[accessRequestRow, string]                     `dbx:"principal_name"`
	PrincipalEmail    columnx.Column[accessRequestRow, *string]                    `dbx:"principal_email"`
	HostName          columnx.Column[accessRequestRow, string]                     `dbx:"host_name"`
	HostAccount       columnx.Column[accessRequestRow, string]                     `dbx:"host_account"`
	Protocol          columnx.Column[accessRequestRow, string]                     `dbx:"protocol"`
	Status            columnx.Column[accessRequestRow, string]                     `dbx:"status"`
	RequestedAt       columnx.Column[accessRequestRow, time.Time]                  `dbx:"requested_at"`
	ReviewedAt        columnx.Column[accessRequestRow, *time.Time]                 `dbx:"reviewed_at"`
	ReviewedBy        columnx.Column[accessRequestRow, *string]                    `dbx:"reviewed_by"`
	ReviewComment     columnx.Column[accessRequestRow, *string]                    `dbx:"review_comment"`
	ApprovedUntil     columnx.Column[accessRequestRow, *time.Time]                 `dbx:"approved_until"`
	ConsumedAt        columnx.Column[accessRequestRow, *time.Time]                 `dbx:"consumed_at"`
	ConsumedSessionID columnx.Column[accessRequestRow, *int64]                     `dbx:"consumed_session_id"`
}

type accessRequestRepo struct {
	ars  accessRequestSchema
	repo *repository.Base[accessRequestRow, accessRequestSchema]
}

func NewAccessRequestRepository(db *dbx.DB) ports.AccessRequestRepository {
	ars := schemax.MustSchema("bastion_access_requests", accessRequestSchema{})
	return &accessRequestRepo{ars: ars, repo: repository.New[accessRequestRow](db, ars)}
}

func (r *accessRequestRepo) ListRequests(ctx context.Context, in ports.ListAccessRequestsInput) ([]ports.AccessRequestRecord, int, error) {
	specs := []repository.Spec{repository.OrderBy(r.ars.RequestedAt.Desc())}
	if status := strings.TrimSpace(in.Status); status != "" {
		specs = append(specs, repository.Where(r.ars.Status.Eq(status)))
	}
	if in.Limit <= 0 {
		in.Limit = 10
	}
	if in.Offset < 0 {
		in.Offset = 0
	}
	total, err := r.repo.CountSpec(ctx, specs...)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.repo.ListSpec(ctx, append(specs, repository.Limit(in.Limit), repository.Offset(in.Offset))...)
	if err != nil {
		return nil, 0, err
	}
	return collectionx.MapList(rows, func(_ int, row accessRequestRow) ports.AccessRequestRecord {
		return toAccessRequestRecord(row)
	}).Values(), int(total), nil
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

func (r *accessRequestRepo) UpdateRequestStatus(ctx context.Context, id, status string, reviewedAt time.Time, reviewedBy string, reviewComment *string, approvedUntil *time.Time) (mo.Option[ports.AccessRequestRecord], error) {
	requestID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return mo.None[ports.AccessRequestRecord](), err
	}
	assignments := []querydsl.Assignment{
		r.ars.Status.Set(status),
		r.ars.ReviewedAt.Set(&reviewedAt),
		r.ars.ReviewedBy.Set(&reviewedBy),
		r.ars.ReviewComment.Set(reviewComment),
		r.ars.ApprovedUntil.Set(approvedUntil),
		r.ars.ConsumedAt.Set(nil),
		r.ars.ConsumedSessionID.Set(nil),
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

func (r *accessRequestRepo) ConsumeRequest(ctx context.Context, id string, consumedAt time.Time, consumedSessionID *string) (mo.Option[ports.AccessRequestRecord], error) {
	requestID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return mo.None[ports.AccessRequestRecord](), err
	}
	var consumedSessionIDInt *int64
	if consumedSessionID != nil && strings.TrimSpace(*consumedSessionID) != "" {
		value, parseErr := strconv.ParseInt(strings.TrimSpace(*consumedSessionID), 10, 64)
		if parseErr != nil {
			return mo.None[ports.AccessRequestRecord](), parseErr
		}
		consumedSessionIDInt = &value
	}
	assignments := []querydsl.Assignment{
		r.ars.Status.Set("consumed"),
		r.ars.ConsumedAt.Set(&consumedAt),
		r.ars.ConsumedSessionID.Set(consumedSessionIDInt),
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
		ID:                strconv.FormatInt(row.ID, 10),
		PolicyID:          strconv.FormatInt(row.PolicyID, 10),
		PrincipalName:     row.PrincipalName,
		PrincipalEmail:    row.PrincipalEmail,
		HostName:          row.HostName,
		HostAccount:       row.HostAccount,
		Protocol:          row.Protocol,
		Status:            row.Status,
		RequestedAt:       row.RequestedAt,
		ReviewedAt:        row.ReviewedAt,
		ReviewedBy:        row.ReviewedBy,
		ReviewComment:     row.ReviewComment,
		ApprovedUntil:     row.ApprovedUntil,
		ConsumedAt:        row.ConsumedAt,
		ConsumedSessionID: formatInt64Ptr(row.ConsumedSessionID),
	}
}

func formatInt64Ptr(v *int64) *string {
	if v == nil {
		return nil
	}
	value := strconv.FormatInt(*v, 10)
	return &value
}
