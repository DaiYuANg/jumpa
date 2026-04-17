package main

import (
	"log/slog"
	"strings"

	cli "github.com/DaiYuANg/jumpa/internal/cli"
	"github.com/spf13/cobra"
)

func newRequestsCmd(log *slog.Logger, flags *rootFlags) *cobra.Command {
	options := cli.ListOptions{
		Page:     1,
		PageSize: 50,
	}

	cmd := &cobra.Command{
		GroupID: reviewCommandGroup,
		Use:     "requests",
		Aliases: []string{"request"},
		Short:   "List and review access requests",
		Long: commandText(
			"List access requests created by bastion policy evaluation.",
			"Use approve or reject to review an individual request.",
		),
		Example: commandText(
			"jumpa requests",
			"jumpa requests --status pending --page 1 --page-size 20",
			"jumpa requests approve req_123 --comment approved",
			"jumpa requests reject req_123 --comment missing-ticket",
		),
		Args:    cobra.NoArgs,
		PreRunE: func(_ *cobra.Command, _ []string) error { return validateRequestListOptions(&options) },
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSubApp(cmd.Context(), log, flags.overrides(cmd), "jumpa-cli-requests", cli.NewRequestsModule(options))
		},
	}

	cmd.Flags().BoolVar(&options.JSON, "json", false, "print machine-readable JSON instead of a table")
	cmd.Flags().StringVar(&options.Status, "status", "", "filter by request status")
	cmd.Flags().IntVar(&options.Page, "page", 1, "page number")
	cmd.Flags().IntVar(&options.PageSize, "page-size", 50, "page size")
	if err := cmd.RegisterFlagCompletionFunc("status", completeAccessRequestStatus); err != nil {
		panic(err)
	}
	cmd.AddCommand(
		newRequestReviewCmd(log, flags, false),
		newRequestReviewCmd(log, flags, true),
	)
	return cmd
}

func newRequestReviewCmd(log *slog.Logger, flags *rootFlags, reject bool) *cobra.Command {
	options := cli.AccessRequestReviewOptions{Reject: reject}
	use := "approve <request-id>"
	short := "Approve an access request"
	long := commandText(
		"Approve an access request and optionally record a review comment.",
		"When --reviewer is omitted, the current login email is used.",
	)
	example := commandText(
		"jumpa requests approve req_123",
		"jumpa requests approve req_123 --comment approved",
		"jumpa requests approve req_123 --reviewer alice@example.com",
	)
	appName := "jumpa-cli-requests-approve"
	if reject {
		use = "reject <request-id>"
		short = "Reject an access request"
		long = commandText(
			"Reject an access request and optionally record a review comment.",
			"When --reviewer is omitted, the current login email is used.",
		)
		example = commandText(
			"jumpa requests reject req_123",
			"jumpa requests reject req_123 --comment missing-ticket",
			"jumpa requests reject req_123 --reviewer alice@example.com",
		)
		appName = "jumpa-cli-requests-reject"
	}

	cmd := &cobra.Command{
		Use:               use,
		Short:             short,
		Long:              long,
		Example:           example,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			options.ID = args[0]
			return runSubApp(cmd.Context(), log, flags.overrides(cmd), appName, cli.NewAccessRequestReviewModule(options))
		},
	}

	cmd.Flags().StringVar(&options.Reviewer, "reviewer", "", "reviewer identity; defaults to the current login email")
	cmd.Flags().StringVar(&options.Comment, "comment", "", "optional review comment")
	cmd.Flags().BoolVar(&options.JSON, "json", false, "print machine-readable JSON instead of a table")
	return cmd
}

func completeAccessRequestStatus(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	matches := make([]string, 0, len(accessRequestStatusValues))
	for _, status := range accessRequestStatusValues {
		if strings.HasPrefix(status, strings.ToLower(toComplete)) {
			matches = append(matches, status)
		}
	}
	return matches, cobra.ShellCompDirectiveNoFileComp
}
