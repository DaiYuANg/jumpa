package identity

import (
	"context"
	"strings"
)

type osAuthenticator struct {
	provider ProviderDescriptor
	backend  osPasswordBackend
}

func NewOSAuthenticator(provider ProviderDescriptor, backendConfig osBackendConfig) Authenticator {
	return &osAuthenticator{provider: provider, backend: newOSPasswordBackend(backendConfig)}
}

func (a *osAuthenticator) Descriptor() ProviderDescriptor {
	return a.provider
}

func (a *osAuthenticator) SupportsPassword() bool {
	return a.backend != nil && a.backend.Available()
}

func (a *osAuthenticator) AuthenticatePassword(ctx context.Context, credentials PasswordCredentials) (Authentication, error) {
	if strings.TrimSpace(credentials.Username) == "" || strings.TrimSpace(credentials.Password) == "" {
		return Authentication{}, ErrInvalidCredentials
	}

	if a.backend == nil {
		return Authentication{}, ErrUnsupportedIdentityBackend
	}
	return a.backend.AuthenticatePassword(ctx, a.provider, credentials)
}
