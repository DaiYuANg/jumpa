package application

import (
	"strings"

	bastiondomain "github.com/DaiYuANg/jumpa/internal/modules/bastion/domain"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/ports"
)

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
		return "passthrough"
	}
	return strings.ToLower(value)
}

func coalesceRecordingPolicy(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "always"
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
