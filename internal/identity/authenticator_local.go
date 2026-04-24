package identity

import (
	"context"
	"strings"
	"time"

	"github.com/arcgolabs/collectionx"
	"github.com/arcgolabs/dbx"
	columnx "github.com/arcgolabs/dbx/column"
	mapperx "github.com/arcgolabs/dbx/mapper"
	"github.com/arcgolabs/dbx/querydsl"
	schemax "github.com/arcgolabs/dbx/schema"
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
	schemax.Schema[localUserRow]
	ID           columnx.Column[localUserRow, string]    `dbx:"id,pk"`
	Username     columnx.Column[localUserRow, string]    `dbx:"username"`
	Email        columnx.Column[localUserRow, string]    `dbx:"email"`
	PasswordHash columnx.Column[localUserRow, string]    `dbx:"password_hash"`
	IsActive     columnx.Column[localUserRow, bool]      `dbx:"is_active"`
	CreatedAt    columnx.Column[localUserRow, time.Time] `dbx:"created_at"`
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
	us := schemax.MustSchema("users", localUserSchema{})
	value := strings.TrimSpace(usernameOrEmail)

	rows, err := dbx.QueryAll[localUserRow](
		ctx,
		a.db,
		querydsl.Select(querydsl.AllColumns(us).Values()...).
			From(us).
			Where(querydsl.Or(
				us.Username.Eq(value),
				us.Email.Eq(value),
			)).
			Limit(1),
		mapperx.MustMapper[localUserRow](us),
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
