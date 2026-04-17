package cli

import "testing"

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
