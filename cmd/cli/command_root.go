package main

import (
	"context"
	"log/slog"
	"strings"

	cli "github.com/DaiYuANg/jumpa/internal/cli"
	"github.com/arcgolabs/dix"
	"github.com/spf13/cobra"
)

const (
	workflowCommandGroup = "workflow"
	resourceCommandGroup = "resources"
	reviewCommandGroup   = "review"
)

var accessRequestStatusValues = []string{"pending", "approved", "rejected"}

func newRootCommand(log *slog.Logger) *cobra.Command {
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
			return runSubApp(cmd.Context(), log, cmd, "jumpa-cli-ui", cli.NewUIModule())
		},
	}

	root.AddGroup(
		&cobra.Group{ID: workflowCommandGroup, Title: "Workflow Commands"},
		&cobra.Group{ID: resourceCommandGroup, Title: "Resource Commands"},
		&cobra.Group{ID: reviewCommandGroup, Title: "Review Commands"},
	)

	bindRootFlags(root)
	root.AddCommand(
		newUICmd(log),
		newHostsCmd(log),
		newSessionsCmd(log),
		newRequestsCmd(log),
		newGatewaysCmd(log),
		newConnectCmd(log),
	)

	return root
}

func bindRootFlags(cmd *cobra.Command) {
	flags := cmd.PersistentFlags()
	flags.String("api", "", "jumpa API base URL")
	flags.String("gateway", "", "override SSH gateway host:port")
	flags.String("email", "", "login email for control-plane authentication")
	flags.String("password", "", "login password for control-plane authentication")
	flags.String("principal", "", "SSH principal used when launching gateway sessions")
	flags.String("ssh-config", "", "override the ssh config file used for gateway connection defaults")
	flags.String("ssh-key", "", "private key path used for SSH authentication")
	flags.String("ssh-key-passphrase", "", "private key passphrase used for SSH authentication")
	flags.Bool("ssh-agent", false, "enable SSH agent authentication using SSH_AUTH_SOCK or --ssh-agent-sock")
	flags.String("ssh-agent-sock", "", "override the SSH agent socket path")
	flags.Bool("alt-screen", true, "run the Bubble Tea UI in the terminal alternate screen")
}

func runSubApp(ctx context.Context, log *slog.Logger, cmd *cobra.Command, name string, commandModule dix.Module) error {
	cfg, err := cli.LoadAndValidateConfig(cmd.Root().PersistentFlags())
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
