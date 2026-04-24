package main

import (
	"log/slog"

	cli "github.com/DaiYuANg/jumpa/internal/cli"
	"github.com/spf13/cobra"
)

func newUICmd(log *slog.Logger) *cobra.Command {
	return &cobra.Command{
		GroupID: workflowCommandGroup,
		Use:     "ui",
		Short:   "Launch the interactive terminal UI",
		Long: commandText(
			"Launch the Bubble Tea based terminal UI for browsing hosts,",
			"sessions, access requests, and gateways from a single screen.",
		),
		Example: commandText(
			"jumpa ui",
			"jumpa ui --alt-screen=false",
			"jumpa ui --api http://127.0.0.1:8080",
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSubApp(cmd.Context(), log, cmd, "jumpa-cli-ui", cli.NewUIModule())
		},
	}
}
