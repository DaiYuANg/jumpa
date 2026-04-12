package cli

import (
	"context"
	"log/slog"

	collectionlist "github.com/DaiYuANg/arcgo/collectionx/list"
	"github.com/DaiYuANg/arcgo/dix"
	"github.com/samber/mo"
	"github.com/spf13/cobra"
)

type rootFlags struct {
	APIURL      string
	GatewayAddr string
	Email       string
	Password    string
	Principal   string
	SSHBinary   string
	AltScreen   bool
}

func NewRootCommand(log *slog.Logger) *cobra.Command {
	flags := &rootFlags{}

	root := &cobra.Command{
		Use:           "jumpa",
		Short:         "jumpa control-plane cli",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSubApp(cmd.Context(), log, "jumpa-cli-ui", flags.overrides(cmd), NewUIModule())
		},
	}

	bindRootFlags(root, flags)
	commands := collectionlist.NewList(
		newUICmd(log, flags),
		newHostsCmd(log, flags),
		newSessionsCmd(log, flags),
		newRequestsCmd(log, flags),
		newGatewaysCmd(log, flags),
		newConnectCmd(log, flags),
	)
	commands.Range(func(_ int, command *cobra.Command) bool {
		root.AddCommand(command)
		return true
	})

	return root
}

func bindRootFlags(cmd *cobra.Command, flags *rootFlags) {
	cmd.PersistentFlags().StringVar(&flags.APIURL, "api", "", "jumpa api base url")
	cmd.PersistentFlags().StringVar(&flags.GatewayAddr, "gateway", "", "override ssh gateway host:port")
	cmd.PersistentFlags().StringVar(&flags.Email, "email", "", "login email")
	cmd.PersistentFlags().StringVar(&flags.Password, "password", "", "login password")
	cmd.PersistentFlags().StringVar(&flags.Principal, "principal", "", "ssh principal used for gateway login")
	cmd.PersistentFlags().StringVar(&flags.SSHBinary, "ssh", "", "ssh binary path")
	cmd.PersistentFlags().BoolVar(&flags.AltScreen, "alt-screen", true, "run bubbletea with alt screen")
}

func (f *rootFlags) overrides(cmd *cobra.Command) Overrides {
	if f == nil || cmd == nil {
		return Overrides{}
	}
	return Overrides{
		APIURL:      StringOverride(f.APIURL),
		GatewayAddr: StringOverride(f.GatewayAddr),
		Email:       StringOverride(f.Email),
		Password:    StringOverride(f.Password),
		Principal:   StringOverride(f.Principal),
		SSHBinary:   StringOverride(f.SSHBinary),
		AltScreen:   boolOverride(cmd, "alt-screen", f.AltScreen),
	}
}

func boolOverride(cmd *cobra.Command, name string, value bool) mo.Option[bool] {
	flag := cmd.Flags().Lookup(name)
	if flag == nil || !flag.Changed {
		return mo.None[bool]()
	}
	return mo.Some(value)
}

func runSubApp(ctx context.Context, log *slog.Logger, name string, overrides Overrides, commandModule dix.Module) error {
	app := dix.New(
		name,
		dix.WithVersion("0.1.0"),
		dix.WithLogger(log),
		dix.WithModules(
			NewCommonModule(overrides),
			commandModule,
		),
	)

	runtime, err := app.Start(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = runtime.Stop(context.WithoutCancel(ctx)) }()

	runner, err := dix.ResolveAs[CommandRunner](runtime.Container())
	if err != nil {
		return err
	}

	return runner.Run(ctx)
}

func newUICmd(log *slog.Logger, flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "ui",
		Short: "launch the interactive terminal UI",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSubApp(cmd.Context(), log, "jumpa-cli-ui", flags.overrides(cmd), NewUIModule())
		},
	}
}

func newHostsCmd(log *slog.Logger, flags *rootFlags) *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "hosts",
		Short: "list bastion hosts",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSubApp(cmd.Context(), log, "jumpa-cli-hosts", flags.overrides(cmd), NewHostsModule(ListOptions{
				JSON: jsonOutput,
			}))
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "print JSON")
	return cmd
}

func newSessionsCmd(log *slog.Logger, flags *rootFlags) *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "sessions",
		Short: "list bastion sessions",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSubApp(cmd.Context(), log, "jumpa-cli-sessions", flags.overrides(cmd), NewSessionsModule(ListOptions{
				JSON: jsonOutput,
			}))
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "print JSON")
	return cmd
}

func newRequestsCmd(log *slog.Logger, flags *rootFlags) *cobra.Command {
	options := ListOptions{
		Page:     1,
		PageSize: 50,
	}
	cmd := &cobra.Command{
		Use:   "requests",
		Short: "list and review access requests",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSubApp(cmd.Context(), log, "jumpa-cli-requests", flags.overrides(cmd), NewRequestsModule(options))
		},
	}
	cmd.Flags().BoolVar(&options.JSON, "json", false, "print JSON")
	cmd.Flags().StringVar(&options.Status, "status", "", "filter by request status")
	cmd.Flags().IntVar(&options.Page, "page", 1, "page number")
	cmd.Flags().IntVar(&options.PageSize, "page-size", 50, "page size")
	cmd.AddCommand(
		newRequestReviewCmd(log, flags, false),
		newRequestReviewCmd(log, flags, true),
	)
	return cmd
}

func newRequestReviewCmd(log *slog.Logger, flags *rootFlags, reject bool) *cobra.Command {
	options := AccessRequestReviewOptions{Reject: reject}
	use := "approve <request-id>"
	short := "approve an access request"
	appName := "jumpa-cli-requests-approve"
	if reject {
		use = "reject <request-id>"
		short = "reject an access request"
		appName = "jumpa-cli-requests-reject"
	}

	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.ID = args[0]
			return runSubApp(cmd.Context(), log, appName, flags.overrides(cmd), NewAccessRequestReviewModule(options))
		},
	}
	cmd.Flags().StringVar(&options.Reviewer, "reviewer", "", "reviewer identity; defaults to current login email")
	cmd.Flags().StringVar(&options.Comment, "comment", "", "optional review comment")
	cmd.Flags().BoolVar(&options.JSON, "json", false, "print JSON")
	return cmd
}

func newGatewaysCmd(log *slog.Logger, flags *rootFlags) *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "gateways",
		Short: "list registered gateways",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSubApp(cmd.Context(), log, "jumpa-cli-gateways", flags.overrides(cmd), NewGatewaysModule(ListOptions{
				JSON: jsonOutput,
			}))
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "print JSON")
	return cmd
}

func newConnectCmd(log *slog.Logger, flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "connect <host> [account]",
		Short: "launch local ssh through the jumpa gateway",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			account := ""
			if len(args) > 1 {
				account = args[1]
			}
			return runSubApp(cmd.Context(), log, "jumpa-cli-connect", flags.overrides(cmd), NewConnectModule(ConnectOptions{
				Host:    args[0],
				Account: account,
			}))
		},
	}
}
