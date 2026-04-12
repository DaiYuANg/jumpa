package application

import (
	"context"
	"time"

	bastiondomain "github.com/DaiYuANg/jumpa/internal/modules/bastion/domain"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/ports"
)

type sessionService struct {
	sessionRepo ports.SessionRepository
}

func NewSessionService(sessionRepo ports.SessionRepository) SessionService {
	return &sessionService{sessionRepo: sessionRepo}
}

func NewSessionRuntimeService(sessionRepo ports.SessionRepository) SessionRuntimeService {
	return &sessionService{sessionRepo: sessionRepo}
}

func (s *sessionService) ListSessions(ctx context.Context) ([]bastiondomain.Session, error) {
	items, err := s.sessionRepo.ListSessions(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]bastiondomain.Session, len(items))
	for i, it := range items {
		out[i] = toDomainSession(it)
	}
	return out, nil
}

func (s *sessionService) Start(ctx context.Context, in StartSessionInput) (bastiondomain.Session, error) {
	now := time.Now().UTC()
	sessionID, err := s.sessionRepo.CreateSession(ctx, ports.CreateSessionInput{
		HostID:        in.HostID,
		HostAccountID: in.HostAccountID,
		PrincipalID:   in.PrincipalName,
		Protocol:      in.Protocol,
		Status:        "opening",
		SourceAddr:    emptyStringToNil(in.SourceAddr),
		StartedAt:     now,
	})
	if err != nil {
		return bastiondomain.Session{}, err
	}

	return bastiondomain.Session{
		ID:            sessionID,
		HostName:      in.HostName,
		HostAccount:   in.HostAccount,
		PrincipalName: in.PrincipalName,
		Protocol:      in.Protocol,
		Status:        "opening",
		StartedAt:     now,
	}, nil
}

func (s *sessionService) MarkActive(ctx context.Context, sessionID string) error {
	return s.sessionRepo.UpdateSessionStatus(ctx, sessionID, "active", nil)
}

func (s *sessionService) Finish(ctx context.Context, sessionID, status string) error {
	now := time.Now().UTC()
	return s.sessionRepo.UpdateSessionStatus(ctx, sessionID, status, &now)
}
