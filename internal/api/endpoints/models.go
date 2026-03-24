package endpoints

import "time"

type pageResponse[T any] struct {
	Items    []T `json:"items"`
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}
type dynamicOutput struct{ Body any `json:"body"` }
type Result[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    T      `json:"data"`
}
type PageRequest struct {
	Page     int `query:"page" validate:"omitempty,min=1"`
	PageSize int `query:"pageSize" validate:"omitempty,min=1,max=200"`
}
type PageResult[T any] struct {
	Items    []T `json:"items"`
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}
type meResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Email       string   `json:"email,omitempty"`
	Roles       []idName `json:"roles"`
	Permissions []string `json:"permissions"`
}
type idName struct{ ID, Name string }
type userDTO struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Age       int       `json:"age,omitempty"`
	RoleIDs   []string  `json:"roleIds,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
type roleDTO struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	Description        string    `json:"description,omitempty"`
	PermissionGroupIDs []string  `json:"permissionGroupIds,omitempty"`
	CreatedAt          time.Time `json:"createdAt"`
}
type permissionDTO struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	GroupID   *string   `json:"groupId,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}
type permissionGroupDTO struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

type ListResourceInput struct {
	PageRequest
	Q         string `query:"q" validate:"omitempty,max=200"`
	ID        string `query:"id" validate:"omitempty,max=1000"`
	NameLike  string `query:"name_like" validate:"omitempty,max=200"`
	EmailLike string `query:"email_like" validate:"omitempty,max=200"`
	Sort      string `query:"sort" validate:"omitempty,max=200"`
	Order     string `query:"order" validate:"omitempty,max=16"`
}
type ByIDInput struct{ ID string `path:"id" validate:"required"` }
type DeleteManyInput struct{ ID string `query:"id" validate:"required"` }
type HealthOutput struct {
	Body struct{ Status string `json:"status"` } `json:"body"`
}
type LoginInput struct {
	Body struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=1,max=128"`
	} `json:"body"`
}
type RefreshInput struct{ Body struct{ RefreshToken string `json:"refreshToken"` } `json:"body"` }
type LogoutInput struct{ Body struct{ RefreshToken string `json:"refreshToken"` } `json:"body"` }
type CreateUserInput struct {
	Body struct {
		Name    string   `json:"name" validate:"required,min=2,max=64"`
		Email   string   `json:"email" validate:"required,email"`
		Age     int      `json:"age" validate:"gte=0,lte=130"`
		RoleIDs []string `json:"roleIds"`
	} `json:"body"`
}
type PatchUserInput struct {
	ID   string `path:"id" validate:"required"`
	Body struct {
		Name    *string  `json:"name,omitempty" validate:"omitempty,min=2,max=64"`
		Email   *string  `json:"email,omitempty" validate:"omitempty,email"`
		Age     *int     `json:"age,omitempty" validate:"omitempty,gte=0,lte=130"`
		RoleIDs []string `json:"roleIds,omitempty"`
	} `json:"body"`
}
type BulkPatchInput struct {
	ID   string `query:"id" validate:"required"`
	Body struct{ GroupID *string `json:"groupId"` } `json:"body"`
}
type CreateRoleInput struct {
	Body struct {
		Name               string   `json:"name" validate:"required,min=2,max=128"`
		Description        string   `json:"description"`
		PermissionGroupIDs []string `json:"permissionGroupIds"`
	} `json:"body"`
}
type PatchRoleInput struct {
	ID   string `path:"id" validate:"required"`
	Body struct {
		Name               *string  `json:"name,omitempty" validate:"omitempty,min=2,max=128"`
		Description        *string  `json:"description,omitempty"`
		PermissionGroupIDs []string `json:"permissionGroupIds,omitempty"`
	} `json:"body"`
}
type CreatePermissionInput struct {
	Body struct {
		Name    string  `json:"name" validate:"required,min=2,max=128"`
		Code    string  `json:"code" validate:"required,min=2,max=128"`
		GroupID *string `json:"groupId"`
	} `json:"body"`
}
type PatchPermissionInput struct {
	ID   string `path:"id" validate:"required"`
	Body struct {
		Name    *string `json:"name,omitempty" validate:"omitempty,min=2,max=128"`
		Code    *string `json:"code,omitempty" validate:"omitempty,min=2,max=128"`
		GroupID *string `json:"groupId"`
	} `json:"body"`
}
type CreatePermissionGroupInput struct {
	Body struct {
		Name        string `json:"name" validate:"required,min=2,max=128"`
		Description string `json:"description"`
	} `json:"body"`
}
type PatchPermissionGroupInput struct {
	ID   string `path:"id" validate:"required"`
	Body struct {
		Name        *string `json:"name,omitempty" validate:"omitempty,min=2,max=128"`
		Description *string `json:"description,omitempty"`
	} `json:"body"`
}
