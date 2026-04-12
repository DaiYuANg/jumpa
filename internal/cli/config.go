package cli

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
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

type flagOverrides struct {
	APIURL      string
	GatewayAddr string
	Email       string
	Password    string
	Principal   string
	SSHBinary   string
	AltScreen   mo.Option[bool]
}

func DefaultConfig() Config {
	return Config{
		APIURL:    "http://127.0.0.1:8080",
		SSHBinary: "ssh",
		AltScreen: true,
	}
}

func loadConfig(log *slog.Logger) Config {
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

	overrides, err := parseFlagOverrides()
	if err != nil {
		log.Error("cli flag parse failed", slog.String("error", err.Error()))
		panic(err)
	}
	applyFlagOverrides(&cfg, overrides)
	return cfg
}

func parseFlagOverrides() (flagOverrides, error) {
	fs := flag.NewFlagSet("jumpa-cli", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var overrides flagOverrides
	fs.StringVar(&overrides.APIURL, "api", "", "jumpa api base url")
	fs.StringVar(&overrides.GatewayAddr, "gateway", "", "override ssh gateway host:port")
	fs.StringVar(&overrides.Email, "email", "", "login email")
	fs.StringVar(&overrides.Password, "password", "", "login password")
	fs.StringVar(&overrides.Principal, "principal", "", "ssh principal used for gateway login")
	fs.StringVar(&overrides.SSHBinary, "ssh", "", "ssh binary path")
	fs.Var(&boolOptionValue{target: &overrides.AltScreen}, "alt-screen", "run bubbletea with alt screen")

	err := fs.Parse(os.Args[1:])
	return overrides, err
}

func applyFlagOverrides(cfg *Config, overrides flagOverrides) {
	if cfg == nil {
		return
	}

	trimmedOption(overrides.APIURL).ForEach(func(value string) {
		cfg.APIURL = value
	})
	trimmedOption(overrides.GatewayAddr).ForEach(func(value string) {
		cfg.GatewayAddr = value
	})
	trimmedOption(overrides.Email).ForEach(func(value string) {
		cfg.Email = value
	})
	trimmedOption(overrides.Password).ForEach(func(value string) {
		cfg.Password = value
	})
	trimmedOption(overrides.Principal).ForEach(func(value string) {
		cfg.Principal = value
	})
	trimmedOption(overrides.SSHBinary).ForEach(func(value string) {
		cfg.SSHBinary = value
	})
	overrides.AltScreen.ForEach(func(value bool) {
		cfg.AltScreen = value
	})
}

func trimmedOption(value string) mo.Option[string] {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return mo.None[string]()
	}
	return mo.Some(trimmed)
}

type boolOptionValue struct {
	target *mo.Option[bool]
}

func (v *boolOptionValue) String() string {
	if v == nil || v.target == nil || v.target.IsAbsent() {
		return ""
	}
	return fmt.Sprintf("%t", v.target.OrElse(false))
}

func (v *boolOptionValue) Set(raw string) error {
	if v == nil || v.target == nil {
		return nil
	}
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "t", "true", "y", "yes", "on":
		*v.target = mo.Some(true)
	case "0", "f", "false", "n", "no", "off":
		*v.target = mo.Some(false)
	default:
		return fmt.Errorf("invalid boolean value %q", raw)
	}
	return nil
}
