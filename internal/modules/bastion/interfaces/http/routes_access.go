package http

import (
	"context"

	"github.com/DaiYuANg/arcgo/httpx"
	apiendpoints "github.com/DaiYuANg/jumpa/internal/api/endpoints"
	"github.com/DaiYuANg/jumpa/internal/modules/bastion/application"
	bastiondomain "github.com/DaiYuANg/jumpa/internal/modules/bastion/domain"
	"github.com/danielgtaylor/huma/v2"
)

func (e *BastionEndpoint) registerAccessRoutes(api *httpx.Group) {
	httpx.MustGroupGet(api, "/access-policies", func(ctx context.Context, _ *struct{}) (*apiendpoints.DynamicOutput, error) {
		items, err := e.policySvc.ListPolicies(ctx)
		if err != nil {
			return nil, err
		}
		return &apiendpoints.DynamicOutput{Body: apiendpoints.OK(toPolicyDTOs(items))}, nil
	}, huma.OperationTags("access"))

	httpx.MustGroupGet(api, "/access-policies/{id}", func(ctx context.Context, input *apiendpoints.ByIDInput) (*apiendpoints.DynamicOutput, error) {
		item, err := e.policySvc.GetPolicy(ctx, input.ID)
		if err != nil {
			return nil, err
		}
		if item.IsAbsent() {
			return nil, httpx.NewError(404, "access policy not found")
		}
		return &apiendpoints.DynamicOutput{Body: apiendpoints.OK(toPolicyDTOs([]bastiondomain.AccessPolicy{item.MustGet()})[0])}, nil
	}, huma.OperationTags("access"))

	httpx.MustGroupPost(api, "/access-policies", func(ctx context.Context, input *createPolicyInput) (*apiendpoints.DynamicOutput, error) {
		item, err := e.policySvc.CreatePolicy(ctx, application.CreatePolicyInput{
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
		item, err := e.policySvc.UpdatePolicy(ctx, input.ID, application.UpdatePolicyInput{
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
		deleted, err := e.policySvc.DeletePolicy(ctx, input.ID)
		if err != nil {
			return nil, err
		}
		if !deleted {
			return nil, httpx.NewError(404, "access policy not found")
		}
		return &apiendpoints.DynamicOutput{Body: apiendpoints.OK(map[string]bool{"deleted": true})}, nil
	}, huma.OperationTags("access"))

	httpx.MustGroupGet(api, "/access-requests", func(ctx context.Context, input *listAccessRequestsInput) (*apiendpoints.DynamicOutput, error) {
		page, pageSize, offset := normalizePageRequest(input.Page, input.PageSize)
		items, total, err := e.requestSvc.ListRequests(ctx, application.ListAccessRequestsInput{
			Status: input.Status,
			Limit:  pageSize,
			Offset: offset,
		})
		if err != nil {
			return nil, err
		}
		result := pageResult[accessRequestDTO]{
			Items:    toAccessRequestDTOs(items),
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		}
		return &apiendpoints.DynamicOutput{Body: apiendpoints.OK(result)}, nil
	}, huma.OperationTags("access"))

	httpx.MustGroupGet(api, "/access-requests/{id}", func(ctx context.Context, input *apiendpoints.ByIDInput) (*apiendpoints.DynamicOutput, error) {
		item, err := e.requestSvc.GetRequest(ctx, input.ID)
		if err != nil {
			return nil, err
		}
		if item.IsAbsent() {
			return nil, httpx.NewError(404, "access request not found")
		}
		return &apiendpoints.DynamicOutput{Body: apiendpoints.OK(toAccessRequestDTOs([]bastiondomain.AccessRequest{item.MustGet()})[0])}, nil
	}, huma.OperationTags("access"))

	httpx.MustGroupPost(api, "/access-requests/{id}/approve", func(ctx context.Context, input *reviewAccessRequestInput) (*apiendpoints.DynamicOutput, error) {
		item, err := e.requestSvc.Approve(ctx, input.ID, input.Body.Reviewer, input.Body.Comment)
		if err != nil {
			return nil, err
		}
		if item.IsAbsent() {
			return nil, httpx.NewError(404, "access request not found")
		}
		return &apiendpoints.DynamicOutput{Body: apiendpoints.OK(toAccessRequestDTOs([]bastiondomain.AccessRequest{item.MustGet()})[0])}, nil
	}, huma.OperationTags("access"))

	httpx.MustGroupPost(api, "/access-requests/{id}/reject", func(ctx context.Context, input *reviewAccessRequestInput) (*apiendpoints.DynamicOutput, error) {
		item, err := e.requestSvc.Reject(ctx, input.ID, input.Body.Reviewer, input.Body.Comment)
		if err != nil {
			return nil, err
		}
		if item.IsAbsent() {
			return nil, httpx.NewError(404, "access request not found")
		}
		return &apiendpoints.DynamicOutput{Body: apiendpoints.OK(toAccessRequestDTOs([]bastiondomain.AccessRequest{item.MustGet()})[0])}, nil
	}, huma.OperationTags("access"))
}
