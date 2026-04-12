package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"strings"

	collectionlist "github.com/DaiYuANg/arcgo/collectionx/list"
	cliapi "github.com/DaiYuANg/jumpa/internal/cli/api"
	"github.com/samber/lo"
	"github.com/samber/mo"
	"golang.org/x/term"
)

type SessionResolver struct {
	cfg    Config
	log    *slog.Logger
	client *cliapi.Client
	stdin  io.Reader
	stdout io.Writer
}

type ResolveOptions struct {
	NeedPrincipal bool
	NeedOverview  bool
}

type SessionContext struct {
	Login       cliapi.LoginResponse
	Overview    mo.Option[cliapi.Overview]
	Principal   mo.Option[string]
	GatewayAddr mo.Option[string]
}

type promptField struct {
	Label    string
	Fallback string
	Secret   bool
	Assign   func(string)
}

func NewSessionResolver(cfg Config, log *slog.Logger, client *cliapi.Client, streams stdio) *SessionResolver {
	return &SessionResolver{
		cfg:    cfg,
		log:    log,
		client: client,
		stdin:  streams.In,
		stdout: streams.Out,
	}
}

func (r *SessionResolver) Resolve(ctx context.Context, opts ResolveOptions) (SessionContext, error) {
	creds, err := r.resolveCredentials(opts.NeedPrincipal)
	if err != nil {
		return SessionContext{}, err
	}

	r.log.Debug("authenticating control-plane session", slog.String("email", creds.Email))
	login, err := r.client.Login(ctx, creds.Email, creds.Password)
	if err != nil {
		return SessionContext{}, fmt.Errorf("cli login: %w", err)
	}

	result := SessionContext{
		Login:     login,
		Principal: optionString(creds.Principal),
	}

	if !opts.NeedOverview {
		return result, nil
	}

	overview, err := r.client.Overview(ctx)
	if err != nil {
		return SessionContext{}, fmt.Errorf("cli overview: %w", err)
	}

	result.Overview = mo.Some(overview)
	result.GatewayAddr = optionString(resolveGatewayAddr(r.cfg.GatewayAddr, r.client.BaseURL(), overview.SSHListenAddr))
	return result, nil
}

type credentials struct {
	Email     string
	Password  string
	Principal string
}

func (r *SessionResolver) resolveCredentials(needPrincipal bool) (credentials, error) {
	input := credentials{
		Email:     strings.TrimSpace(r.cfg.Email),
		Password:  r.cfg.Password,
		Principal: strings.TrimSpace(r.cfg.Principal),
	}

	defaultPrincipal := lo.Ternary(optionString(input.Principal).IsPresent(), input.Principal, localPart(input.Email))
	fields := collectionlist.NewList(
		promptField{
			Label:    "Email",
			Fallback: input.Email,
			Assign: func(value string) {
				input.Email = value
			},
		},
		promptField{
			Label:    "Password",
			Fallback: input.Password,
			Secret:   true,
			Assign: func(value string) {
				input.Password = value
			},
		},
	)
	if needPrincipal {
		fields.Add(promptField{
			Label:    "SSH Principal",
			Fallback: defaultPrincipal,
			Assign: func(value string) {
				input.Principal = value
			},
		})
	}

	reader := bufio.NewReader(r.stdin)
	fields.Range(func(_ int, field promptField) bool {
		if optionString(field.Fallback).IsPresent() {
			field.Assign(strings.TrimSpace(field.Fallback))
			return true
		}
		if field.Secret {
			field.Assign(promptSecret(r.stdin, r.stdout, field.Label))
			return true
		}
		field.Assign(promptText(reader, r.stdout, field.Label, field.Fallback))
		return true
	})

	if optionString(input.Email).IsAbsent() {
		return credentials{}, errors.New("email is required")
	}
	if optionString(input.Password).IsAbsent() {
		return credentials{}, errors.New("password is required")
	}
	if needPrincipal && optionString(input.Principal).IsAbsent() {
		input.Principal = localPart(input.Email)
	}
	if needPrincipal && optionString(input.Principal).IsAbsent() {
		return credentials{}, errors.New("principal is required")
	}

	return input, nil
}

func promptText(reader *bufio.Reader, writer io.Writer, label, fallback string) string {
	if strings.TrimSpace(fallback) != "" {
		_, _ = fmt.Fprintf(writer, "%s [%s]: ", label, fallback)
	} else {
		_, _ = fmt.Fprintf(writer, "%s: ", label)
	}

	value, _ := reader.ReadString('\n')
	value = strings.TrimSpace(value)
	if value == "" {
		return strings.TrimSpace(fallback)
	}
	return value
}

func promptSecret(reader io.Reader, writer io.Writer, label string) string {
	_, _ = fmt.Fprintf(writer, "%s: ", label)
	file, ok := reader.(interface{ Fd() uintptr })
	if !ok {
		_, _ = fmt.Fprintln(writer)
		return ""
	}
	raw, _ := term.ReadPassword(int(file.Fd()))
	_, _ = fmt.Fprintln(writer)
	return strings.TrimSpace(string(raw))
}

func localPart(email string) string {
	value := strings.TrimSpace(email)
	if idx := strings.IndexByte(value, '@'); idx > 0 {
		return value[:idx]
	}
	return value
}

func resolveGatewayAddr(override, apiBase, sshListen string) string {
	if strings.TrimSpace(override) != "" {
		return strings.TrimSpace(override)
	}
	if strings.TrimSpace(sshListen) == "" {
		return "127.0.0.1:2222"
	}
	if strings.HasPrefix(sshListen, ":") {
		u, err := url.Parse(apiBase)
		if err != nil || u.Hostname() == "" {
			return "127.0.0.1" + sshListen
		}
		return u.Hostname() + sshListen
	}
	return sshListen
}

func optionString(value string) mo.Option[string] {
	return StringOverride(value)
}
