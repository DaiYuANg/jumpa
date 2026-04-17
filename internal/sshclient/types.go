package sshclient

import (
	"io"
	"log/slog"
	"strings"
	"time"
)

const (
	HostKeyPolicyKnownHosts = "known_hosts"
	HostKeyPolicyInsecure   = "insecure"
	defaultSSHPort          = "22"
	DefaultConnectTimeout   = 15 * time.Second
)

type Config struct {
	HostKeyPolicy  string
	KnownHostsPath string
	ConfigPath     string
	ConnectTimeout time.Duration
}

type Streams struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer
}

type Request struct {
	User            string
	Host            string
	Port            string
	Password        string
	PrivateKey      *PrivateKey
	AgentSocket     string
	ProxyJumps      []ProxyJump
	LocalForwards   []LocalForward
	RemoteForwards  []RemoteForward
	DynamicForwards []DynamicForward
	Terminal        *Terminal
}

type Terminal struct {
	Term    string
	Width   int
	Height  int
	MakeRaw bool
	Resize  <-chan WindowSize
}

type WindowSize struct {
	Width  int
	Height int
}

type Client struct {
	cfg    Config
	log    *slog.Logger
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func DefaultConfig() Config {
	return Config{
		HostKeyPolicy:  HostKeyPolicyKnownHosts,
		ConnectTimeout: DefaultConnectTimeout,
	}
}

func New(cfg Config, log *slog.Logger, streams Streams) *Client {
	return &Client{
		cfg:    applyDefaults(cfg),
		log:    log,
		stdin:  streams.In,
		stdout: streams.Out,
		stderr: streams.Err,
	}
}

func applyDefaults(cfg Config) Config {
	cfg.HostKeyPolicy = strings.ToLower(strings.TrimSpace(cfg.HostKeyPolicy))
	cfg.KnownHostsPath = strings.TrimSpace(cfg.KnownHostsPath)
	cfg.ConfigPath = strings.TrimSpace(cfg.ConfigPath)
	if cfg.HostKeyPolicy == "" {
		cfg.HostKeyPolicy = HostKeyPolicyKnownHosts
	}
	if cfg.ConnectTimeout <= 0 {
		cfg.ConnectTimeout = DefaultConnectTimeout
	}
	return cfg
}

func normalizeRequest(req Request) Request {
	req.User = strings.TrimSpace(req.User)
	req.Host = strings.TrimSpace(req.Host)
	req.Port = strings.TrimSpace(req.Port)
	req.Password = strings.TrimSpace(req.Password)
	req.AgentSocket = strings.TrimSpace(req.AgentSocket)
	req.PrivateKey = normalizePrivateKey(req.PrivateKey)
	req.ProxyJumps = normalizeProxyJumps(req.ProxyJumps)
	return req
}

func trimBrackets(value string) string {
	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		return strings.TrimSuffix(strings.TrimPrefix(value, "["), "]")
	}
	return value
}
