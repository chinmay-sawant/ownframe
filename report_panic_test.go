package ownframe

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/crash"
)

func TestPanicReportKeepsChain(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "r")
	if err := crash.SetDir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = crash.SetDir(t.TempDir()) })
	boom := errors.New("boom")
	var got error
	func() {
		defer savePanic(nil, &got)
		panic(boom)
	}()
	if !errors.Is(got, boom) {
		t.Fatalf("boom err = %v", got)
	}
	if !errors.Is(got, ErrPanicked) {
		t.Fatalf("panic err = %v", got)
	}
	if err := reportError(nil, boom); !errors.Is(err, boom) {
		t.Fatalf("report err = %v", err)
	}
}
