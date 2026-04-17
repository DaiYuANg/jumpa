package cli

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	cliapp "github.com/DaiYuANg/jumpa/internal/cli/app"
	"github.com/DaiYuANg/jumpa/internal/sshclient"
)

type SSHLauncher struct {
	cfg    Config
	client *sshclient.Client
}

type SSHLaunchOptions struct {
	Password        string
	LocalForwards   []sshclient.LocalForward
	RemoteForwards  []sshclient.RemoteForward
	DynamicForwards []sshclient.DynamicForward
}

func NewSSHLauncher(cfg Config, log *slog.Logger, streams stdio) *SSHLauncher {
	return &SSHLauncher{
		cfg: cfg,
		client: sshclient.New(sshclient.Config{
			HostKeyPolicy:  cfg.SSHHostKeyPolicy,
			KnownHostsPath: cfg.SSHKnownHostsPath,
			ConfigPath:     cfg.SSHConfigPath,
			ConnectTimeout: 15 * time.Second,
		}, log, sshclient.Streams{
			In:  streams.In,
			Out: streams.Out,
			Err: streams.Err,
		}),
	}
}

func (l *SSHLauncher) Launch(req cliapp.LaunchRequest, options SSHLaunchOptions) error {
	request := sshclient.Request{
		User:            req.Target,
		Host:            req.GatewayHost,
		Port:            req.GatewayPort,
		Password:        options.Password,
		LocalForwards:   options.LocalForwards,
		RemoteForwards:  options.RemoteForwards,
		DynamicForwards: options.DynamicForwards,
	}

	if key := resolveSSHPrivateKey(l.cfg); key != nil {
		request.PrivateKey = key
	}

	agentSocket, err := resolveSSHAgentSocket(l.cfg)
	if err != nil {
		return err
	}
	request.AgentSocket = agentSocket

	return l.client.Launch(request)
}

func BuildLaunchRequest(principal, gatewayAddr, hostName, account string) (cliapp.LaunchRequest, error) {
	targetParts := []string{strings.TrimSpace(principal), strings.TrimSpace(hostName)}
	if accountValue := strings.TrimSpace(account); accountValue != "" {
		targetParts = append(targetParts, accountValue)
	}

	gatewayHost, gatewayPort, err := cliapp.ParseGatewayAddress(gatewayAddr)
	if err != nil {
		return cliapp.LaunchRequest{}, err
	}

	return cliapp.LaunchRequest{
		Target:      strings.Join(targetParts, "#"),
		GatewayHost: gatewayHost,
		GatewayPort: gatewayPort,
	}, nil
}

func resolveSSHPrivateKey(cfg Config) *sshclient.PrivateKey {
	path := strings.TrimSpace(cfg.SSHPrivateKeyPath)
	if path == "" {
		return nil
	}

	return &sshclient.PrivateKey{
		Path:       path,
		Passphrase: strings.TrimSpace(cfg.SSHPrivateKeyPassphrase),
	}
}

func resolveSSHAgentSocket(cfg Config) (string, error) {
	if !cfg.SSHAgentEnabled {
		return "", nil
	}

	if socket := strings.TrimSpace(cfg.SSHAgentSocket); socket != "" {
		return socket, nil
	}

	if socket := strings.TrimSpace(os.Getenv("SSH_AUTH_SOCK")); socket != "" {
		return socket, nil
	}

	return "", fmt.Errorf("ssh agent is enabled but no socket was configured and SSH_AUTH_SOCK is empty")
}
