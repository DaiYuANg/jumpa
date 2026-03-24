package repo

import (
	"context"

	"github.com/DaiYuANg/arcgo/dbx"
)

func (r *authPrincipalRepo) UpsertAuthPrincipal(ctx context.Context, userID int64, email string) error {
	id := principalIDByUser(userID)
	rows, err := dbx.QueryAll[authPrincipalRow](ctx, r.db,
		dbx.Select(r.aps.AllColumns()...).From(r.aps).Where(r.aps.ID.Eq(id)),
		dbx.MustMapper[authPrincipalRow](r.aps),
	)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		_, err = dbx.Exec(ctx, r.db, dbx.InsertInto(r.aps).Columns(r.aps.ID, r.aps.Email).Values(
			r.aps.ID.Set(id), r.aps.Email.Set(email),
		))
		return err
	}
	_, err = dbx.Exec(ctx, r.db, dbx.Update(r.aps).Set(r.aps.Email.Set(email)).Where(r.aps.ID.Eq(id)))
	return err
}

func (r *authPrincipalRepo) DeleteAuthPrincipal(ctx context.Context, userID int64) error {
	id := principalIDByUser(userID)
	_, _ = dbx.Exec(ctx, r.db, dbx.DeleteFrom(r.apr).Where(r.apr.PrincipalID.Eq(id)))
	_, err := dbx.Exec(ctx, r.db, dbx.DeleteFrom(r.aps).Where(r.aps.ID.Eq(id)))
	return err
}

func (r *authPrincipalRepo) SetAuthPrincipalRoles(ctx context.Context, userID int64, roleIDs []string) error {
	id := principalIDByUser(userID)
	if _, err := dbx.Exec(ctx, r.db, dbx.DeleteFrom(r.apr).Where(r.apr.PrincipalID.Eq(id))); err != nil {
		return err
	}
	for _, roleID := range roleIDs {
		if roleID == "" {
			continue
		}
		if _, err := dbx.Exec(ctx, r.db, dbx.InsertInto(r.apr).Columns(r.apr.PrincipalID, r.apr.Role).Values(
			r.apr.PrincipalID.Set(id), r.apr.Role.Set(roleID),
		)); err != nil {
			return err
		}
	}
	return nil
}
