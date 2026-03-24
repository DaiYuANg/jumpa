package domain

import "time"

type User struct {
	ID        int64
	Name      string
	Email     string
	Age       int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CreateUserInput struct {
	Name  string
	Email string
	Age   int
}

type UpdateUserInput struct {
	Name  *string
	Email *string
	Age   *int
}

type Role struct {
	ID                 string
	Name               string
	Description        string
	PermissionGroupIDs []string
	CreatedAt          time.Time
}

type PermissionGroup struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
}

type Permission struct {
	ID        string
	Name      string
	Code      string
	GroupID   *string
	CreatedAt time.Time
}
