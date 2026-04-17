package sshclient

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func hostKeyCallback(cfg Config) (ssh.HostKeyCallback, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.HostKeyPolicy)) {
	case HostKeyPolicyInsecure:
		return ssh.InsecureIgnoreHostKey(), nil
	case "", HostKeyPolicyKnownHosts:
		path, err := knownHostsPath(cfg)
		if err != nil {
			return nil, err
		}
		return knownhosts.New(path)
	default:
		return nil, fmt.Errorf("unsupported ssh host key policy %q", cfg.HostKeyPolicy)
	}
}

func knownHostsPath(cfg Config) (string, error) {
	if trimmed := strings.TrimSpace(cfg.KnownHostsPath); trimmed != "" {
		return trimmed, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve known_hosts path: %w", err)
	}
	return filepath.Join(home, ".ssh", "known_hosts"), nil
}
