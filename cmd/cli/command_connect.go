package main

import (
	"log/slog"

	cli "github.com/DaiYuANg/jumpa/internal/cli"
	"github.com/spf13/cobra"
)

func newConnectCmd(log *slog.Logger, flags *rootFlags) *cobra.Command {
	connectFlags := struct {
		LocalForwards   []string
		RemoteForwards  []string
		DynamicForwards []string
	}{}

	cmd := &cobra.Command{
		GroupID: workflowCommandGroup,
		Use:     "connect <host> [account]",
		Short:   "Launch an SSH session through the Jumpa gateway",
		Long: commandText(
			"Authenticate with the Jumpa control plane, resolve the bastion",
			"gateway address, and launch an SSH session for the selected",
			"host and optional host account.",
			"",
			"Use -L/--local-forward with [bind_address:]port:host:hostport",
			"to expose remote services on the local machine while the SSH",
			"session is active.",
			"",
			"Use -R/--remote-forward with [bind_address:]port:host:hostport",
			"to expose local services to the remote side, and -D/--dynamic-forward",
			"with [bind_address:]port to start a local SOCKS5 proxy.",
		),
		Example: commandText(
			"jumpa connect prod-web-01",
			"jumpa connect prod-web-01 ubuntu",
			"jumpa connect prod-web-01 ubuntu --principal alice",
			"jumpa connect prod-web-01 --gateway 127.0.0.1:2222",
			"jumpa connect prod-db-01 --ssh-key ~/.ssh/id_ed25519",
			"jumpa connect prod-db-01 -L 15432:127.0.0.1:5432",
			"jumpa connect prod-web-01 -R 18080:127.0.0.1:8080",
			"jumpa connect prod-web-01 -D 1080",
		),
		Args:              cobra.RangeArgs(1, 2),
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			account := ""
			if len(args) > 1 {
				account = args[1]
			}

			localForwards, err := parseLocalForwards(connectFlags.LocalForwards)
			if err != nil {
				return err
			}
			remoteForwards, err := parseRemoteForwards(connectFlags.RemoteForwards)
			if err != nil {
				return err
			}
			dynamicForwards, err := parseDynamicForwards(connectFlags.DynamicForwards)
			if err != nil {
				return err
			}

			return runSubApp(cmd.Context(), log, flags.overrides(cmd), "jumpa-cli-connect", cli.NewConnectModule(cli.ConnectOptions{
				Host:            args[0],
				Account:         account,
				LocalForwards:   localForwards,
				RemoteForwards:  remoteForwards,
				DynamicForwards: dynamicForwards,
			}))
		},
	}

	cmd.Flags().StringArrayVarP(&connectFlags.LocalForwards, "local-forward", "L", nil, "local port forward in [bind_address:]port:host:hostport form")
	cmd.Flags().StringArrayVarP(&connectFlags.RemoteForwards, "remote-forward", "R", nil, "remote port forward in [bind_address:]port:host:hostport form")
	cmd.Flags().StringArrayVarP(&connectFlags.DynamicForwards, "dynamic-forward", "D", nil, "dynamic SOCKS5 forward in [bind_address:]port form")
	return cmd
}
