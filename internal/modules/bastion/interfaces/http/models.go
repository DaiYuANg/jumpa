package http

import (
	"time"

	"github.com/DaiYuANg/jumpa/internal/identity"
)

type overviewDTO struct {
	ProductName        string                      `json:"productName"`
	DatabaseDriver     string                      `json:"databaseDriver"`
	CacheEnabled       bool                        `json:"cacheEnabled"`
	BastionEnabled     bool                        `json:"bastionEnabled"`
	SSHListenAddr      string                      `json:"sshListenAddr"`
	RecordingDir       string                      `json:"recordingDir"`
	IdentityProvider   identity.ProviderDescriptor `json:"identityProvider"`
	IdentityModes      []string                    `json:"identityModes"`
	PasswordAuthReady  bool                        `json:"passwordAuthReady"`
	SupportedDrivers   []string                    `json:"supportedDrivers"`
	SupportedProtocols []string                    `json:"supportedProtocols"`
	CapabilityNotes    []string                    `json:"capabilityNotes,omitempty"`
	GeneratedAt        time.Time                   `json:"generatedAt"`
}

type hostDTO struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Address         string    `json:"address"`
	Port            int       `json:"port"`
	Protocol        string    `json:"protocol"`
	Environment     string    `json:"environment,omitempty"`
	Platform        string    `json:"platform,omitempty"`
	Authentication  string    `json:"authentication"`
	JumpEnabled     bool      `json:"jumpEnabled"`
	RecordingPolicy string    `json:"recordingPolicy"`
	CreatedAt       time.Time `json:"createdAt"`
}

type hostAccountDTO struct {
	ID                 string    `json:"id"`
	HostID             string    `json:"hostId"`
	AccountName        string    `json:"accountName"`
	AuthenticationType string    `json:"authenticationType"`
	CredentialRef      *string   `json:"credentialRef,omitempty"`
	CreatedAt          time.Time `json:"createdAt"`
}

type policyDTO struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	SubjectType       string    `json:"subjectType"`
	SubjectName       string    `json:"subjectName"`
	TargetType        string    `json:"targetType"`
	TargetName        string    `json:"targetName"`
	AccountPattern    string    `json:"accountPattern"`
	Protocol          string    `json:"protocol"`
	ApprovalRequired  bool      `json:"approvalRequired"`
	RecordingRequired bool      `json:"recordingRequired"`
	CreatedAt         time.Time `json:"createdAt"`
}

type sessionDTO struct {
	ID            string     `json:"id"`
	HostName      string     `json:"hostName"`
	HostAccount   string     `json:"hostAccount"`
	PrincipalName string     `json:"principalName"`
	Protocol      string     `json:"protocol"`
	Status        string     `json:"status"`
	StartedAt     time.Time  `json:"startedAt"`
	EndedAt       *time.Time `json:"endedAt,omitempty"`
}

type accessRequestDTO struct {
	ID             string     `json:"id"`
	PolicyID       string     `json:"policyId"`
	PrincipalName  string     `json:"principalName"`
	PrincipalEmail string     `json:"principalEmail,omitempty"`
	HostName       string     `json:"hostName"`
	HostAccount    string     `json:"hostAccount"`
	Protocol       string     `json:"protocol"`
	Status         string     `json:"status"`
	RequestedAt    time.Time  `json:"requestedAt"`
	ReviewedAt     *time.Time `json:"reviewedAt,omitempty"`
	ReviewedBy     *string    `json:"reviewedBy,omitempty"`
	ReviewComment  *string    `json:"reviewComment,omitempty"`
}

type createHostInput struct {
	Body struct {
		Name            string  `json:"name" validate:"required,min=1,max=128"`
		Address         string  `json:"address" validate:"required,min=1,max=255"`
		Port            int     `json:"port" validate:"omitempty,min=1,max=65535"`
		Protocol        string  `json:"protocol" validate:"omitempty,min=1,max=32"`
		Environment     *string `json:"environment"`
		Platform        *string `json:"platform"`
		Authentication  string  `json:"authentication" validate:"omitempty,min=1,max=32"`
		CredentialRef   *string `json:"credentialRef"`
		JumpEnabled     *bool   `json:"jumpEnabled"`
		RecordingPolicy string  `json:"recordingPolicy" validate:"omitempty,min=1,max=32"`
	} `json:"body"`
}

type patchHostInput struct {
	ID   string `path:"id" validate:"required"`
	Body struct {
		Name            *string `json:"name,omitempty" validate:"omitempty,min=1,max=128"`
		Address         *string `json:"address,omitempty" validate:"omitempty,min=1,max=255"`
		Port            *int    `json:"port,omitempty" validate:"omitempty,min=1,max=65535"`
		Protocol        *string `json:"protocol,omitempty" validate:"omitempty,min=1,max=32"`
		Environment     *string `json:"environment,omitempty"`
		Platform        *string `json:"platform,omitempty"`
		Authentication  *string `json:"authentication,omitempty" validate:"omitempty,min=1,max=32"`
		CredentialRef   *string `json:"credentialRef,omitempty"`
		JumpEnabled     *bool   `json:"jumpEnabled,omitempty"`
		RecordingPolicy *string `json:"recordingPolicy,omitempty" validate:"omitempty,min=1,max=32"`
	} `json:"body"`
}

type hostAccountByIDInput struct {
	HostID    string `path:"hostId" validate:"required"`
	AccountID string `path:"accountId" validate:"required"`
}

type hostAccountsByHostInput struct {
	HostID string `path:"hostId" validate:"required"`
}

type createHostAccountInput struct {
	HostID string `path:"hostId" validate:"required"`
	Body   struct {
		AccountName        string  `json:"accountName" validate:"required,min=1,max=128"`
		AuthenticationType string  `json:"authenticationType" validate:"omitempty,min=1,max=32"`
		CredentialRef      *string `json:"credentialRef"`
	} `json:"body"`
}

type patchHostAccountInput struct {
	HostID    string `path:"hostId" validate:"required"`
	AccountID string `path:"accountId" validate:"required"`
	Body      struct {
		AccountName        *string `json:"accountName,omitempty" validate:"omitempty,min=1,max=128"`
		AuthenticationType *string `json:"authenticationType,omitempty" validate:"omitempty,min=1,max=32"`
		CredentialRef      *string `json:"credentialRef,omitempty"`
	} `json:"body"`
}

type createPolicyInput struct {
	Body struct {
		Name              string `json:"name" validate:"required,min=1,max=128"`
		SubjectType       string `json:"subjectType" validate:"omitempty,min=1,max=32"`
		SubjectName       string `json:"subjectName" validate:"omitempty,max=128"`
		TargetType        string `json:"targetType" validate:"omitempty,min=1,max=32"`
		TargetName        string `json:"targetName" validate:"omitempty,max=128"`
		AccountPattern    string `json:"accountPattern" validate:"omitempty,max=128"`
		Protocol          string `json:"protocol" validate:"omitempty,min=1,max=32"`
		ApprovalRequired  *bool  `json:"approvalRequired"`
		RecordingRequired *bool  `json:"recordingRequired"`
	} `json:"body"`
}

type patchPolicyInput struct {
	ID   string `path:"id" validate:"required"`
	Body struct {
		Name              *string `json:"name,omitempty" validate:"omitempty,min=1,max=128"`
		SubjectType       *string `json:"subjectType,omitempty" validate:"omitempty,min=1,max=32"`
		SubjectName       *string `json:"subjectName,omitempty" validate:"omitempty,max=128"`
		TargetType        *string `json:"targetType,omitempty" validate:"omitempty,min=1,max=32"`
		TargetName        *string `json:"targetName,omitempty" validate:"omitempty,max=128"`
		AccountPattern    *string `json:"accountPattern,omitempty" validate:"omitempty,max=128"`
		Protocol          *string `json:"protocol,omitempty" validate:"omitempty,min=1,max=32"`
		ApprovalRequired  *bool   `json:"approvalRequired,omitempty"`
		RecordingRequired *bool   `json:"recordingRequired,omitempty"`
	} `json:"body"`
}

type reviewAccessRequestInput struct {
	ID   string `path:"id" validate:"required"`
	Body struct {
		Reviewer string  `json:"reviewer" validate:"required,min=1,max=128"`
		Comment  *string `json:"comment,omitempty" validate:"omitempty,max=500"`
	} `json:"body"`
}
