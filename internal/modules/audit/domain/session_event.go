package domain

import "time"

type SessionEvent struct {
	ID        string
	SessionID string
	EventType string
	Payload   *string
	CreatedAt time.Time
}
