package sshclient

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type ProxyJump struct {
	User string
	Host string
	Port string
}

func (c *Client) dialClient(req Request, cfg Config) (*ssh.Client, func(), error) {
	if len(req.ProxyJumps) == 0 {
		return c.dialSSHClient(req, cfg, nil)
	}

	cleanups := make([]func(), 0, len(req.ProxyJumps)+1)
	var dialer sshDialer

	for i, jump := range req.ProxyJumps {
		jumpReq, jumpCfg, err := c.resolveProxyJumpRequest(jump, req)
		if err != nil {
			runClosers(cleanups)
			return nil, nil, fmt.Errorf("resolve proxy jump %d: %w", i+1, err)
		}

		client, cleanup, err := c.dialSSHClient(jumpReq, jumpCfg, dialer)
		if err != nil {
			runClosers(cleanups)
			return nil, nil, fmt.Errorf("dial proxy jump %d (%s): %w", i+1, jumpReq.Host, err)
		}
		cleanups = append(cleanups, cleanup)
		dialer = client
	}

	client, cleanup, err := c.dialSSHClient(req, cfg, dialer)
	if err != nil {
		runClosers(cleanups)
		return nil, nil, err
	}
	cleanups = append(cleanups, cleanup)

	return client, func() {
		runClosers(cleanups)
	}, nil
}

func (c *Client) resolveProxyJumpRequest(jump ProxyJump, base Request) (Request, Config, error) {
	req := Request{
		User:        strings.TrimSpace(jump.User),
		Host:        strings.TrimSpace(jump.Host),
		Port:        strings.TrimSpace(jump.Port),
		Password:    base.Password,
		PrivateKey:  base.PrivateKey,
		AgentSocket: base.AgentSocket,
	}

	req, cfg, err := c.applySSHConfig(req)
	if err != nil {
		return Request{}, Config{}, err
	}
	if req.User == "" {
		req.User = defaultSSHUser()
	}
	if req.User == "" {
		return Request{}, Config{}, fmt.Errorf("proxy jump user is required for host %q", jump.Host)
	}
	if req.Host == "" {
		return Request{}, Config{}, fmt.Errorf("proxy jump host is required")
	}
	if req.Port == "" {
		req.Port = defaultSSHPort
	}
	req.ProxyJumps = nil

	return normalizeRequest(req), cfg, nil
}

func (c *Client) dialSSHClient(req Request, cfg Config, dialer sshDialer) (*ssh.Client, func(), error) {
	callback, err := hostKeyCallback(cfg)
	if err != nil {
		return nil, nil, err
	}

	authMethods, cleanupAuth, err := c.authMethods(req)
	if err != nil {
		return nil, nil, err
	}

	addr := net.JoinHostPort(trimBrackets(req.Host), req.Port)
	conn, err := dialNetwork(addr, cfg.ConnectTimeout, dialer)
	if err != nil {
		if cleanupAuth != nil {
			cleanupAuth()
		}
		return nil, nil, err
	}

	config := &ssh.ClientConfig{
		User:            req.User,
		Auth:            authMethods,
		HostKeyCallback: callback,
		Timeout:         cfg.ConnectTimeout,
	}

	if c.log != nil {
		c.log.Info("dialing ssh connection",
			"user", req.User,
			"host", addr,
			"via_proxy", dialer != nil,
		)
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		_ = conn.Close()
		if cleanupAuth != nil {
			cleanupAuth()
		}
		return nil, nil, err
	}

	client := ssh.NewClient(sshConn, chans, reqs)
	return client, func() {
		_ = client.Close()
		if cleanupAuth != nil {
			cleanupAuth()
		}
	}, nil
}

type sshDialer interface {
	Dial(network, addr string) (net.Conn, error)
}

func dialNetwork(addr string, timeout time.Duration, dialer sshDialer) (net.Conn, error) {
	if dialer != nil {
		return dialer.Dial("tcp", addr)
	}
	return net.DialTimeout("tcp", addr, timeout)
}

func defaultSSHUser() string {
	for _, name := range []string{"USER", "LOGNAME"} {
		if value := strings.TrimSpace(os.Getenv(name)); value != "" {
			return value
		}
	}
	return ""
}

func normalizeProxyJumps(jumps []ProxyJump) []ProxyJump {
	if len(jumps) == 0 {
		return nil
	}

	normalized := make([]ProxyJump, 0, len(jumps))
	for _, jump := range jumps {
		host := strings.TrimSpace(jump.Host)
		if host == "" {
			continue
		}
		normalized = append(normalized, ProxyJump{
			User: strings.TrimSpace(jump.User),
			Host: host,
			Port: strings.TrimSpace(jump.Port),
		})
	}

	if len(normalized) == 0 {
		return nil
	}
	return normalized
}
