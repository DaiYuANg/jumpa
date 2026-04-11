package identity

import (
	"context"
	"errors"

	"github.com/DaiYuANg/arcgo/collectionx"
)

var (
	ErrUnsupportedIdentityBackend = errors.New("identity backend is not implemented")
	ErrInvalidCredentials         = errors.New("invalid credentials")
)

type PasswordCredentials struct {
	Username   string
	Password   string
	RemoteAddr string
}

type Authentication struct {
	Username   string
	Provider   ProviderDescriptor
	Attributes collectionx.Map[string, any]
}

type Authenticator interface {
	Descriptor() ProviderDescriptor
	SupportsPassword() bool
	AuthenticatePassword(ctx context.Context, credentials PasswordCredentials) (Authentication, error)
}
