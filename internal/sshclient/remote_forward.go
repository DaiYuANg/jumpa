package sshclient

import (
	"fmt"
	"net"
	"strings"

	"golang.org/x/crypto/ssh"
)

const defaultRemoteForwardBindHost = "127.0.0.1"

type RemoteForward struct {
	BindHost  string
	BindPort  string
	LocalHost string
	LocalPort string
}

func ParseRemoteForward(spec string) (RemoteForward, error) {
	trimmed := strings.TrimSpace(spec)
	if trimmed == "" {
		return RemoteForward{}, fmt.Errorf("value is required")
	}

	parts := splitSSHAddressSpec(trimmed)
	switch len(parts) {
	case 3:
		return normalizeRemoteForward(RemoteForward{
			BindHost:  defaultRemoteForwardBindHost,
			BindPort:  parts[0],
			LocalHost: parts[1],
			LocalPort: parts[2],
		})
	case 4:
		return normalizeRemoteForward(RemoteForward{
			BindHost:  parts[0],
			BindPort:  parts[1],
			LocalHost: parts[2],
			LocalPort: parts[3],
		})
	default:
		return RemoteForward{}, fmt.Errorf("expected [bind_address:]port:host:hostport")
	}
}

func (c *Client) startRemoteForwards(client *ssh.Client, forwards []RemoteForward) (func(), error) {
	if len(forwards) == 0 {
		return nil, nil
	}

	listeners := make([]net.Listener, 0, len(forwards))
	done := make(chan struct{})

	for _, forward := range forwards {
		normalized, err := normalizeRemoteForward(forward)
		if err != nil {
			close(done)
			closeListeners(listeners)
			return nil, err
		}

		bindAddr := net.JoinHostPort(trimBrackets(normalized.BindHost), normalized.BindPort)
		listener, err := client.Listen("tcp", bindAddr)
		if err != nil {
			close(done)
			closeListeners(listeners)
			return nil, fmt.Errorf("listen on remote forward %s -> %s: %w", formatAddress(normalized.BindHost, normalized.BindPort), formatAddress(normalized.LocalHost, normalized.LocalPort), err)
		}

		listeners = append(listeners, listener)
		if c.log != nil {
			c.log.Info("started remote ssh forward",
				"remote", formatAddress(normalized.BindHost, normalized.BindPort),
				"local", formatAddress(normalized.LocalHost, normalized.LocalPort),
			)
		}

		go c.serveRemoteForward(done, listener, normalized)
	}

	return func() {
		close(done)
		closeListeners(listeners)
	}, nil
}

func (c *Client) serveRemoteForward(done <-chan struct{}, listener net.Listener, forward RemoteForward) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-done:
				return
			default:
			}

			if c.log != nil {
				c.log.Warn("accept remote ssh forward failed",
					"remote", formatAddress(forward.BindHost, forward.BindPort),
					"error", err.Error(),
				)
			}
			return
		}

		go c.handleRemoteForwardConn(conn, forward)
	}
}

func (c *Client) handleRemoteForwardConn(remoteConn net.Conn, forward RemoteForward) {
	localAddr := net.JoinHostPort(trimBrackets(forward.LocalHost), forward.LocalPort)
	localConn, err := net.Dial("tcp", localAddr)
	if err != nil {
		_ = remoteConn.Close()
		if c.log != nil {
			c.log.Warn("open local leg for remote ssh forward failed",
				"local", formatAddress(forward.LocalHost, forward.LocalPort),
				"error", err.Error(),
			)
		}
		return
	}

	proxyConnections(remoteConn, localConn)
}

func normalizeRemoteForward(forward RemoteForward) (RemoteForward, error) {
	normalized := RemoteForward{
		BindHost:  trimBrackets(strings.TrimSpace(forward.BindHost)),
		BindPort:  strings.TrimSpace(forward.BindPort),
		LocalHost: trimBrackets(strings.TrimSpace(forward.LocalHost)),
		LocalPort: strings.TrimSpace(forward.LocalPort),
	}

	if normalized.BindHost == "" {
		normalized.BindHost = defaultRemoteForwardBindHost
	}
	if normalized.LocalHost == "" {
		return RemoteForward{}, fmt.Errorf("remote forward requires a local host")
	}
	if err := validateTCPPort(normalized.BindPort); err != nil {
		return RemoteForward{}, fmt.Errorf("invalid remote forward bind port %q: %w", normalized.BindPort, err)
	}
	if err := validateTCPPort(normalized.LocalPort); err != nil {
		return RemoteForward{}, fmt.Errorf("invalid remote forward local port %q: %w", normalized.LocalPort, err)
	}

	return normalized, nil
}
