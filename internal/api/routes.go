package api

import (
	"context"
	"net/http"
	"time"

	"github.com/DaiYuANg/arcgo-rbac-template/internal/domain"
	"github.com/DaiYuANg/arcgo/httpx"
	"github.com/danielgtaylor/huma/v2"
)

type ListUsersInput struct {
	Limit int    `query:"limit" validate:"omitempty,min=1,max=100"`
	Page  int    `query:"page" validate:"omitempty,min=1"`
	Q     string `query:"q" validate:"omitempty,max=100"`
}

type ListUsersOutput struct {
	Body struct {
		Items []domain.User `json:"items"`
		Total int           `json:"total"`
		Page  int           `json:"page"`
		Limit int           `json:"limit"`
	} `json:"body"`
}

type GetUserInput struct {
	ID int64 `path:"id" validate:"required,min=1"`
}

type GetUserOutput struct {
	Body domain.User `json:"body"`
}

type CreateUserInput struct {
	Body domain.CreateUserInput `json:"body"`
}

type CreateUserOutput struct {
	Body domain.User `json:"body"`
}

type UpdateUserInput struct {
	ID   int64                  `path:"id"`
	Body domain.UpdateUserInput `json:"body"`
}

type UpdateUserOutput struct {
	Body domain.User `json:"body"`
}

type DeleteUserInput struct {
	ID int64 `path:"id"`
}

type DeleteUserOutput struct {
	Body struct {
		Deleted bool `json:"deleted"`
	} `json:"body"`
}

type HealthOutput struct {
	Body struct {
		Status string `json:"status"`
		Time   string `json:"time"`
	} `json:"body"`
}

type UserService interface {
	List(ctx context.Context, search string, limit, offset int) ([]domain.User, int, error)
	Get(ctx context.Context, id int64) (domain.User, bool, error)
	Create(ctx context.Context, in domain.CreateUserInput) (domain.User, error)
	Update(ctx context.Context, id int64, in domain.UpdateUserInput) (domain.User, bool, error)
	Delete(ctx context.Context, id int64) (bool, error)
}

func RegisterRoutes(server httpx.ServerRuntime, svc UserService) {
	httpx.MustGet(server, "/health", func(ctx context.Context, _ *struct{}) (*HealthOutput, error) {
		out := &HealthOutput{}
		out.Body.Status = "ok"
		out.Body.Time = time.Now().UTC().Format(time.RFC3339)
		return out, nil
	}, huma.OperationTags("system"))

	api := server.Group("/api/v1")

	httpx.MustGroupGet(api, "/users", func(ctx context.Context, input *ListUsersInput) (*ListUsersOutput, error) {
		limit, page := input.Limit, input.Page
		if limit <= 0 {
			limit = 10
		}
		if limit > 100 {
			limit = 100
		}
		if page <= 0 {
			page = 1
		}
		offset := (page - 1) * limit
		items, total, err := svc.List(ctx, input.Q, limit, offset)
		if err != nil {
			return nil, err
		}
		out := &ListUsersOutput{}
		out.Body.Items = items
		out.Body.Total = total
		out.Body.Page = page
		out.Body.Limit = limit
		return out, nil
	}, huma.OperationTags("users"))

	httpx.MustGroupGet(api, "/users/{id}", func(ctx context.Context, input *GetUserInput) (*GetUserOutput, error) {
		user, ok, err := svc.Get(ctx, input.ID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, httpx.NewError(http.StatusNotFound, "user not found")
		}
		return &GetUserOutput{Body: user}, nil
	}, huma.OperationTags("users"))

	httpx.MustGroupPost(api, "/users", func(ctx context.Context, input *CreateUserInput) (*CreateUserOutput, error) {
		user, err := svc.Create(ctx, input.Body)
		if err != nil {
			return nil, err
		}
		return &CreateUserOutput{Body: user}, nil
	}, huma.OperationTags("users"))

	httpx.MustGroupPut(api, "/users/{id}", func(ctx context.Context, input *UpdateUserInput) (*UpdateUserOutput, error) {
		user, ok, err := svc.Update(ctx, input.ID, input.Body)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, httpx.NewError(http.StatusNotFound, "user not found")
		}
		return &UpdateUserOutput{Body: user}, nil
	}, huma.OperationTags("users"))

	httpx.MustGroupDelete(api, "/users/{id}", func(ctx context.Context, input *DeleteUserInput) (*DeleteUserOutput, error) {
		deleted, err := svc.Delete(ctx, input.ID)
		if err != nil {
			return nil, err
		}
		if !deleted {
			return nil, httpx.NewError(http.StatusNotFound, "user not found")
		}
		out := &DeleteUserOutput{}
		out.Body.Deleted = true
		return out, nil
	}, huma.OperationTags("users"))
}
