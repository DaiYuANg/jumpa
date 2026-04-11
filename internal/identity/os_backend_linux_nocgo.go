//go:build linux && !cgo

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
	return "pam"
}

func (b platformOSPasswordBackend) Available() bool {
	return false
}

func (b platformOSPasswordBackend) AuthenticatePassword(_ context.Context, provider ProviderDescriptor, _ PasswordCredentials) (Authentication, error) {
	return Authentication{}, fmt.Errorf("%w: linux backend %q requires cgo and libpam development headers", ErrUnsupportedIdentityBackend, provider.Backend)
}
