package main

import (
	"testing"

	cli "github.com/DaiYuANg/jumpa/internal/cli"
)

func TestValidateRequestListOptions(t *testing.T) {
	t.Parallel()

	t.Run("normalizes status", func(t *testing.T) {
		t.Parallel()

		options := &cli.ListOptions{
			Status:   " Pending ",
			Page:     1,
			PageSize: 50,
		}
		if err := validateRequestListOptions(options); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if options.Status != "pending" {
			t.Fatalf("expected normalized status, got %q", options.Status)
		}
	})

	t.Run("rejects invalid status", func(t *testing.T) {
		t.Parallel()

		options := &cli.ListOptions{
			Status:   "done",
			Page:     1,
			PageSize: 50,
		}
		if err := validateRequestListOptions(options); err == nil {
			t.Fatal("expected invalid status error")
		}
	})

	t.Run("rejects invalid paging", func(t *testing.T) {
		t.Parallel()

		options := &cli.ListOptions{
			Page:     0,
			PageSize: 201,
		}
		if err := validateRequestListOptions(options); err == nil {
			t.Fatal("expected invalid paging error")
		}
	})
}

func TestParseLocalForwards(t *testing.T) {
	t.Parallel()

	t.Run("parses repeated flags", func(t *testing.T) {
		t.Parallel()

		forwards, err := parseLocalForwards([]string{"15432:db.internal:5432", "127.0.0.1:18080:127.0.0.1:8080"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(forwards) != 2 {
			t.Fatalf("expected 2 forwards, got %d", len(forwards))
		}
	})

	t.Run("rejects malformed flag", func(t *testing.T) {
		t.Parallel()

		if _, err := parseLocalForwards([]string{"not-a-forward"}); err == nil {
			t.Fatal("expected malformed local forward error")
		}
	})
}

func TestParseRemoteForwards(t *testing.T) {
	t.Parallel()

	t.Run("parses repeated flags", func(t *testing.T) {
		t.Parallel()

		forwards, err := parseRemoteForwards([]string{"15432:127.0.0.1:5432", "0.0.0.0:18080:127.0.0.1:8080"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(forwards) != 2 {
			t.Fatalf("expected 2 forwards, got %d", len(forwards))
		}
	})

	t.Run("rejects malformed flag", func(t *testing.T) {
		t.Parallel()

		if _, err := parseRemoteForwards([]string{"not-a-forward"}); err == nil {
			t.Fatal("expected malformed remote forward error")
		}
	})
}

func TestParseDynamicForwards(t *testing.T) {
	t.Parallel()

	t.Run("parses repeated flags", func(t *testing.T) {
		t.Parallel()

		forwards, err := parseDynamicForwards([]string{"1080", "127.0.0.1:2080"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(forwards) != 2 {
			t.Fatalf("expected 2 forwards, got %d", len(forwards))
		}
	})

	t.Run("rejects malformed flag", func(t *testing.T) {
		t.Parallel()

		if _, err := parseDynamicForwards([]string{"127.0.0.1:2080:extra"}); err == nil {
			t.Fatal("expected malformed dynamic forward error")
		}
	})
}
