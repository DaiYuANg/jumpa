package sshclient

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type PrivateKey struct {
	Path       string
	PEM        []byte
	Passphrase string
}

func (c *Client) authMethods(req Request) ([]ssh.AuthMethod, func(), error) {
	methods := make([]ssh.AuthMethod, 0, 4)
	cleanup := make([]func(), 0, 1)

	if key := normalizePrivateKey(req.PrivateKey); key != nil {
		signer, err := loadPrivateKeySigner(*key)
		if err != nil {
			runClosers(cleanup)
			return nil, nil, err
		}
		methods = append(methods, ssh.PublicKeys(signer))
	}

	methods = append(methods, passwordAuthMethods(req.Password)...)

	if req.AgentSocket != "" {
		conn, err := net.Dial("unix", req.AgentSocket)
		if err != nil {
			runClosers(cleanup)
			return nil, nil, fmt.Errorf("connect ssh agent %q: %w", req.AgentSocket, err)
		}

		methods = append(methods, ssh.PublicKeysCallback(agent.NewClient(conn).Signers))
		cleanup = append(cleanup, func() {
			_ = conn.Close()
		})
	}

	if len(methods) == 0 {
		runClosers(cleanup)
		return nil, nil, fmt.Errorf("ssh client requires at least one auth method")
	}

	return methods, func() {
		runClosers(cleanup)
	}, nil
}

func loadPrivateKeySigner(key PrivateKey) (ssh.Signer, error) {
	raw, source, err := privateKeyBytes(key)
	if err != nil {
		return nil, err
	}

	passphrase := strings.TrimSpace(key.Passphrase)
	if passphrase != "" {
		signer, parseErr := ssh.ParsePrivateKeyWithPassphrase(raw, []byte(passphrase))
		if parseErr != nil {
			return nil, fmt.Errorf("parse private key %s: %w", source, parseErr)
		}
		return signer, nil
	}

	signer, parseErr := ssh.ParsePrivateKey(raw)
	if parseErr == nil {
		return signer, nil
	}

	var missing *ssh.PassphraseMissingError
	if errors.As(parseErr, &missing) {
		return nil, fmt.Errorf("private key %s requires a passphrase", source)
	}
	return nil, fmt.Errorf("parse private key %s: %w", source, parseErr)
}

func privateKeyBytes(key PrivateKey) ([]byte, string, error) {
	if raw := bytes.TrimSpace(key.PEM); len(raw) > 0 {
		return raw, "inline key", nil
	}

	path := strings.TrimSpace(key.Path)
	if path == "" {
		return nil, "", fmt.Errorf("private key requires a path or PEM payload")
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("read private key %q: %w", path, err)
	}
	return raw, path, nil
}

func normalizePrivateKey(key *PrivateKey) *PrivateKey {
	if key == nil {
		return nil
	}

	normalized := &PrivateKey{
		Path:       strings.TrimSpace(key.Path),
		PEM:        bytes.TrimSpace(key.PEM),
		Passphrase: strings.TrimSpace(key.Passphrase),
	}
	if normalized.Path == "" && len(normalized.PEM) == 0 {
		return nil
	}
	return normalized
}

func passwordAuthMethods(password string) []ssh.AuthMethod {
	trimmed := strings.TrimSpace(password)
	if trimmed == "" {
		return nil
	}

	return []ssh.AuthMethod{
		ssh.Password(trimmed),
		ssh.KeyboardInteractive(func(_ string, _ string, questions []string, _ []bool) ([]string, error) {
			answers := make([]string, len(questions))
			for i := range questions {
				answers[i] = trimmed
			}
			return answers, nil
		}),
	}
}

func runClosers(closers []func()) {
	for i := len(closers) - 1; i >= 0; i-- {
		closers[i]()
	}
}
