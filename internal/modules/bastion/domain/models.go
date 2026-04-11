package domain

import (
	"time"

	"github.com/DaiYuANg/jumpa/internal/identity"
)

type Host struct {
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

type HostAccount struct {
	ID                 string     `json:"id"`
	HostID             string     `json:"hostId"`
	AccountName        string     `json:"accountName"`
	AuthenticationType string     `json:"authenticationType"`
	CredentialRef      *string    `json:"credentialRef,omitempty"`
	CreatedAt          time.Time  `json:"createdAt"`
}

type AccessPolicy struct {
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

type Session struct {
	ID            string     `json:"id"`
	HostName      string     `json:"hostName"`
	HostAccount   string     `json:"hostAccount"`
	PrincipalName string     `json:"principalName"`
	Protocol      string     `json:"protocol"`
	Status        string     `json:"status"`
	StartedAt     time.Time  `json:"startedAt"`
	EndedAt       *time.Time `json:"endedAt,omitempty"`
}

type Overview struct {
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
