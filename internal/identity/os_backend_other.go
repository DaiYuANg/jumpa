//go:build !linux && !windows && !darwin

package identity

import (
	"context"
	"fmt"
	"runtime"
)

type platformOSPasswordBackend struct{}

func newOSPasswordBackend(_ osBackendConfig) osPasswordBackend {
	return platformOSPasswordBackend{}
}

func (platformOSPasswordBackend) Name() string {
	return runtime.GOOS
}

func (platformOSPasswordBackend) Available() bool {
	return false
}

func (platformOSPasswordBackend) AuthenticatePassword(_ context.Context, provider ProviderDescriptor, _ PasswordCredentials) (Authentication, error) {
	return Authentication{}, fmt.Errorf("%w: OS backend %q is not supported on %s", ErrUnsupportedIdentityBackend, provider.Backend, runtime.GOOS)
}
