package sshclient

import (
	"path/filepath"
	"testing"
)

func TestParseSSHConfigProfile(t *testing.T) {
	t.Parallel()

	config := `
Host *
  UserKnownHostsFile ~/.ssh/global_known_hosts

Host jumpa-gateway
  User jumpa
  HostName bastion.internal
  Port 2200
  IdentityFile ~/.ssh/id_jumpa
  IdentityAgent SSH_AUTH_SOCK
  StrictHostKeyChecking no
  ProxyJump ops@jumphost-a:2022,jumphost-b

Host ignored
  HostName ignored.internal
`

	profile, err := parseSSHConfigProfile(config, "jumpa-gateway")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if profile.HostName != "bastion.internal" {
		t.Fatalf("expected hostname override, got %q", profile.HostName)
	}
	if profile.User != "jumpa" {
		t.Fatalf("expected user override, got %q", profile.User)
	}
	if profile.Port != "2200" {
		t.Fatalf("expected port override, got %q", profile.Port)
	}
	if filepath.Base(profile.IdentityFile) != "id_jumpa" {
		t.Fatalf("expected identity file override, got %q", profile.IdentityFile)
	}
	if profile.IdentityAgent != sshConfigEnvAgentSocket {
		t.Fatalf("expected agent env marker, got %q", profile.IdentityAgent)
	}
	if profile.UserKnownHostsFile == "" {
		t.Fatal("expected known_hosts override")
	}
	if profile.StrictHostKeyChecking != "no" {
		t.Fatalf("expected strict host key checking override, got %q", profile.StrictHostKeyChecking)
	}
	if profile.ProxyJump != "ops@jumphost-a:2022,jumphost-b" {
		t.Fatalf("expected proxy jump override, got %q", profile.ProxyJump)
	}
}

func TestMatchSSHConfigPatterns(t *testing.T) {
	t.Parallel()

	if !matchSSHConfigPatterns("bastion-1", []string{"bastion-*", "!bastion-2"}) {
		t.Fatal("expected host to match positive pattern")
	}
	if matchSSHConfigPatterns("bastion-2", []string{"bastion-*", "!bastion-2"}) {
		t.Fatal("expected negated pattern to exclude host")
	}
}

func TestShouldApplyConfigPort(t *testing.T) {
	t.Parallel()

	if !shouldApplyConfigPort("") || !shouldApplyConfigPort("22") {
		t.Fatal("expected default ports to be overridable by ssh config")
	}
	if shouldApplyConfigPort("2222") {
		t.Fatal("expected explicit non-default port to win over ssh config")
	}
}

func TestParseProxyJumpList(t *testing.T) {
	t.Parallel()

	jumps, configured, err := parseProxyJumpList("ops@jump-a:2022,jump-b,[2001:db8::10]:2200")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !configured {
		t.Fatal("expected proxy jump list to be marked configured")
	}
	if len(jumps) != 3 {
		t.Fatalf("expected 3 proxy jumps, got %d", len(jumps))
	}
	if jumps[0].User != "ops" || jumps[0].Host != "jump-a" || jumps[0].Port != "2022" {
		t.Fatalf("unexpected first proxy jump: %+v", jumps[0])
	}
	if jumps[1].User != "" || jumps[1].Host != "jump-b" || jumps[1].Port != defaultSSHPort {
		t.Fatalf("unexpected second proxy jump: %+v", jumps[1])
	}
	if jumps[2].Host != "2001:db8::10" || jumps[2].Port != "2200" {
		t.Fatalf("unexpected third proxy jump: %+v", jumps[2])
	}
}
