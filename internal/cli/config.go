package cli

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/DaiYuANg/arcgo/configx"
	cliapp "github.com/DaiYuANg/jumpa/internal/cli/app"
	"github.com/DaiYuANg/jumpa/internal/sshclient"
	"github.com/samber/mo"
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

type Overrides struct {
	APIURL                  mo.Option[string]
	GatewayAddr             mo.Option[string]
	Email                   mo.Option[string]
	Password                mo.Option[string]
	Principal               mo.Option[string]
	SSHConfigPath           mo.Option[string]
	SSHPrivateKeyPath       mo.Option[string]
	SSHPrivateKeyPassphrase mo.Option[string]
	SSHAgentEnabled         mo.Option[bool]
	SSHAgentSocket          mo.Option[string]
	AltScreen               mo.Option[bool]
}

func DefaultConfig() Config {
	return Config{
		APIURL:           "http://127.0.0.1:8080",
		SSHHostKeyPolicy: sshclient.HostKeyPolicyKnownHosts,
		AltScreen:        true,
	}
}

func LoadConfig() (Config, error) {
	result := configx.NewT[Config](
		configx.WithTypedDefaults(DefaultConfig()),
		configx.WithDotenv(".env"),
		configx.WithEnvPrefix("APP_CLI_"),
	).Load()

	return result.Get()
}

func ResolveConfig(overrides Overrides) (Config, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return Config{}, fmt.Errorf("load cli config: %w", err)
	}

	cfg = ApplyOverrides(cfg, overrides)
	if err := ValidateConfig(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
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

func ApplyOverrides(cfg Config, overrides Overrides) Config {
	overrides.APIURL.ForEach(func(value string) {
		cfg.APIURL = value
	})
	overrides.GatewayAddr.ForEach(func(value string) {
		cfg.GatewayAddr = value
	})
	overrides.Email.ForEach(func(value string) {
		cfg.Email = value
	})
	overrides.Password.ForEach(func(value string) {
		cfg.Password = value
	})
	overrides.Principal.ForEach(func(value string) {
		cfg.Principal = value
	})
	overrides.SSHConfigPath.ForEach(func(value string) {
		cfg.SSHConfigPath = value
	})
	overrides.SSHPrivateKeyPath.ForEach(func(value string) {
		cfg.SSHPrivateKeyPath = value
	})
	overrides.SSHPrivateKeyPassphrase.ForEach(func(value string) {
		cfg.SSHPrivateKeyPassphrase = value
	})
	overrides.SSHAgentEnabled.ForEach(func(value bool) {
		cfg.SSHAgentEnabled = value
	})
	overrides.SSHAgentSocket.ForEach(func(value string) {
		cfg.SSHAgentSocket = value
	})
	overrides.AltScreen.ForEach(func(value bool) {
		cfg.AltScreen = value
	})
	return cfg
}

func StringOverride(value string) mo.Option[string] {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return mo.None[string]()
	}
	return mo.Some(trimmed)
}
