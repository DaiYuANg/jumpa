package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ClientOption func(*Client)

type Client struct {
	baseURL      string
	httpClient   *http.Client
	accessToken  string
	refreshToken string
}

func NewClient(baseURL string, opts ...ClientOption) *Client {
	client := &Client{
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
	for _, opt := range opts {
		if opt != nil {
			opt(client)
		}
	}
	return client
}

func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(client *Client) {
		if client == nil || httpClient == nil {
			return
		}
		client.httpClient = httpClient
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
	if err := c.doJSON(ctx, http.MethodPost, "/api/auth/login", payload, false, &out); err != nil {
		return LoginResponse{}, err
	}
	c.accessToken = out.Data.AccessToken
	c.refreshToken = out.Data.RefreshToken
	return out.Data, nil
}

func (c *Client) Overview(ctx context.Context) (Overview, error) {
	var out Result[Overview]
	if err := c.doJSON(ctx, http.MethodGet, "/api/bastion/overview", nil, true, &out); err != nil {
		return Overview{}, err
	}
	return out.Data, nil
}

func (c *Client) Hosts(ctx context.Context) ([]Host, error) {
	var out Result[[]Host]
	if err := c.doJSON(ctx, http.MethodGet, "/api/assets/hosts", nil, true, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
}

func (c *Client) AccessRequests(ctx context.Context) ([]AccessRequest, error) {
	var out Result[PageResult[AccessRequest]]
	if err := c.doJSON(ctx, http.MethodGet, "/api/access-requests?page=1&pageSize=50", nil, true, &out); err != nil {
		return nil, err
	}
	return out.Data.Items, nil
}

func (c *Client) Sessions(ctx context.Context) ([]Session, error) {
	var out Result[[]Session]
	if err := c.doJSON(ctx, http.MethodGet, "/api/sessions", nil, true, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
}

func (c *Client) Gateways(ctx context.Context) ([]Gateway, error) {
	var out Result[[]Gateway]
	if err := c.doJSON(ctx, http.MethodGet, "/api/gateways", nil, true, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, body any, withAuth bool, out any) error {
	target, err := url.JoinPath(c.baseURL, path)
	if err != nil {
		return err
	}
	if strings.Contains(path, "?") {
		target = c.baseURL + path
	}

	var reader io.Reader
	if body != nil {
		raw, marshalErr := json.Marshal(map[string]any{"body": body})
		if marshalErr != nil {
			return marshalErr
		}
		reader = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, target, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if withAuth && c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("request failed: %s", strings.TrimSpace(string(raw)))
	}
	return json.Unmarshal(raw, out)
}
