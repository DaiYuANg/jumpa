package sshclient

import (
	"fmt"
	"log/slog"
	"net"
)

func (c *Client) Launch(req Request) error {
	req = normalizeRequest(req)
	req, effectiveCfg, err := c.applySSHConfig(req)
	if err != nil {
		return err
	}
	if req.User == "" {
		return fmt.Errorf("ssh client requires a user")
	}
	if req.Host == "" {
		return fmt.Errorf("ssh client requires a host")
	}
	if req.Port == "" {
		return fmt.Errorf("ssh client requires a port")
	}
	client, cleanupClient, err := c.dialClient(req, effectiveCfg)
	if err != nil {
		return err
	}
	if cleanupClient != nil {
		defer cleanupClient()
	}

	if c.log != nil {
		c.log.Info("launching ssh session",
			slog.String("user", req.User),
			slog.String("gateway", net.JoinHostPort(trimBrackets(req.Host), req.Port)),
			slog.Int("proxy_hops", len(req.ProxyJumps)),
		)
	}

	stopForwards, err := c.startForwards(client, req)
	if err != nil {
		return err
	}
	if stopForwards != nil {
		defer stopForwards()
	}

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer func() { _ = session.Close() }()

	session.Stdin = c.stdin
	session.Stdout = c.stdout
	session.Stderr = c.stderr

	restore, err := c.prepareInteractiveSession(session, req.Terminal)
	if err != nil {
		return err
	}
	if restore != nil {
		defer restore()
	}

	if err := session.Shell(); err != nil {
		return err
	}
	return session.Wait()
}
