//go:build darwin && !cgo

package identity

import (
	"context"
	"fmt"
)

type platformOSPasswordBackend struct {
	cfg osBackendConfig
}

func newOSPasswordBackend(cfg osBackendConfig) osPasswordBackend {
	return platformOSPasswordBackend{cfg: cfg}
}

func (b platformOSPasswordBackend) Name() string {
	return "opendirectory"
}

func (b platformOSPasswordBackend) Available() bool {
	return false
}

func (b platformOSPasswordBackend) AuthenticatePassword(_ context.Context, provider ProviderDescriptor, _ PasswordCredentials) (Authentication, error) {
	return Authentication{}, fmt.Errorf("%w: macOS backend %q requires cgo and OpenDirectory framework access", ErrUnsupportedIdentityBackend, provider.Backend)
}
