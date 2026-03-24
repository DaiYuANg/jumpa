package event

import "time"

type UserCreatedEvent struct {
	UserID    int64
	UserName  string
	Email     string
	CreatedAt time.Time
}

func (UserCreatedEvent) Name() string { return "user.created" }
