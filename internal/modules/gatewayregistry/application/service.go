package application

import (
	"context"
	"errors"
	"strings"
	"time"

	config2 "github.com/DaiYuANg/jumpa/internal/config"
	gatewaydomain "github.com/DaiYuANg/jumpa/internal/modules/gatewayregistry/domain"
	"github.com/DaiYuANg/jumpa/internal/modules/gatewayregistry/ports"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

type gatewayService struct {
	cfg  config2.AppConfig
	repo ports.GatewayRepository
}

func NewGatewayService(cfg config2.AppConfig, repo ports.GatewayRepository) GatewayService {
	return &gatewayService{cfg: cfg, repo: repo}
}

func (s *gatewayService) List(ctx context.Context) ([]gatewaydomain.Gateway, error) {
	items, err := s.repo.ListGateways(ctx)
	if err != nil {
		return nil, err
	}
	return lo.Map(items, func(it ports.GatewayRecord, _ int) gatewaydomain.Gateway {
		return toDomainGateway(it, s.cfg)
	}), nil
}

func (s *gatewayService) Get(ctx context.Context, id string) (mo.Option[gatewaydomain.Gateway], error) {
	item, err := s.repo.GetGatewayByID(ctx, id)
	if err != nil {
		return mo.None[gatewaydomain.Gateway](), err
	}
	if item.IsAbsent() {
		return mo.None[gatewaydomain.Gateway](), nil
	}
	return mo.Some(toDomainGateway(item.MustGet(), s.cfg)), nil
}

func (s *gatewayService) RegisterHeartbeat(ctx context.Context, in RegisterHeartbeatInput) (gatewaydomain.Gateway, error) {
	now := time.Now().UTC()
	nodeKey := strings.TrimSpace(in.NodeKey)
	item, err := s.repo.GetGatewayByNodeKey(ctx, nodeKey)
	if err != nil {
		return gatewaydomain.Gateway{}, err
	}

	tagsCSV := strings.Join(lo.FilterMap(in.Tags, func(it string, _ int) (string, bool) {
		v := strings.TrimSpace(it)
		return v, v != ""
	}), ",")

	if item.IsAbsent() {
		created, createErr := s.repo.CreateGateway(ctx, ports.CreateGatewayInput{
			NodeKey:       nodeKey,
			NodeName:      strings.TrimSpace(in.NodeName),
			RuntimeType:   strings.TrimSpace(in.RuntimeType),
			AdvertiseAddr: strings.TrimSpace(in.AdvertiseAddr),
			SSHListenAddr: strings.TrimSpace(in.SSHListenAddr),
			Zone:          strings.TrimSpace(in.Zone),
			TagsCSV:       tagsCSV,
			State:         "online",
			RegisteredAt:  now,
			LastSeenAt:    now,
			UpdatedAt:     now,
		})
		if createErr != nil {
			return gatewaydomain.Gateway{}, createErr
		}
		return toDomainGateway(created, s.cfg), nil
	}

	updated, updateErr := s.repo.UpdateGatewayHeartbeat(ctx, item.MustGet().ID, ports.UpdateGatewayHeartbeatInput{
		NodeName:      strings.TrimSpace(in.NodeName),
		AdvertiseAddr: strings.TrimSpace(in.AdvertiseAddr),
		SSHListenAddr: strings.TrimSpace(in.SSHListenAddr),
		Zone:          strings.TrimSpace(in.Zone),
		TagsCSV:       tagsCSV,
		State:         "online",
		LastSeenAt:    now,
		UpdatedAt:     now,
	})
	if updateErr != nil {
		return gatewaydomain.Gateway{}, updateErr
	}
	if updated.IsAbsent() {
		return gatewaydomain.Gateway{}, errors.New("gateway heartbeat update lost target node")
	}
	return toDomainGateway(updated.MustGet(), s.cfg), nil
}

func (s *gatewayService) MarkOffline(ctx context.Context, nodeKey string) error {
	item, err := s.repo.GetGatewayByNodeKey(ctx, strings.TrimSpace(nodeKey))
	if err != nil {
		return err
	}
	if item.IsAbsent() {
		return nil
	}
	_, err = s.repo.UpdateGatewayState(ctx, item.MustGet().ID, "offline", time.Now().UTC())
	return err
}

func toDomainGateway(it ports.GatewayRecord, cfg config2.AppConfig) gatewaydomain.Gateway {
	status := it.State
	offlineAfter := cfg.Gateway.Registry.OfflineAfterSec
	if offlineAfter <= 0 {
		offlineAfter = 60
	}
	if status != "offline" && time.Since(it.LastSeenAt.UTC()) > time.Duration(offlineAfter)*time.Second {
		status = "stale"
	}
	return gatewaydomain.Gateway{
		ID:              it.ID,
		NodeKey:         it.NodeKey,
		NodeName:        it.NodeName,
		RuntimeType:     it.RuntimeType,
		AdvertiseAddr:   it.AdvertiseAddr,
		SSHListenAddr:   it.SSHListenAddr,
		Zone:            it.Zone,
		Tags:            parseTags(it.TagsCSV),
		State:           it.State,
		EffectiveStatus: status,
		RegisteredAt:    it.RegisteredAt,
		LastSeenAt:      it.LastSeenAt,
		UpdatedAt:       it.UpdatedAt,
	}
}

func parseTags(raw string) []string {
	return lo.FilterMap(strings.Split(strings.TrimSpace(raw), ","), func(it string, _ int) (string, bool) {
		v := strings.TrimSpace(it)
		return v, v != ""
	})
}
