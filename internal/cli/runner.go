package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"strings"

	tea "charm.land/bubbletea/v2"
	collectionlist "github.com/DaiYuANg/arcgo/collectionx/list"
	cliapi "github.com/DaiYuANg/jumpa/internal/cli/api"
	cliapp "github.com/DaiYuANg/jumpa/internal/cli/app"
	"github.com/samber/lo"
	"golang.org/x/term"
)

type Runner struct {
	cfg    Config
	log    *slog.Logger
	client *cliapi.Client
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

type loginInput struct {
	Email     string
	Password  string
	Principal string
}

type promptField struct {
	Label    string
	Fallback string
	Secret   bool
	Assign   func(string)
}

func NewRunner(cfg Config, log *slog.Logger, client *cliapi.Client) *Runner {
	return &Runner{
		cfg:    cfg,
		log:    log,
		client: client,
		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

func (r *Runner) Run(ctx context.Context) error {
	input, err := r.resolveLoginInput()
	if err != nil {
		return err
	}

	login, err := r.client.Login(ctx, input.Email, input.Password)
	if err != nil {
		return fmt.Errorf("cli login: %w", err)
	}

	overview, err := r.client.Overview(ctx)
	if err != nil {
		return fmt.Errorf("cli overview: %w", err)
	}

	gatewayAddr := resolveGatewayAddr(r.cfg.GatewayAddr, r.client.BaseURL(), overview.SSHListenAddr)
	model := cliapp.New(r.client, cliapp.Options{
		Principal:   input.Principal,
		GatewayAddr: gatewayAddr,
		Me:          login.User,
		AltScreen:   r.cfg.AltScreen,
	})

	finalModel, err := tea.NewProgram(model).Run()
	if err != nil {
		return fmt.Errorf("cli ui: %w", err)
	}

	typed, ok := finalModel.(cliapp.Model)
	if !ok {
		return nil
	}

	launch := typed.LaunchRequest()
	if launch == nil {
		return nil
	}

	return r.runSSH(*launch)
}

func (r *Runner) resolveLoginInput() (loginInput, error) {
	input := loginInput{
		Email:     strings.TrimSpace(r.cfg.Email),
		Password:  r.cfg.Password,
		Principal: strings.TrimSpace(r.cfg.Principal),
	}

	defaultPrincipal := lo.Ternary(
		trimmedOption(input.Principal).IsPresent(),
		input.Principal,
		localPart(input.Email),
	)

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
		promptField{
			Label:    "SSH Principal",
			Fallback: defaultPrincipal,
			Assign: func(value string) {
				input.Principal = value
			},
		},
	)

	reader := bufio.NewReader(r.stdin)
	fields.Range(func(_ int, field promptField) bool {
		if trimmedOption(field.Fallback).IsPresent() {
			field.Assign(field.Fallback)
			return true
		}
		if field.Secret {
			field.Assign(promptSecret(r.stdout, field.Label))
			return true
		}

		field.Assign(promptText(reader, r.stdout, field.Label, field.Fallback))
		return true
	})

	if trimmedOption(input.Email).IsAbsent() {
		return loginInput{}, errors.New("email is required")
	}
	if trimmedOption(input.Password).IsAbsent() {
		return loginInput{}, errors.New("password is required")
	}
	if trimmedOption(input.Principal).IsAbsent() {
		input.Principal = localPart(input.Email)
	}
	if trimmedOption(input.Principal).IsAbsent() {
		return loginInput{}, errors.New("principal is required")
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

func promptSecret(writer io.Writer, label string) string {
	_, _ = fmt.Fprintf(writer, "%s: ", label)
	raw, _ := term.ReadPassword(int(os.Stdin.Fd()))
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

func (r *Runner) runSSH(req cliapp.LaunchRequest) error {
	target := fmt.Sprintf("%s@%s", req.Target, req.GatewayHost)
	binary := trimmedOption(r.cfg.SSHBinary).OrElse("ssh")
	cmd := exec.Command(binary, "-p", req.GatewayPort, target)
	cmd.Stdin = r.stdin
	cmd.Stdout = r.stdout
	cmd.Stderr = r.stderr

	r.log.Info("launching ssh session",
		slog.String("binary", binary),
		slog.String("target", target),
		slog.String("port", req.GatewayPort),
	)

	return cmd.Run()
}
