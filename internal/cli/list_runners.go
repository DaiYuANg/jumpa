package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/jumpa/internal/cli/api"
	"github.com/samber/lo"
)

type ListOptions struct {
	JSON     bool
	Status   string
	Page     int
	PageSize int
}

type HostsRunner struct {
	out      io.Writer
	client   *api.Client
	sessions *SessionResolver
	json     bool
}

type SessionsRunner struct {
	out      io.Writer
	client   *api.Client
	sessions *SessionResolver
	json     bool
}

type GatewaysRunner struct {
	out      io.Writer
	client   *api.Client
	sessions *SessionResolver
	json     bool
}

type RequestsRunner struct {
	out      io.Writer
	client   *api.Client
	sessions *SessionResolver
	options  ListOptions
}

type AccessRequestReviewOptions struct {
	ID       string
	Reviewer string
	Comment  string
	JSON     bool
	Reject   bool
}

type AccessRequestReviewRunner struct {
	out      io.Writer
	client   *api.Client
	sessions *SessionResolver
	options  AccessRequestReviewOptions
}

type DetailOptions struct {
	ID   string
	JSON bool
}

type HostDetailRunner struct {
	out      io.Writer
	client   *api.Client
	sessions *SessionResolver
	options  DetailOptions
}

type SessionDetailRunner struct {
	out      io.Writer
	client   *api.Client
	sessions *SessionResolver
	options  DetailOptions
}

type GatewayDetailRunner struct {
	out      io.Writer
	client   *api.Client
	sessions *SessionResolver
	options  DetailOptions
}

func NewHostsModule(options ListOptions) dix.Module {
	return dix.NewModule("cli-hosts",
		dix.WithModuleProviders(
			dix.Provider3(func(streams stdio, client *api.Client, sessions *SessionResolver) CommandRunner {
				return &HostsRunner{out: streams.Out, client: client, sessions: sessions, json: options.JSON}
			}),
		),
	)
}

func NewSessionsModule(options ListOptions) dix.Module {
	return dix.NewModule("cli-sessions",
		dix.WithModuleProviders(
			dix.Provider3(func(streams stdio, client *api.Client, sessions *SessionResolver) CommandRunner {
				return &SessionsRunner{out: streams.Out, client: client, sessions: sessions, json: options.JSON}
			}),
		),
	)
}

func NewGatewaysModule(options ListOptions) dix.Module {
	return dix.NewModule("cli-gateways",
		dix.WithModuleProviders(
			dix.Provider3(func(streams stdio, client *api.Client, sessions *SessionResolver) CommandRunner {
				return &GatewaysRunner{out: streams.Out, client: client, sessions: sessions, json: options.JSON}
			}),
		),
	)
}

func NewRequestsModule(options ListOptions) dix.Module {
	return dix.NewModule("cli-requests",
		dix.WithModuleProviders(
			dix.Provider3(func(streams stdio, client *api.Client, sessions *SessionResolver) CommandRunner {
				return &RequestsRunner{out: streams.Out, client: client, sessions: sessions, options: options}
			}),
		),
	)
}

func NewAccessRequestReviewModule(options AccessRequestReviewOptions) dix.Module {
	return dix.NewModule("cli-requests-review",
		dix.WithModuleProviders(
			dix.Provider3(func(streams stdio, client *api.Client, sessions *SessionResolver) CommandRunner {
				return &AccessRequestReviewRunner{out: streams.Out, client: client, sessions: sessions, options: options}
			}),
		),
	)
}

func NewHostDetailModule(options DetailOptions) dix.Module {
	return dix.NewModule("cli-host-detail",
		dix.WithModuleProviders(
			dix.Provider3(func(streams stdio, client *api.Client, sessions *SessionResolver) CommandRunner {
				return &HostDetailRunner{out: streams.Out, client: client, sessions: sessions, options: options}
			}),
		),
	)
}

func NewSessionDetailModule(options DetailOptions) dix.Module {
	return dix.NewModule("cli-session-detail",
		dix.WithModuleProviders(
			dix.Provider3(func(streams stdio, client *api.Client, sessions *SessionResolver) CommandRunner {
				return &SessionDetailRunner{out: streams.Out, client: client, sessions: sessions, options: options}
			}),
		),
	)
}

func NewGatewayDetailModule(options DetailOptions) dix.Module {
	return dix.NewModule("cli-gateway-detail",
		dix.WithModuleProviders(
			dix.Provider3(func(streams stdio, client *api.Client, sessions *SessionResolver) CommandRunner {
				return &GatewayDetailRunner{out: streams.Out, client: client, sessions: sessions, options: options}
			}),
		),
	)
}

func (r *HostsRunner) Run(ctx context.Context) error {
	if _, err := r.sessions.Resolve(ctx, ResolveOptions{}); err != nil {
		return err
	}
	hosts, err := r.client.Hosts(ctx)
	if err != nil {
		return err
	}
	if r.json {
		return WriteJSON(r.out, hosts)
	}
	rows := lo.Map(hosts, func(host api.Host, _ int) []string {
		return []string{
			host.Name,
			fmt.Sprintf("%s:%d", host.Address, host.Port),
			host.Protocol,
			host.Environment,
			host.Platform,
			lo.Ternary(host.JumpEnabled, "on", "off"),
			host.Authentication,
		}
	})
	return WriteTable(r.out, []string{"NAME", "ADDRESS", "PROTO", "ENV", "PLATFORM", "JUMP", "AUTH"}, rows)
}

func (r *SessionsRunner) Run(ctx context.Context) error {
	if _, err := r.sessions.Resolve(ctx, ResolveOptions{}); err != nil {
		return err
	}
	sessions, err := r.client.Sessions(ctx)
	if err != nil {
		return err
	}
	if r.json {
		return WriteJSON(r.out, sessions)
	}
	rows := lo.Map(sessions, func(session api.Session, _ int) []string {
		return []string{
			session.ID,
			session.Status,
			session.HostName,
			session.HostAccount,
			session.PrincipalName,
			session.StartedAt.Local().Format("2006-01-02 15:04:05"),
		}
	})
	return WriteTable(r.out, []string{"ID", "STATUS", "HOST", "ACCOUNT", "PRINCIPAL", "STARTED"}, rows)
}

func (r *GatewaysRunner) Run(ctx context.Context) error {
	if _, err := r.sessions.Resolve(ctx, ResolveOptions{}); err != nil {
		return err
	}
	gateways, err := r.client.Gateways(ctx)
	if err != nil {
		return err
	}
	if r.json {
		return WriteJSON(r.out, gateways)
	}
	rows := lo.Map(gateways, func(gateway api.Gateway, _ int) []string {
		return []string{
			gateway.NodeName,
			gateway.Zone,
			gateway.EffectiveStatus,
			gateway.AdvertiseAddr,
			strings.Join(gateway.Tags, ","),
			gateway.LastSeenAt.Local().Format("2006-01-02 15:04:05"),
		}
	})
	return WriteTable(r.out, []string{"NAME", "ZONE", "STATUS", "ADDR", "TAGS", "LAST_SEEN"}, rows)
}

func (r *RequestsRunner) Run(ctx context.Context) error {
	if _, err := r.sessions.Resolve(ctx, ResolveOptions{}); err != nil {
		return err
	}
	result, err := r.client.AccessRequests(ctx, api.AccessRequestQuery{
		Status:   StringOverride(r.options.Status),
		Page:     r.options.Page,
		PageSize: r.options.PageSize,
	})
	if err != nil {
		return err
	}
	if r.options.JSON {
		return WriteJSON(r.out, result)
	}
	rows := lo.Map(result.Items, func(request api.AccessRequest, _ int) []string {
		return []string{
			request.ID,
			request.Status,
			request.HostName,
			request.HostAccount,
			request.Protocol,
			request.PrincipalName,
			request.RequestedAt.Local().Format("2006-01-02 15:04:05"),
		}
	})
	return WriteTable(r.out, []string{"ID", "STATUS", "HOST", "ACCOUNT", "PROTO", "PRINCIPAL", "REQUESTED"}, rows)
}

func (r *AccessRequestReviewRunner) Run(ctx context.Context) error {
	session, err := r.sessions.Resolve(ctx, ResolveOptions{})
	if err != nil {
		return err
	}

	reviewer := lo.Ternary(
		StringOverride(r.options.Reviewer).IsPresent(),
		strings.TrimSpace(r.options.Reviewer),
		strings.TrimSpace(session.Login.User.Email),
	)
	if StringOverride(reviewer).IsAbsent() {
		return errors.New("reviewer is required")
	}

	comment := StringOverride(r.options.Comment)
	requestID := strings.TrimSpace(r.options.ID)

	var item api.AccessRequest
	if r.options.Reject {
		item, err = r.client.RejectAccessRequest(ctx, requestID, reviewer, comment)
	} else {
		item, err = r.client.ApproveAccessRequest(ctx, requestID, reviewer, comment)
	}
	if err != nil {
		return err
	}

	if r.options.JSON {
		return WriteJSON(r.out, item)
	}

	rows := [][]string{{
		item.ID,
		item.Status,
		item.HostName,
		item.HostAccount,
		item.Protocol,
		lo.FromPtr(item.ReviewedBy),
		formatOptionalTime(item.ReviewedAt),
		formatOptionalTime(item.ApprovedUntil),
	}}
	return WriteTable(r.out, []string{"ID", "STATUS", "HOST", "ACCOUNT", "PROTO", "REVIEWER", "REVIEWED_AT", "APPROVED_UNTIL"}, rows)
}

func (r *HostDetailRunner) Run(ctx context.Context) error {
	if _, err := r.sessions.Resolve(ctx, ResolveOptions{}); err != nil {
		return err
	}
	item, err := r.client.Host(ctx, r.options.ID)
	if err != nil {
		return err
	}
	if r.options.JSON {
		return WriteJSON(r.out, item)
	}
	rows := [][]string{
		{"ID", item.ID},
		{"NAME", item.Name},
		{"ADDRESS", item.Address},
		{"PORT", fmt.Sprintf("%d", item.Port)},
		{"PROTOCOL", item.Protocol},
		{"ENVIRONMENT", item.Environment},
		{"PLATFORM", item.Platform},
		{"AUTHENTICATION", item.Authentication},
		{"JUMP_ENABLED", fmt.Sprintf("%t", item.JumpEnabled)},
		{"RECORDING_POLICY", item.RecordingPolicy},
		{"CREATED_AT", item.CreatedAt.Local().Format("2006-01-02 15:04:05")},
	}
	return WriteKeyValueTable(r.out, rows)
}

func (r *SessionDetailRunner) Run(ctx context.Context) error {
	if _, err := r.sessions.Resolve(ctx, ResolveOptions{}); err != nil {
		return err
	}
	item, err := r.client.Session(ctx, r.options.ID)
	if err != nil {
		return err
	}
	if r.options.JSON {
		return WriteJSON(r.out, item)
	}
	rows := [][]string{
		{"ID", item.ID},
		{"STATUS", item.Status},
		{"HOST", item.HostName},
		{"ACCOUNT", item.HostAccount},
		{"PRINCIPAL", item.PrincipalName},
		{"PROTOCOL", item.Protocol},
		{"STARTED_AT", item.StartedAt.Local().Format("2006-01-02 15:04:05")},
		{"ENDED_AT", formatOptionalTime(item.EndedAt)},
	}
	return WriteKeyValueTable(r.out, rows)
}

func (r *GatewayDetailRunner) Run(ctx context.Context) error {
	if _, err := r.sessions.Resolve(ctx, ResolveOptions{}); err != nil {
		return err
	}
	item, err := r.client.Gateway(ctx, r.options.ID)
	if err != nil {
		return err
	}
	if r.options.JSON {
		return WriteJSON(r.out, item)
	}
	rows := [][]string{
		{"ID", item.ID},
		{"NODE_KEY", item.NodeKey},
		{"NODE_NAME", item.NodeName},
		{"RUNTIME_TYPE", item.RuntimeType},
		{"ADVERTISE_ADDR", item.AdvertiseAddr},
		{"SSH_LISTEN_ADDR", item.SSHListenAddr},
		{"ZONE", item.Zone},
		{"TAGS", strings.Join(item.Tags, ",")},
		{"STATE", item.State},
		{"EFFECTIVE_STATUS", item.EffectiveStatus},
		{"REGISTERED_AT", item.RegisteredAt.Local().Format("2006-01-02 15:04:05")},
		{"LAST_SEEN_AT", item.LastSeenAt.Local().Format("2006-01-02 15:04:05")},
		{"UPDATED_AT", item.UpdatedAt.Local().Format("2006-01-02 15:04:05")},
	}
	return WriteKeyValueTable(r.out, rows)
}

func formatOptionalTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Local().Format("2006-01-02 15:04:05")
}
