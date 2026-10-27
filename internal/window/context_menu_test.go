package window

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/clipboard"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/page"
)

func TestDevToolsRightClickFileTabAndCloseAll(t *testing.T) {
	ctx := context.Background()
	p, err := page.New(page.Config{HTML: `<div style="position:fixed;top:0;left:0"><button id="tab-0" style="width:100px;height:30px"><span>README.md</span></button></div>`, Width: 300, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	var action string
	p.Handle(page.Handlers{
		ContextMenuFor: func(id string) []page.MenuItem {
			if id != "tab-0" {
				return nil
			}
			return []page.MenuItem{{ID: "code-view:tab-close-all:0", Label: "Close All", Enabled: true}}
		},
		MenuAction: func(_ context.Context, id string) error { action = id; return nil },
	})
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	s := &shell{app: p, ctx: ctx, screenW: 300, screenH: 200}
	s.dev.on = true
	s.dev.panel = devRect{X: 200, Y: 0, W: 100, H: 200}
	handled, err := s.devMenuPointer(20, 15, true, false)
	if err != nil || !handled || !s.menu.open || len(s.menu.items) != 1 || s.menu.items[0].Label != "Close All" {
		t.Fatalf("right click: handled=%v err=%v menu=%+v", handled, err, s.menu)
	}
	handled, err = s.devMenuPointer(25, 20, false, true)
	if err != nil || !handled || action != "code-view:tab-close-all:0" || s.menu.open {
		t.Fatalf("menu click: handled=%v err=%v action=%q menu=%+v", handled, err, action, s.menu)
	}
	handled, err = s.devMenuPointer(250, 20, true, false)
	if err != nil || handled || s.menu.open {
		t.Fatalf("dock right click: handled=%v err=%v menu=%+v", handled, err, s.menu)
	}
}

// menuForScreen is a menuScreen that names rows for one element and records
// the custom action the window runs.
type menuForScreen struct {
	*menuScreen
	rows  map[string][]host.MenuItem
	asked string
	acted string
}

type menuAtScreen struct {
	*menuForScreen
	called bool
}

func (f *menuAtScreen) ContextMenuAt(_, _ float64) []host.MenuItem {
	f.called = true
	return []host.MenuItem{{ID: "close-all", Label: "Close All", Enabled: true}}
}

type menuPresentScreen struct {
	*menuScreen
	items []host.MenuItem
}

func (m *menuPresentScreen) PresentMenu(_, _ int, items []host.MenuItem) bool {
	m.items = append([]host.MenuItem(nil), items...)

	return true
}

func TestPresentedMenuSkipsShellChrome(t *testing.T) {
	t.Parallel()

	app := &menuPresentScreen{menuScreen: &menuScreen{
		fakeScreen: &fakeScreen{width: 300, height: 300},
		items:      []host.MenuItem{{ID: "close", Label: "Close", Enabled: true}},
	}}
	s := &shell{app: app, ctx: context.Background(), screenW: 300, screenH: 300}
	s.openMenu(12, 18)

	if s.menu.open {
		t.Fatal("shell menu opened after the page claimed the rows")
	}

	if len(app.items) != 1 || app.items[0].ID != "close" {
		t.Fatalf("presented rows = %+v", app.items)
	}
}

func TestMenuUsesPageHitTesting(t *testing.T) {
	app := &menuAtScreen{menuForScreen: &menuForScreen{menuScreen: &menuScreen{fakeScreen: &fakeScreen{width: 300, height: 300}}}}
	s := &shell{app: app, screenW: 300, screenH: 300}
	s.openMenu(10, 20)
	if !app.called || !s.menu.open || s.menu.items[0].ID != "close-all" {
		t.Fatalf("called = %v, menu = %+v", app.called, s.menu)
	}
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

	if app.pasted != "hello\n" {
		t.Fatalf("pasted = %q", app.pasted)
	}

	if s.menu.open {
		t.Fatal("menu stayed open")
	}
}

func TestClipboardShortcutKeepsMultilineText(t *testing.T) {
	clipboard.UseMemory(true)
	clipboard.Write("first\nsecond\n")
	app := &menuScreen{fakeScreen: &fakeScreen{}}
	s := &shell{app: app, ctx: context.Background()}
	if err := s.applyChord(chordPaste); err != nil {
		t.Fatal(err)
	}
	if app.pasted != "first\nsecond\n" {
		t.Fatalf("pasted = %q", app.pasted)
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
