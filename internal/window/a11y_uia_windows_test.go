//go:build windows && (amd64 || arm64)

package window

import (
	"context"
	"testing"
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
