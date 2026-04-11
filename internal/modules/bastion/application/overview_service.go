package application

import (
	"context"
	"time"

	config2 "github.com/DaiYuANg/jumpa/internal/config"
	"github.com/DaiYuANg/jumpa/internal/identity"
	bastiondomain "github.com/DaiYuANg/jumpa/internal/modules/bastion/domain"
)

type overviewService struct {
	cfg           config2.AppConfig
	identity      identity.ProviderDescriptor
	authenticator identity.Authenticator
}

func NewOverviewService(cfg config2.AppConfig, provider identity.ProviderDescriptor, authenticator identity.Authenticator) OverviewService {
	return &overviewService{cfg: cfg, identity: provider, authenticator: authenticator}
}

func (s *overviewService) Get(_ context.Context) (bastiondomain.Overview, error) {
	return bastiondomain.Overview{
		ProductName:      s.cfg.App.Name,
		DatabaseDriver:   s.cfg.DB.Driver,
		CacheEnabled:     s.cfg.Valkey.Enabled,
		BastionEnabled:   s.cfg.Bastion.Enabled,
		SSHListenAddr:    s.cfg.Bastion.SSH.ListenAddr,
		RecordingDir:     s.cfg.Bastion.Session.RecordingDirectory,
		IdentityProvider: s.identity,
		IdentityModes: []string{
			"local",
			"os",
		},
		PasswordAuthReady: s.authenticator.SupportsPassword(),
		SupportedDrivers:  []string{"sqlite", "mariadb", "postgres"},
		SupportedProtocols: []string{
			"ssh",
			"sftp",
		},
		CapabilityNotes: []string{
			"Current landing includes a dedicated SSH gateway runtime with downstream SSH proxying and persisted session lifecycle records.",
			"Keep OS-backed login as an authentication source while storing bastion authorization, target mapping, and audit state inside the application database.",
			"The gateway listener is now split out as a dedicated runtime in cmd/gateway.",
		},
		GeneratedAt: time.Now().UTC(),
	}, nil
}
