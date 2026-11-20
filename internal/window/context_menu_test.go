package window

import (
	"context"
	"errors"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/clipboard"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/page"
	"io/fs"
	"testing"
	"testing/fstest"
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

// TestMenuClaimActsOnRowPress gives the open menu first refusal on a left
// press. A custom-chrome draggable strip runs before the menu in pointer(),
// so without this a press on a menu row inside the strip starts a window
// drag: the row never acts and the release lands a stray page click.
func TestMenuClaimActsOnRowPress(t *testing.T) {
	t.Parallel()

	app := &menuForScreen{
		menuScreen: &menuScreen{
			fakeScreen: &fakeScreen{width: 300, height: 300},
			items: []host.MenuItem{
				{ID: "do-thing", Label: "Do thing", Enabled: true},
			},
		},
		rows: map[string][]host.MenuItem{},
	}
	s := &shell{app: app, ctx: context.Background(), screenW: 300, screenH: 300}
	s.openMenu(10, 20)
	if !s.menu.open {
		t.Fatal("menu did not open")
	}

	mx, my, mw, _ := s.menuRect()
	cx, cy := int(mx+mw/2), int(my+menuRowH/2)
	if got := s.menuIndex(cx, cy); got != 0 {
		t.Fatalf("row centre index = %d, want 0", got)
	}

	handled, err := s.menuClaim(cx, cy, true, true)
	if err != nil || !handled {
		t.Fatalf("claim = %v, %v", handled, err)
	}

	if app.acted != "do-thing" {
		t.Fatalf("acted = %q, want do-thing", app.acted)
	}

	if s.menu.open {
		t.Fatal("menu stayed open after a row press")
	}

	if !s.mouseDown {
		t.Fatal("menu claim left mouseDown false, so the press would fire twice")
	}
}

// TestMenuClaimOutsideDismisses checks that a left press outside the open
// menu still dismisses it through the same first-refusal path, instead of
// reaching the window drag or the page.
func TestMenuClaimOutsideDismisses(t *testing.T) {
	t.Parallel()

	app := &menuScreen{fakeScreen: &fakeScreen{}, items: []host.MenuItem{
		{ID: "copy", Label: "Copy", Enabled: true},
	}}
	s := &shell{app: app, ctx: context.Background(), screenW: 300, screenH: 300}
	s.openMenu(10, 20)

	handled, err := s.menuClaim(250, 250, true, true)
	if err != nil || !handled {
		t.Fatalf("claim = %v, %v", handled, err)
	}

	if s.menu.open {
		t.Fatal("menu stayed open after an outside press")
	}
}

// TestMenuClaimIgnoresWithoutPress checks the claim stays quiet when the
// menu is closed or the press is not a fresh left edge, so normal pointer
// flow is untouched.
func TestMenuClaimIgnoresWithoutPress(t *testing.T) {
	t.Parallel()

	app := &menuScreen{fakeScreen: &fakeScreen{}, items: []host.MenuItem{
		{ID: "copy", Label: "Copy", Enabled: true},
	}}
	s := &shell{app: app, ctx: context.Background(), screenW: 300, screenH: 300}

	if handled, err := s.menuClaim(15, 25, true, true); err != nil || handled {
		t.Fatalf("closed menu claim = %v, %v", handled, err)
	}

	s.openMenu(10, 20)

	if handled, err := s.menuClaim(15, 25, false, true); err != nil || handled {
		t.Fatalf("held level claim = %v, %v", handled, err)
	}

	if !s.menu.open {
		t.Fatal("claim disturbed the menu without a fresh press")
	}
}

// TestMenuDockGeometryClicksRow checks that with the devtools dock open the
// menu draws and hit-tests in the same coordinates: a left click at the
// drawn centre of a row acts, and a click outside closes and consumes.
func TestMenuDockGeometryClicksRow(t *testing.T) {
	t.Parallel()

	app := &menuForScreen{
		menuScreen: &menuScreen{
			fakeScreen: &fakeScreen{width: 940, height: 800},
			items: []host.MenuItem{
				{ID: "do-thing", Label: "Do thing", Enabled: true},
				{ID: "do-other", Label: "Do other", Enabled: true},
				{ID: "do-more", Label: "Do more", Enabled: true},
			},
		},
		rows: map[string][]host.MenuItem{},
	}
	const screenW, screenH = 1280, 800
	s := &shell{app: app, ctx: context.Background(), screenW: screenW, screenH: screenH}
	s.dev.on = true
	s.dev.panel = devDockRect(screenW, screenH, s.devDockWidth())

	viewW, _ := s.viewSize()
	if s.dev.panel.X != float64(viewW) {
		t.Fatalf("panel X = %v, view width = %d", s.dev.panel.X, viewW)
	}

	// Right-click near the right and bottom edges of the content strip so
	// the menu must clamp inside the shrunken viewport.
	handled, err := s.devMenuPointer(viewW-10, 780, true, false)
	if err != nil || !handled || !s.menu.open {
		t.Fatalf("right click: handled=%v err=%v menu=%+v", handled, err, s.menu)
	}

	mx, my, mw, mh := s.menuRect()
	if mx+mw > float64(viewW) || my+mh > float64(screenH) {
		t.Fatalf("menu rect (%v, %v, %v, %v) escapes the content strip", mx, my, mw, mh)
	}

	for i := range s.menu.items {
		cx := int(mx + mw/2)
		cy := int(my + float64(i)*menuRowH + menuRowH/2)
		if got := s.menuIndex(cx, cy); got != i {
			t.Fatalf("row %d centre index = %d", i, got)
		}
	}

	cx := int(mx + mw/2)
	cy := int(my + menuRowH + menuRowH/2)
	handled, err = s.devMenuPointer(cx, cy, false, true)
	if err != nil || !handled {
		t.Fatalf("menu click: handled=%v err=%v", handled, err)
	}

	if s.menu.open {
		t.Fatal("menu stayed open after a row click")
	}

	if app.acted != "do-other" {
		t.Fatalf("acted = %q, want do-other", app.acted)
	}

	s.openMenu(10, 20)

	handled, err = s.devMenuPointer(viewW-2, screenH-2, false, true)
	if err != nil || !handled {
		t.Fatalf("outside click: handled=%v err=%v", handled, err)
	}

	if s.menu.open {
		t.Fatal("menu stayed open after an outside click")
	}

	if app.acted != "do-other" {
		t.Fatalf("outside click acted: %q", app.acted)
	}
}

func TestMenuOutsideClickCloses(t *testing.T) {
	t.Parallel()

	app := &menuScreen{fakeScreen: &fakeScreen{}, items: []host.MenuItem{
		{ID: "copy", Label: "Copy", Enabled: true},
	}}
	s := &shell{app: app, ctx: context.Background(), screenW: 300, screenH: 300}
	s.openMenu(10, 20)

	handled, err := s.menuClick(250, 250)
	if err != nil || !handled {
		t.Fatalf("click = %v, %v", handled, err)
	}

	if s.menu.open {
		t.Fatal("menu stayed open")
	}
}

func TestMenuEmptyRowsStayClosed(t *testing.T) {
	t.Parallel()

	app := &menuScreen{fakeScreen: &fakeScreen{}}
	s := &shell{app: app, ctx: context.Background(), screenW: 300, screenH: 300}
	s.openMenu(10, 20)

	if s.menu.open {
		t.Fatal("menu opened with no rows")
	}
}

func TestMenuPlainScreenStaysClosed(t *testing.T) {
	t.Parallel()

	s := &shell{app: &fakeScreen{}, ctx: context.Background(), screenW: 300, screenH: 300}
	s.openMenu(10, 20)

	if s.menu.open {
		t.Fatal("menu opened without a context menu")
	}
}

func TestMenuRectPullsInsideTheScreen(t *testing.T) {
	t.Parallel()

	app := &menuScreen{fakeScreen: &fakeScreen{}, items: []host.MenuItem{
		{ID: "copy", Label: "Copy", Enabled: true},
	}}
	s := &shell{app: app, ctx: context.Background(), screenW: 120, screenH: 80}
	s.openMenu(110, 70)

	x, y, w, h := s.menuRect()
	if x+w > 120 || y+h > 80 {
		t.Fatalf("rect = %v, %v, %v, %v", x, y, w, h)
	}

	if s.menuIndex(int(x)+1, int(y)+1) != 0 {
		t.Fatal("first row missed")
	}
}

// dropScreen is a fakeScreen that accepts drops.
type dropScreen struct {
	*fakeScreen
	calls int
	files []host.Drop
	err   error
}

func (s *dropScreen) Drop(_ context.Context, files []host.Drop) error {
	s.calls++
	s.files = files

	return s.err
}

func TestDropPassCallsScreenOnce(t *testing.T) {
	t.Parallel()

	screen := &dropScreen{fakeScreen: &fakeScreen{}}
	s := &shell{app: screen, ctx: context.Background()}

	fsys := fstest.MapFS{
		"a.txt": {Data: []byte("hello")},
		"b.png": {Data: []byte("bytes")},
	}

	if err := s.dropPass(fsys); err != nil {
		t.Fatalf("dropPass: %v", err)
	}

	if screen.calls != 1 {
		t.Fatalf("calls = %d, want 1", screen.calls)
	}

	if len(screen.files) != 2 {
		t.Fatalf("files = %d, want 2", len(screen.files))
	}

	first := screen.files[0]
	if first.Name != "a.txt" || first.Size != 5 || first.IsDir {
		t.Fatalf("first = %+v", first)
	}

	if first.Path != "" {
		t.Fatalf("path = %q, want empty without AbsPath", first.Path)
	}

	data, err := first.Read()
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != "hello" {
		t.Fatalf("data = %q", data)
	}
}

func TestDropPassKeepsDirectoryAsOneEntry(t *testing.T) {
	t.Parallel()

	screen := &dropScreen{fakeScreen: &fakeScreen{}}
	s := &shell{app: screen, ctx: context.Background()}

	fsys := fstest.MapFS{"docs": {Mode: fs.ModeDir | 0o755}}

	if err := s.dropPass(fsys); err != nil {
		t.Fatalf("dropPass: %v", err)
	}

	if len(screen.files) != 1 {
		t.Fatalf("files = %d, want 1", len(screen.files))
	}

	file := screen.files[0]
	if !file.IsDir || file.Name != "docs" {
		t.Fatalf("file = %+v", file)
	}
}

func TestDropPassSkipsPlainScreen(t *testing.T) {
	t.Parallel()

	s := &shell{app: &fakeScreen{}, ctx: context.Background()}

	if err := s.dropPass(fstest.MapFS{"a.txt": {Data: []byte("x")}}); err != nil {
		t.Fatalf("dropPass: %v", err)
	}
}

func TestDropPassEmptyFSCallsNothing(t *testing.T) {
	t.Parallel()

	screen := &dropScreen{fakeScreen: &fakeScreen{}}
	s := &shell{app: screen, ctx: context.Background()}

	if err := s.dropPass(fstest.MapFS{}); err != nil {
		t.Fatalf("dropPass: %v", err)
	}

	if screen.calls != 0 {
		t.Fatalf("calls = %d, want 0", screen.calls)
	}
}

func TestDropPassReturnsScreenError(t *testing.T) {
	t.Parallel()

	want := errors.New("drop failed")
	screen := &dropScreen{fakeScreen: &fakeScreen{}, err: want}
	s := &shell{app: screen, ctx: context.Background()}

	if err := s.dropPass(fstest.MapFS{"a.txt": {Data: []byte("x")}}); !errors.Is(err, want) {
		t.Fatalf("err = %v", err)
	}
}

// absFS adds AbsPath to the entries of a MapFS, the way Ebiten's desktop
// dropped-files system does.
type absFS struct{ fstest.MapFS }

func (f absFS) ReadDir(name string) ([]fs.DirEntry, error) {
	entries, err := f.MapFS.ReadDir(name)
	if err != nil {
		return nil, err
	}

	for i, entry := range entries {
		entries[i] = absEntry{entry}
	}

	return entries, nil
}

type absEntry struct{ fs.DirEntry }

func (e absEntry) AbsPath() string { return "/real/" + e.Name() }

func TestDropPassCarriesAbsPath(t *testing.T) {
	t.Parallel()

	screen := &dropScreen{fakeScreen: &fakeScreen{}}
	s := &shell{app: screen, ctx: context.Background()}

	fsys := absFS{MapFS: fstest.MapFS{"a.txt": {Data: []byte("x")}}}

	if err := s.dropPass(fsys); err != nil {
		t.Fatalf("dropPass: %v", err)
	}

	if got := screen.files[0].Path; got != "/real/a.txt" {
		t.Fatalf("path = %q", got)
	}
}
