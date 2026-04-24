package cli

import (
	"testing"

	"github.com/spf13/pflag"
)

func TestValidateConfig(t *testing.T) {
	t.Parallel()

	t.Run("normalizes gateway address", func(t *testing.T) {
		t.Parallel()

		cfg := Config{
			APIURL:           "http://127.0.0.1:8080",
			GatewayAddr:      "::1",
			SSHHostKeyPolicy: "known_hosts",
		}
		if err := ValidateConfig(&cfg); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.GatewayAddr != "[::1]:22" {
			t.Fatalf("expected normalized gateway address, got %q", cfg.GatewayAddr)
		}
	})

	t.Run("rejects invalid api url", func(t *testing.T) {
		t.Parallel()

		cfg := Config{
			APIURL:           "not-a-url",
			SSHHostKeyPolicy: "known_hosts",
		}
		if err := ValidateConfig(&cfg); err == nil {
			t.Fatal("expected invalid api url error")
		}
	})

	t.Run("rejects invalid host key policy", func(t *testing.T) {
		t.Parallel()

		cfg := Config{
			APIURL:           "http://127.0.0.1:8080",
			SSHHostKeyPolicy: "trust-me",
		}
		if err := ValidateConfig(&cfg); err == nil {
			t.Fatal("expected invalid host key policy error")
		}
	})

	t.Run("rejects key passphrase without key path", func(t *testing.T) {
		t.Parallel()

		cfg := Config{
			APIURL:                  "http://127.0.0.1:8080",
			SSHHostKeyPolicy:        "known_hosts",
			SSHPrivateKeyPassphrase: "secret",
		}
		if err := ValidateConfig(&cfg); err == nil {
			t.Fatal("expected invalid private key config error")
		}
	})

	t.Run("enables ssh agent when socket is set", func(t *testing.T) {
		t.Parallel()

		cfg := Config{
			APIURL:           "http://127.0.0.1:8080",
			SSHHostKeyPolicy: "known_hosts",
			SSHAgentSocket:   "/tmp/agent.sock",
		}
		if err := ValidateConfig(&cfg); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !cfg.SSHAgentEnabled {
			t.Fatal("expected ssh agent to be enabled")
		}
	})
}

func TestLoadAndValidateConfig(t *testing.T) {
	t.Parallel()

	fs := pflag.NewFlagSet("jumpa-cli", pflag.ContinueOnError)
	fs.String("api", "", "")
	fs.String("gateway", "", "")
	fs.String("email", "", "")
	fs.String("password", "", "")
	fs.String("principal", "", "")
	fs.String("ssh-config", "", "")
	fs.String("ssh-key", "", "")
	fs.String("ssh-key-passphrase", "", "")
	fs.Bool("ssh-agent", false, "")
	fs.String("ssh-agent-sock", "", "")
	fs.Bool("alt-screen", true, "")

	if err := fs.Parse([]string{
		"--api=https://jumpa.example.com",
		"--gateway=::1",
		"--email=alice@example.com",
		"--ssh-agent-sock=/tmp/agent.sock",
		"--alt-screen=false",
	}); err != nil {
		t.Fatalf("parse flags: %v", err)
	}

	cfg, err := LoadAndValidateConfig(fs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.APIURL != "https://jumpa.example.com" {
		t.Fatalf("expected api url from flags, got %q", cfg.APIURL)
	}
	if cfg.GatewayAddr != "[::1]:22" {
		t.Fatalf("expected normalized gateway address from flags, got %q", cfg.GatewayAddr)
	}
	if cfg.Email != "alice@example.com" {
		t.Fatalf("expected email from flags, got %q", cfg.Email)
	}
	if !cfg.SSHAgentEnabled {
		t.Fatal("expected ssh agent to be enabled from socket flag")
	}
	if cfg.AltScreen {
		t.Fatal("expected alt screen to be disabled from flags")
	}
}
