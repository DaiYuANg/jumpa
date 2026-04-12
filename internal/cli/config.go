package cli

import (
	"log/slog"
	"strings"

	"github.com/DaiYuANg/arcgo/configx"
	"github.com/samber/mo"
)

type Config struct {
	APIURL      string `koanf:"api_url"`
	GatewayAddr string `koanf:"gateway_addr"`
	Email       string `koanf:"email"`
	Password    string `koanf:"password"`
	Principal   string `koanf:"principal"`
	SSHBinary   string `koanf:"ssh_binary"`
	AltScreen   bool   `koanf:"alt_screen"`
}

type Overrides struct {
	APIURL      mo.Option[string]
	GatewayAddr mo.Option[string]
	Email       mo.Option[string]
	Password    mo.Option[string]
	Principal   mo.Option[string]
	SSHBinary   mo.Option[string]
	AltScreen   mo.Option[bool]
}

func DefaultConfig() Config {
	return Config{
		APIURL:    "http://127.0.0.1:8080",
		SSHBinary: "ssh",
		AltScreen: true,
	}
}

func LoadConfig(log *slog.Logger) Config {
	result := configx.NewT[Config](
		configx.WithTypedDefaults(DefaultConfig()),
		configx.WithDotenv(".env"),
		configx.WithEnvPrefix("APP_CLI_"),
	).Load()

	cfg, err := result.Get()
	if err != nil {
		log.Error("cli config load failed", slog.String("error", err.Error()))
		panic(err)
	}

	return cfg
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
	overrides.SSHBinary.ForEach(func(value string) {
		cfg.SSHBinary = value
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
