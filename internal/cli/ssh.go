package cli

import (
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"strings"

	cliapp "github.com/DaiYuANg/jumpa/internal/cli/app"
)

type SSHLauncher struct {
	cfg    Config
	log    *slog.Logger
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func NewSSHLauncher(cfg Config, log *slog.Logger, streams stdio) *SSHLauncher {
	return &SSHLauncher{
		cfg:    cfg,
		log:    log,
		stdin:  streams.In,
		stdout: streams.Out,
		stderr: streams.Err,
	}
}

func (l *SSHLauncher) Launch(req cliapp.LaunchRequest) error {
	target := fmt.Sprintf("%s@%s", req.Target, req.GatewayHost)
	binary := StringOverride(l.cfg.SSHBinary).OrElse("ssh")
	cmd := exec.Command(binary, "-p", req.GatewayPort, target)
	cmd.Stdin = l.stdin
	cmd.Stdout = l.stdout
	cmd.Stderr = l.stderr

	l.log.Info("launching ssh session",
		slog.String("binary", binary),
		slog.String("target", target),
		slog.String("port", req.GatewayPort),
	)

	return cmd.Run()
}

func BuildLaunchRequest(principal, gatewayAddr, hostName, account string) cliapp.LaunchRequest {
	targetParts := []string{strings.TrimSpace(principal), strings.TrimSpace(hostName)}
	if accountValue := strings.TrimSpace(account); accountValue != "" {
		targetParts = append(targetParts, accountValue)
	}

	gatewayHost, gatewayPort := splitGatewayAddress(gatewayAddr)
	return cliapp.LaunchRequest{
		Target:      strings.Join(targetParts, "#"),
		GatewayHost: gatewayHost,
		GatewayPort: gatewayPort,
	}
}

func splitGatewayAddress(raw string) (string, string) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "127.0.0.1", "2222"
	}
	parts := strings.Split(value, ":")
	if len(parts) == 1 {
		return parts[0], "22"
	}
	host := strings.Join(parts[:len(parts)-1], ":")
	port := parts[len(parts)-1]
	if host == "" {
		host = "127.0.0.1"
	}
	if port == "" {
		port = "22"
	}
	return host, port
}
