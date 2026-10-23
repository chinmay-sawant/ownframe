//go:build linux && !android && !ios && !js

package window

import (
	"context"
	"testing"
)

func TestLinuxA11yDoesNotClaimLiveBus(t *testing.T) {
	s := newShell(context.Background(), &fakeScreen{width: 10, height: 10})
	if s.a11y.pub == nil {
		t.Fatal("publisher")
	}
	if s.a11y.pub.Live() {
		t.Fatal("live AT-SPI client is not implemented")
	}
	s.syncA11y()
	if s.a11y.pub.Status().Live {
		t.Fatal("status")
	}
}
