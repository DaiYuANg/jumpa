package dbx

import (
	"context"

	"github.com/DaiYuANg/arcgo-rbac-template/internal/modules/iam/infrastructure/persistence"
	"github.com/DaiYuANg/arcgo/dbx"
	"github.com/DaiYuANg/arcgo/dbx/repository"
)

type authPrincipalRow struct {
	ID    string `dbx:"id"`
	Email string `dbx:"email"`
}

type authPrincipalSchema struct {
	dbx.Schema[authPrincipalRow]
	ID    dbx.Column[authPrincipalRow, string] `dbx:"id,pk"`
	Email dbx.Column[authPrincipalRow, string] `dbx:"email"`
}

type authPrincipalRoleRow struct {
	PrincipalID string `dbx:"principal_id"`
	Role        string `dbx:"role"`
}

type authPrincipalRoleSchema struct {
	dbx.Schema[authPrincipalRoleRow]
	PrincipalID dbx.Column[authPrincipalRoleRow, string] `dbx:"principal_id"`
	Role        dbx.Column[authPrincipalRoleRow, string] `dbx:"role"`
}

type authPrincipalRepo struct {
	aps              authPrincipalSchema
	apr              authPrincipalRoleSchema
	principalRepo     *repository.Base[authPrincipalRow, authPrincipalSchema]
	principalRoleRepo *repository.Base[authPrincipalRoleRow, authPrincipalRoleSchema]
}

func NewAuthPrincipalRepository(db *dbx.DB) persistence.AuthPrincipalRepository {
	aps := dbx.MustSchema("app_auth_principals", authPrincipalSchema{})
	apr := dbx.MustSchema("app_auth_principal_roles", authPrincipalRoleSchema{})
	return &authPrincipalRepo{
		aps:               aps,
		apr:               apr,
		principalRepo:     repository.New[authPrincipalRow](db, aps),
		principalRoleRepo: repository.New[authPrincipalRoleRow](db, apr),
	}
}

func (r *authPrincipalRepo) UpsertAuthPrincipal(ctx context.Context, userID int64, email string) error {
	id := principalIDByUser(userID)
	row := authPrincipalRow{ID: id, Email: email}
	return r.principalRepo.Upsert(ctx, &row, "id")
}

func (r *authPrincipalRepo) DeleteAuthPrincipal(ctx context.Context, userID int64) error {
	id := principalIDByUser(userID)
	_, _ = r.principalRoleRepo.Delete(ctx, dbx.DeleteFrom(r.apr).Where(r.apr.PrincipalID.Eq(id)))
	_, err := r.principalRepo.DeleteByID(ctx, id)
	return err
}

func (r *authPrincipalRepo) SetAuthPrincipalRoles(ctx context.Context, userID int64, roleIDs []string) error {
	id := principalIDByUser(userID)
	if _, err := r.principalRoleRepo.Delete(ctx, dbx.DeleteFrom(r.apr).Where(r.apr.PrincipalID.Eq(id))); err != nil {
		return err
	}
	for _, roleID := range roleIDs {
		if roleID == "" {
			continue
		}
		row := authPrincipalRoleRow{PrincipalID: id, Role: roleID}
		if err := r.principalRoleRepo.Create(ctx, &row); err != nil {
			return err
		}
	}
	return nil
}

