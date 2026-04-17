package sshclient

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"

	"golang.org/x/crypto/ssh"
)

const socksVersion5 = 0x05

type DynamicForward struct {
	LocalHost string
	LocalPort string
}

func ParseDynamicForward(spec string) (DynamicForward, error) {
	trimmed := strings.TrimSpace(spec)
	if trimmed == "" {
		return DynamicForward{}, fmt.Errorf("value is required")
	}

	parts := splitSSHAddressSpec(trimmed)
	switch len(parts) {
	case 1:
		return normalizeDynamicForward(DynamicForward{
			LocalHost: defaultLocalForwardHost,
			LocalPort: parts[0],
		})
	case 2:
		return normalizeDynamicForward(DynamicForward{
			LocalHost: parts[0],
			LocalPort: parts[1],
		})
	default:
		return DynamicForward{}, fmt.Errorf("expected [bind_address:]port")
	}
}

func (c *Client) startDynamicForwards(client *ssh.Client, forwards []DynamicForward) (func(), error) {
	if len(forwards) == 0 {
		return nil, nil
	}

	listeners := make([]net.Listener, 0, len(forwards))
	done := make(chan struct{})

	for _, forward := range forwards {
		normalized, err := normalizeDynamicForward(forward)
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
			return nil, fmt.Errorf("listen on dynamic forward %s: %w", formatAddress(normalized.LocalHost, normalized.LocalPort), err)
		}

		listeners = append(listeners, listener)
		if c.log != nil {
			c.log.Info("started dynamic ssh forward",
				"local", formatAddress(normalized.LocalHost, normalized.LocalPort),
			)
		}

		go c.serveDynamicForward(done, client, listener, normalized)
	}

	return func() {
		close(done)
		closeListeners(listeners)
	}, nil
}

func (c *Client) serveDynamicForward(done <-chan struct{}, client *ssh.Client, listener net.Listener, forward DynamicForward) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-done:
				return
			default:
			}

			if c.log != nil {
				c.log.Warn("accept dynamic ssh forward failed",
					"local", formatAddress(forward.LocalHost, forward.LocalPort),
					"error", err.Error(),
				)
			}
			return
		}

		go c.handleDynamicForwardConn(client, conn, forward)
	}
}

func (c *Client) handleDynamicForwardConn(client *ssh.Client, localConn net.Conn, forward DynamicForward) {
	defer func() { _ = localConn.Close() }()

	targetAddr, err := negotiateSOCKS5(localConn)
	if err != nil {
		if c.log != nil {
			c.log.Warn("dynamic ssh forward handshake failed",
				"local", formatAddress(forward.LocalHost, forward.LocalPort),
				"error", err.Error(),
			)
		}
		return
	}

	remoteConn, err := client.Dial("tcp", targetAddr)
	if err != nil {
		_ = writeSOCKS5Reply(localConn, 0x05)
		if c.log != nil {
			c.log.Warn("open remote leg for dynamic ssh forward failed",
				"target", targetAddr,
				"error", err.Error(),
			)
		}
		return
	}
	if err := writeSOCKS5Reply(localConn, 0x00); err != nil {
		_ = remoteConn.Close()
		if c.log != nil {
			c.log.Warn("write dynamic ssh forward success reply failed",
				"target", targetAddr,
				"error", err.Error(),
			)
		}
		return
	}

	proxyConnections(localConn, remoteConn)
}

func normalizeDynamicForward(forward DynamicForward) (DynamicForward, error) {
	normalized := DynamicForward{
		LocalHost: trimBrackets(strings.TrimSpace(forward.LocalHost)),
		LocalPort: strings.TrimSpace(forward.LocalPort),
	}

	if normalized.LocalHost == "" {
		normalized.LocalHost = defaultLocalForwardHost
	}
	if err := validateTCPPort(normalized.LocalPort); err != nil {
		return DynamicForward{}, fmt.Errorf("invalid dynamic forward local port %q: %w", normalized.LocalPort, err)
	}
	return normalized, nil
}

func negotiateSOCKS5(conn net.Conn) (string, error) {
	header := make([]byte, 2)
	if _, err := io.ReadFull(conn, header); err != nil {
		return "", fmt.Errorf("read socks greeting: %w", err)
	}
	if header[0] != socksVersion5 {
		return "", fmt.Errorf("unsupported socks version %d", header[0])
	}

	methods := make([]byte, int(header[1]))
	if _, err := io.ReadFull(conn, methods); err != nil {
		return "", fmt.Errorf("read socks methods: %w", err)
	}

	if !supportsSOCKSNoAuth(methods) {
		_, _ = conn.Write([]byte{socksVersion5, 0xFF})
		return "", fmt.Errorf("no supported socks auth methods")
	}

	if _, err := conn.Write([]byte{socksVersion5, 0x00}); err != nil {
		return "", fmt.Errorf("write socks method selection: %w", err)
	}

	requestHeader := make([]byte, 4)
	if _, err := io.ReadFull(conn, requestHeader); err != nil {
		return "", fmt.Errorf("read socks request: %w", err)
	}
	if requestHeader[0] != socksVersion5 {
		return "", fmt.Errorf("unsupported socks request version %d", requestHeader[0])
	}
	if requestHeader[1] != 0x01 {
		_ = writeSOCKS5Reply(conn, 0x07)
		return "", fmt.Errorf("unsupported socks command %d", requestHeader[1])
	}

	host, err := readSOCKS5Address(conn, requestHeader[3])
	if err != nil {
		_ = writeSOCKS5Reply(conn, 0x08)
		return "", err
	}

	portBytes := make([]byte, 2)
	if _, err := io.ReadFull(conn, portBytes); err != nil {
		return "", fmt.Errorf("read socks target port: %w", err)
	}

	return net.JoinHostPort(trimBrackets(host), fmt.Sprintf("%d", binary.BigEndian.Uint16(portBytes))), nil
}

func supportsSOCKSNoAuth(methods []byte) bool {
	for _, method := range methods {
		if method == 0x00 {
			return true
		}
	}
	return false
}

func readSOCKS5Address(conn net.Conn, atyp byte) (string, error) {
	switch atyp {
	case 0x01:
		buf := make([]byte, 4)
		if _, err := io.ReadFull(conn, buf); err != nil {
			return "", fmt.Errorf("read socks ipv4 target: %w", err)
		}
		return net.IP(buf).String(), nil
	case 0x03:
		size := make([]byte, 1)
		if _, err := io.ReadFull(conn, size); err != nil {
			return "", fmt.Errorf("read socks domain length: %w", err)
		}
		buf := make([]byte, int(size[0]))
		if _, err := io.ReadFull(conn, buf); err != nil {
			return "", fmt.Errorf("read socks domain target: %w", err)
		}
		return string(buf), nil
	case 0x04:
		buf := make([]byte, 16)
		if _, err := io.ReadFull(conn, buf); err != nil {
			return "", fmt.Errorf("read socks ipv6 target: %w", err)
		}
		return net.IP(buf).String(), nil
	default:
		return "", fmt.Errorf("unsupported socks address type %d", atyp)
	}
}

func writeSOCKS5Reply(conn net.Conn, status byte) error {
	_, err := conn.Write([]byte{socksVersion5, status, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
	if err != nil {
		return fmt.Errorf("write socks reply: %w", err)
	}
	return nil
}
