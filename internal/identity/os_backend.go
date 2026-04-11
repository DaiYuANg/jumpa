package identity

import (
	"context"

	config2 "github.com/DaiYuANg/jumpa/internal/config"
)

type osPasswordBackend interface {
	Name() string
	Available() bool
	AuthenticatePassword(ctx context.Context, provider ProviderDescriptor, credentials PasswordCredentials) (Authentication, error)
}

type osBackendConfig struct {
	PAMService    string
	DirectoryNode string
}

func toOSBackendConfig(cfg config2.AppConfig) osBackendConfig {
	return osBackendConfig{
		PAMService:    cfg.Identity.OS.PAMService,
		DirectoryNode: cfg.Identity.OS.DirectoryNode,
	}
}
