package ui

import (
	"testing"
	"time"
)

func TestFormatBytes(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{-1, "--"},
		{0, "0 B"},
		{512, "512 B"},
		{1023, "1023 B"},
		{1024, "1.0 KiB"},
		{1536, "1.5 KiB"},
		{10 * 1024 * 1024, "10.0 MiB"},
		{3 * 1024 * 1024 * 1024, "3.0 GiB"},
	}

	for _, c := range cases {
		if got := formatBytes(c.in); got != c.want {
			t.Errorf("formatBytes(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestFormatSpeed(t *testing.T) {
	if got := formatSpeed(0); got != "--" {
		t.Errorf("zero speed = %q", got)
	}

	if got := formatSpeed(1536); got != "1.5 KiB/s" {
		t.Errorf("1536 B/s = %q", got)
	}
}

func TestFormatETA(t *testing.T) {
	cases := []struct {
		in   time.Duration
		want string
	}{
		{-1, "--"},
		{5 * time.Second, "<10s"},
		{45 * time.Second, "45s"},
		{90 * time.Second, "1m 30s"},
		{2*time.Hour + 5*time.Minute, "2h 05m"},
		{50 * time.Hour, "2d 02h"},
	}

	for _, c := range cases {
		if got := formatETA(c.in); got != c.want {
			t.Errorf("formatETA(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestRedactURL(t *testing.T) {
	secret := "https://bob:hunter2@example.com/file.bin" //nolint:gosec // reason: test-only fake URL for redact check, not a credential.
	if got := redactURL(secret); got != "https://example.com/file.bin" {
		t.Errorf("redacted = %q", got)
	}

	plain := "https://example.com/file.bin?token=1"
	if got := redactURL(plain); got != plain {
		t.Errorf("plain changed to %q", got)
	}

	if got := redactURL("not a url"); got != "not a url" {
		t.Errorf("bad url changed to %q", got)
	}
}
