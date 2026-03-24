package domain

import "time"

type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Age       int       `json:"age"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateUserInput struct {
	Name  string `json:"name" validate:"required,min=2,max=64"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"gte=0,lte=130"`
}

type UpdateUserInput struct {
	Name  *string `json:"name,omitempty" validate:"omitempty,min=2,max=64"`
	Email *string `json:"email,omitempty" validate:"omitempty,email"`
	Age   *int    `json:"age,omitempty" validate:"omitempty,gte=0,lte=130"`
}
