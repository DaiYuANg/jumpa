package sshclient

import "testing"

func TestParseLocalForward(t *testing.T) {
	t.Parallel()

	t.Run("defaults bind host", func(t *testing.T) {
		t.Parallel()

		forward, err := ParseLocalForward("15432:db.internal:5432")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if forward.LocalHost != defaultLocalForwardHost {
			t.Fatalf("expected default local host %q, got %q", defaultLocalForwardHost, forward.LocalHost)
		}
		if forward.LocalPort != "15432" || forward.RemoteHost != "db.internal" || forward.RemotePort != "5432" {
			t.Fatalf("unexpected forward: %+v", forward)
		}
	})

	t.Run("supports ipv6", func(t *testing.T) {
		t.Parallel()

		forward, err := ParseLocalForward("[::1]:8080:[2001:db8::10]:80")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if forward.LocalHost != "::1" || forward.RemoteHost != "2001:db8::10" {
			t.Fatalf("expected normalized ipv6 hosts, got %+v", forward)
		}
	})

	t.Run("rejects malformed spec", func(t *testing.T) {
		t.Parallel()

		if _, err := ParseLocalForward("8080:only-two-parts"); err == nil {
			t.Fatal("expected malformed local forward error")
		}
	})
}

func TestParseRemoteForward(t *testing.T) {
	t.Parallel()

	t.Run("defaults bind host", func(t *testing.T) {
		t.Parallel()

		forward, err := ParseRemoteForward("15432:127.0.0.1:5432")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if forward.BindHost != defaultRemoteForwardBindHost {
			t.Fatalf("expected default bind host %q, got %q", defaultRemoteForwardBindHost, forward.BindHost)
		}
		if forward.BindPort != "15432" || forward.LocalHost != "127.0.0.1" || forward.LocalPort != "5432" {
			t.Fatalf("unexpected forward: %+v", forward)
		}
	})

	t.Run("supports ipv6", func(t *testing.T) {
		t.Parallel()

		forward, err := ParseRemoteForward("[::]:8080:[::1]:80")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if forward.BindHost != "::" || forward.LocalHost != "::1" {
			t.Fatalf("expected normalized ipv6 hosts, got %+v", forward)
		}
	})

	t.Run("rejects malformed spec", func(t *testing.T) {
		t.Parallel()

		if _, err := ParseRemoteForward("8080:only-two-parts"); err == nil {
			t.Fatal("expected malformed remote forward error")
		}
	})
}
