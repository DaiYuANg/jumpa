package identity

import (
	"context"
	"strings"
	"time"

	"github.com/DaiYuANg/arcgo/collectionx"
	"github.com/DaiYuANg/arcgo/dbx"
	"github.com/samber/mo"
	"golang.org/x/crypto/bcrypt"
)

type localAuthenticator struct {
	provider ProviderDescriptor
	db       *dbx.DB
}

type localUserRow struct {
	ID           string    `dbx:"id"`
	Username     string    `dbx:"username"`
	Email        string    `dbx:"email"`
	PasswordHash string    `dbx:"password_hash"`
	IsActive     bool      `dbx:"is_active"`
	CreatedAt    time.Time `dbx:"created_at"`
}

type localUserSchema struct {
	dbx.Schema[localUserRow]
	ID           dbx.Column[localUserRow, string]    `dbx:"id,pk"`
	Username     dbx.Column[localUserRow, string]    `dbx:"username"`
	Email        dbx.Column[localUserRow, string]    `dbx:"email"`
	PasswordHash dbx.Column[localUserRow, string]    `dbx:"password_hash"`
	IsActive     dbx.Column[localUserRow, bool]      `dbx:"is_active"`
	CreatedAt    dbx.Column[localUserRow, time.Time] `dbx:"created_at"`
}

func NewLocalAuthenticator(provider ProviderDescriptor, db *dbx.DB) Authenticator {
	return &localAuthenticator{provider: provider, db: db}
}

func (a *localAuthenticator) Descriptor() ProviderDescriptor {
	return a.provider
}

func (a *localAuthenticator) SupportsPassword() bool {
	return a.db != nil
}

func (a *localAuthenticator) AuthenticatePassword(ctx context.Context, credentials PasswordCredentials) (Authentication, error) {
	if strings.TrimSpace(credentials.Username) == "" || strings.TrimSpace(credentials.Password) == "" {
		return Authentication{}, ErrInvalidCredentials
	}

	if a.db == nil {
		return Authentication{}, ErrUnsupportedIdentityBackend
	}

	user, err := a.findUser(ctx, credentials.Username)
	if err != nil {
		return Authentication{}, err
	}
	if user.IsAbsent() {
		return Authentication{}, ErrInvalidCredentials
	}

	row := user.MustGet()
	if !row.IsActive {
		return Authentication{}, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(row.PasswordHash), []byte(credentials.Password)); err != nil {
		return Authentication{}, ErrInvalidCredentials
	}

	auth := newAuthentication(row.Username, a.provider, credentials.RemoteAddr)
	auth.Attributes.Set("email", row.Email)
	auth.Attributes.Set("userID", row.ID)
	return auth, nil
}

func (a *localAuthenticator) findUser(ctx context.Context, usernameOrEmail string) (mo.Option[localUserRow], error) {
	us := dbx.MustSchema("users", localUserSchema{})
	value := strings.TrimSpace(usernameOrEmail)

	rows, err := dbx.QueryAll[localUserRow](
		ctx,
		a.db,
		dbx.Select(us.AllColumns().Values()...).
			From(us).
			Where(dbx.Or(
				us.Username.Eq(value),
				us.Email.Eq(value),
			)).
			Limit(1),
		dbx.MustMapper[localUserRow](us),
	)
	if err != nil {
		return mo.None[localUserRow](), err
	}

	return rows.GetFirstOption(), nil
}

func newAuthentication(username string, provider ProviderDescriptor, remoteAddr string) Authentication {
	return Authentication{
		Username: username,
		Provider: provider,
		Attributes: collectionx.NewMapFrom(map[string]any{
			"remoteAddr": remoteAddr,
		}),
	}
}
