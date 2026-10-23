//go:build !android && !ios

package ownframe

import (
	"errors"
	"os"
	"strings"
	"testing"
)

// TestBindMobileSetGame checks BindMobile reaches ebitenmobile.SetGame.
// Desktop SetGame panics. savePanic returns ErrPanicked and a crash file.
// The call does not start the desktop loop. Android and iOS are excluded
// because SetGame starts the mobile view there.
func TestBindMobileSetGame(t *testing.T) {
	dir := withCrashDir(t)
	page, err := New(Config{
		HTML:   `<p>hi</p>`,
		Width:  1,
		Height: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = BindMobile(t.Context(), page)
	if err == nil {
		t.Fatal("BindMobile returned nil without calling SetGame")
	}
	if !errors.Is(err, ErrPanicked) {
		t.Fatalf("err = %v, want ErrPanicked from SetGame", err)
	}
	if !strings.Contains(err.Error(), "setGame") {
		t.Fatalf("err = %v, want the desktop SetGame rejection", err)
	}

	files := reportFiles(t, dir)
	if len(files) != 1 {
		t.Fatalf("reports = %v", files)
	}
	if got := LastReportPath(); got != files[0] {
		t.Fatalf("LastReportPath = %q, want %q", got, files[0])
	}
	body, rerr := os.ReadFile(files[0])
	if rerr != nil {
		t.Fatal(rerr)
	}
	if !strings.Contains(string(body), "setGame") {
		t.Fatalf("report body = %q", body)
	}
}
