package dbx

import (
	"context"
	"database/sql"

	"github.com/DaiYuANg/arcgo-rbac-template/internal/modules/iam/infrastructure/persistence"
	"github.com/DaiYuANg/arcgo/dbx"
)

type unitOfWork struct {
	db *dbx.DB
}

type uowTx struct {
	tx   *dbx.Tx
	role persistence.RoleRepository
	rpg  persistence.RolePermissionGroupRepository
}

func NewUnitOfWork(db *dbx.DB) persistence.UnitOfWork {
	return &unitOfWork{db: db}
}

func (u *unitOfWork) InTx(ctx context.Context, opts *sql.TxOptions, fn func(ctx context.Context, tx persistence.UnitOfWorkTx) error) error {
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

func (t *uowTx) Roles() persistence.RoleRepository { return t.role }
func (t *uowTx) RolePermissionGroups() persistence.RolePermissionGroupRepository { return t.rpg }

