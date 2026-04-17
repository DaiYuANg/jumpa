package main

import (
	"context"
	"log/slog"
	"strings"

	"github.com/DaiYuANg/arcgo/dix"
	cli "github.com/DaiYuANg/jumpa/internal/cli"
	"github.com/samber/mo"
	"github.com/spf13/cobra"
)

const (
	workflowCommandGroup = "workflow"
	resourceCommandGroup = "resources"
	reviewCommandGroup   = "review"
)

var accessRequestStatusValues = []string{"pending", "approved", "rejected"}

type rootFlags struct {
	APIURL                  string
	GatewayAddr             string
	Email                   string
	Password                string
	Principal               string
	SSHConfigPath           string
	SSHPrivateKeyPath       string
	SSHPrivateKeyPassphrase string
	SSHAgentEnabled         bool
	SSHAgentSocket          string
	AltScreen               bool
}

func newRootCommand(log *slog.Logger) *cobra.Command {
	flags := &rootFlags{}

	root := &cobra.Command{
		Use:   "jumpa",
		Short: "Operate the Jumpa control plane from a terminal",
		Long: commandText(
			"Operate the Jumpa control plane from a terminal.",
			"",
			"Most commands authenticate against the HTTP API using credentials from",
			"APP_CLI_* environment variables, .env, or interactive prompts.",
			"",
			"The connect command launches an SSH session through the bastion",
			"gateway after resolving login state from the control plane.",
		),
		Example: commandText(
			"jumpa ui",
			"jumpa hosts --json",
			"jumpa requests --status pending",
			"jumpa connect prod-web-01 ubuntu",
		),
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSubApp(cmd.Context(), log, flags.overrides(cmd), "jumpa-cli-ui", cli.NewUIModule())
		},
	}

	root.AddGroup(
		&cobra.Group{ID: workflowCommandGroup, Title: "Workflow Commands"},
		&cobra.Group{ID: resourceCommandGroup, Title: "Resource Commands"},
		&cobra.Group{ID: reviewCommandGroup, Title: "Review Commands"},
	)

	bindRootFlags(root, flags)
	root.AddCommand(
		newUICmd(log, flags),
		newHostsCmd(log, flags),
		newSessionsCmd(log, flags),
		newRequestsCmd(log, flags),
		newGatewaysCmd(log, flags),
		newConnectCmd(log, flags),
	)

	return root
}

func bindRootFlags(cmd *cobra.Command, flags *rootFlags) {
	cmd.PersistentFlags().StringVar(&flags.APIURL, "api", "", "jumpa API base URL")
	cmd.PersistentFlags().StringVar(&flags.GatewayAddr, "gateway", "", "override SSH gateway host:port")
	cmd.PersistentFlags().StringVar(&flags.Email, "email", "", "login email for control-plane authentication")
	cmd.PersistentFlags().StringVar(&flags.Password, "password", "", "login password for control-plane authentication")
	cmd.PersistentFlags().StringVar(&flags.Principal, "principal", "", "SSH principal used when launching gateway sessions")
	cmd.PersistentFlags().StringVar(&flags.SSHConfigPath, "ssh-config", "", "override the ssh config file used for gateway connection defaults")
	cmd.PersistentFlags().StringVar(&flags.SSHPrivateKeyPath, "ssh-key", "", "private key path used for SSH authentication")
	cmd.PersistentFlags().StringVar(&flags.SSHPrivateKeyPassphrase, "ssh-key-passphrase", "", "private key passphrase used for SSH authentication")
	cmd.PersistentFlags().BoolVar(&flags.SSHAgentEnabled, "ssh-agent", false, "enable SSH agent authentication using SSH_AUTH_SOCK or --ssh-agent-sock")
	cmd.PersistentFlags().StringVar(&flags.SSHAgentSocket, "ssh-agent-sock", "", "override the SSH agent socket path")
	cmd.PersistentFlags().BoolVar(&flags.AltScreen, "alt-screen", true, "run the Bubble Tea UI in the terminal alternate screen")
}

func (f *rootFlags) overrides(cmd *cobra.Command) cli.Overrides {
	if f == nil || cmd == nil {
		return cli.Overrides{}
	}
	return cli.Overrides{
		APIURL:                  cli.StringOverride(f.APIURL),
		GatewayAddr:             cli.StringOverride(f.GatewayAddr),
		Email:                   cli.StringOverride(f.Email),
		Password:                cli.StringOverride(f.Password),
		Principal:               cli.StringOverride(f.Principal),
		SSHConfigPath:           cli.StringOverride(f.SSHConfigPath),
		SSHPrivateKeyPath:       cli.StringOverride(f.SSHPrivateKeyPath),
		SSHPrivateKeyPassphrase: cli.StringOverride(f.SSHPrivateKeyPassphrase),
		SSHAgentEnabled:         boolOverride(cmd, "ssh-agent", f.SSHAgentEnabled),
		SSHAgentSocket:          cli.StringOverride(f.SSHAgentSocket),
		AltScreen:               boolOverride(cmd, "alt-screen", f.AltScreen),
	}
}

func boolOverride(cmd *cobra.Command, name string, value bool) mo.Option[bool] {
	flag := cmd.Flags().Lookup(name)
	if flag == nil || !flag.Changed {
		return mo.None[bool]()
	}
	return mo.Some(value)
}

func runSubApp(ctx context.Context, log *slog.Logger, overrides cli.Overrides, name string, commandModule dix.Module) error {
	cfg, err := cli.ResolveConfig(overrides)
	if err != nil {
		return err
	}

	app := dix.New(
		name,
		dix.WithVersion("0.1.0"),
		dix.WithLogger(log),
		dix.WithModules(
			cli.NewCommonModule(cfg),
			commandModule,
		),
	)

	runtime, err := app.Start(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = runtime.Stop(context.WithoutCancel(ctx)) }()

	runner, err := dix.ResolveAs[cli.CommandRunner](runtime.Container())
	if err != nil {
		return err
	}

	return runner.Run(ctx)
}

func commandText(lines ...string) string {
	return strings.TrimSpace(strings.Join(lines, "\n"))
}
