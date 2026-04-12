package application

import "context"

type RecordSessionEventInput struct {
	SessionID string
	EventType string
	Payload   map[string]string
}

type SessionEventService interface {
	Record(ctx context.Context, in RecordSessionEventInput) error
}
