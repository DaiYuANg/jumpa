package http

import (
	"context"

	"github.com/DaiYuANg/arcgo/httpx"
	apiendpoints "github.com/DaiYuANg/jumpa/internal/api/endpoints"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/application"
	bastiondomain "github.com/DaiYuANg/jumpa/internal/modules/bastion/domain"
	"github.com/danielgtaylor/huma/v2"
)

func registerAssetRoutes(api *httpx.Group, assetSvc application.AssetService) {
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
}
