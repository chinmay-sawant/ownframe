//go:build linux && !android && !ios && !js

package window

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/atspi"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/page"
)

func TestA11yActionsUsePageLoopAndRejectRemovedTargets(t *testing.T) {
	ctx := t.Context()
	p, err := page.New(page.Config{Width: 320, Height: 200,
		HTML: `<input id="name"><button id="go">Go</button><p>{{.}}</p>`})
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	p.SetData(count)
	p.Handle(page.Handlers{Click: func(_ context.Context, box page.Box) error {
		if box.ID == "go" {
			count++
			p.SetData(count)
		}
		return nil
	}})
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	s := newShell(ctx, p)
	t.Cleanup(s.cancel)
	t.Cleanup(s.closeA11y)
	wake := make(chan struct{}, 1)
	s.a11y.actions.Close()
	s.a11y.actions = atspi.NewActions(ctx, func() { wake <- struct{}{} })
	enqueue := func(id host.AccessibilityNodeID, kind host.AccessibilityActionKind) <-chan error {
		tree, err := p.AccessibilityTree()
		if err != nil {
			t.Fatal(err)
		}
		result := make(chan error, 1)
		go func() {
			result <- s.a11y.actions.Request(ctx, host.AccessibilityRequest{Document: tree.Document, Node: id, Kind: kind})
		}()
		select {
		case <-wake:
		case <-time.After(time.Second):
			t.Fatal("request did not wake the host")
		}
		return result
	}
	focus := enqueue("id:name", host.AccessibilityFocus)
	s.runA11yActions()
	if err := <-focus; err != nil || p.FocusID() != "name" {
		t.Fatal("focus action did not reach the page")
	}
	activate := enqueue("id:go", host.AccessibilityActivate)
	s.runA11yActions()
	if err := <-activate; err != nil || count != 1 {
		t.Fatal("activation did not reach the application handler")
	}
	removed := enqueue("id:go", host.AccessibilityActivate)
	p.SetData(count)
	if err := p.Load(ctx, `<p>Removed controls</p>`); err != nil {
		t.Fatal(err)
	}
	s.runA11yActions()
	if err := <-removed; !errors.Is(err, host.ErrAccessibilityStale) || count != 1 {
		t.Fatal("removed target reached another control")
	}
}
