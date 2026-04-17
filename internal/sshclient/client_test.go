package sshclient

import "testing"

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	if cfg.HostKeyPolicy != HostKeyPolicyKnownHosts {
		t.Fatalf("expected default host key policy %q, got %q", HostKeyPolicyKnownHosts, cfg.HostKeyPolicy)
	}
	if cfg.ConnectTimeout != DefaultConnectTimeout {
		t.Fatalf("expected default connect timeout %s, got %s", DefaultConnectTimeout, cfg.ConnectTimeout)
	}
}

func TestNormalizeRequest(t *testing.T) {
	t.Parallel()

	req := normalizeRequest(Request{
		User:        " alice ",
		Host:        " [::1] ",
		Port:        " 2222 ",
		Password:    " secret ",
		AgentSocket: " /tmp/agent.sock ",
		ProxyJumps: []ProxyJump{
			{User: " jump ", Host: " bastion ", Port: " 2200 "},
			{Host: " "},
		},
		PrivateKey: &PrivateKey{
			Path:       " ~/.ssh/id_ed25519 ",
			Passphrase: " passphrase ",
		},
	})

	if req.User != "alice" {
		t.Fatalf("expected trimmed user, got %q", req.User)
	}
	if req.Host != "[::1]" {
		t.Fatalf("expected trimmed host, got %q", req.Host)
	}
	if req.Port != "2222" {
		t.Fatalf("expected trimmed port, got %q", req.Port)
	}
	if req.Password != "secret" {
		t.Fatalf("expected trimmed password, got %q", req.Password)
	}
	if req.AgentSocket != "/tmp/agent.sock" {
		t.Fatalf("expected trimmed agent socket, got %q", req.AgentSocket)
	}
	if req.PrivateKey == nil || req.PrivateKey.Path != "~/.ssh/id_ed25519" || req.PrivateKey.Passphrase != "passphrase" {
		t.Fatalf("expected normalized private key, got %+v", req.PrivateKey)
	}
	if len(req.ProxyJumps) != 1 || req.ProxyJumps[0].User != "jump" || req.ProxyJumps[0].Host != "bastion" || req.ProxyJumps[0].Port != "2200" {
		t.Fatalf("expected normalized proxy jumps, got %+v", req.ProxyJumps)
	}
}

func TestNormalizeTerminal(t *testing.T) {
	t.Parallel()

	terminal := normalizeTerminal(&Terminal{})
	if terminal == nil {
		t.Fatal("expected terminal config")
	}
	if terminal.Term != "xterm-256color" {
		t.Fatalf("expected default term, got %q", terminal.Term)
	}
	if terminal.Width != 80 || terminal.Height != 24 {
		t.Fatalf("expected default size 80x24, got %dx%d", terminal.Width, terminal.Height)
	}
}
