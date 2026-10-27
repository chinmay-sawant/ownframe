//go:build linux && !android && !ios && !js

package window

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/atspi"
	"github.com/chinmay-sawant/ownframe/internal/host"
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

type accessibleFakeScreen struct {
	fakeScreen
	tree host.AccessibilityTree
}

func (a *accessibleFakeScreen) AccessibilityTree() (host.AccessibilityTree, error) {
	return a.tree, nil
}

func (a *accessibleFakeScreen) AccessibilityPerform(context.Context, host.AccessibilityRequest) error {
	return nil
}

func TestLinuxA11yPublishesWhenLive(t *testing.T) {
	fakeBus := atspi.NewFakeBus()
	scr := &accessibleFakeScreen{
		fakeScreen: fakeScreen{width: 10, height: 10},
		tree: host.AccessibilityTree{
			Document: 1,
			Revision: 1,
			Root:     "id:root",
			Nodes: []host.AccessibilityNode{{
				ID:   "id:root",
				Role: host.AccessibilityButton,
			}},
		},
	}
	s := newShell(context.Background(), scr)
	s.a11y.pub = atspi.NewPublisher(fakeBus)
	if !s.a11y.pub.Live() {
		t.Fatal("expected publisher to be live")
	}
	s.syncA11y()
	if len(fakeBus.Apps()) != 1 {
		t.Fatalf("expected 1 published app, got %d", len(fakeBus.Apps()))
	}
}
