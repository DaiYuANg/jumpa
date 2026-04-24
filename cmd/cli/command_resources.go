package main

import (
	"log/slog"

	cli "github.com/DaiYuANg/jumpa/internal/cli"
	"github.com/spf13/cobra"
)

func newHostsCmd(log *slog.Logger) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		GroupID: resourceCommandGroup,
		Use:     "hosts",
		Aliases: []string{"host"},
		Short:   "List bastion hosts",
		Long: commandText(
			"List bastion hosts registered in the control plane.",
			"Use the get subcommand to inspect one host in detail.",
		),
		Example: commandText(
			"jumpa hosts",
			"jumpa hosts --json",
			"jumpa hosts get host_123",
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSubApp(cmd.Context(), log, cmd, "jumpa-cli-hosts", cli.NewHostsModule(cli.ListOptions{
				JSON: jsonOutput,
			}))
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "print machine-readable JSON instead of a table")
	cmd.AddCommand(newHostGetCmd(log))
	return cmd
}

func newSessionsCmd(log *slog.Logger) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		GroupID: resourceCommandGroup,
		Use:     "sessions",
		Aliases: []string{"session"},
		Short:   "List bastion sessions",
		Long: commandText(
			"List bastion sessions recorded by the control plane.",
			"Use the get subcommand to inspect one session in detail.",
		),
		Example: commandText(
			"jumpa sessions",
			"jumpa sessions --json",
			"jumpa sessions get session_123",
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSubApp(cmd.Context(), log, cmd, "jumpa-cli-sessions", cli.NewSessionsModule(cli.ListOptions{
				JSON: jsonOutput,
			}))
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "print machine-readable JSON instead of a table")
	cmd.AddCommand(newSessionGetCmd(log))
	return cmd
}

func newGatewaysCmd(log *slog.Logger) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		GroupID: resourceCommandGroup,
		Use:     "gateways",
		Aliases: []string{"gateway"},
		Short:   "List registered gateways",
		Long: commandText(
			"List bastion gateway nodes registered with the control plane.",
			"Use the get subcommand to inspect one gateway in detail.",
		),
		Example: commandText(
			"jumpa gateways",
			"jumpa gateways --json",
			"jumpa gateways get gateway_123",
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSubApp(cmd.Context(), log, cmd, "jumpa-cli-gateways", cli.NewGatewaysModule(cli.ListOptions{
				JSON: jsonOutput,
			}))
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "print machine-readable JSON instead of a table")
	cmd.AddCommand(newGatewayGetCmd(log))
	return cmd
}

func newHostGetCmd(log *slog.Logger) *cobra.Command {
	options := cli.DetailOptions{}

	cmd := &cobra.Command{
		Use:   "get <host-id>",
		Short: "Show host details",
		Long:  "Show detailed metadata for a single bastion host.",
		Example: commandText(
			"jumpa hosts get host_123",
			"jumpa hosts get host_123 --json",
		),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			options.ID = args[0]
			return runSubApp(cmd.Context(), log, cmd, "jumpa-cli-host-get", cli.NewHostDetailModule(options))
		},
	}

	cmd.Flags().BoolVar(&options.JSON, "json", false, "print machine-readable JSON instead of a table")
	return cmd
}

func newSessionGetCmd(log *slog.Logger) *cobra.Command {
	options := cli.DetailOptions{}

	cmd := &cobra.Command{
		Use:   "get <session-id>",
		Short: "Show session details",
		Long:  "Show detailed metadata for a single bastion session.",
		Example: commandText(
			"jumpa sessions get session_123",
			"jumpa sessions get session_123 --json",
		),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			options.ID = args[0]
			return runSubApp(cmd.Context(), log, cmd, "jumpa-cli-session-get", cli.NewSessionDetailModule(options))
		},
	}

	cmd.Flags().BoolVar(&options.JSON, "json", false, "print machine-readable JSON instead of a table")
	return cmd
}

func newGatewayGetCmd(log *slog.Logger) *cobra.Command {
	options := cli.DetailOptions{}

	cmd := &cobra.Command{
		Use:   "get <gateway-id>",
		Short: "Show gateway details",
		Long:  "Show detailed metadata for a single registered bastion gateway.",
		Example: commandText(
			"jumpa gateways get gateway_123",
			"jumpa gateways get gateway_123 --json",
		),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			options.ID = args[0]
			return runSubApp(cmd.Context(), log, cmd, "jumpa-cli-gateway-get", cli.NewGatewayDetailModule(options))
		},
	}

	cmd.Flags().BoolVar(&options.JSON, "json", false, "print machine-readable JSON instead of a table")
	return cmd
}
