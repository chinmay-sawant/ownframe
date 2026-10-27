package page

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

var (
	_ host.ContextMenuFor = (*Page)(nil)
	_ host.MenuActioner   = (*Page)(nil)
)

// TestContextMenuWireUsesHandler proves the hit-tested id reaches the
// callback registered through Handle and that its rows replace the built-in
// editing menu.
func TestContextMenuWireUsesHandler(t *testing.T) {
	t.Parallel()

	p := bftPage(t, `<div id="tab-0">tab</div>`)
	bftDraw(t, p, nil)

	want := []MenuItem{
		{ID: "code-view:tab-close-all:1", Label: "Close All", Enabled: true},
	}

	var asked string

	p.Handle(Handlers{
		ContextMenuFor: func(id string) []MenuItem {
			asked = id

			return want
		},
	})

	got := p.ContextMenuFor("tab-0")
	if asked != "tab-0" {
		t.Fatalf("handler id = %q, want tab-0", asked)
	}

	if len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("rows = %+v, want %+v", got, want)
	}
}

// TestContextMenuWireFallsBackWhenHandlerEmpty keeps the built-in menu when
// the custom handler has nothing for the element.
func TestContextMenuWireFallsBackWhenHandlerEmpty(t *testing.T) {
	t.Parallel()

	p := bftPage(t, `<div id="tab-0">tab</div>`)
	bftDraw(t, p, nil)

	p.Handle(Handlers{
		ContextMenuFor: func(string) []MenuItem { return nil },
	})

	rows := p.ContextMenuFor("editor")
	if len(rows) == 0 {
		t.Fatal("empty handler rows did not fall back to the built-in menu")
	}

	if rows[0].ID != "cut" {
		t.Fatalf("first fallback row = %q, want cut", rows[0].ID)
	}
}

// TestMenuWireActionGetsCustomID proves a non-built-in row id reaches the
// registered callback with the caller's context.
func TestMenuWireActionGetsCustomID(t *testing.T) {
	t.Parallel()

	p := bftPage(t, `<div id="tab-0">tab</div>`)
	bftDraw(t, p, nil)

	type menuCtxKey struct{}

	ctx := context.WithValue(context.Background(), menuCtxKey{}, "menu-wire")

	var (
		gotID  string
		gotCtx context.Context
	)

	p.Handle(Handlers{
		MenuAction: func(ctx context.Context, id string) error {
			gotCtx = ctx
			gotID = id

			return nil
		},
	})

	const want = "code-view:tab-close-all:1"
	if err := p.MenuAction(ctx, want); err != nil {
		t.Fatalf("MenuAction: %v", err)
	}

	if gotID != want {
		t.Fatalf("handler id = %q, want %q", gotID, want)
	}

	if gotCtx != ctx {
		t.Fatal("handler did not receive the caller's context")
	}
}

// TestContextMenuAddHandlersMergesMenu covers the partial-registration path:
// AddHandlers documents that non-nil callbacks join the existing set.
func TestContextMenuAddHandlersMergesMenu(t *testing.T) {
	t.Parallel()

	p := bftPage(t, `<div id="tab-0">tab</div>`)
	bftDraw(t, p, nil)

	var (
		asked bool
		acted string
	)

	p.AddHandlers(Handlers{
		ContextMenuFor: func(string) []MenuItem {
			asked = true

			return []MenuItem{
				{ID: "code-view:tab-close-all:1", Label: "Close All", Enabled: true},
			}
		},
		MenuAction: func(_ context.Context, id string) error {
			acted = id

			return nil
		},
	})

	rows := p.ContextMenuFor("tab-0")
	if !asked || len(rows) != 1 || rows[0].ID != "code-view:tab-close-all:1" {
		t.Errorf("AddHandlers dropped ContextMenuFor: asked=%v rows=%+v", asked, rows)
	}

	if err := p.MenuAction(context.Background(), "code-view:tab-close-all:1"); err != nil {
		t.Fatalf("MenuAction: %v", err)
	}

	if acted != "code-view:tab-close-all:1" {
		t.Errorf("AddHandlers dropped MenuAction: id=%q, want code-view:tab-close-all:1", acted)
	}
}
