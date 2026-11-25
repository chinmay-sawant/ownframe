//go:build linux && !android

package window

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseDRMModeLineHz(t *testing.T) {
	t.Parallel()

	cases := []struct {
		line   string
		wantHz int
		wantOK bool
	}{
		{"2560x1440@144", 144, true},
		{"1920x1080@59.94", 60, true},
		{"3840x2160@120p", 120, true},
		{"1920x1080@60i", 60, true},
		{"1920x1080", 0, false},
		{"1920x1080p", 0, false},
		{"1920x1080@", 0, false},
		{"1920x1080@0", 0, false},
		{"1920x1080@999", 0, false},
		{"1920x1080@abc", 0, false},
		{"", 0, false},
	}

	for _, tc := range cases {
		hz, ok := parseDRMModeLineHz(tc.line)
		if hz != tc.wantHz || ok != tc.wantOK {
			t.Errorf("parseDRMModeLineHz(%q) = (%d, %v), want (%d, %v)", tc.line, hz, ok, tc.wantHz, tc.wantOK)
		}
	}
}

func TestDetectMonitorHzDRM(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Disconnected output should be skipped.
	disconn := filepath.Join(dir, "card0-HDMI-A-1")
	if err := os.Mkdir(disconn, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(disconn, "status"), []byte("disconnected\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(disconn, "modes"), []byte("1920x1080@240\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Connected output with preferred mode at 165 Hz.
	conn := filepath.Join(dir, "card0-DP-1")
	if err := os.Mkdir(conn, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(conn, "status"), []byte("connected\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(conn, "modes"), []byte("2560x1440\n2560x1440@165\n1920x1080@60\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	hz, ok := detectMonitorHzDRM(dir)
	if !ok || hz != 165 {
		t.Fatalf("detectMonitorHzDRM = (%d, %v), want (165, true)", hz, ok)
	}

	// Missing directory falls back cleanly without panicking.
	if hz, ok := detectMonitorHzDRM(filepath.Join(dir, "nonexistent")); ok || hz != 0 {
		t.Fatalf("missing dir = (%d, %v), want (0, false)", hz, ok)
	}
}

func TestDetectMonitorHzLinuxSafeFallback(t *testing.T) {
	t.Setenv("DISPLAY", "")

	hz := detectMonitorHz(100, 100, 800, 600)
	if hz < minValidMonitorHz || hz > maxValidMonitorHz {
		t.Fatalf("detectMonitorHz returned out-of-range %d", hz)
	}
}
