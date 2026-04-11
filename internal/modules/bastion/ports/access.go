package ports

import (
	"context"
	"time"

	"github.com/samber/mo"
)

type AccessPolicyRecord struct {
	ID                string
	Name              string
	SubjectType       string
	SubjectRef        string
	TargetType        string
	TargetRef         string
	AccountPattern    string
	Protocol          string
	ApprovalRequired  bool
	RecordingRequired bool
	CreatedAt         time.Time
}

type CreateAccessPolicyRecordInput struct {
	Name              string
	SubjectType       string
	SubjectRef        string
	TargetType        string
	TargetRef         string
	AccountPattern    string
	Protocol          string
	ApprovalRequired  bool
	RecordingRequired bool
	CreatedAt         time.Time
}

type PatchAccessPolicyRecordInput struct {
	Name              *string
	SubjectType       *string
	SubjectRef        *string
	TargetType        *string
	TargetRef         *string
	AccountPattern    *string
	Protocol          *string
	ApprovalRequired  *bool
	RecordingRequired *bool
}

type AccessRequestRecord struct {
	ID                string
	PolicyID          string
	PrincipalName     string
	PrincipalEmail    *string
	HostName          string
	HostAccount       string
	Protocol          string
	Status            string
	RequestedAt       time.Time
	ReviewedAt        *time.Time
	ReviewedBy        *string
	ReviewComment     *string
	ApprovedUntil     *time.Time
	ConsumedAt        *time.Time
	ConsumedSessionID *string
}

type FindAccessRequestInput struct {
	PolicyID       string
	PrincipalName  string
	PrincipalEmail *string
	HostName       string
	HostAccount    string
	Protocol       string
}

type CreateAccessRequestInput struct {
	PolicyID       string
	PrincipalName  string
	PrincipalEmail *string
	HostName       string
	HostAccount    string
	Protocol       string
	RequestedAt    time.Time
}

type ListAccessRequestsInput struct {
	Status string
	Limit  int
	Offset int
}

type PolicyRepository interface {
	ListPolicies(ctx context.Context) ([]AccessPolicyRecord, error)
	GetPolicyByID(ctx context.Context, id string) (mo.Option[AccessPolicyRecord], error)
	CreatePolicy(ctx context.Context, in CreateAccessPolicyRecordInput) (AccessPolicyRecord, error)
	UpdatePolicy(ctx context.Context, id string, in PatchAccessPolicyRecordInput) (mo.Option[AccessPolicyRecord], error)
	DeletePolicy(ctx context.Context, id string) (bool, error)
}

type PrincipalAccessRepository interface {
	ListRoleIDsByEmail(ctx context.Context, email string) ([]string, error)
}

type AccessRequestRepository interface {
	ListRequests(ctx context.Context, in ListAccessRequestsInput) ([]AccessRequestRecord, int, error)
	GetRequestByID(ctx context.Context, id string) (mo.Option[AccessRequestRecord], error)
	FindLatestRequest(ctx context.Context, in FindAccessRequestInput) (mo.Option[AccessRequestRecord], error)
	CreateRequest(ctx context.Context, in CreateAccessRequestInput) (AccessRequestRecord, error)
	UpdateRequestStatus(ctx context.Context, id, status string, reviewedAt time.Time, reviewedBy string, reviewComment *string, approvedUntil *time.Time) (mo.Option[AccessRequestRecord], error)
	ConsumeRequest(ctx context.Context, id string, consumedAt time.Time, consumedSessionID *string) (mo.Option[AccessRequestRecord], error)
}
