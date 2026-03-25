package dbx

import (
	"context"
	"errors"
	"strings"
	"time"

	legacydomain "github.com/DaiYuANg/arcgo-rbac-template/internal/domain"
	"github.com/DaiYuANg/arcgo-rbac-template/internal/schema"
	"github.com/DaiYuANg/arcgo/dbx"
	"github.com/DaiYuANg/arcgo/dbx/repository"
)

type UserRow = schema.UserRow
type UserSchema = schema.UserSchema

type UserRepository interface {
	List(ctx context.Context, search string, limit, offset int) ([]legacydomain.User, int, error)
	GetByID(ctx context.Context, id int64) (legacydomain.User, bool, error)
	Create(ctx context.Context, in legacydomain.CreateUserInput) (legacydomain.User, error)
	Update(ctx context.Context, id int64, in legacydomain.UpdateUserInput) (legacydomain.User, bool, error)
	Delete(ctx context.Context, id int64) (bool, error)
}

type userRepo struct {
	db     *dbx.DB
	schema UserSchema
	repo   *repository.Base[UserRow, UserSchema]
}

func NewUserRepository(db *dbx.DB, s UserSchema) UserRepository {
	return &userRepo{db: db, schema: s, repo: repository.New[UserRow](db, s)}
}

func (r *userRepo) List(ctx context.Context, search string, limit, offset int) ([]legacydomain.User, int, error) {
	s := r.schema
	specs := []repository.Spec{repository.OrderBy(s.ID.Asc())}
	if search != "" {
		pattern := "%" + strings.TrimSpace(search) + "%"
		specs = append(specs, repository.Where(dbx.Or(dbx.Like(s.Name, pattern), dbx.Like(s.Email, pattern))))
	}
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	total, err := r.repo.CountSpec(ctx, specs...)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.repo.ListSpec(ctx, append(specs, repository.Limit(limit), repository.Offset(offset))...)
	if err != nil {
		return nil, 0, err
	}
	users := make([]legacydomain.User, len(rows))
	for i, row := range rows {
		users[i] = legacydomain.User{ID: row.ID, Name: row.Name, Email: row.Email, Age: row.Age, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
	}
	return users, int(total), nil
}

func (r *userRepo) GetByID(ctx context.Context, id int64) (legacydomain.User, bool, error) {
	s := r.schema
	row, err := r.repo.FirstSpec(ctx, repository.Where(s.ID.Eq(id)))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return legacydomain.User{}, false, nil
		}
		return legacydomain.User{}, false, err
	}
	return legacydomain.User{ID: row.ID, Name: row.Name, Email: row.Email, Age: row.Age, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}, true, nil
}

func (r *userRepo) Create(ctx context.Context, in legacydomain.CreateUserInput) (legacydomain.User, error) {
	now := time.Now().UTC()
	row := UserRow{
		Name:      in.Name,
		Email:     in.Email,
		Age:       in.Age,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := r.repo.Create(ctx, &row); err != nil {
		return legacydomain.User{}, err
	}
	return legacydomain.User{ID: row.ID, Name: row.Name, Email: row.Email, Age: row.Age, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}, nil
}

func (r *userRepo) Update(ctx context.Context, id int64, in legacydomain.UpdateUserInput) (legacydomain.User, bool, error) {
	s := r.schema
	assignments := []dbx.Assignment{s.UpdatedAt.Set(time.Now().UTC())}
	if in.Name != nil {
		assignments = append(assignments, s.Name.Set(*in.Name))
	}
	if in.Email != nil {
		assignments = append(assignments, s.Email.Set(*in.Email))
	}
	if in.Age != nil {
		assignments = append(assignments, s.Age.Set(*in.Age))
	}
	res, err := r.repo.UpdateByID(ctx, id, assignments...)
	if err != nil {
		return legacydomain.User{}, false, err
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return legacydomain.User{}, false, nil
	}
	out, ok, err := r.GetByID(ctx, id)
	return out, ok, err
}

func (r *userRepo) Delete(ctx context.Context, id int64) (bool, error) {
	res, err := r.repo.DeleteByID(ctx, id)
	if err != nil {
		return false, err
	}
	ra, _ := res.RowsAffected()
	return ra > 0, nil
}

