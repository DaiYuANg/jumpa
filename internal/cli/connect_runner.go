package cli

import (
	"context"

	"github.com/DaiYuANg/arcgo/dix"
)

type ConnectOptions struct {
	Host    string
	Account string
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
	request := BuildLaunchRequest(principal, gatewayAddr, r.options.Host, r.options.Account)
	return r.ssh.Launch(request)
}
