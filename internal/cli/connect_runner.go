package cli

import (
	"context"

	"github.com/arcgolabs/dix"
	"github.com/DaiYuANg/jumpa/internal/sshclient"
)

type ConnectOptions struct {
	Host            string
	Account         string
	LocalForwards   []sshclient.LocalForward
	RemoteForwards  []sshclient.RemoteForward
	DynamicForwards []sshclient.DynamicForward
}

type ConnectRunner struct {
	sessions *SessionResolver
	ssh      *SSHLauncher
	options  ConnectOptions
}

func NewConnectModule(options ConnectOptions) dix.Module {
	return dix.NewModule("cli-connect",
		dix.WithModuleProviders(
			dix.Provider2(func(sessions *SessionResolver, ssh *SSHLauncher) CommandRunner {
				return &ConnectRunner{
					sessions: sessions,
					ssh:      ssh,
					options:  options,
				}
			}),
		),
	)
}

func (r *ConnectRunner) Run(ctx context.Context) error {
	session, err := r.sessions.Resolve(ctx, ResolveOptions{
		NeedPrincipal: true,
		NeedOverview:  true,
	})
	if err != nil {
		return err
	}

	principal, _ := session.Principal.Get()
	gatewayAddr, _ := session.GatewayAddr.Get()
	loginPassword, _ := session.LoginPassword.Get()
	request, err := BuildLaunchRequest(principal, gatewayAddr, r.options.Host, r.options.Account)
	if err != nil {
		return err
	}
	return r.ssh.Launch(request, SSHLaunchOptions{
		Password:        loginPassword,
		LocalForwards:   r.options.LocalForwards,
		RemoteForwards:  r.options.RemoteForwards,
		DynamicForwards: r.options.DynamicForwards,
	})
}
