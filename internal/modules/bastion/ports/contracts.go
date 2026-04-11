package ports

import (
	"context"
	"time"

	"github.com/samber/mo"
)

type HostRecord struct {
	ID                 string
	Name               string
	Address            string
	Port               int
	Protocol           string
	Environment        *string
	Platform           *string
	AuthenticationType string
	CredentialRef      *string
	JumpEnabled        bool
	RecordingPolicy    string
	CreatedAt          time.Time
}

type HostAccountRecord struct {
	ID                 string
	HostID             string
	AccountName        string
	AuthenticationType string
	CredentialRef      *string
	CreatedAt          time.Time
}

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

type CreateHostRecordInput struct {
	Name               string
	Address            string
	Port               int
	Protocol           string
	Environment        *string
	Platform           *string
	AuthenticationType string
	CredentialRef      *string
	JumpEnabled        bool
	RecordingPolicy    string
	CreatedAt          time.Time
}

type PatchHostRecordInput struct {
	Name               *string
	Address            *string
	Port               *int
	Protocol           *string
	Environment        *string
	Platform           *string
	AuthenticationType *string
	CredentialRef      *string
	JumpEnabled        *bool
	RecordingPolicy    *string
}

type CreateHostAccountRecordInput struct {
	HostID             string
	AccountName        string
	AuthenticationType string
	CredentialRef      *string
	CreatedAt          time.Time
}

type PatchHostAccountRecordInput struct {
	AccountName        *string
	AuthenticationType *string
	CredentialRef      *string
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

type SessionRecord struct {
	ID            string
	HostID        string
	HostName      string
	HostAccountID *string
	HostAccount   string
	PrincipalID   string
	PrincipalName string
	Protocol      string
	Status        string
	SourceAddr    *string
	StartedAt     time.Time
	EndedAt       *time.Time
}

type CreateSessionInput struct {
	HostID        string
	HostAccountID *string
	PrincipalID   string
	Protocol      string
	Status        string
	SourceAddr    *string
	StartedAt     time.Time
}

type CreateSessionEventInput struct {
	SessionID string
	EventType string
	Payload   *string
	CreatedAt time.Time
}

type HostRepository interface {
	ListHosts(ctx context.Context) ([]HostRecord, error)
	GetHostByID(ctx context.Context, id string) (mo.Option[HostRecord], error)
	GetHostByName(ctx context.Context, name string) (mo.Option[HostRecord], error)
	CreateHost(ctx context.Context, in CreateHostRecordInput) (HostRecord, error)
	UpdateHost(ctx context.Context, id string, in PatchHostRecordInput) (mo.Option[HostRecord], error)
	DeleteHost(ctx context.Context, id string) (bool, error)
}

type HostAccountRepository interface {
	GetHostAccountByID(ctx context.Context, hostID, accountID string) (mo.Option[HostAccountRecord], error)
	GetHostAccountByName(ctx context.Context, hostID, accountName string) (mo.Option[HostAccountRecord], error)
	ListHostAccountsByHostID(ctx context.Context, hostID string) ([]HostAccountRecord, error)
	CreateHostAccount(ctx context.Context, in CreateHostAccountRecordInput) (HostAccountRecord, error)
	UpdateHostAccount(ctx context.Context, hostID, accountID string, in PatchHostAccountRecordInput) (mo.Option[HostAccountRecord], error)
	DeleteHostAccount(ctx context.Context, hostID, accountID string) (bool, error)
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

type SessionRepository interface {
	ListSessions(ctx context.Context) ([]SessionRecord, error)
	CreateSession(ctx context.Context, in CreateSessionInput) (string, error)
	UpdateSessionStatus(ctx context.Context, id, status string, endedAt *time.Time) error
}

type SessionEventRepository interface {
	CreateSessionEvent(ctx context.Context, in CreateSessionEventInput) error
}
