package window

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/clipboard"
	"github.com/chinmay-sawant/ownframe/internal/host"
)

// menuForScreen is a menuScreen that names rows for one element and records
// the custom action the window runs.
type menuForScreen struct {
	*menuScreen
	rows  map[string][]host.MenuItem
	asked string
	acted string
}

func (f *menuForScreen) ContextMenuFor(id string) []host.MenuItem {
	f.asked = id

	return f.rows[id]
}

func (f *menuForScreen) MenuAction(_ context.Context, id string) error {
	f.acted = id

	return nil
}

func TestMenuPasteFromRow(t *testing.T) {
	clipboard.UseMemory(true)
	clipboard.Write("hello\n")

	app := &menuScreen{fakeScreen: &fakeScreen{width: 300, height: 300}, items: []host.MenuItem{
		{ID: "cut", Label: "Cut", Enabled: false},
		{ID: "paste", Label: "Paste", Enabled: true},
	}}
	s := &shell{app: app, ctx: context.Background(), screenW: 300, screenH: 300}
	s.openMenu(10, 20)

	if !s.menu.open || len(s.menu.items) != 2 {
		t.Fatalf("menu = %+v", s.menu)
	}

	row := s.menuIndex(15, 20+menuRowH+5)
	if row != 1 {
		t.Fatalf("row = %d", row)
	}

	handled, err := s.menuClick(15, 20+menuRowH+5)
	if err != nil || !handled {
		t.Fatalf("click = %v, %v", handled, err)
	}

	if app.pasted != "hello" {
		t.Fatalf("pasted = %q", app.pasted)
	}

	if s.menu.open {
		t.Fatal("menu stayed open")
	}
}

func TestMenuSkipsDisabledRow(t *testing.T) {
	t.Parallel()

	app := &menuScreen{fakeScreen: &fakeScreen{}, items: []host.MenuItem{
		{ID: "undo", Label: "Undo", Enabled: false},
	}}
	s := &shell{app: app, ctx: context.Background(), screenW: 300, screenH: 300}
	s.openMenu(10, 20)

	handled, err := s.menuClick(15, 25)
	if err != nil || !handled {
		t.Fatalf("click = %v, %v", handled, err)
	}

	if app.pasted != "" || s.menu.open {
		t.Fatal("disabled row acted or kept the menu")
	}
}

func TestEscapeClosesMenu(t *testing.T) {
	t.Parallel()

	app := &menuScreen{fakeScreen: &fakeScreen{}, items: []host.MenuItem{
		{ID: "copy", Label: "Copy", Enabled: true},
	}}
	s := &shell{app: app, ctx: context.Background(), screenW: 300, screenH: 300}
	s.openMenu(5, 5)

	if err := s.escape(); err != nil {
		t.Fatal(err)
	}

	if s.menu.open {
		t.Fatal("menu stayed open after escape")
	}
}

func TestMenuContextMenuForGetsHitIDAndAction(t *testing.T) {
	t.Parallel()

	const rowID = "code-view:tab-close:0:1"

	app := &menuForScreen{
		menuScreen: &menuScreen{
			fakeScreen: &fakeScreen{
				width:  300,
				height: 300,
				boxes:  []layout.Box{{ID: "tab-1", Tag: "button", X: 0, Y: 0, W: 100, H: 40}},
			},
		},
		rows: map[string][]host.MenuItem{
			"tab-1": {{ID: rowID, Label: "Close", Enabled: true}},
		},
	}
	s := &shell{app: app, ctx: context.Background(), screenW: 300, screenH: 300}
	s.openMenu(10, 20)

	if app.asked != "tab-1" {
		t.Fatalf("ContextMenuFor id = %q, want tab-1", app.asked)
	}

	if !s.menu.open || len(s.menu.items) != 1 || s.menu.items[0].ID != rowID {
		t.Fatalf("menu = %+v", s.menu)
	}

	handled, err := s.menuClick(15, 25)
	if err != nil || !handled {
		t.Fatalf("click = %v, %v", handled, err)
	}

	if app.acted != rowID {
		t.Fatalf("MenuAction id = %q, want %q", app.acted, rowID)
	}
}

func TestMenuContextMenuForEmptyFallsBack(t *testing.T) {
	t.Parallel()

	app := &menuForScreen{
		menuScreen: &menuScreen{
			fakeScreen: &fakeScreen{
				width:  300,
				height: 300,
				boxes:  []layout.Box{{ID: "editor", Tag: "div", X: 0, Y: 0, W: 200, H: 200}},
			},
			items: []host.MenuItem{{ID: "copy", Label: "Copy", Enabled: true}},
		},
	}
	s := &shell{app: app, ctx: context.Background(), screenW: 300, screenH: 300}
	s.openMenu(10, 20)

	if app.asked != "editor" {
		t.Fatalf("ContextMenuFor id = %q, want editor", app.asked)
	}

	if !s.menu.open || len(s.menu.items) != 1 || s.menu.items[0].ID != "copy" {
		t.Fatalf("menu = %+v", s.menu)
	}

	handled, err := s.menuClick(15, 25)
	if err != nil || !handled {
		t.Fatalf("click = %v, %v", handled, err)
	}

	if app.acted != "" {
		t.Fatalf("built-in row reached MenuAction: %q", app.acted)
	}
}

func TestMenuContextMenuForWithoutHitPassesEmptyID(t *testing.T) {
	t.Parallel()

	app := &menuForScreen{
		menuScreen: &menuScreen{
			fakeScreen: &fakeScreen{width: 300, height: 300},
		},
		rows: map[string][]host.MenuItem{
			"": {{ID: "code-view:tab-close-all:0", Label: "Close All", Enabled: true}},
		},
	}
	s := &shell{app: app, ctx: context.Background(), screenW: 300, screenH: 300}
	s.openMenu(10, 20)

	if !s.menu.open || len(s.menu.items) != 1 {
		t.Fatalf("menu = %+v", s.menu)
	}
}
