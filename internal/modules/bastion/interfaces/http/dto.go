package http

import (
	bastiondomain "github.com/DaiYuANg/jumpa/internal/modules/bastion/domain"
	"github.com/samber/lo"
)

func toOverviewDTO(it bastiondomain.Overview) overviewDTO {
	return overviewDTO{
		ProductName:        it.ProductName,
		DatabaseDriver:     it.DatabaseDriver,
		CacheEnabled:       it.CacheEnabled,
		BastionEnabled:     it.BastionEnabled,
		SSHListenAddr:      it.SSHListenAddr,
		RecordingDir:       it.RecordingDir,
		IdentityProvider:   it.IdentityProvider,
		IdentityModes:      it.IdentityModes,
		PasswordAuthReady:  it.PasswordAuthReady,
		SupportedDrivers:   it.SupportedDrivers,
		SupportedProtocols: it.SupportedProtocols,
		CapabilityNotes:    it.CapabilityNotes,
		GeneratedAt:        it.GeneratedAt,
	}
}

func toHostDTOs(items []bastiondomain.Host) []hostDTO {
	return lo.Map(items, func(it bastiondomain.Host, _ int) hostDTO {
		return hostDTO{
			ID:              it.ID,
			Name:            it.Name,
			Address:         it.Address,
			Port:            it.Port,
			Protocol:        it.Protocol,
			Environment:     it.Environment,
			Platform:        it.Platform,
			Authentication:  it.Authentication,
			JumpEnabled:     it.JumpEnabled,
			RecordingPolicy: it.RecordingPolicy,
			CreatedAt:       it.CreatedAt,
		}
	})
}

func toPolicyDTOs(items []bastiondomain.AccessPolicy) []policyDTO {
	return lo.Map(items, func(it bastiondomain.AccessPolicy, _ int) policyDTO {
		return policyDTO{
			ID:                it.ID,
			Name:              it.Name,
			SubjectType:       it.SubjectType,
			SubjectName:       it.SubjectName,
			TargetType:        it.TargetType,
			TargetName:        it.TargetName,
			AccountPattern:    it.AccountPattern,
			Protocol:          it.Protocol,
			ApprovalRequired:  it.ApprovalRequired,
			RecordingRequired: it.RecordingRequired,
			CreatedAt:         it.CreatedAt,
		}
	})
}

func toHostAccountDTOs(items []bastiondomain.HostAccount) []hostAccountDTO {
	return lo.Map(items, func(it bastiondomain.HostAccount, _ int) hostAccountDTO {
		return hostAccountDTO{
			ID:                 it.ID,
			HostID:             it.HostID,
			AccountName:        it.AccountName,
			AuthenticationType: it.AuthenticationType,
			CredentialRef:      it.CredentialRef,
			CreatedAt:          it.CreatedAt,
		}
	})
}

func toSessionDTOs(items []bastiondomain.Session) []sessionDTO {
	return lo.Map(items, func(it bastiondomain.Session, _ int) sessionDTO {
		return sessionDTO{
			ID:            it.ID,
			HostName:      it.HostName,
			HostAccount:   it.HostAccount,
			PrincipalName: it.PrincipalName,
			Protocol:      it.Protocol,
			Status:        it.Status,
			StartedAt:     it.StartedAt,
			EndedAt:       it.EndedAt,
		}
	})
}

func toAccessRequestDTOs(items []bastiondomain.AccessRequest) []accessRequestDTO {
	return lo.Map(items, func(it bastiondomain.AccessRequest, _ int) accessRequestDTO {
		return accessRequestDTO{
			ID:                it.ID,
			PolicyID:          it.PolicyID,
			PrincipalName:     it.PrincipalName,
			PrincipalEmail:    it.PrincipalEmail,
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
	})
}
