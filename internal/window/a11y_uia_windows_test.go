//go:build windows && (amd64 || arm64)

package window

import (
	"context"
	"testing"
	"time"
)

func TestWindowsA11yHasProviderAndStaysDetachedWithoutTree(t *testing.T) {
	s := newShell(context.Background(), &fakeScreen{width: 10, height: 10})
	defer s.closeA11y()
	if s.a11y.prov == nil || s.a11y.actions == nil {
		t.Fatal("UIA provider and action queue must exist")
	}
	// fakeScreen is not host.Accessible, so sync must stay a no-op.
	s.syncA11y()
	if s.a11y.host != nil {
		t.Fatal("attached without a tree")
	}
}

// TestWindowsA11yDetachesOnContextCancel mirrors the AT-SPI lifecycle test: a
// NewGame host that never calls runWindow must still detach UIA when its
// context ends.
func TestWindowsA11yDetachesOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	s := newShell(ctx, &fakeScreen{width: 10, height: 10})
	defer s.closeA11y()
	cancel()

	deadline := time.Now().Add(2 * time.Second)
	for {
		s.a11y.mu.Lock()
		closed := s.a11y.closed
		s.a11y.mu.Unlock()
		if closed {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("UIA stayed attached after the context was canceled")
		}
		time.Sleep(time.Millisecond)
	}
}
