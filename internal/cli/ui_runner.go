package cli

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/DaiYuANg/arcgo/dix"
	"github.com/DaiYuANg/jumpa/internal/cli/api"
	cliapp "github.com/DaiYuANg/jumpa/internal/cli/app"
)

type UIRunner struct {
	cfg      Config
	client   *api.Client
	sessions *SessionResolver
	ssh      *SSHLauncher
}

func NewUIModule() dix.Module {
	return dix.NewModule("cli-ui",
		dix.WithModuleProviders(
			dix.Provider4(func(cfg Config, client *api.Client, sessions *SessionResolver, ssh *SSHLauncher) CommandRunner {
				return &UIRunner{
					cfg:      cfg,
					client:   client,
					sessions: sessions,
					ssh:      ssh,
				}
			}),
		),
	)
}

func (r *UIRunner) Run(ctx context.Context) error {
	session, err := r.sessions.Resolve(ctx, ResolveOptions{
		NeedPrincipal: true,
		NeedOverview:  true,
	})
	if err != nil {
		return err
	}

	gatewayAddr, _ := session.GatewayAddr.Get()
	principal, _ := session.Principal.Get()
	loginPassword, _ := session.LoginPassword.Get()
	model := cliapp.New(r.client, cliapp.Options{
		Principal:   principal,
		GatewayAddr: gatewayAddr,
		Me:          session.Login.User,
		AltScreen:   r.cfg.AltScreen,
	})

	finalModel, err := tea.NewProgram(model).Run()
	if err != nil {
		return fmt.Errorf("cli ui: %w", err)
	}

	typed, ok := finalModel.(cliapp.Model)
	if !ok || typed.LaunchRequest() == nil {
		return nil
	}

	return r.ssh.Launch(*typed.LaunchRequest(), SSHLaunchOptions{
		Password: loginPassword,
	})
}
