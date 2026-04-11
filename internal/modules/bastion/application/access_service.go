package application

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	config2 "github.com/DaiYuANg/jumpa/internal/config"
	bastiondomain "github.com/DaiYuANg/jumpa/internal/modules/bastion/domain"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/ports"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

type policyService struct {
	policyRepo ports.PolicyRepository
}

type accessService struct {
	policyRepo        ports.PolicyRepository
	principalRepo     ports.PrincipalAccessRepository
	accessRequestRepo ports.AccessRequestRepository
}

type accessRequestService struct {
	cfg  config2.AppConfig
	repo ports.AccessRequestRepository
}

func NewPolicyService(policyRepo ports.PolicyRepository) PolicyService {
	return &policyService{policyRepo: policyRepo}
}

func NewAccessService(policyRepo ports.PolicyRepository, principalRepo ports.PrincipalAccessRepository, accessRequestRepo ports.AccessRequestRepository) AccessService {
	return &accessService{policyRepo: policyRepo, principalRepo: principalRepo, accessRequestRepo: accessRequestRepo}
}

func NewAccessRequestService(cfg config2.AppConfig, repo ports.AccessRequestRepository) AccessRequestService {
	return &accessRequestService{cfg: cfg, repo: repo}
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
		return AccessDecision{Allowed: false, Reason: "no matching access policy"}, nil
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

func (s *accessRequestService) ListRequests(ctx context.Context, in ListAccessRequestsInput) ([]bastiondomain.AccessRequest, int, error) {
	items, total, err := s.repo.ListRequests(ctx, ports.ListAccessRequestsInput{
		Status: in.Status,
		Limit:  in.Limit,
		Offset: in.Offset,
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]bastiondomain.AccessRequest, len(items))
	for i, it := range items {
		out[i] = toDomainAccessRequest(it)
	}
	return out, total, nil
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
