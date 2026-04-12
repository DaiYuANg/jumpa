package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	clienthttp "github.com/DaiYuANg/arcgo/clientx/http"
	"github.com/samber/mo"
)

type Client struct {
	baseURL      string
	httpClient   clienthttp.Client
	accessToken  string
	refreshToken string
}

func NewClient(baseURL string, httpClient clienthttp.Client) *Client {
	return &Client{
		baseURL:    strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		httpClient: httpClient,
	}
}

func (c *Client) BaseURL() string {
	return c.baseURL
}

func (c *Client) Login(ctx context.Context, email, password string) (LoginResponse, error) {
	payload := map[string]any{
		"email":    strings.TrimSpace(email),
		"password": password,
	}
	var out Result[LoginResponse]
	if err := c.doJSON(ctx, "POST", "/api/auth/login", payload, false, &out); err != nil {
		return LoginResponse{}, err
	}
	c.accessToken = out.Data.AccessToken
	c.refreshToken = out.Data.RefreshToken
	return out.Data, nil
}

func (c *Client) Overview(ctx context.Context) (Overview, error) {
	var out Result[Overview]
	if err := c.doJSON(ctx, "GET", "/api/bastion/overview", nil, true, &out); err != nil {
		return Overview{}, err
	}
	return out.Data, nil
}

func (c *Client) Hosts(ctx context.Context) ([]Host, error) {
	var out Result[[]Host]
	if err := c.doJSON(ctx, "GET", "/api/assets/hosts", nil, true, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
}

func (c *Client) Host(ctx context.Context, id string) (Host, error) {
	var out Result[Host]
	if err := c.doJSON(ctx, "GET", "/api/assets/hosts/"+strings.TrimSpace(id), nil, true, &out); err != nil {
		return Host{}, err
	}
	return out.Data, nil
}

type AccessRequestQuery struct {
	Status   mo.Option[string]
	Page     int
	PageSize int
}

func (c *Client) AccessRequests(ctx context.Context, query AccessRequestQuery) (PageResult[AccessRequest], error) {
	params := url.Values{}
	query.Status.ForEach(func(status string) {
		params.Set("status", status)
	})
	if query.Page > 0 {
		params.Set("page", fmt.Sprintf("%d", query.Page))
	}
	if query.PageSize > 0 {
		params.Set("pageSize", fmt.Sprintf("%d", query.PageSize))
	}

	path := "/api/access-requests"
	if queryString := params.Encode(); queryString != "" {
		path += "?" + queryString
	}

	var out Result[PageResult[AccessRequest]]
	if err := c.doJSON(ctx, "GET", path, nil, true, &out); err != nil {
		return PageResult[AccessRequest]{}, err
	}
	return out.Data, nil
}

func (c *Client) ApproveAccessRequest(ctx context.Context, id, reviewer string, comment mo.Option[string]) (AccessRequest, error) {
	return c.reviewAccessRequest(ctx, id, "/approve", reviewer, comment)
}

func (c *Client) RejectAccessRequest(ctx context.Context, id, reviewer string, comment mo.Option[string]) (AccessRequest, error) {
	return c.reviewAccessRequest(ctx, id, "/reject", reviewer, comment)
}

func (c *Client) Sessions(ctx context.Context) ([]Session, error) {
	var out Result[[]Session]
	if err := c.doJSON(ctx, "GET", "/api/sessions", nil, true, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
}

func (c *Client) Session(ctx context.Context, id string) (Session, error) {
	var out Result[Session]
	if err := c.doJSON(ctx, "GET", "/api/sessions/"+strings.TrimSpace(id), nil, true, &out); err != nil {
		return Session{}, err
	}
	return out.Data, nil
}

func (c *Client) Gateways(ctx context.Context) ([]Gateway, error) {
	var out Result[[]Gateway]
	if err := c.doJSON(ctx, "GET", "/api/gateways", nil, true, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
}

func (c *Client) Gateway(ctx context.Context, id string) (Gateway, error) {
	var out Result[Gateway]
	if err := c.doJSON(ctx, "GET", "/api/gateways/"+strings.TrimSpace(id), nil, true, &out); err != nil {
		return Gateway{}, err
	}
	return out.Data, nil
}

func (c *Client) reviewAccessRequest(ctx context.Context, id, action, reviewer string, comment mo.Option[string]) (AccessRequest, error) {
	payload := map[string]any{
		"reviewer": strings.TrimSpace(reviewer),
	}
	comment.ForEach(func(value string) {
		payload["comment"] = value
	})

	var out Result[AccessRequest]
	if err := c.doJSON(ctx, "POST", "/api/access-requests/"+strings.TrimSpace(id)+action, payload, true, &out); err != nil {
		return AccessRequest{}, err
	}
	return out.Data, nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, body any, withAuth bool, out any) error {
	request := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Accept", "application/json")

	if body != nil {
		request.SetHeader("Content-Type", "application/json")
		request.SetBody(map[string]any{"body": body})
	}
	if withAuth && c.accessToken != "" {
		request.SetHeader("Authorization", "Bearer "+c.accessToken)
	}

	response, err := c.httpClient.Execute(ctx, request, method, path)
	if err != nil {
		return err
	}
	if response.IsError() {
		return fmt.Errorf("request failed: %s", strings.TrimSpace(response.String()))
	}
	return json.Unmarshal(response.Bytes(), out)
}
