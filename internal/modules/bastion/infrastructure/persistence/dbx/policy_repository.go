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

type policyRow struct {
	ID                int64     `dbx:"id"`
	Name              string    `dbx:"name"`
	SubjectType       string    `dbx:"subject_type"`
	SubjectRef        string    `dbx:"subject_ref"`
	TargetType        string    `dbx:"target_type"`
	TargetRef         string    `dbx:"target_ref"`
	AccountPattern    string    `dbx:"account_pattern"`
	Protocol          string    `dbx:"protocol"`
	ApprovalRequired  bool      `dbx:"approval_required"`
	RecordingRequired bool      `dbx:"recording_required"`
	CreatedAt         time.Time `dbx:"created_at"`
}

type policySchema struct {
	schemax.Schema[policyRow]
	ID                columnx.IDColumn[policyRow, int64, idgen.IDSnowflake] `dbx:"id,pk"`
	Name              columnx.Column[policyRow, string]                     `dbx:"name"`
	SubjectType       columnx.Column[policyRow, string]                     `dbx:"subject_type"`
	SubjectRef        columnx.Column[policyRow, string]                     `dbx:"subject_ref"`
	TargetType        columnx.Column[policyRow, string]                     `dbx:"target_type"`
	TargetRef         columnx.Column[policyRow, string]                     `dbx:"target_ref"`
	AccountPattern    columnx.Column[policyRow, string]                     `dbx:"account_pattern"`
	Protocol          columnx.Column[policyRow, string]                     `dbx:"protocol"`
	ApprovalRequired  columnx.Column[policyRow, bool]                       `dbx:"approval_required"`
	RecordingRequired columnx.Column[policyRow, bool]                       `dbx:"recording_required"`
	CreatedAt         columnx.Column[policyRow, time.Time]                  `dbx:"created_at"`
}

type policyRepo struct {
	ps   policySchema
	repo *repository.Base[policyRow, policySchema]
}

func NewPolicyRepository(db *dbx.DB) ports.PolicyRepository {
	ps := schemax.MustSchema("bastion_access_policies", policySchema{})
	return &policyRepo{ps: ps, repo: repository.New[policyRow](db, ps)}
}

func (r *policyRepo) ListPolicies(ctx context.Context) ([]ports.AccessPolicyRecord, error) {
	rows, err := r.repo.ListSpec(ctx, repository.OrderBy(r.ps.Name.Asc()))
	if err != nil {
		return nil, err
	}
	return collectionx.MapList(rows, func(_ int, row policyRow) ports.AccessPolicyRecord {
		return ports.AccessPolicyRecord{
			ID:                strconv.FormatInt(row.ID, 10),
			Name:              row.Name,
			SubjectType:       row.SubjectType,
			SubjectRef:        row.SubjectRef,
			TargetType:        row.TargetType,
			TargetRef:         row.TargetRef,
			AccountPattern:    row.AccountPattern,
			Protocol:          row.Protocol,
			ApprovalRequired:  row.ApprovalRequired,
			RecordingRequired: row.RecordingRequired,
			CreatedAt:         row.CreatedAt,
		}
	}).Values(), nil
}

func (r *policyRepo) GetPolicyByID(ctx context.Context, id string) (mo.Option[ports.AccessPolicyRecord], error) {
	policyID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return mo.None[ports.AccessPolicyRecord](), err
	}
	row, err := r.repo.FirstSpec(ctx, repository.Where(r.ps.ID.Eq(policyID)))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return mo.None[ports.AccessPolicyRecord](), nil
		}
		return mo.None[ports.AccessPolicyRecord](), err
	}
	return mo.Some(toPolicyRecord(row)), nil
}

func (r *policyRepo) CreatePolicy(ctx context.Context, in ports.CreateAccessPolicyRecordInput) (ports.AccessPolicyRecord, error) {
	row := &policyRow{
		Name:              in.Name,
		SubjectType:       in.SubjectType,
		SubjectRef:        in.SubjectRef,
		TargetType:        in.TargetType,
		TargetRef:         in.TargetRef,
		AccountPattern:    in.AccountPattern,
		Protocol:          in.Protocol,
		ApprovalRequired:  in.ApprovalRequired,
		RecordingRequired: in.RecordingRequired,
		CreatedAt:         in.CreatedAt,
	}
	if err := r.repo.Create(ctx, row); err != nil {
		return ports.AccessPolicyRecord{}, err
	}
	return toPolicyRecord(*row), nil
}

func (r *policyRepo) UpdatePolicy(ctx context.Context, id string, in ports.PatchAccessPolicyRecordInput) (mo.Option[ports.AccessPolicyRecord], error) {
	policyID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return mo.None[ports.AccessPolicyRecord](), err
	}
	assignments := make([]querydsl.Assignment, 0, 9)
	if in.Name != nil {
		assignments = append(assignments, r.ps.Name.Set(*in.Name))
	}
	if in.SubjectType != nil {
		assignments = append(assignments, r.ps.SubjectType.Set(*in.SubjectType))
	}
	if in.SubjectRef != nil {
		assignments = append(assignments, r.ps.SubjectRef.Set(*in.SubjectRef))
	}
	if in.TargetType != nil {
		assignments = append(assignments, r.ps.TargetType.Set(*in.TargetType))
	}
	if in.TargetRef != nil {
		assignments = append(assignments, r.ps.TargetRef.Set(*in.TargetRef))
	}
	if in.AccountPattern != nil {
		assignments = append(assignments, r.ps.AccountPattern.Set(*in.AccountPattern))
	}
	if in.Protocol != nil {
		assignments = append(assignments, r.ps.Protocol.Set(*in.Protocol))
	}
	if in.ApprovalRequired != nil {
		assignments = append(assignments, r.ps.ApprovalRequired.Set(*in.ApprovalRequired))
	}
	if in.RecordingRequired != nil {
		assignments = append(assignments, r.ps.RecordingRequired.Set(*in.RecordingRequired))
	}
	if len(assignments) == 0 {
		return r.GetPolicyByID(ctx, id)
	}
	res, err := r.repo.UpdateByID(ctx, policyID, assignments...)
	if err != nil {
		return mo.None[ports.AccessPolicyRecord](), err
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return mo.None[ports.AccessPolicyRecord](), nil
	}
	return r.GetPolicyByID(ctx, id)
}

func (r *policyRepo) DeletePolicy(ctx context.Context, id string) (bool, error) {
	policyID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return false, err
	}
	res, err := r.repo.DeleteByID(ctx, policyID)
	if err != nil {
		return false, err
	}
	ra, _ := res.RowsAffected()
	return ra > 0, nil
}

func toPolicyRecord(row policyRow) ports.AccessPolicyRecord {
	return ports.AccessPolicyRecord{
		ID:                strconv.FormatInt(row.ID, 10),
		Name:              row.Name,
		SubjectType:       row.SubjectType,
		SubjectRef:        row.SubjectRef,
		TargetType:        row.TargetType,
		TargetRef:         row.TargetRef,
		AccountPattern:    row.AccountPattern,
		Protocol:          row.Protocol,
		ApprovalRequired:  row.ApprovalRequired,
		RecordingRequired: row.RecordingRequired,
		CreatedAt:         row.CreatedAt,
	}
}
