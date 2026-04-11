package application

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"time"

	config2 "github.com/DaiYuANg/jumpa/internal/config"
	"github.com/DaiYuANg/jumpa/internal/identity"
	bastiondomain "github.com/DaiYuANg/jumpa/internal/modules/bastion/domain"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/ports"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

type overviewService struct {
	cfg           config2.AppConfig
	identity      identity.ProviderDescriptor
	authenticator identity.Authenticator
}

type assetService struct {
	hostRepo        ports.HostRepository
	hostAccountRepo ports.HostAccountRepository
}
type targetService struct {
	hostRepo        ports.HostRepository
	hostAccountRepo ports.HostAccountRepository
}
type policyService struct {
	policyRepo ports.PolicyRepository
}
type accessService struct {
	policyRepo        ports.PolicyRepository
	principalRepo     ports.PrincipalAccessRepository
	accessRequestRepo ports.AccessRequestRepository
}
type sessionService struct {
	sessionRepo ports.SessionRepository
	eventRepo   ports.SessionEventRepository
}
type accessRequestService struct {
	cfg  config2.AppConfig
	repo ports.AccessRequestRepository
}

func NewOverviewService(cfg config2.AppConfig, provider identity.ProviderDescriptor, authenticator identity.Authenticator) OverviewService {
	return &overviewService{cfg: cfg, identity: provider, authenticator: authenticator}
}

func NewAssetService(hostRepo ports.HostRepository, hostAccountRepo ports.HostAccountRepository) AssetService {
	return &assetService{hostRepo: hostRepo, hostAccountRepo: hostAccountRepo}
}

func NewTargetService(hostRepo ports.HostRepository, hostAccountRepo ports.HostAccountRepository) TargetService {
	return &targetService{hostRepo: hostRepo, hostAccountRepo: hostAccountRepo}
}

func NewPolicyService(policyRepo ports.PolicyRepository) PolicyService {
	return &policyService{policyRepo: policyRepo}
}

func NewAccessService(policyRepo ports.PolicyRepository, principalRepo ports.PrincipalAccessRepository, accessRequestRepo ports.AccessRequestRepository) AccessService {
	return &accessService{policyRepo: policyRepo, principalRepo: principalRepo, accessRequestRepo: accessRequestRepo}
}

func NewSessionService(sessionRepo ports.SessionRepository) SessionService {
	return &sessionService{sessionRepo: sessionRepo}
}

func NewSessionRuntimeService(sessionRepo ports.SessionRepository, eventRepo ports.SessionEventRepository) SessionRuntimeService {
	return &sessionService{sessionRepo: sessionRepo, eventRepo: eventRepo}
}

func NewAccessRequestService(cfg config2.AppConfig, repo ports.AccessRequestRepository) AccessRequestService {
	return &accessRequestService{cfg: cfg, repo: repo}
}

func (s *overviewService) Get(_ context.Context) (bastiondomain.Overview, error) {
	return bastiondomain.Overview{
		ProductName:      s.cfg.App.Name,
		DatabaseDriver:   s.cfg.DB.Driver,
		CacheEnabled:     s.cfg.Valkey.Enabled,
		BastionEnabled:   s.cfg.Bastion.Enabled,
		SSHListenAddr:    s.cfg.Bastion.SSH.ListenAddr,
		RecordingDir:     s.cfg.Bastion.Session.RecordingDirectory,
		IdentityProvider: s.identity,
		IdentityModes: []string{
			"local",
			"os",
		},
		PasswordAuthReady: s.authenticator.SupportsPassword(),
		SupportedDrivers:  []string{"sqlite", "mariadb", "postgres"},
		SupportedProtocols: []string{
			"ssh",
			"sftp",
		},
		CapabilityNotes: []string{
			"Current landing includes a dedicated SSH gateway runtime with downstream SSH proxying and persisted session lifecycle records.",
			"Keep OS-backed login as an authentication source while storing bastion authorization, target mapping, and audit state inside the application database.",
			"The gateway listener is now split out as a dedicated runtime in cmd/gateway.",
		},
		GeneratedAt: time.Now().UTC(),
	}, nil
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

func (s *policyService) ListPolicies(ctx context.Context) ([]bastiondomain.AccessPolicy, error) {
	items, err := s.policyRepo.ListPolicies(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]bastiondomain.AccessPolicy, len(items))
	for i, it := range items {
		out[i] = toDomainPolicy(it)
	}
	return out, nil
}

func (s *policyService) GetPolicy(ctx context.Context, id string) (mo.Option[bastiondomain.AccessPolicy], error) {
	item, err := s.policyRepo.GetPolicyByID(ctx, id)
	if err != nil || item.IsAbsent() {
		return mo.None[bastiondomain.AccessPolicy](), err
	}
	return mo.Some(toDomainPolicy(item.MustGet())), nil
}

func (s *policyService) CreatePolicy(ctx context.Context, in CreatePolicyInput) (bastiondomain.AccessPolicy, error) {
	item, err := s.policyRepo.CreatePolicy(ctx, ports.CreateAccessPolicyRecordInput{
		Name:              strings.TrimSpace(in.Name),
		SubjectType:       coalesceSubjectType(in.SubjectType),
		SubjectRef:        coalescePattern(in.SubjectRef),
		TargetType:        coalesceTargetType(in.TargetType),
		TargetRef:         coalescePattern(in.TargetRef),
		AccountPattern:    coalescePattern(in.AccountPattern),
		Protocol:          coalescePolicyProtocol(in.Protocol),
		ApprovalRequired:  in.ApprovalRequired,
		RecordingRequired: in.RecordingRequired,
		CreatedAt:         time.Now().UTC(),
	})
	if err != nil {
		return bastiondomain.AccessPolicy{}, err
	}
	return toDomainPolicy(item), nil
}

func (s *policyService) UpdatePolicy(ctx context.Context, id string, in UpdatePolicyInput) (mo.Option[bastiondomain.AccessPolicy], error) {
	item, err := s.policyRepo.UpdatePolicy(ctx, id, ports.PatchAccessPolicyRecordInput{
		Name:              normalizeOptionalString(in.Name),
		SubjectType:       normalizeOptionalString(in.SubjectType),
		SubjectRef:        normalizeOptionalString(in.SubjectRef),
		TargetType:        normalizeOptionalString(in.TargetType),
		TargetRef:         normalizeOptionalString(in.TargetRef),
		AccountPattern:    normalizeOptionalString(in.AccountPattern),
		Protocol:          normalizeOptionalString(in.Protocol),
		ApprovalRequired:  in.ApprovalRequired,
		RecordingRequired: in.RecordingRequired,
	})
	if err != nil || item.IsAbsent() {
		return mo.None[bastiondomain.AccessPolicy](), err
	}
	return mo.Some(toDomainPolicy(item.MustGet())), nil
}

func (s *policyService) DeletePolicy(ctx context.Context, id string) (bool, error) {
	return s.policyRepo.DeletePolicy(ctx, id)
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

func toDomainHost(it ports.HostRecord) bastiondomain.Host {
	return bastiondomain.Host{
		ID:              it.ID,
		Name:            it.Name,
		Address:         it.Address,
		Port:            it.Port,
		Protocol:        it.Protocol,
		Environment:     valueOrEmpty(it.Environment),
		Platform:        valueOrEmpty(it.Platform),
		Authentication:  it.AuthenticationType,
		JumpEnabled:     it.JumpEnabled,
		RecordingPolicy: it.RecordingPolicy,
		CreatedAt:       it.CreatedAt,
	}
}

func toDomainHostAccount(it ports.HostAccountRecord) bastiondomain.HostAccount {
	return bastiondomain.HostAccount{
		ID:                 it.ID,
		HostID:             it.HostID,
		AccountName:        it.AccountName,
		AuthenticationType: it.AuthenticationType,
		CredentialRef:      it.CredentialRef,
		CreatedAt:          it.CreatedAt,
	}
}

func toDomainPolicy(it ports.AccessPolicyRecord) bastiondomain.AccessPolicy {
	return bastiondomain.AccessPolicy{
		ID:                it.ID,
		Name:              it.Name,
		SubjectType:       it.SubjectType,
		SubjectName:       it.SubjectRef,
		TargetType:        it.TargetType,
		TargetName:        it.TargetRef,
		AccountPattern:    it.AccountPattern,
		Protocol:          it.Protocol,
		ApprovalRequired:  it.ApprovalRequired,
		RecordingRequired: it.RecordingRequired,
		CreatedAt:         it.CreatedAt,
	}
}

func (s *accessService) Authorize(ctx context.Context, in AccessCheckInput) (AccessDecision, error) {
	policies, err := s.policyRepo.ListPolicies(ctx)
	if err != nil {
		return AccessDecision{}, err
	}
	roleIDs := []string{}
	if strings.TrimSpace(in.PrincipalEmail) != "" && s.principalRepo != nil {
		roleIDs, err = s.principalRepo.ListRoleIDsByEmail(ctx, in.PrincipalEmail)
		if err != nil {
			return AccessDecision{}, err
		}
	}

	matched, ok := lo.Find(policies, func(policy ports.AccessPolicyRecord) bool {
		return matchesAccessPolicy(policy, in, roleIDs)
	})
	if !ok {
		return AccessDecision{
			Allowed: false,
			Reason:  "no matching access policy",
		}, nil
	}

	if matched.ApprovalRequired {
		request, requestErr := s.ensureAccessRequest(ctx, matched, in)
		if requestErr != nil {
			return AccessDecision{}, requestErr
		}
		if isAccessRequestApprovedUsable(time.Now().UTC(), request) {
			return AccessDecision{
				Allowed:           true,
				ApprovalRequired:  false,
				RecordingRequired: matched.RecordingRequired,
				MatchedPolicyID:   matched.ID,
				RequestID:         request.ID,
				Reason:            "matched approved access request",
			}, nil
		}
		return AccessDecision{
			Allowed:           false,
			ApprovalRequired:  true,
			RecordingRequired: matched.RecordingRequired,
			MatchedPolicyID:   matched.ID,
			RequestID:         request.ID,
			Reason:            "matching policy requires approval workflow",
		}, nil
	}

	return AccessDecision{
		Allowed:           true,
		ApprovalRequired:  false,
		RecordingRequired: matched.RecordingRequired,
		MatchedPolicyID:   matched.ID,
		Reason:            "matched access policy",
	}, nil
}

func (s *accessRequestService) ListRequests(ctx context.Context) ([]bastiondomain.AccessRequest, error) {
	items, err := s.repo.ListRequests(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]bastiondomain.AccessRequest, len(items))
	for i, it := range items {
		out[i] = toDomainAccessRequest(it)
	}
	return out, nil
}

func (s *accessRequestService) GetRequest(ctx context.Context, id string) (mo.Option[bastiondomain.AccessRequest], error) {
	item, err := s.repo.GetRequestByID(ctx, id)
	if err != nil || item.IsAbsent() {
		return mo.None[bastiondomain.AccessRequest](), err
	}
	return mo.Some(toDomainAccessRequest(item.MustGet())), nil
}

func (s *accessRequestService) Approve(ctx context.Context, id, reviewer string, comment *string) (mo.Option[bastiondomain.AccessRequest], error) {
	approvedUntil := approvalExpiry(time.Now().UTC(), s.cfg)
	item, err := s.repo.UpdateRequestStatus(ctx, id, "approved", time.Now().UTC(), strings.TrimSpace(reviewer), normalizeOptionalString(comment), approvedUntil)
	if err != nil || item.IsAbsent() {
		return mo.None[bastiondomain.AccessRequest](), err
	}
	return mo.Some(toDomainAccessRequest(item.MustGet())), nil
}

func (s *accessRequestService) Reject(ctx context.Context, id, reviewer string, comment *string) (mo.Option[bastiondomain.AccessRequest], error) {
	item, err := s.repo.UpdateRequestStatus(ctx, id, "rejected", time.Now().UTC(), strings.TrimSpace(reviewer), normalizeOptionalString(comment), nil)
	if err != nil || item.IsAbsent() {
		return mo.None[bastiondomain.AccessRequest](), err
	}
	return mo.Some(toDomainAccessRequest(item.MustGet())), nil
}

func (s *accessService) ConsumeApprovedRequest(ctx context.Context, requestID, sessionID string) error {
	if strings.TrimSpace(requestID) == "" {
		return nil
	}
	item, err := s.accessRequestRepo.GetRequestByID(ctx, requestID)
	if err != nil {
		return err
	}
	if item.IsAbsent() {
		return fmt.Errorf("access request %q not found", requestID)
	}
	request := item.MustGet()
	if !isAccessRequestApprovedUsable(time.Now().UTC(), request) {
		return fmt.Errorf("access request %q is not usable", requestID)
	}
	sessionID = strings.TrimSpace(sessionID)
	var consumedSessionID *string
	if sessionID != "" {
		consumedSessionID = &sessionID
	}
	updated, err := s.accessRequestRepo.ConsumeRequest(ctx, requestID, time.Now().UTC(), consumedSessionID)
	if err != nil {
		return err
	}
	if updated.IsAbsent() {
		return fmt.Errorf("access request %q could not be consumed", requestID)
	}
	return nil
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

func (s *sessionService) RecordEvent(ctx context.Context, sessionID, eventType string, payload map[string]string) error {
	if s.eventRepo == nil {
		return nil
	}

	var data *string
	if len(payload) > 0 {
		raw, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		value := string(raw)
		data = &value
	}

	return s.eventRepo.CreateSessionEvent(ctx, ports.CreateSessionEventInput{
		SessionID: sessionID,
		EventType: eventType,
		Payload:   data,
		CreatedAt: time.Now().UTC(),
	})
}

func (s *sessionService) Finish(ctx context.Context, sessionID, status string) error {
	now := time.Now().UTC()
	return s.sessionRepo.UpdateSessionStatus(ctx, sessionID, status, &now)
}

func toDomainSession(it ports.SessionRecord) bastiondomain.Session {
	return bastiondomain.Session{
		ID:            it.ID,
		HostName:      it.HostName,
		HostAccount:   it.HostAccount,
		PrincipalName: it.PrincipalName,
		Protocol:      it.Protocol,
		Status:        it.Status,
		StartedAt:     it.StartedAt,
		EndedAt:       it.EndedAt,
	}
}

func toDomainAccessRequest(it ports.AccessRequestRecord) bastiondomain.AccessRequest {
	return bastiondomain.AccessRequest{
		ID:                it.ID,
		PolicyID:          it.PolicyID,
		PrincipalName:     it.PrincipalName,
		PrincipalEmail:    valueOrEmpty(it.PrincipalEmail),
		HostName:          it.HostName,
		HostAccount:       it.HostAccount,
		Protocol:          it.Protocol,
		Status:            it.Status,
		RequestedAt:       it.RequestedAt,
		ReviewedAt:        it.ReviewedAt,
		ReviewedBy:        it.ReviewedBy,
		ReviewComment:     it.ReviewComment,
		ApprovedUntil:     it.ApprovedUntil,
		ConsumedAt:        it.ConsumedAt,
		ConsumedSessionID: it.ConsumedSessionID,
	}
}

func valueOrEmpty(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func emptyStringToNil(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

func matchesAccessPolicy(policy ports.AccessPolicyRecord, in AccessCheckInput, roleIDs []string) bool {
	return matchesSubject(policy.SubjectType, policy.SubjectRef, in.PrincipalName, in.PrincipalEmail, roleIDs) &&
		matchesTarget(policy.TargetType, policy.TargetRef, in.HostName) &&
		matchesProtocol(policy.Protocol, in.Protocol) &&
		matchesPattern(policy.AccountPattern, in.AccountName)
}

func matchesSubject(subjectType, subjectRef, principalName, principalEmail string, roleIDs []string) bool {
	switch strings.ToLower(strings.TrimSpace(subjectType)) {
	case "", "*", "user", "principal":
		return matchesPattern(subjectRef, principalName)
	case "email":
		return matchesPattern(subjectRef, principalEmail)
	case "role":
		return lo.SomeBy(roleIDs, func(roleID string) bool {
			return matchesPattern(subjectRef, roleID)
		})
	default:
		return false
	}
}

func matchesTarget(targetType, targetRef, hostName string) bool {
	switch strings.ToLower(strings.TrimSpace(targetType)) {
	case "", "*", "host":
		return matchesPattern(targetRef, hostName)
	default:
		return false
	}
}

func matchesProtocol(policyProtocol, actualProtocol string) bool {
	policyProtocol = strings.TrimSpace(policyProtocol)
	if policyProtocol == "" || policyProtocol == "*" {
		return true
	}
	return strings.EqualFold(policyProtocol, actualProtocol)
}

func matchesPattern(pattern, value string) bool {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" || pattern == "*" {
		return true
	}
	ok, err := path.Match(pattern, value)
	if err != nil {
		return strings.EqualFold(pattern, value)
	}
	return ok
}

func (s *accessService) ensureAccessRequest(ctx context.Context, policy ports.AccessPolicyRecord, in AccessCheckInput) (ports.AccessRequestRecord, error) {
	findInput := ports.FindAccessRequestInput{
		PolicyID:      policy.ID,
		PrincipalName: in.PrincipalName,
		HostName:      in.HostName,
		HostAccount:   in.AccountName,
		Protocol:      in.Protocol,
	}
	if strings.TrimSpace(in.PrincipalEmail) != "" {
		findInput.PrincipalEmail = &in.PrincipalEmail
	}

	request, err := s.accessRequestRepo.FindLatestRequest(ctx, findInput)
	if err != nil {
		return ports.AccessRequestRecord{}, err
	}
	if request.IsPresent() {
		current := request.MustGet()
		now := time.Now().UTC()
		if strings.EqualFold(current.Status, "pending") || isAccessRequestApprovedUsable(now, current) {
			return current, nil
		}
	}

	return s.accessRequestRepo.CreateRequest(ctx, ports.CreateAccessRequestInput{
		PolicyID:       policy.ID,
		PrincipalName:  in.PrincipalName,
		PrincipalEmail: findInput.PrincipalEmail,
		HostName:       in.HostName,
		HostAccount:    in.AccountName,
		Protocol:       in.Protocol,
		RequestedAt:    time.Now().UTC(),
	})
}

func approvalExpiry(now time.Time, cfg config2.AppConfig) *time.Time {
	ttlMin := cfg.Bastion.Access.ApprovalTTLMin
	if ttlMin <= 0 {
		return nil
	}
	expiresAt := now.Add(time.Duration(ttlMin) * time.Minute)
	return &expiresAt
}

func isAccessRequestApprovedUsable(now time.Time, request ports.AccessRequestRecord) bool {
	if !strings.EqualFold(request.Status, "approved") {
		return false
	}
	if request.ConsumedAt != nil {
		return false
	}
	if request.ApprovedUntil != nil && now.After(*request.ApprovedUntil) {
		return false
	}
	return true
}

func normalizeOptionalString(v *string) *string {
	if v == nil {
		return nil
	}
	value := strings.TrimSpace(*v)
	return &value
}

func coalescePort(port int) int {
	if port <= 0 {
		return 22
	}
	return port
}

func coalesceProtocol(protocol string) string {
	value := strings.TrimSpace(protocol)
	if value == "" {
		return "ssh"
	}
	return strings.ToLower(value)
}

func coalesceAuthentication(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "managed"
	}
	return strings.ToLower(value)
}

func coalesceRecordingPolicy(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "required"
	}
	return strings.ToLower(value)
}

func coalesceSubjectType(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "user"
	}
	return strings.ToLower(value)
}

func coalesceTargetType(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "host"
	}
	return strings.ToLower(value)
}

func coalescePattern(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "*"
	}
	return value
}

func coalescePolicyProtocol(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "ssh"
	}
	return strings.ToLower(value)
}
