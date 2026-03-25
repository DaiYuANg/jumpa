package dbx

import (
	"context"
	"database/sql"

	"github.com/DaiYuANg/arcgo-rbac-template/internal/modules/iam/ports"
	"github.com/DaiYuANg/arcgo/dbx"
)

type unitOfWork struct {
	db *dbx.DB
}

type uowTx struct {
	tx   *dbx.Tx
	role ports.RoleRepository
	rpg  ports.RolePermissionGroupRepository
}

func NewUnitOfWork(db *dbx.DB) ports.UnitOfWork {
	return &unitOfWork{db: db}
}

func (u *unitOfWork) InTx(ctx context.Context, opts *sql.TxOptions, fn func(ctx context.Context, tx ports.UnitOfWorkTx) error) error {
	if u == nil || u.db == nil {
		return dbx.ErrNilDB
	}
	if fn == nil {
		return nil
	}
	tx, err := u.db.BeginTx(ctx, opts)
	if err != nil {
		return err
	}
	committed := false
	defer func() { if !committed { _ = tx.Rollback() } }()

	scope := &uowTx{
		tx:   tx,
		role: NewRoleRepository(tx),
		rpg:  NewRolePermissionGroupRepository(tx),
	}
	if runErr := fn(ctx, scope); runErr != nil {
		return runErr
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}

func (t *uowTx) Roles() ports.RoleRepository { return t.role }
func (t *uowTx) RolePermissionGroups() ports.RolePermissionGroupRepository { return t.rpg }

