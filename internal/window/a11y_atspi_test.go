//go:build linux && !android && !ios && !js

package window

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/atspi"
	"github.com/chinmay-sawant/ownframe/internal/host"
)

func TestLinuxA11yHasPublisher(t *testing.T) {
	s := newShell(context.Background(), &fakeScreen{width: 10, height: 10})
	if s.a11y.pub == nil {
		t.Fatal("publisher")
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
	s.a11yOrigin = func() (int, int) { return 0, 0 }
	s.a11y.pub = atspi.NewPublisher(fakeBus)
	if !s.a11y.pub.Live() {
		t.Fatal("expected publisher to be live")
	}
	s.syncA11y()
	if len(fakeBus.Apps()) != 1 {
		t.Fatalf("expected 1 published app, got %d", len(fakeBus.Apps()))
	}
	s.syncA11y()
	if len(fakeBus.Apps()) != 1 {
		t.Fatalf("unchanged generation republished: %d", len(fakeBus.Apps()))
	}
	scr.tree.Nodes[0].Name = "changed"
	scr.redraws++
	s.syncA11y()
	if len(fakeBus.Apps()) != 2 {
		t.Fatalf("changed generation did not publish: %d", len(fakeBus.Apps()))
	}
	s.closeA11y()
	if s.a11y.pub.Live() {
		t.Fatal("publisher stayed live after close")
	}
	s.syncA11y()
	if len(fakeBus.Apps()) != 2 {
		t.Fatalf("closed publisher published again: %d", len(fakeBus.Apps()))
	}
}
