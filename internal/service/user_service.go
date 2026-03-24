package service

import (
	"context"
	"log/slog"

	"github.com/DaiYuANg/arcgo-rbac-template/internal/domain"
	"github.com/DaiYuANg/arcgo-rbac-template/internal/event"
	"github.com/DaiYuANg/arcgo-rbac-template/internal/repo"
	"github.com/DaiYuANg/arcgo/eventx"
)

type UserService interface {
	List(ctx context.Context, search string, limit, offset int) ([]domain.User, int, error)
	Get(ctx context.Context, id int64) (domain.User, bool, error)
	Create(ctx context.Context, in domain.CreateUserInput) (domain.User, error)
	Update(ctx context.Context, id int64, in domain.UpdateUserInput) (domain.User, bool, error)
	Delete(ctx context.Context, id int64) (bool, error)
}

type userService struct {
	repo repo.UserRepository
	bus  eventx.BusRuntime
	log  *slog.Logger
}

func NewUserService(repo repo.UserRepository, bus eventx.BusRuntime, log *slog.Logger) UserService {
	return &userService{repo: repo, bus: bus, log: log}
}

func (s *userService) List(ctx context.Context, search string, limit, offset int) ([]domain.User, int, error) {
	return s.repo.List(ctx, search, limit, offset)
}

func (s *userService) Get(ctx context.Context, id int64) (domain.User, bool, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *userService) Create(ctx context.Context, in domain.CreateUserInput) (domain.User, error) {
	user, err := s.repo.Create(ctx, in)
	if err != nil {
		return domain.User{}, err
	}
	_ = s.bus.PublishAsync(ctx, event.UserCreatedEvent{
		UserID: user.ID, UserName: user.Name, Email: user.Email, CreatedAt: user.CreatedAt,
	})
	return user, nil
}

func (s *userService) Update(ctx context.Context, id int64, in domain.UpdateUserInput) (domain.User, bool, error) {
	return s.repo.Update(ctx, id, in)
}

func (s *userService) Delete(ctx context.Context, id int64) (bool, error) {
	return s.repo.Delete(ctx, id)
}
