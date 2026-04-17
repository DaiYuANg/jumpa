package sshclient

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
)

const defaultLocalForwardHost = "127.0.0.1"

type LocalForward struct {
	LocalHost  string
	LocalPort  string
	RemoteHost string
	RemotePort string
}

func ParseLocalForward(spec string) (LocalForward, error) {
	trimmed := strings.TrimSpace(spec)
	if trimmed == "" {
		return LocalForward{}, fmt.Errorf("value is required")
	}

	parts := splitSSHAddressSpec(trimmed)
	switch len(parts) {
	case 3:
		return normalizeLocalForward(LocalForward{
			LocalHost:  defaultLocalForwardHost,
			LocalPort:  parts[0],
			RemoteHost: parts[1],
			RemotePort: parts[2],
		})
	case 4:
		return normalizeLocalForward(LocalForward{
			LocalHost:  parts[0],
			LocalPort:  parts[1],
			RemoteHost: parts[2],
			RemotePort: parts[3],
		})
	default:
		return LocalForward{}, fmt.Errorf("expected [bind_address:]port:host:hostport")
	}
}

func (c *Client) startLocalForwards(client *ssh.Client, forwards []LocalForward) (func(), error) {
	if len(forwards) == 0 {
		return nil, nil
	}

	listeners := make([]net.Listener, 0, len(forwards))
	done := make(chan struct{})

	for _, forward := range forwards {
		normalized, err := normalizeLocalForward(forward)
		if err != nil {
			close(done)
			closeListeners(listeners)
			return nil, err
		}

		localAddr := net.JoinHostPort(trimBrackets(normalized.LocalHost), normalized.LocalPort)
		listener, err := net.Listen("tcp", localAddr)
		if err != nil {
			close(done)
			closeListeners(listeners)
			return nil, fmt.Errorf("listen on local forward %s -> %s: %w", formatAddress(normalized.LocalHost, normalized.LocalPort), formatAddress(normalized.RemoteHost, normalized.RemotePort), err)
		}

		listeners = append(listeners, listener)
		if c.log != nil {
			c.log.Info("started local ssh forward",
				"local", formatAddress(normalized.LocalHost, normalized.LocalPort),
				"remote", formatAddress(normalized.RemoteHost, normalized.RemotePort),
			)
		}

		go c.serveLocalForward(done, client, listener, normalized)
	}

	return func() {
		close(done)
		closeListeners(listeners)
	}, nil
}

func (c *Client) startForwards(client *ssh.Client, req Request) (func(), error) {
	stops := make([]func(), 0, 3)

	stopLocal, err := c.startLocalForwards(client, req.LocalForwards)
	if err != nil {
		runClosers(stops)
		return nil, err
	}
	if stopLocal != nil {
		stops = append(stops, stopLocal)
	}

	stopRemote, err := c.startRemoteForwards(client, req.RemoteForwards)
	if err != nil {
		runClosers(stops)
		return nil, err
	}
	if stopRemote != nil {
		stops = append(stops, stopRemote)
	}

	stopDynamic, err := c.startDynamicForwards(client, req.DynamicForwards)
	if err != nil {
		runClosers(stops)
		return nil, err
	}
	if stopDynamic != nil {
		stops = append(stops, stopDynamic)
	}

	if len(stops) == 0 {
		return nil, nil
	}

	return func() {
		runClosers(stops)
	}, nil
}

func (c *Client) serveLocalForward(done <-chan struct{}, client *ssh.Client, listener net.Listener, forward LocalForward) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-done:
				return
			default:
			}

			if c.log != nil {
				c.log.Warn("accept local ssh forward failed",
					"local", formatAddress(forward.LocalHost, forward.LocalPort),
					"error", err.Error(),
				)
			}
			return
		}

		go c.handleLocalForwardConn(client, conn, forward)
	}
}

func (c *Client) handleLocalForwardConn(client *ssh.Client, localConn net.Conn, forward LocalForward) {
	remoteAddr := net.JoinHostPort(trimBrackets(forward.RemoteHost), forward.RemotePort)
	remoteConn, err := client.Dial("tcp", remoteAddr)
	if err != nil {
		_ = localConn.Close()
		if c.log != nil {
			c.log.Warn("open remote leg for local ssh forward failed",
				"remote", formatAddress(forward.RemoteHost, forward.RemotePort),
				"error", err.Error(),
			)
		}
		return
	}

	proxyConnections(localConn, remoteConn)
}

func normalizeLocalForward(forward LocalForward) (LocalForward, error) {
	normalized := LocalForward{
		LocalHost:  trimBrackets(strings.TrimSpace(forward.LocalHost)),
		LocalPort:  strings.TrimSpace(forward.LocalPort),
		RemoteHost: trimBrackets(strings.TrimSpace(forward.RemoteHost)),
		RemotePort: strings.TrimSpace(forward.RemotePort),
	}

	if normalized.LocalHost == "" {
		normalized.LocalHost = defaultLocalForwardHost
	}
	if normalized.RemoteHost == "" {
		return LocalForward{}, fmt.Errorf("local forward requires a remote host")
	}
	if err := validateTCPPort(normalized.LocalPort); err != nil {
		return LocalForward{}, fmt.Errorf("invalid local forward local port %q: %w", normalized.LocalPort, err)
	}
	if err := validateTCPPort(normalized.RemotePort); err != nil {
		return LocalForward{}, fmt.Errorf("invalid local forward remote port %q: %w", normalized.RemotePort, err)
	}

	return normalized, nil
}

func splitSSHAddressSpec(value string) []string {
	parts := make([]string, 0, 4)
	start := 0
	depth := 0

	for i, r := range value {
		switch r {
		case '[':
			depth++
		case ']':
			if depth > 0 {
				depth--
			}
		case ':':
			if depth == 0 {
				parts = append(parts, value[start:i])
				start = i + 1
			}
		}
	}

	parts = append(parts, value[start:])
	return parts
}

func validateTCPPort(value string) error {
	port, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("must be a valid TCP port")
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("must be between 1 and 65535")
	}
	return nil
}

func formatAddress(host, port string) string {
	if strings.TrimSpace(host) == "" {
		return ":" + strings.TrimSpace(port)
	}
	return net.JoinHostPort(trimBrackets(host), strings.TrimSpace(port))
}

func closeListeners(listeners []net.Listener) {
	for _, listener := range listeners {
		_ = listener.Close()
	}
}

func proxyConnections(left net.Conn, right net.Conn) {
	defer func() { _ = left.Close() }()
	defer func() { _ = right.Close() }()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		copyAndCloseWrite(right, left)
	}()
	go func() {
		defer wg.Done()
		copyAndCloseWrite(left, right)
	}()

	wg.Wait()
}

func copyAndCloseWrite(dst net.Conn, src net.Conn) {
	_, _ = io.Copy(dst, src)
	if closer, ok := dst.(interface{ CloseWrite() error }); ok {
		_ = closer.CloseWrite()
	}
}
