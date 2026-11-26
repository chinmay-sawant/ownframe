package ui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/ui"

	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/driver"
	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/monitor"
	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/sqlkit"
	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/store"
)

// Window geometry shared by the shell, tests, and the drag guard.
const (
	DefaultWidth  = 1440
	DefaultHeight = 960
	MinWidth      = 960
	MinHeight     = 640

	titleBarHeight = 36
	railWidth      = 48
	statusHeight   = 24
	gridRowHeight  = 26
	gridHeadHeight = 28
	gutterWidth    = 52
)

// LaunchOptions configures NewApp.
type LaunchOptions struct {
	// DBPath is an optional SQLite path or PostgreSQL DSN. Empty opens the
	// seeded demo database.
	DBPath string
	// DataDir overrides the config directory; ":memory:" keeps state in memory.
	DataDir string
	// DemoDir holds the seeded demo database; empty uses the config dir.
	DemoDir string
	Width   int
	Height  int
	// Perf turns on page perf counters.
	Perf bool
}

// App owns all Quarry UI state. Every panel is a pure function of this state
// and is rebuilt through View.Update after each action.
type App struct {
	page *ownframe.Page
	view *ui.View
	ctx  context.Context

	width, height int
	dark          bool

	// Connection and schema.
	drv       driver.Driver
	cfg       driver.ConnectionConfig
	version   string
	connErr   string
	schema    []driver.SchemaNode
	sqlSchema sqlkit.Schema
	expanded  map[string]bool
	filter    string
	selNode   string

	// Layout.
	sidebarOpen   bool
	sidebarW      int
	editorH       int
	inspectorW    int
	activeRail    string
	prodUnlocked  bool
	splitDragging string

	// Workspace.
	tabs      []*workTab
	activeTab int
	nextTabID int

	// Overlays. Only one modal at a time; popovers listed separately.
	modal      string // "", "conn", "export", "import", "history", "settings", "palette", "preview", "confirm", "sqlpreview"
	modalErr   string
	modalNote  string
	menuOpen   string // "" or "conn" (title pill dropdown)
	ctxMenu    *contextMenu
	confirm    *confirmState
	pal        paletteState
	connDraft  connDraft
	connTest   connTestState
	mf         modalFocus
	exportOpts exportOpts
	importSt   importState

	// Editor popover.
	ac acState

	// Async query pipeline.
	doneCh    chan queryDone
	running   *runState
	lastQuery lastQueryInfo

	// Monitor.
	mon monitorState

	// Persistence and recents.
	prefs   *store.Store
	slow    *monitor.SlowLog
	lat     *monitor.LatencyTracker
	demoDir string

	// Input.
	focus              string // editor | grid | ""
	suppressSubmit     bool
	pressedEditor      bool
	edDrag             bool
	edClicks           int
	edClickAt          time.Time
	edClickOff         int
	edPX, edPY         float64
	edSumX, edSumY     float64
	find               findState
	nodeByID           map[string]driver.SchemaNode
	histQuery          string
	lastType           time.Time
	modsFn             func() (shift, ctrl bool)
	clock              func() time.Time
	lastClickID        string
	lastClickAt        time.Time
	toast              string
	frameHook          func(context.Context)
	regions            map[string]*region
	livePanes          map[string]bool
	thumbDrag          string
	syncing            bool
	maximizedFn        func() bool
	lastMax            bool
	imagePending       map[string][]byte
	defLimit           int
	confirmDestructive bool
	imageSent          map[string]bool
	toastAt            time.Time
}

// readMods reads Shift and Ctrl from the Ebiten window; page handlers carry no
// modifier state for letter and Enter keys.
func (a *App) readMods() (shift, ctrl bool) {
	if a.modsFn != nil {
		return a.modsFn()
	}
	return ebiten.IsKeyPressed(ebiten.KeyShift),
		ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)
}

// SetModifiers replaces the modifier source; tests use it to simulate chords.
func (a *App) SetModifiers(fn func() (shift, ctrl bool)) { a.modsFn = fn }

// SetFrameHook registers fn to run after every page tick (the window snapshot
// server answers requests from it).
func (a *App) SetFrameHook(fn func(context.Context)) { a.frameHook = fn }

// Page returns the mounted page for Run, Serve, and tests.
func (a *App) Page() *ownframe.Page { return a.page }

// Dark reports whether QuarryDark is active.
func (a *App) Dark() bool { return a.dark }

// NewApp builds Quarry, opens the demo or flagged database, and mounts the page.
func NewApp(ctx context.Context, opts LaunchOptions) (*App, error) {
	a := &App{
		ctx:                ctx,
		width:              orInt(opts.Width, DefaultWidth),
		height:             orInt(opts.Height, DefaultHeight),
		dark:               true,
		sidebarOpen:        true,
		sidebarW:           defaultSidebarW,
		editorH:            defaultEditorH,
		inspectorW:         defaultInspectorW,
		activeRail:         "explorer",
		expanded:           map[string]bool{},
		doneCh:             make(chan queryDone, doneChanBuffer),
		clock:              time.Now,
		defLimit:           thousandRows,
		confirmDestructive: true,
	}
	a.mon.pollEvery = defaultPollEvery
	a.demoDir = opts.DemoDir
	if err := a.initPrefs(ctx, opts.DataDir); err != nil {
		return nil, err
	}
	if p, ok := a.loadPref("theme"); ok && p == "light" {
		a.dark = false
	}
	a.restoreLayout()
	a.restoreSettings()
	if err := a.openInitial(ctx, opts.DBPath); err != nil {
		a.connErr = err.Error()
	}
	a.newQueryTab("")
	view, err := ui.Mount(ownframe.Config{
		Title:       "Quarry",
		Width:       a.width,
		Height:      a.height,
		MinWidth:    MinWidth,
		MinHeight:   MinHeight,
		Theme:       ThemeCSS(a.dark),
		LockView:    true,
		RetainState: true,
		Scrollbar:   ownframe.ScrollbarHidden,
	}, a.screen())
	if err != nil {
		return nil, err
	}
	a.view = view
	a.page = view.Page
	a.page.Window().SetPerf(opts.Perf)
	a.flushImages()
	a.view.Handle(a.handlers())
	a.page.SetTick(a.tickHook)
	return a, nil
}

func orInt(v, def int) int {
	if v > 0 {
		return v
	}
	return def
}

// Close stops background work and releases the driver and store.
func (a *App) Close() {
	a.stopMonitor()
	if a.running != nil {
		a.running.cancel()
	}
	if a.drv != nil {
		_ = a.drv.Close()
	}
	a.closePrefs()
}

// handlers wires every ownframe input callback.
func (a *App) handlers() ownframe.Handlers {
	return ownframe.Handlers{
		Click:          a.onClick,
		Hover:          a.onHoverAll,
		Press:          a.onPress,
		DragStart:      a.onDragStart,
		DragMove:       a.onDragMove,
		DragEnd:        a.onDragEnd,
		Wheel:          a.onWheel,
		KeyDown:        a.onKeyDown,
		Change:         a.onChange,
		Type:           a.onType,
		Backspace:      a.onBackspace,
		Copy:           a.onCopy,
		Cut:            a.onCut,
		Paste:          a.onPaste,
		Undo:           a.onUndo,
		SelectAll:      a.onSelectAll,
		Redo:           a.onRedo,
		Submit:         a.onSubmit,
		Drop:           a.onDrop,
		ContextMenuFor: a.contextMenuFor,
		MenuAction:     a.menuAction,
	}
}

// redraw rebuilds the tree and updates the page.
func (a *App) redraw(ctx context.Context) error {
	if a.view == nil {
		return nil
	}
	a.noteModalOpener()
	root := a.screen()
	a.flushImages()
	if err := a.view.Update(ctx, root); err != nil {
		return err
	}
	if err := a.syncGridView(ctx); err != nil {
		return err
	}
	if err := a.syncPanes(ctx); err != nil {
		return err
	}
	return a.syncModalFocus(ctx)
}

// syncGridView measures the laid-out grid body and, when it differs from the
// height the window was computed for, rebuilds once so the visible rows match
// the real viewport.
func (a *App) syncGridView(ctx context.Context) error {
	t := a.activeTabPtr()
	if t == nil || t.grid == nil || a.syncing {
		return nil
	}
	for _, b := range a.page.Boxes() {
		if b.ID != "grid" {
			continue
		}
		// Whole rows only: a partly visible row would paint past the clip.
		h := (int(b.H) - gridHeadHeight) / gridRowHeight * gridRowHeight
		if h <= 0 || h == t.grid.viewH {
			return nil
		}
		t.grid.viewH = h
		a.syncing = true
		defer func() { a.syncing = false }()
		return a.view.Update(ctx, a.screen())
	}
	return nil
}

// overlaysOpen reports whether a modal, popover, or menu owns the window. The
// drag guard yields while it is true.
func (a *App) overlaysOpen() bool {
	return a.modal != "" || a.menuOpen != "" || a.ac.open
}

// Draggable is the window drag guard: only the title strip drags, never while
// an overlay is open and never over the controls on its right edge.
func (a *App) Draggable(x, y int) bool {
	if a.overlaysOpen() {
		return false
	}
	return x >= 0 && y >= 0 && y < titleBarHeight && x < a.width-3*44
}

func (a *App) toggleTheme(ctx context.Context) error { return a.setThemeDark(ctx, !a.dark) }

// setThemeDark switches theme, saves the choice, and repaints.
func (a *App) setThemeDark(ctx context.Context, dark bool) error {
	if dark == a.dark {
		return nil
	}
	a.dark = dark
	name := "light"
	if a.dark {
		name = "dark"
	}
	a.savePref("theme", name)
	if err := a.page.SetTheme(ThemeCSS(a.dark)); err != nil {
		return err
	}
	return a.redraw(ctx)
}

// ---- tick and async pipeline ----

// tickHook is the page SetTick callback: it drains completions and repaints
// only when something changed.
func (a *App) tickHook(ctx context.Context) error {
	if a.frameHook != nil {
		defer a.frameHook(ctx)
	}
	if a.onTick(ctx) {
		return a.redraw(ctx)
	}
	return nil
}

// maximized reports the desktop window state (tests replace maximizedFn).
func (a *App) maximized() bool {
	if a.maximizedFn != nil {
		return a.maximizedFn()
	}
	return a.page != nil && a.page.Window().Maximized()
}

// trackWindow notices a window resize or a maximize/restore and reports
// whether the layout inputs changed. The grid re-measures after a resize.
func (a *App) trackWindow() bool {
	changed := false
	if a.page != nil {
		if w, h := a.page.Size(); w > 0 && h > 0 && (w != a.width || h != a.height) {
			a.width, a.height = w, h
			changed = true
			if t := a.activeTabPtr(); t != nil && t.grid != nil {
				t.grid.viewH = 0
			}
		}
	}
	if m := a.maximized(); m != a.lastMax {
		a.lastMax = m
		changed = true
	}
	return changed
}

// onTick drains at most a bounded batch of query completions and monitor
// snapshots, and refreshes the running timer. It reports whether the UI is
// dirty; an idle app returns false so the window can park.
func (a *App) onTick(ctx context.Context) bool {
	dirty := false
	for i := 0; i < 8; i++ {
		select {
		case res := <-a.doneCh:
			a.applyQueryDone(res)
			dirty = true
			continue
		default:
		}
		break
	}
	if a.drainMonitor() {
		dirty = true
	}
	if a.takeChartHover() {
		dirty = true
	}
	if a.drainConnTest() {
		dirty = true
	}
	if a.running != nil {
		now := a.clock()
		if now.Sub(a.running.shown) >= 100*time.Millisecond {
			a.running.shown = now
			dirty = true
		}
	}
	if a.trackWindow() {
		dirty = true
	}
	if a.toast != "" && a.clock().Sub(a.toastAt) > 3*time.Second {
		a.toast = ""
		dirty = true
	}
	return dirty
}

func (a *App) notify(msg string) {
	a.toast = msg
	a.toastAt = a.clock()
}

// ---- key handling ----

func (a *App) onType(ctx context.Context, text string) error {
	a.lastType = a.clock()
	if a.editorFocused() {
		return a.editorType(ctx, text)
	}
	return nil
}

func (a *App) onBackspace(ctx context.Context) error {
	if a.editorFocused() {
		return a.editorBackspace(ctx)
	}
	return nil
}

func (a *App) onCopy(ctx context.Context) (string, bool, error) {
	if a.editorFocused() {
		if ed := a.activeTabPtr().ed; ed.hasSelection() {
			return ed.selectionText(), true, nil
		}
		return "", false, nil
	}
	if t := a.activeTabPtr(); t != nil && t.grid != nil && a.focus == "grid" && a.modal == "" && a.page.Form().FocusID() == "" {
		if txt := a.gridCopyText(t); txt != "" {
			return txt, true, nil
		}
	}
	return "", false, nil
}

func (a *App) onCut(ctx context.Context) (string, bool, error) {
	if a.editorFocused() {
		t := a.activeTabPtr()
		if t.ed.hasSelection() {
			s := t.ed.selectionText()
			t.ed.insert("")
			t.dirty = true
			t.ed.ensureVisible(a.editorRows())
			return s, true, a.redraw(ctx)
		}
	}
	return "", false, nil
}

func (a *App) onPaste(ctx context.Context, text string) error {
	if a.editorFocused() {
		return a.editorType(ctx, text)
	}
	return nil
}

func (a *App) onUndo(ctx context.Context) error {
	if a.editorFocused() {
		a.activeTabPtr().ed.undoStep()
		return a.redraw(ctx)
	}
	return nil
}

func (a *App) onRedo(ctx context.Context) error {
	if a.editorFocused() {
		a.activeTabPtr().ed.redoStep()
		return a.redraw(ctx)
	}
	return nil
}

func (a *App) onSelectAll(ctx context.Context) error {
	if a.editorFocused() {
		ed := a.activeTabPtr().ed
		ed.anchor, ed.caret = 0, len(ed.text)
		return a.redraw(ctx)
	}
	return nil
}

// onKeyDown handles global shortcuts. Letter and Enter chords arrive without
// a modifier name, so the Ebiten modifier state decides.
func (a *App) onKeyDown(ctx context.Context, key string) error {
	shift, ctrl := a.readMods()
	if key == "escape" {
		return a.escape(ctx)
	}
	if a.modal == "palette" {
		return a.paletteKey(ctx, key)
	}
	if a.ac.open && !ctrl {
		if handled, err := a.autocompleteKey(ctx, key); handled || err != nil {
			return err
		}
	}
	if a.find.open && key == "enter" && a.page.Form().FocusID() == "find-input" {
		a.suppressSubmit = true
		if shift {
			return a.findStep(ctx, -1)
		}
		return a.findStep(ctx, 1)
	}
	if key == "f3" {
		return a.findStep(ctx, 1)
	}
	if a.editorFocused() && (!ctrl || strings.HasPrefix(key, "ctrl+")) {
		if handled, err := a.editorKey(ctx, key, shift); handled || err != nil {
			return err
		}
	}
	if ctrl {
		switch key {
		case "k", "p":
			return a.openPalette(ctx)
		case "enter":
			return a.runFromKey(ctx, shift)
		case "space":
			return a.openAutocomplete(ctx)
		case "n":
			a.newQueryTab("")
			return a.redraw(ctx)
		case "w":
			return a.closeTab(ctx, a.activeTab)
		case "t":
			if shift {
				return a.toggleTheme(ctx)
			}
		case "tab":
			return a.cycleTab(ctx, 1)
		case "s":
			return a.saveSnippet(ctx)
		case "f":
			return a.openFind(ctx)
		case "slash":
			if a.editorFocused() {
				t := a.activeTabPtr()
				t.ed.toggleComment()
				t.dirty = true
				return a.redraw(ctx)
			}
		}
		return nil
	}
	switch key {
	case "f5":
		return a.runFromKey(ctx, false)
	case "f6":
		return a.runFromKey(ctx, true)
	case "f7":
		return a.startExplain(ctx, false)
	}
	if a.modal == "" {
		if t := a.activeTabPtr(); t != nil && t.grid != nil && a.focus == "grid" && a.page.Form().FocusID() == "" {
			return a.gridKey(ctx, t, key, shift)
		}
	}
	return nil
}

func (a *App) escape(ctx context.Context) error {
	switch {
	case a.find.open && a.modal == "":
		return a.closeFind(ctx)
	case a.ac.open:
		a.ac.open = false
	case a.ctxMenu != nil:
		a.ctxMenu = nil
	case a.menuOpen != "":
		a.menuOpen = ""
	case a.modal != "":
		a.modal = ""
		a.modalErr = ""
		_ = a.page.Form().Focus(ctx, "")
	case a.running != nil:
		a.cancelRunning()
		return a.redraw(ctx)
	default:
		if t := a.activeTabPtr(); t != nil && t.editing != nil {
			t.editing = nil
		} else if t != nil && t.inspectorOpen {
			t.inspectorOpen = false
		} else {
			return nil
		}
	}
	return a.redraw(ctx)
}

// errNoTab is returned by actions that need an active tab.
var errNoTab = errors.New("quarry: no active tab")

func (a *App) onSubmit(ctx context.Context) error {
	if a.modal != "" && a.modal != "palette" {
		return a.modalEnter(ctx)
	}
	if a.suppressSubmit {
		a.suppressSubmit = false
		return nil
	}
	if t := a.activeTabPtr(); t != nil && t.editing != nil {
		return a.commitInlineEdit(ctx, t)
	}
	return nil
}

// onChange runs after any control edit.
func (a *App) onChange(ctx context.Context, box ownframe.Box) error {
	switch {
	case box.ID == "sb-filter":
		a.filter = a.page.Form().FormValue("sb-filter")
		return a.redraw(ctx)
	case box.ID == "editor-input":
		return a.editorChanged(ctx)
	case box.ID == "grid-filter":
		if t := a.activeTabPtr(); t != nil && t.grid != nil {
			t.quick = a.page.Form().FormValue("grid-filter")
			t.grid.rebuild(t.quick, t.sortCol, t.sortDir)
		}
		return a.redraw(ctx)
	case box.ID == "find-input":
		a.find.query = a.page.Form().FormValue("find-input")
		if t := a.activeTabPtr(); t != nil && t.ed != nil {
			t.ed.caret = min(t.ed.anchor, t.ed.caret)
		}
		return a.findStep(ctx, 1)
	case box.ID == "pal-input":
		a.pal.query = a.page.Form().FormValue("pal-input")
		a.pal.active, a.pal.top = 0, 0
		return a.redraw(ctx)
	case box.ID == "hist-search":
		a.histQuery = a.page.Form().FormValue("hist-search")
		return a.redraw(ctx)
	case strings.HasPrefix(box.ID, "tbl-filter"):
		return nil
	case strings.HasPrefix(box.ID, "cd-"):
		return a.connFieldChanged(ctx, box.ID)
	}
	return nil
}

// stamp formats a duration for the status bar.
func fmtDur(d time.Duration) string {
	switch {
	case d < time.Millisecond:
		return fmt.Sprintf("%d us", d.Microseconds())
	case d < time.Second:
		return fmt.Sprintf("%.1f ms", float64(d.Microseconds())/msPerSecond)
	default:
		return fmt.Sprintf("%.2f s", d.Seconds())
	}
}
