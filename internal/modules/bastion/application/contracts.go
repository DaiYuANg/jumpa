package application

import (
	"context"

	bastiondomain "github.com/DaiYuANg/jumpa/internal/modules/bastion/domain"
	"github.com/samber/mo"
)

type OverviewService interface {
	Get(ctx context.Context) (bastiondomain.Overview, error)
}

type AssetService interface {
	ListHosts(ctx context.Context) ([]bastiondomain.Host, error)
	GetHost(ctx context.Context, id string) (mo.Option[bastiondomain.Host], error)
	CreateHost(ctx context.Context, in CreateHostInput) (bastiondomain.Host, error)
	UpdateHost(ctx context.Context, id string, in UpdateHostInput) (mo.Option[bastiondomain.Host], error)
	DeleteHost(ctx context.Context, id string) (bool, error)
	ListHostAccounts(ctx context.Context, hostID string) ([]bastiondomain.HostAccount, error)
	GetHostAccount(ctx context.Context, hostID, accountID string) (mo.Option[bastiondomain.HostAccount], error)
	CreateHostAccount(ctx context.Context, hostID string, in CreateHostAccountInput) (bastiondomain.HostAccount, error)
	UpdateHostAccount(ctx context.Context, hostID, accountID string, in UpdateHostAccountInput) (mo.Option[bastiondomain.HostAccount], error)
	DeleteHostAccount(ctx context.Context, hostID, accountID string) (bool, error)
}

type TargetService interface {
	GetHostByName(ctx context.Context, name string) (mo.Option[bastiondomain.Host], error)
	GetHostAccountByName(ctx context.Context, hostID, accountName string) (mo.Option[bastiondomain.HostAccount], error)
}

type PolicyService interface {
	ListPolicies(ctx context.Context) ([]bastiondomain.AccessPolicy, error)
	GetPolicy(ctx context.Context, id string) (mo.Option[bastiondomain.AccessPolicy], error)
	CreatePolicy(ctx context.Context, in CreatePolicyInput) (bastiondomain.AccessPolicy, error)
	UpdatePolicy(ctx context.Context, id string, in UpdatePolicyInput) (mo.Option[bastiondomain.AccessPolicy], error)
	DeletePolicy(ctx context.Context, id string) (bool, error)
}

type AccessCheckInput struct {
	PrincipalName  string
	PrincipalEmail string
	HostName       string
	AccountName    string
	Protocol       string
}

type AccessDecision struct {
	Allowed           bool
	ApprovalRequired  bool
	RecordingRequired bool
	MatchedPolicyID   string
	RequestID         string
	Reason            string
}

type AccessService interface {
	Authorize(ctx context.Context, in AccessCheckInput) (AccessDecision, error)
	ConsumeApprovedRequest(ctx context.Context, requestID, sessionID string) error
}

type CreateHostInput struct {
	Name            string
	Address         string
	Port            int
	Protocol        string
	Environment     *string
	Platform        *string
	Authentication  string
	CredentialRef   *string
	JumpEnabled     bool
	RecordingPolicy string
}

type UpdateHostInput struct {
	Name            *string
	Address         *string
	Port            *int
	Protocol        *string
	Environment     *string
	Platform        *string
	Authentication  *string
	CredentialRef   *string
	JumpEnabled     *bool
	RecordingPolicy *string
}

type CreateHostAccountInput struct {
	AccountName        string
	AuthenticationType string
	CredentialRef      *string
}

type UpdateHostAccountInput struct {
	AccountName        *string
	AuthenticationType *string
	CredentialRef      *string
}

type CreatePolicyInput struct {
	Name              string
	SubjectType       string
	SubjectRef        string
	TargetType        string
	TargetRef         string
	AccountPattern    string
	Protocol          string
	ApprovalRequired  bool
	RecordingRequired bool
}

type UpdatePolicyInput struct {
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

type SessionService interface {
	ListSessions(ctx context.Context) ([]bastiondomain.Session, error)
}

type ListAccessRequestsInput struct {
	Status string
	Limit  int
	Offset int
}

type AccessRequestService interface {
	ListRequests(ctx context.Context, in ListAccessRequestsInput) ([]bastiondomain.AccessRequest, int, error)
	GetRequest(ctx context.Context, id string) (mo.Option[bastiondomain.AccessRequest], error)
	Approve(ctx context.Context, id, reviewer string, comment *string) (mo.Option[bastiondomain.AccessRequest], error)
	Reject(ctx context.Context, id, reviewer string, comment *string) (mo.Option[bastiondomain.AccessRequest], error)
}

type StartSessionInput struct {
	PrincipalName string
	HostID        string
	HostName      string
	HostAccountID *string
	HostAccount   string
	Protocol      string
	SourceAddr    string
}

type SessionRuntimeService interface {
	Start(ctx context.Context, in StartSessionInput) (bastiondomain.Session, error)
	MarkActive(ctx context.Context, sessionID string) error
	Finish(ctx context.Context, sessionID, status string) error
}
