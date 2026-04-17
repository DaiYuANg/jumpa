package sshclient

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const sshConfigEnvAgentSocket = "SSH_AUTH_SOCK"

type sshConfigProfile struct {
	User                  string
	HostName              string
	Port                  string
	IdentityFile          string
	IdentityAgent         string
	UserKnownHostsFile    string
	StrictHostKeyChecking string
	ProxyJump             string
}

func (c *Client) applySSHConfig(req Request) (Request, Config, error) {
	effectiveCfg := c.cfg
	hostAlias := trimBrackets(req.Host)
	if hostAlias == "" {
		return req, effectiveCfg, nil
	}

	profile, err := loadSSHConfigProfile(hostAlias, c.cfg.ConfigPath)
	if err != nil {
		return req, effectiveCfg, err
	}

	if profile.HostName != "" {
		req.Host = profile.HostName
	}
	if req.User == "" && profile.User != "" {
		req.User = profile.User
	}
	if profile.Port != "" && shouldApplyConfigPort(req.Port) {
		req.Port = profile.Port
	}
	if req.PrivateKey == nil && profile.IdentityFile != "" {
		req.PrivateKey = &PrivateKey{Path: profile.IdentityFile}
	}
	if req.AgentSocket == "" {
		switch strings.ToLower(strings.TrimSpace(profile.IdentityAgent)) {
		case "", "none":
		case sshConfigEnvAgentSocket:
			req.AgentSocket = strings.TrimSpace(os.Getenv(sshConfigEnvAgentSocket))
		default:
			req.AgentSocket = profile.IdentityAgent
		}
	}

	switch strings.ToLower(strings.TrimSpace(profile.StrictHostKeyChecking)) {
	case "no", "off", "false":
		effectiveCfg.HostKeyPolicy = HostKeyPolicyInsecure
	case "", "yes", "on", "true", "ask", "accept-new":
	default:
		return req, effectiveCfg, fmt.Errorf("unsupported StrictHostKeyChecking value %q in ssh config", profile.StrictHostKeyChecking)
	}

	if profile.UserKnownHostsFile != "" {
		effectiveCfg.KnownHostsPath = profile.UserKnownHostsFile
	}
	if len(req.ProxyJumps) == 0 {
		proxyJumps, configured, err := parseProxyJumpList(profile.ProxyJump)
		if err != nil {
			return req, effectiveCfg, err
		}
		if configured {
			req.ProxyJumps = proxyJumps
		}
	}

	return normalizeRequest(req), applyDefaults(effectiveCfg), nil
}

func loadSSHConfigProfile(host, overridePath string) (sshConfigProfile, error) {
	configPath, err := resolveSSHConfigPath(overridePath)
	if err != nil {
		return sshConfigProfile{}, err
	}

	raw, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return sshConfigProfile{}, nil
		}
		return sshConfigProfile{}, fmt.Errorf("read ssh config %q: %w", configPath, err)
	}

	profile, err := parseSSHConfigProfile(string(raw), host)
	if err != nil {
		return sshConfigProfile{}, fmt.Errorf("parse ssh config %q: %w", configPath, err)
	}
	return profile, nil
}

func resolveSSHConfigPath(overridePath string) (string, error) {
	if trimmed := strings.TrimSpace(overridePath); trimmed != "" {
		return expandSSHPath(trimmed)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve ssh config path: %w", err)
	}
	return filepath.Join(home, ".ssh", "config"), nil
}

func parseSSHConfigProfile(content, host string) (sshConfigProfile, error) {
	profile := sshConfigProfile{}
	currentPatterns := []string{"*"}
	scanner := bufio.NewScanner(strings.NewReader(content))

	for lineNo := 1; scanner.Scan(); lineNo++ {
		line := stripSSHConfigComment(scanner.Text())
		if strings.TrimSpace(line) == "" {
			continue
		}

		key, value, ok := parseSSHConfigKV(line)
		if !ok {
			return sshConfigProfile{}, fmt.Errorf("line %d: invalid config directive", lineNo)
		}

		if strings.EqualFold(key, "Host") {
			currentPatterns = strings.Fields(value)
			continue
		}
		if !matchSSHConfigPatterns(host, currentPatterns) {
			continue
		}

		switch strings.ToLower(key) {
		case "user":
			if profile.User == "" {
				profile.User = strings.TrimSpace(stripSSHConfigQuotes(value))
			}
		case "hostname":
			if profile.HostName == "" {
				profile.HostName = trimBrackets(strings.TrimSpace(stripSSHConfigQuotes(value)))
			}
		case "port":
			if profile.Port == "" {
				profile.Port = strings.TrimSpace(stripSSHConfigQuotes(value))
			}
		case "identityfile":
			if profile.IdentityFile == "" {
				expanded, err := expandSSHPath(stripSSHConfigQuotes(value))
				if err != nil {
					return sshConfigProfile{}, fmt.Errorf("line %d: %w", lineNo, err)
				}
				profile.IdentityFile = expanded
			}
		case "identityagent":
			if profile.IdentityAgent == "" {
				agentValue, err := expandSSHConfigAgent(stripSSHConfigQuotes(value))
				if err != nil {
					return sshConfigProfile{}, fmt.Errorf("line %d: %w", lineNo, err)
				}
				profile.IdentityAgent = agentValue
			}
		case "userknownhostsfile":
			if profile.UserKnownHostsFile == "" {
				firstPath := firstSSHConfigField(value)
				if firstPath != "" {
					expanded, err := expandSSHPath(stripSSHConfigQuotes(firstPath))
					if err != nil {
						return sshConfigProfile{}, fmt.Errorf("line %d: %w", lineNo, err)
					}
					profile.UserKnownHostsFile = expanded
				}
			}
		case "stricthostkeychecking":
			if profile.StrictHostKeyChecking == "" {
				profile.StrictHostKeyChecking = strings.ToLower(strings.TrimSpace(stripSSHConfigQuotes(value)))
			}
		case "proxyjump":
			if profile.ProxyJump == "" {
				profile.ProxyJump = strings.TrimSpace(stripSSHConfigQuotes(value))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return sshConfigProfile{}, err
	}
	return profile, nil
}

func parseSSHConfigKV(line string) (string, string, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return "", "", false
	}

	keyEnd := strings.IndexAny(trimmed, " \t=")
	if keyEnd <= 0 {
		return "", "", false
	}

	key := strings.TrimSpace(trimmed[:keyEnd])
	rest := strings.TrimSpace(trimmed[keyEnd:])
	if strings.HasPrefix(rest, "=") {
		rest = strings.TrimSpace(strings.TrimPrefix(rest, "="))
	}
	if key == "" || rest == "" {
		return "", "", false
	}
	return key, rest, true
}

func matchSSHConfigPatterns(host string, patterns []string) bool {
	if len(patterns) == 0 {
		return false
	}

	normalizedHost := strings.ToLower(strings.TrimSpace(host))
	matched := false
	for _, rawPattern := range patterns {
		pattern := strings.ToLower(strings.TrimSpace(rawPattern))
		if pattern == "" {
			continue
		}

		negated := strings.HasPrefix(pattern, "!")
		if negated {
			pattern = strings.TrimPrefix(pattern, "!")
		}

		ok, err := path.Match(pattern, normalizedHost)
		if err != nil || !ok {
			continue
		}
		if negated {
			return false
		}
		matched = true
	}

	return matched
}

func stripSSHConfigComment(line string) string {
	inSingleQuote := false
	inDoubleQuote := false

	for i, r := range line {
		switch r {
		case '\'':
			if !inDoubleQuote {
				inSingleQuote = !inSingleQuote
			}
		case '"':
			if !inSingleQuote {
				inDoubleQuote = !inDoubleQuote
			}
		case '#':
			if !inSingleQuote && !inDoubleQuote {
				return strings.TrimSpace(line[:i])
			}
		}
	}

	return strings.TrimSpace(line)
}

func stripSSHConfigQuotes(value string) string {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) >= 2 {
		if strings.HasPrefix(trimmed, "\"") && strings.HasSuffix(trimmed, "\"") {
			return strings.TrimSuffix(strings.TrimPrefix(trimmed, "\""), "\"")
		}
		if strings.HasPrefix(trimmed, "'") && strings.HasSuffix(trimmed, "'") {
			return strings.TrimSuffix(strings.TrimPrefix(trimmed, "'"), "'")
		}
	}
	return trimmed
}

func expandSSHPath(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if trimmed == "~" || strings.HasPrefix(trimmed, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home dir: %w", err)
		}
		if trimmed == "~" {
			return home, nil
		}
		return filepath.Join(home, strings.TrimPrefix(trimmed, "~/")), nil
	}
	return trimmed, nil
}

func expandSSHConfigAgent(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if strings.EqualFold(trimmed, "none") || trimmed == sshConfigEnvAgentSocket {
		return trimmed, nil
	}
	return expandSSHPath(trimmed)
}

func firstSSHConfigField(value string) string {
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

func shouldApplyConfigPort(port string) bool {
	trimmed := strings.TrimSpace(port)
	return trimmed == "" || trimmed == "22"
}

func parseProxyJumpList(raw string) ([]ProxyJump, bool, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, false, nil
	}
	if strings.EqualFold(trimmed, "none") {
		return nil, true, nil
	}

	parts := splitProxyJumpSpecs(trimmed)
	jumps := make([]ProxyJump, 0, len(parts))
	for _, part := range parts {
		jump, err := parseProxyJumpSpec(part)
		if err != nil {
			return nil, true, err
		}
		jumps = append(jumps, jump)
	}
	return jumps, true, nil
}

func parseProxyJumpSpec(raw string) (ProxyJump, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ProxyJump{}, fmt.Errorf("proxy jump value is required")
	}

	user := ""
	hostPart := trimmed
	if at := strings.LastIndex(trimmed, "@"); at > 0 {
		user = strings.TrimSpace(trimmed[:at])
		hostPart = strings.TrimSpace(trimmed[at+1:])
	}

	host, port, err := parseSSHHostPort(hostPart)
	if err != nil {
		return ProxyJump{}, err
	}

	return ProxyJump{
		User: user,
		Host: host,
		Port: port,
	}, nil
}

func splitProxyJumpSpecs(raw string) []string {
	parts := make([]string, 0, 4)
	start := 0
	depth := 0

	for i, r := range raw {
		switch r {
		case '[':
			depth++
		case ']':
			if depth > 0 {
				depth--
			}
		case ',':
			if depth == 0 {
				parts = append(parts, raw[start:i])
				start = i + 1
			}
		}
	}

	parts = append(parts, raw[start:])
	return parts
}

func parseSSHHostPort(raw string) (string, string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", "", fmt.Errorf("ssh host is required")
	}

	host, port, err := net.SplitHostPort(value)
	if err == nil {
		if strings.TrimSpace(host) == "" {
			return "", "", fmt.Errorf("ssh host is required")
		}
		if strings.TrimSpace(port) == "" {
			port = defaultSSHPort
		}
		return trimBrackets(host), strings.TrimSpace(port), nil
	}

	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		return trimBrackets(value), defaultSSHPort, nil
	}
	if strings.Count(value, ":") > 1 {
		return "", "", fmt.Errorf("wrap IPv6 addresses with a port as [addr]:port")
	}
	if strings.Contains(value, ":") {
		host, port, ok := strings.Cut(value, ":")
		if !ok || strings.TrimSpace(host) == "" {
			return "", "", fmt.Errorf("ssh host is required")
		}
		return strings.TrimSpace(host), strings.TrimSpace(port), nil
	}
	return value, defaultSSHPort, nil
}
