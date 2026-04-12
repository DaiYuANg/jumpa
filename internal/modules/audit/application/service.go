package application

import (
	"context"
	"encoding/json"
	"time"

	"github.com/DaiYuANg/jumpa/internal/modules/audit/ports"
)

type sessionEventService struct {
	repo ports.SessionEventRepository
}

func NewSessionEventService(repo ports.SessionEventRepository) SessionEventService {
	return &sessionEventService{repo: repo}
}

func (s *sessionEventService) Record(ctx context.Context, in RecordSessionEventInput) error {
	var data *string
	if len(in.Payload) > 0 {
		raw, err := json.Marshal(in.Payload)
		if err != nil {
			return err
		}
		value := string(raw)
		data = &value
	}

	return s.repo.CreateSessionEvent(ctx, ports.CreateSessionEventInput{
		SessionID: in.SessionID,
		EventType: in.EventType,
		Payload:   data,
		CreatedAt: time.Now().UTC(),
	})
}
