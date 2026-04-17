package sshclient

import "testing"

func TestParseDynamicForward(t *testing.T) {
	t.Parallel()

	t.Run("defaults bind host", func(t *testing.T) {
		t.Parallel()

		forward, err := ParseDynamicForward("1080")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if forward.LocalHost != defaultLocalForwardHost || forward.LocalPort != "1080" {
			t.Fatalf("unexpected forward: %+v", forward)
		}
	})

	t.Run("supports explicit bind host", func(t *testing.T) {
		t.Parallel()

		forward, err := ParseDynamicForward("[::1]:1080")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if forward.LocalHost != "::1" || forward.LocalPort != "1080" {
			t.Fatalf("unexpected forward: %+v", forward)
		}
	})

	t.Run("rejects malformed spec", func(t *testing.T) {
		t.Parallel()

		if _, err := ParseDynamicForward("127.0.0.1:1080:extra"); err == nil {
			t.Fatal("expected malformed dynamic forward error")
		}
	})
}
