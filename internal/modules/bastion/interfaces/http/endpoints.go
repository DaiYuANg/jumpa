package http

import (
	"context"

	"github.com/DaiYuANg/arcgo/httpx"
	apiendpoints "github.com/DaiYuANg/jumpa/internal/api/endpoints"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/application"
	bastiondomain "github.com/DaiYuANg/jumpa/internal/modules/bastion/domain"
	"github.com/danielgtaylor/huma/v2"
	"github.com/samber/lo"
)

type BastionEndpoint struct {
	httpx.BaseEndpoint
	overviewSvc application.OverviewService
	assetSvc    application.AssetService
	policySvc   application.PolicyService
	requestSvc  application.AccessRequestService
	sessionSvc  application.SessionService
}

func NewBastionEndpoint(
	overviewSvc application.OverviewService,
	assetSvc application.AssetService,
	policySvc application.PolicyService,
	requestSvc application.AccessRequestService,
	sessionSvc application.SessionService,
) *BastionEndpoint {
	return &BastionEndpoint{
		overviewSvc: overviewSvc,
		assetSvc:    assetSvc,
		policySvc:   policySvc,
		requestSvc:  requestSvc,
		sessionSvc:  sessionSvc,
	}
}

func (e *BastionEndpoint) RegisterRoutes(server httpx.ServerRuntime) {
	registerBastionEndpoints(server.Group("/api"), e.overviewSvc, e.assetSvc, e.policySvc, e.requestSvc, e.sessionSvc)
}

func registerBastionEndpoints(
	api *httpx.Group,
	overviewSvc application.OverviewService,
	assetSvc application.AssetService,
	policySvc application.PolicyService,
	requestSvc application.AccessRequestService,
	sessionSvc application.SessionService,
) {
	httpx.MustGroupGet(api, "/bastion/overview", func(ctx context.Context, _ *struct{}) (*apiendpoints.DynamicOutput, error) {
		data, err := overviewSvc.Get(ctx)
		if err != nil {
			return nil, err
		}
		return &apiendpoints.DynamicOutput{Body: apiendpoints.OK(toOverviewDTO(data))}, nil
	}, huma.OperationTags("bastion"))

	httpx.MustGroupGet(api, "/assets/hosts", func(ctx context.Context, _ *struct{}) (*apiendpoints.DynamicOutput, error) {
		items, err := assetSvc.ListHosts(ctx)
		if err != nil {
			return nil, err
		}
		return &apiendpoints.DynamicOutput{Body: apiendpoints.OK(toHostDTOs(items))}, nil
	}, huma.OperationTags("assets"))

	httpx.MustGroupGet(api, "/assets/hosts/{id}", func(ctx context.Context, input *apiendpoints.ByIDInput) (*apiendpoints.DynamicOutput, error) {
		item, err := assetSvc.GetHost(ctx, input.ID)
		if err != nil {
			return nil, err
		}
		if item.IsAbsent() {
			return nil, httpx.NewError(404, "host not found")
		}
		return &apiendpoints.DynamicOutput{Body: apiendpoints.OK(toHostDTOs([]bastiondomain.Host{item.MustGet()})[0])}, nil
	}, huma.OperationTags("assets"))

	httpx.MustGroupPost(api, "/assets/hosts", func(ctx context.Context, input *createHostInput) (*apiendpoints.DynamicOutput, error) {
		item, err := assetSvc.CreateHost(ctx, application.CreateHostInput{
			Name:            input.Body.Name,
			Address:         input.Body.Address,
			Port:            input.Body.Port,
			Protocol:        input.Body.Protocol,
			Environment:     input.Body.Environment,
			Platform:        input.Body.Platform,
			Authentication:  input.Body.Authentication,
			CredentialRef:   input.Body.CredentialRef,
			JumpEnabled:     boolOr(input.Body.JumpEnabled, true),
			RecordingPolicy: input.Body.RecordingPolicy,
		})
		if err != nil {
			return nil, err
		}
		return &apiendpoints.DynamicOutput{Body: apiendpoints.OK(toHostDTOs([]bastiondomain.Host{item})[0])}, nil
	}, huma.OperationTags("assets"))

	httpx.MustGroupPatch(api, "/assets/hosts/{id}", func(ctx context.Context, input *patchHostInput) (*apiendpoints.DynamicOutput, error) {
		item, err := assetSvc.UpdateHost(ctx, input.ID, application.UpdateHostInput{
			Name:            input.Body.Name,
			Address:         input.Body.Address,
			Port:            input.Body.Port,
			Protocol:        input.Body.Protocol,
			Environment:     input.Body.Environment,
			Platform:        input.Body.Platform,
			Authentication:  input.Body.Authentication,
			CredentialRef:   input.Body.CredentialRef,
			JumpEnabled:     input.Body.JumpEnabled,
			RecordingPolicy: input.Body.RecordingPolicy,
		})
		if err != nil {
			return nil, err
		}
		if item.IsAbsent() {
			return nil, httpx.NewError(404, "host not found")
		}
		return &apiendpoints.DynamicOutput{Body: apiendpoints.OK(toHostDTOs([]bastiondomain.Host{item.MustGet()})[0])}, nil
	}, huma.OperationTags("assets"))

	httpx.MustGroupDelete(api, "/assets/hosts/{id}", func(ctx context.Context, input *apiendpoints.ByIDInput) (*apiendpoints.DynamicOutput, error) {
		deleted, err := assetSvc.DeleteHost(ctx, input.ID)
		if err != nil {
			return nil, err
		}
		if !deleted {
			return nil, httpx.NewError(404, "host not found")
		}
		return &apiendpoints.DynamicOutput{Body: apiendpoints.OK(map[string]bool{"deleted": true})}, nil
	}, huma.OperationTags("assets"))

	httpx.MustGroupGet(api, "/assets/hosts/{hostId}/accounts", func(ctx context.Context, input *hostAccountsByHostInput) (*apiendpoints.DynamicOutput, error) {
		items, err := assetSvc.ListHostAccounts(ctx, input.HostID)
		if err != nil {
			return nil, err
		}
		return &apiendpoints.DynamicOutput{Body: apiendpoints.OK(toHostAccountDTOs(items))}, nil
	}, huma.OperationTags("assets"))

	httpx.MustGroupGet(api, "/assets/hosts/{hostId}/accounts/{accountId}", func(ctx context.Context, input *hostAccountByIDInput) (*apiendpoints.DynamicOutput, error) {
		item, err := assetSvc.GetHostAccount(ctx, input.HostID, input.AccountID)
		if err != nil {
			return nil, err
		}
		if item.IsAbsent() {
			return nil, httpx.NewError(404, "host account not found")
		}
		return &apiendpoints.DynamicOutput{Body: apiendpoints.OK(toHostAccountDTOs([]bastiondomain.HostAccount{item.MustGet()})[0])}, nil
	}, huma.OperationTags("assets"))

	httpx.MustGroupPost(api, "/assets/hosts/{hostId}/accounts", func(ctx context.Context, input *createHostAccountInput) (*apiendpoints.DynamicOutput, error) {
		item, err := assetSvc.CreateHostAccount(ctx, input.HostID, application.CreateHostAccountInput{
			AccountName:        input.Body.AccountName,
			AuthenticationType: input.Body.AuthenticationType,
			CredentialRef:      input.Body.CredentialRef,
		})
		if err != nil {
			return nil, err
		}
		return &apiendpoints.DynamicOutput{Body: apiendpoints.OK(toHostAccountDTOs([]bastiondomain.HostAccount{item})[0])}, nil
	}, huma.OperationTags("assets"))

	httpx.MustGroupPatch(api, "/assets/hosts/{hostId}/accounts/{accountId}", func(ctx context.Context, input *patchHostAccountInput) (*apiendpoints.DynamicOutput, error) {
		item, err := assetSvc.UpdateHostAccount(ctx, input.HostID, input.AccountID, application.UpdateHostAccountInput{
			AccountName:        input.Body.AccountName,
			AuthenticationType: input.Body.AuthenticationType,
			CredentialRef:      input.Body.CredentialRef,
		})
		if err != nil {
			return nil, err
		}
		if item.IsAbsent() {
			return nil, httpx.NewError(404, "host account not found")
		}
		return &apiendpoints.DynamicOutput{Body: apiendpoints.OK(toHostAccountDTOs([]bastiondomain.HostAccount{item.MustGet()})[0])}, nil
	}, huma.OperationTags("assets"))

	httpx.MustGroupDelete(api, "/assets/hosts/{hostId}/accounts/{accountId}", func(ctx context.Context, input *hostAccountByIDInput) (*apiendpoints.DynamicOutput, error) {
		deleted, err := assetSvc.DeleteHostAccount(ctx, input.HostID, input.AccountID)
		if err != nil {
			return nil, err
		}
		if !deleted {
			return nil, httpx.NewError(404, "host account not found")
		}
		return &apiendpoints.DynamicOutput{Body: apiendpoints.OK(map[string]bool{"deleted": true})}, nil
	}, huma.OperationTags("assets"))

	httpx.MustGroupGet(api, "/access-policies", func(ctx context.Context, _ *struct{}) (*apiendpoints.DynamicOutput, error) {
		items, err := policySvc.ListPolicies(ctx)
		if err != nil {
			return nil, err
		}
		return &apiendpoints.DynamicOutput{Body: apiendpoints.OK(toPolicyDTOs(items))}, nil
	}, huma.OperationTags("access"))

	httpx.MustGroupGet(api, "/access-policies/{id}", func(ctx context.Context, input *apiendpoints.ByIDInput) (*apiendpoints.DynamicOutput, error) {
		item, err := policySvc.GetPolicy(ctx, input.ID)
		if err != nil {
			return nil, err
		}
		if item.IsAbsent() {
			return nil, httpx.NewError(404, "access policy not found")
		}
		return &apiendpoints.DynamicOutput{Body: apiendpoints.OK(toPolicyDTOs([]bastiondomain.AccessPolicy{item.MustGet()})[0])}, nil
	}, huma.OperationTags("access"))

	httpx.MustGroupPost(api, "/access-policies", func(ctx context.Context, input *createPolicyInput) (*apiendpoints.DynamicOutput, error) {
		item, err := policySvc.CreatePolicy(ctx, application.CreatePolicyInput{
			Name:              input.Body.Name,
			SubjectType:       input.Body.SubjectType,
			SubjectRef:        input.Body.SubjectName,
			TargetType:        input.Body.TargetType,
			TargetRef:         input.Body.TargetName,
			AccountPattern:    input.Body.AccountPattern,
			Protocol:          input.Body.Protocol,
			ApprovalRequired:  boolOr(input.Body.ApprovalRequired, false),
			RecordingRequired: boolOr(input.Body.RecordingRequired, true),
		})
		if err != nil {
			return nil, err
		}
		return &apiendpoints.DynamicOutput{Body: apiendpoints.OK(toPolicyDTOs([]bastiondomain.AccessPolicy{item})[0])}, nil
	}, huma.OperationTags("access"))

	httpx.MustGroupPatch(api, "/access-policies/{id}", func(ctx context.Context, input *patchPolicyInput) (*apiendpoints.DynamicOutput, error) {
		item, err := policySvc.UpdatePolicy(ctx, input.ID, application.UpdatePolicyInput{
			Name:              input.Body.Name,
			SubjectType:       input.Body.SubjectType,
			SubjectRef:        input.Body.SubjectName,
			TargetType:        input.Body.TargetType,
			TargetRef:         input.Body.TargetName,
			AccountPattern:    input.Body.AccountPattern,
			Protocol:          input.Body.Protocol,
			ApprovalRequired:  input.Body.ApprovalRequired,
			RecordingRequired: input.Body.RecordingRequired,
		})
		if err != nil {
			return nil, err
		}
		if item.IsAbsent() {
			return nil, httpx.NewError(404, "access policy not found")
		}
		return &apiendpoints.DynamicOutput{Body: apiendpoints.OK(toPolicyDTOs([]bastiondomain.AccessPolicy{item.MustGet()})[0])}, nil
	}, huma.OperationTags("access"))

	httpx.MustGroupDelete(api, "/access-policies/{id}", func(ctx context.Context, input *apiendpoints.ByIDInput) (*apiendpoints.DynamicOutput, error) {
		deleted, err := policySvc.DeletePolicy(ctx, input.ID)
		if err != nil {
			return nil, err
		}
		if !deleted {
			return nil, httpx.NewError(404, "access policy not found")
		}
		return &apiendpoints.DynamicOutput{Body: apiendpoints.OK(map[string]bool{"deleted": true})}, nil
	}, huma.OperationTags("access"))

	httpx.MustGroupGet(api, "/access-requests", func(ctx context.Context, _ *struct{}) (*apiendpoints.DynamicOutput, error) {
		items, err := requestSvc.ListRequests(ctx)
		if err != nil {
			return nil, err
		}
		return &apiendpoints.DynamicOutput{Body: apiendpoints.OK(toAccessRequestDTOs(items))}, nil
	}, huma.OperationTags("access"))

	httpx.MustGroupGet(api, "/access-requests/{id}", func(ctx context.Context, input *apiendpoints.ByIDInput) (*apiendpoints.DynamicOutput, error) {
		item, err := requestSvc.GetRequest(ctx, input.ID)
		if err != nil {
			return nil, err
		}
		if item.IsAbsent() {
			return nil, httpx.NewError(404, "access request not found")
		}
		return &apiendpoints.DynamicOutput{Body: apiendpoints.OK(toAccessRequestDTOs([]bastiondomain.AccessRequest{item.MustGet()})[0])}, nil
	}, huma.OperationTags("access"))

	httpx.MustGroupPost(api, "/access-requests/{id}/approve", func(ctx context.Context, input *reviewAccessRequestInput) (*apiendpoints.DynamicOutput, error) {
		item, err := requestSvc.Approve(ctx, input.ID, input.Body.Reviewer, input.Body.Comment)
		if err != nil {
			return nil, err
		}
		if item.IsAbsent() {
			return nil, httpx.NewError(404, "access request not found")
		}
		return &apiendpoints.DynamicOutput{Body: apiendpoints.OK(toAccessRequestDTOs([]bastiondomain.AccessRequest{item.MustGet()})[0])}, nil
	}, huma.OperationTags("access"))

	httpx.MustGroupPost(api, "/access-requests/{id}/reject", func(ctx context.Context, input *reviewAccessRequestInput) (*apiendpoints.DynamicOutput, error) {
		item, err := requestSvc.Reject(ctx, input.ID, input.Body.Reviewer, input.Body.Comment)
		if err != nil {
			return nil, err
		}
		if item.IsAbsent() {
			return nil, httpx.NewError(404, "access request not found")
		}
		return &apiendpoints.DynamicOutput{Body: apiendpoints.OK(toAccessRequestDTOs([]bastiondomain.AccessRequest{item.MustGet()})[0])}, nil
	}, huma.OperationTags("access"))

	httpx.MustGroupGet(api, "/sessions", func(ctx context.Context, _ *struct{}) (*apiendpoints.DynamicOutput, error) {
		items, err := sessionSvc.ListSessions(ctx)
		if err != nil {
			return nil, err
		}
		return &apiendpoints.DynamicOutput{Body: apiendpoints.OK(toSessionDTOs(items))}, nil
	}, huma.OperationTags("sessions"))
}

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

func boolOr(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}
