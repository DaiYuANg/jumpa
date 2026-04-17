package app

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

const (
	DefaultGatewayHost = "127.0.0.1"
	DefaultGatewayPort = "2222"
	DefaultSSHPort     = "22"
)

func ParseGatewayAddress(raw string) (string, string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return DefaultGatewayHost, DefaultGatewayPort, nil
	}

	if ip := net.ParseIP(trimBrackets(value)); ip != nil {
		return normalizeGatewayHost(value), DefaultSSHPort, nil
	}

	if strings.HasPrefix(value, ":") {
		port := strings.TrimPrefix(value, ":")
		if port == "" {
			return DefaultGatewayHost, DefaultSSHPort, nil
		}
		if err := validateGatewayPort(port); err != nil {
			return "", "", err
		}
		return DefaultGatewayHost, port, nil
	}

	host, port, err := net.SplitHostPort(value)
	if err == nil {
		if host == "" {
			host = DefaultGatewayHost
		}
		if port == "" {
			port = DefaultSSHPort
		}
		if err := validateGatewayPort(port); err != nil {
			return "", "", err
		}
		return normalizeGatewayHost(host), port, nil
	}

	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		host = trimBrackets(value)
		if host == "" {
			return "", "", fmt.Errorf("invalid gateway address %q: host is required", raw)
		}
		return normalizeGatewayHost(host), DefaultSSHPort, nil
	}

	if strings.HasSuffix(value, ":") {
		host = strings.TrimSuffix(value, ":")
		if host == "" {
			return DefaultGatewayHost, DefaultSSHPort, nil
		}
		return normalizeGatewayHost(host), DefaultSSHPort, nil
	}

	if strings.Contains(value, ":") && strings.Count(value, ":") > 1 {
		return "", "", fmt.Errorf("invalid gateway address %q: wrap IPv6 addresses with a port as [addr]:port", raw)
	}

	return normalizeGatewayHost(value), DefaultSSHPort, nil
}

func SplitGatewayAddress(raw string) (string, string) {
	host, port, err := ParseGatewayAddress(raw)
	if err != nil {
		return DefaultGatewayHost, DefaultGatewayPort
	}
	return host, port
}

func NormalizeGatewayAddress(raw string) (string, error) {
	host, port, err := ParseGatewayAddress(raw)
	if err != nil {
		return "", err
	}
	return FormatGatewayAddress(host, port), nil
}

func FormatGatewayAddress(host, port string) string {
	rawHost := trimBrackets(strings.TrimSpace(host))
	if rawHost == "" {
		rawHost = DefaultGatewayHost
	}

	valuePort := strings.TrimSpace(port)
	if valuePort == "" {
		valuePort = DefaultSSHPort
	}

	return net.JoinHostPort(rawHost, valuePort)
}

func normalizeGatewayHost(host string) string {
	value := trimBrackets(strings.TrimSpace(host))
	if value == "" {
		value = DefaultGatewayHost
	}
	if strings.Contains(value, ":") {
		return "[" + value + "]"
	}
	return value
}

func trimBrackets(value string) string {
	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		return strings.TrimSuffix(strings.TrimPrefix(value, "["), "]")
	}
	return value
}

func validateGatewayPort(port string) error {
	number, err := strconv.Atoi(strings.TrimSpace(port))
	if err != nil || number < 1 || number > 65535 {
		return fmt.Errorf("invalid gateway port %q", port)
	}
	return nil
}
