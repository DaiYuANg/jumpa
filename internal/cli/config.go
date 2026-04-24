package cli

import (
	"fmt"
	"net/url"
	"strings"

	cliapp "github.com/DaiYuANg/jumpa/internal/cli/app"
	"github.com/DaiYuANg/jumpa/internal/sshclient"
	"github.com/arcgolabs/configx"
	"github.com/samber/mo"
	"github.com/spf13/pflag"
)

type Config struct {
	APIURL                  string `koanf:"api_url"`
	GatewayAddr             string `koanf:"gateway_addr"`
	Email                   string `koanf:"email"`
	Password                string `koanf:"password"`
	Principal               string `koanf:"principal"`
	SSHHostKeyPolicy        string `koanf:"ssh_host_key_policy"`
	SSHKnownHostsPath       string `koanf:"ssh_known_hosts_path"`
	SSHConfigPath           string `koanf:"ssh_config_path"`
	SSHPrivateKeyPath       string `koanf:"ssh_private_key_path"`
	SSHPrivateKeyPassphrase string `koanf:"ssh_private_key_passphrase"`
	SSHAgentEnabled         bool   `koanf:"ssh_agent_enabled"`
	SSHAgentSocket          string `koanf:"ssh_agent_socket"`
	AltScreen               bool   `koanf:"alt_screen"`
}

func DefaultConfig() Config {
	return Config{
		APIURL:           "http://127.0.0.1:8080",
		SSHHostKeyPolicy: sshclient.HostKeyPolicyKnownHosts,
		AltScreen:        true,
	}
}

func LoadConfig(flagSet *pflag.FlagSet) (Config, error) {
	options := []configx.Option{
		configx.WithTypedDefaults(DefaultConfig()),
		configx.WithDotenv(".env"),
		configx.WithEnvPrefix("APP_CLI_"),
	}
	if flagSet != nil {
		options = append(options,
			configx.WithFlagSet(flagSet),
			configx.WithArgsNameFunc(cliConfigPathForFlag),
		)
	}

	result := configx.NewT[Config](options...).Load()
	return result.Get()
}

func LoadAndValidateConfig(flagSet *pflag.FlagSet) (Config, error) {
	cfg, err := LoadConfig(flagSet)
	if err != nil {
		return Config{}, fmt.Errorf("load cli config: %w", err)
	}
	if err := ValidateConfig(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func cliConfigPathForFlag(name string) string {
	switch strings.TrimSpace(name) {
	case "api":
		return "api_url"
	case "gateway":
		return "gateway_addr"
	case "email":
		return "email"
	case "password":
		return "password"
	case "principal":
		return "principal"
	case "ssh-config":
		return "ssh_config_path"
	case "ssh-key":
		return "ssh_private_key_path"
	case "ssh-key-passphrase":
		return "ssh_private_key_passphrase"
	case "ssh-agent":
		return "ssh_agent_enabled"
	case "ssh-agent-sock":
		return "ssh_agent_socket"
	case "alt-screen":
		return "alt_screen"
	default:
		return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(name)), "-", "_")
	}
}

func ValidateConfig(cfg *Config) error {
	if cfg == nil {
		return nil
	}

	cfg.APIURL = strings.TrimSpace(cfg.APIURL)
	cfg.GatewayAddr = strings.TrimSpace(cfg.GatewayAddr)
	cfg.Email = strings.TrimSpace(cfg.Email)
	cfg.Principal = strings.TrimSpace(cfg.Principal)
	cfg.SSHHostKeyPolicy = strings.ToLower(strings.TrimSpace(cfg.SSHHostKeyPolicy))
	cfg.SSHKnownHostsPath = strings.TrimSpace(cfg.SSHKnownHostsPath)
	cfg.SSHConfigPath = strings.TrimSpace(cfg.SSHConfigPath)
	cfg.SSHPrivateKeyPath = strings.TrimSpace(cfg.SSHPrivateKeyPath)
	cfg.SSHPrivateKeyPassphrase = strings.TrimSpace(cfg.SSHPrivateKeyPassphrase)
	cfg.SSHAgentSocket = strings.TrimSpace(cfg.SSHAgentSocket)

	if cfg.APIURL == "" {
		return fmt.Errorf("invalid cli api url: value is required")
	}

	parsed, err := url.Parse(cfg.APIURL)
	if err != nil {
		return fmt.Errorf("invalid cli api url %q: %w", cfg.APIURL, err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("invalid cli api url %q: expected an absolute http(s) URL", cfg.APIURL)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("invalid cli api url %q: scheme must be http or https", cfg.APIURL)
	}

	switch cfg.SSHHostKeyPolicy {
	case "", sshclient.HostKeyPolicyKnownHosts, sshclient.HostKeyPolicyInsecure:
		if cfg.SSHHostKeyPolicy == "" {
			cfg.SSHHostKeyPolicy = sshclient.HostKeyPolicyKnownHosts
		}
	default:
		return fmt.Errorf("invalid cli ssh host key policy %q: allowed values are %s, %s", cfg.SSHHostKeyPolicy, sshclient.HostKeyPolicyKnownHosts, sshclient.HostKeyPolicyInsecure)
	}

	if cfg.SSHPrivateKeyPassphrase != "" && cfg.SSHPrivateKeyPath == "" {
		return fmt.Errorf("invalid cli ssh private key passphrase: ssh_private_key_path is required")
	}
	if cfg.SSHAgentSocket != "" {
		cfg.SSHAgentEnabled = true
	}

	if cfg.GatewayAddr == "" {
		return nil
	}

	normalized, err := cliapp.NormalizeGatewayAddress(cfg.GatewayAddr)
	if err != nil {
		return fmt.Errorf("invalid cli gateway address %q: %w", cfg.GatewayAddr, err)
	}
	cfg.GatewayAddr = normalized
	return nil
}

func StringOverride(value string) mo.Option[string] {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return mo.None[string]()
	}
	return mo.Some(trimmed)
}
