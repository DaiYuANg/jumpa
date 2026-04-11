package domain

import (
	"time"

	"github.com/DaiYuANg/jumpa/internal/identity"
)

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
