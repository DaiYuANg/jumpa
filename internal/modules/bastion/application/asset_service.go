package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	bastiondomain "github.com/DaiYuANg/jumpa/internal/modules/bastion/domain"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/ports"
	"github.com/samber/mo"
)

type assetService struct {
	hostRepo        ports.HostRepository
	hostAccountRepo ports.HostAccountRepository
}

type targetService struct {
	hostRepo        ports.HostRepository
	hostAccountRepo ports.HostAccountRepository
}

func NewAssetService(hostRepo ports.HostRepository, hostAccountRepo ports.HostAccountRepository) AssetService {
	return &assetService{hostRepo: hostRepo, hostAccountRepo: hostAccountRepo}
}

func NewTargetService(hostRepo ports.HostRepository, hostAccountRepo ports.HostAccountRepository) TargetService {
	return &targetService{hostRepo: hostRepo, hostAccountRepo: hostAccountRepo}
}

func (s *assetService) ListHosts(ctx context.Context) ([]bastiondomain.Host, error) {
	items, err := s.hostRepo.ListHosts(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]bastiondomain.Host, len(items))
	for i, it := range items {
		out[i] = toDomainHost(it)
	}
	return out, nil
}

func (s *assetService) GetHost(ctx context.Context, id string) (mo.Option[bastiondomain.Host], error) {
	item, err := s.hostRepo.GetHostByID(ctx, id)
	if err != nil || item.IsAbsent() {
		return mo.None[bastiondomain.Host](), err
	}
	return mo.Some(toDomainHost(item.MustGet())), nil
}

func (s *assetService) CreateHost(ctx context.Context, in CreateHostInput) (bastiondomain.Host, error) {
	item, err := s.hostRepo.CreateHost(ctx, ports.CreateHostRecordInput{
		Name:               strings.TrimSpace(in.Name),
		Address:            strings.TrimSpace(in.Address),
		Port:               coalescePort(in.Port),
		Protocol:           coalesceProtocol(in.Protocol),
		Environment:        normalizeOptionalString(in.Environment),
		Platform:           normalizeOptionalString(in.Platform),
		AuthenticationType: coalesceAuthentication(in.Authentication),
		CredentialRef:      normalizeOptionalString(in.CredentialRef),
		JumpEnabled:        in.JumpEnabled,
		RecordingPolicy:    coalesceRecordingPolicy(in.RecordingPolicy),
		CreatedAt:          time.Now().UTC(),
	})
	if err != nil {
		return bastiondomain.Host{}, err
	}
	return toDomainHost(item), nil
}

func (s *assetService) UpdateHost(ctx context.Context, id string, in UpdateHostInput) (mo.Option[bastiondomain.Host], error) {
	item, err := s.hostRepo.UpdateHost(ctx, id, ports.PatchHostRecordInput{
		Name:               normalizeOptionalString(in.Name),
		Address:            normalizeOptionalString(in.Address),
		Port:               in.Port,
		Protocol:           normalizeOptionalString(in.Protocol),
		Environment:        normalizeOptionalString(in.Environment),
		Platform:           normalizeOptionalString(in.Platform),
		AuthenticationType: normalizeOptionalString(in.Authentication),
		CredentialRef:      normalizeOptionalString(in.CredentialRef),
		JumpEnabled:        in.JumpEnabled,
		RecordingPolicy:    normalizeOptionalString(in.RecordingPolicy),
	})
	if err != nil || item.IsAbsent() {
		return mo.None[bastiondomain.Host](), err
	}
	return mo.Some(toDomainHost(item.MustGet())), nil
}

func (s *assetService) DeleteHost(ctx context.Context, id string) (bool, error) {
	return s.hostRepo.DeleteHost(ctx, id)
}

func (s *assetService) ListHostAccounts(ctx context.Context, hostID string) ([]bastiondomain.HostAccount, error) {
	items, err := s.hostAccountRepo.ListHostAccountsByHostID(ctx, hostID)
	if err != nil {
		return nil, err
	}
	out := make([]bastiondomain.HostAccount, len(items))
	for i, it := range items {
		out[i] = toDomainHostAccount(it)
	}
	return out, nil
}

func (s *assetService) GetHostAccount(ctx context.Context, hostID, accountID string) (mo.Option[bastiondomain.HostAccount], error) {
	item, err := s.hostAccountRepo.GetHostAccountByID(ctx, hostID, accountID)
	if err != nil || item.IsAbsent() {
		return mo.None[bastiondomain.HostAccount](), err
	}
	return mo.Some(toDomainHostAccount(item.MustGet())), nil
}

func (s *assetService) CreateHostAccount(ctx context.Context, hostID string, in CreateHostAccountInput) (bastiondomain.HostAccount, error) {
	host, err := s.hostRepo.GetHostByID(ctx, hostID)
	if err != nil {
		return bastiondomain.HostAccount{}, err
	}
	if host.IsAbsent() {
		return bastiondomain.HostAccount{}, fmt.Errorf("host %q not found", hostID)
	}

	item, err := s.hostAccountRepo.CreateHostAccount(ctx, ports.CreateHostAccountRecordInput{
		HostID:             hostID,
		AccountName:        strings.TrimSpace(in.AccountName),
		AuthenticationType: coalesceAuthentication(in.AuthenticationType),
		CredentialRef:      normalizeOptionalString(in.CredentialRef),
		CreatedAt:          time.Now().UTC(),
	})
	if err != nil {
		return bastiondomain.HostAccount{}, err
	}
	return toDomainHostAccount(item), nil
}

func (s *assetService) UpdateHostAccount(ctx context.Context, hostID, accountID string, in UpdateHostAccountInput) (mo.Option[bastiondomain.HostAccount], error) {
	item, err := s.hostAccountRepo.UpdateHostAccount(ctx, hostID, accountID, ports.PatchHostAccountRecordInput{
		AccountName:        normalizeOptionalString(in.AccountName),
		AuthenticationType: normalizeOptionalString(in.AuthenticationType),
		CredentialRef:      normalizeOptionalString(in.CredentialRef),
	})
	if err != nil || item.IsAbsent() {
		return mo.None[bastiondomain.HostAccount](), err
	}
	return mo.Some(toDomainHostAccount(item.MustGet())), nil
}

func (s *assetService) DeleteHostAccount(ctx context.Context, hostID, accountID string) (bool, error) {
	return s.hostAccountRepo.DeleteHostAccount(ctx, hostID, accountID)
}

func (s *targetService) GetHostByName(ctx context.Context, name string) (mo.Option[bastiondomain.Host], error) {
	item, err := s.hostRepo.GetHostByName(ctx, name)
	if err != nil || item.IsAbsent() {
		return mo.None[bastiondomain.Host](), err
	}
	return mo.Some(toDomainHost(item.MustGet())), nil
}

func (s *targetService) GetHostAccountByName(ctx context.Context, hostID, accountName string) (mo.Option[bastiondomain.HostAccount], error) {
	item, err := s.hostAccountRepo.GetHostAccountByName(ctx, hostID, accountName)
	if err != nil || item.IsAbsent() {
		return mo.None[bastiondomain.HostAccount](), err
	}
	return mo.Some(toDomainHostAccount(item.MustGet())), nil
}
