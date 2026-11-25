package main

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	excelPlacement = "excel"
	doubleClickGap = 450 * time.Millisecond
	rowsPerNotch   = 3
	colsPerNotch   = 1
)

// dragKind names what the pointer is doing while a gesture is claimed.
type dragKind int

const (
	dragNone dragKind = iota
	dragCells
	dragColHeader
	dragRowHeader
	dragResizeCol
	dragResizeRow
	dragVThumb
	dragHThumb
	dragRefRange
	dragFill
)

// app owns the spreadsheet state. Every panel is a pure function of this
// state and the workbook; redraw rebuilds the tree and View.Update patches it.
type app struct {
	// bg yields the app's root context. It is a field so the many state
	// helpers that hold no context share it; the window handlers have none to
	// thread through the shared redraw and edit paths.
	bg func() context.Context
	// update refreshes the View under bg; redraw calls it with no context.
	update func() error
	// focus moves keyboard focus to a field id (empty clears it) under bg.
	focus func(id string) error
	view  *ui.View
	page  *ownframe.Page
	wb    *Workbook
	dark  bool
	web   bool

	active CellRef
	sel    Range // always contains active
	anchor CellRef

	editing     bool
	editBuf     string
	editFromBar bool
	editOrig    string
	acSel       int      // highlighted row of the function autocomplete list
	hoverNote   *CellRef // commented cell under the pointer

	scrollRow, scrollCol int // first visible non-frozen row/col

	ribbonTab       string
	ribbonCollapsed bool
	overlay         string // "" or the id of the open dialog, menu or popover

	cutRange  *Range // source of a pending cut
	copyRange *Range // source of the last copy

	// Pointer gesture state.
	drag       dragKind
	dragPressX float64
	dragPressY float64
	dragAccX   float64
	dragAccY   float64
	fillTo     CellRef // cell under the pointer during a fill-handle drag
	fillSet    bool
	dragIndex  int // header index or original size being resized
	dragOrig   int
	dragMoved  bool
	refStart   int // byte offset in editBuf where a dragged reference begins
	refEnd     int
	refActive  bool // the buffer still ends with a point-and-click reference
	lastPress  time.Time
	pressSeen  time.Time // last grid Press, so a Click that follows is not handled twice
	lastCell   CellRef
	wheelRow   float64
	wheelCol   float64

	// modsFn reports Shift and Ctrl. Tests replace it; the web preview has no
	// keyboard modifier source, so it reports neither.
	modsFn func() (shift, ctrl bool)

	lastErr error

	x overlayState // ribbon, dialogs, menus and charts (overlays.go)
}

func newApp() (*app, error) {
	a := &app{
		bg:        context.Background,
		wb:        NewWorkbook(),
		ribbonTab: "Home",
	}
	a.sel = Range{Min: a.active, Max: a.active}
	a.modsFn = a.readMods
	a.focus = func(id string) error { return a.page.Form().Focus(a.bg(), id) }
	a.update = func() error { return a.view.Update(a.bg(), a.screen()) }
	view, err := ui.Mount(ownframe.Config{
		Title:       "Quarterly_Model.xlsx - Excel",
		Width:       windowWidth,
		Height:      windowHeight,
		MinWidth:    minWinWidth,
		MinHeight:   minWinHeight,
		LockView:    true,
		Theme:       a.themeSheet(),
		DevTools:    false,
		RetainState: true,
		Scrollbar:   ownframe.ScrollbarHidden,
	}, a.screen())
	if err != nil {
		return nil, err
	}
	a.view = view
	a.page = view.Page
	a.page.Scroll().SetAllowScroll(false)
	view.Handle(ownframe.Handlers{
		Click:     a.handleClick,
		Press:     a.handlePress,
		DragStart: a.handleDragStart,
		DragMove:  a.handleDragMove,
		DragEnd:   a.handleDragEnd,
		Wheel:     a.handleWheel,
		KeyDown:   a.handleKey,
		Change:    a.handleChange,
		Type:      a.handleType,
		Backspace: a.handleBackspace,
		Copy:      a.handleCopy,
		Cut:       a.handleCut,
		Paste:     a.handlePaste,
		SelectAll: a.handleSelectAll,
		Undo:      a.handleUndo,
		Redo:      a.handleRedo,

		Hover:          a.handleHover,
		LongPress:      a.handleLongPress,
		ContextMenuFor: a.rightClick,
		MenuAction:     a.menuAction,
		Drop:           a.handleDrop,
	})
	return a, nil
}

func (a *app) themeSheet() string {
	if a.dark {
		return ExcelDark
	}
	return ExcelLight
}

func (a *app) sheet() *Sheet { return a.wb.Sheet() }

// readMods reads Shift and Ctrl from the Ebiten window. The page handlers
// carry no modifier state for pointer and arrow-key events.
func (a *app) readMods() (shift, ctrl bool) {
	if a.web {
		return false, false
	}
	shift = ebiten.IsKeyPressed(ebiten.KeyShift)
	ctrl = ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)
	return shift, ctrl
}

func (a *app) mods() (shift, ctrl bool) {
	if a.modsFn == nil {
		return false, false
	}
	return a.modsFn()
}

// overlaysOpen reports whether a dialog, menu or popover owns the window. The
// draggable title strip yields while it is true.
func (a *app) overlaysOpen() bool { return a.overlay != "" }

// screen composes the full-window tree.
func (a *app) screen() *ui.Element {
	root := ui.Div().ID("xl-shell").Class("xl-shell").Children(
		a.titleBar(),
		ui.Div().ID("xl-ribbon-slot").Class("xl-ribbon-slot").
			CSSRaw("height", pxStr(a.ribbonHeight())).Children(a.ribbon()),
		a.formulaBar(),
		a.gridView(),
		a.sheetBar(),
		a.statusBar(),
	)
	root.Children(a.overlays()...)
	return root
}

// redraw rebuilds the tree after any state change, then pushes live values
// into the inputs the user is not currently typing in.
func (a *app) redraw() {
	if a.view == nil {
		return
	}
	if err := a.update(); err != nil {
		a.lastErr = err
		return
	}
	a.syncInputs()
}

func (a *app) toggleTheme() {
	a.dark = !a.dark
	if err := a.page.SetTheme(a.themeSheet()); err != nil {
		a.lastErr = err
	}
	a.redraw()
}

func (a *app) setOverlay(id string) {
	if a.editing && id != "" {
		a.commitEdit()
	}
	a.overlay = id
	a.redraw()
}

func (a *app) ribbonHeight() int {
	if a.ribbonCollapsed {
		return ribbonHeightMin
	}
	return ribbonHeightFull
}

// handleClick routes data-action clicks. Ribbon and overlay actions go to the
// hooks that agent C's files own.
func (a *app) handleClick(ctx context.Context, box ownframe.Box) error {
	act := box.Action
	switch {
	case act == "win-min":
		a.page.Window().Minimize()
		return nil
	case act == "win-max":
		a.page.Window().ToggleMaximize()
		return nil
	case act == "win-close":
		a.page.Window().Close()
		return nil
	case act == "theme:toggle":
		a.toggleTheme()
		return nil
	case act == "qat:save":
		a.setOverlay("save")
		return nil
	case act == "qat:undo":
		a.undo()
		return nil
	case act == "qat:redo":
		a.redo()
		return nil
	case act == "search:open":
		a.setOverlay("command")
		return nil
	case act == "fx:cancel":
		a.cancelEdit()
		return nil
	case act == "fx:commit":
		a.commitEdit()
		a.moveActive(1, 0, false, false)
		return nil
	case act == "fx:insert":
		a.setOverlay("insert-function")
		return nil
	case strings.HasPrefix(act, "sheet:"):
		return a.sheetAction(strings.TrimPrefix(act, "sheet:"))
	}
	if a.clickAutocomplete(box) {
		return nil
	}
	if time.Since(a.pressSeen) >= 300*time.Millisecond && a.overlay == "" && a.clickFilterChevron(box.ID) {
		return nil
	}
	if a.clickCellBox(box) {
		return nil
	}
	if handled, err := a.overlayClick(ctx, box); handled || err != nil {
		return err
	}
	if handled, err := a.ribbonClick(ctx, box); handled || err != nil {
		return err
	}
	return nil
}

func (a *app) handleChange(ctx context.Context, box ownframe.Box) error {
	if handled, err := a.overlayChange(ctx, box); handled || err != nil {
		return err
	}
	switch box.ID {
	case "xl-editor":
		a.refActive = false
		a.acSel = 0
		a.editBuf = a.page.Form().FormValue("xl-editor")
		a.page.Form().SetFormValue("xl-fx", a.editBuf)
		a.redraw()
	case "xl-fx":
		if !a.editing {
			a.editing = true
			a.editFromBar = true
			a.editOrig = a.sheet().Raw(a.active.Row, a.active.Col)
		}
		a.refActive = false
		a.acSel = 0
		a.editBuf = a.page.Form().FormValue("xl-fx")
		a.page.Form().SetFormValue("xl-editor", a.editBuf)
		a.redraw()
	}
	return nil
}

// fieldFocused reports whether any text field owns the keyboard.
func (a *app) fieldFocused() bool { return a.page.Form().FocusID() != "" }

// handleType starts an edit when typing arrives with no field focused. The
// editor mounts empty and focused, and the page then inserts the typed text
// into it, so the first character is never doubled.
func (a *app) handleType(ctx context.Context, text string) error {
	if a.editing || a.fieldFocused() || a.overlay != "" || text == "" {
		return nil
	}
	a.beginEdit("")
	return nil
}

func (a *app) handleBackspace(ctx context.Context) error {
	if a.editing || a.fieldFocused() || a.overlay != "" {
		return nil
	}
	a.clearSelection()
	return nil
}

func (a *app) handleCopy(ctx context.Context) (string, bool, error) {
	if a.editing || a.overlay != "" {
		return "", false, nil
	}
	return a.copyText(false), true, nil
}

func (a *app) handleCut(ctx context.Context) (string, bool, error) {
	if a.editing || a.overlay != "" {
		return "", false, nil
	}
	return a.copyText(true), true, nil
}

func (a *app) handlePaste(ctx context.Context, text string) error {
	if a.editing || a.fieldFocused() || a.overlay != "" || text == "" {
		return nil
	}
	a.pasteText(text)
	return nil
}

func (a *app) handleSelectAll(ctx context.Context) error {
	if a.editing || a.fieldFocused() {
		return nil
	}
	a.selectRange(Range{Max: CellRef{Row: MaxRows - 1, Col: MaxCols - 1}})
	return nil
}

func (a *app) handleUndo(ctx context.Context) error {
	if a.editing || a.fieldFocused() {
		return nil
	}
	a.undo()
	return nil
}

func (a *app) handleRedo(ctx context.Context) error {
	if a.editing || a.fieldFocused() {
		return nil
	}
	a.redo()
	return nil
}

// syncInputs pushes the Name Box and formula input values into the page's
// live form state, skipping the one that currently has focus.
func (a *app) syncInputs() {
	focus := a.page.Form().FocusID()
	form := a.page.Form()
	if focus != "xl-name" {
		if v := a.nameBoxText(); form.FormValue("xl-name") != v {
			form.SetFormValue("xl-name", v)
		}
	}
	if focus != "xl-fx" {
		if v := a.formulaText(); form.FormValue("xl-fx") != v {
			form.SetFormValue("xl-fx", v)
		}
	}
	if a.editing && !a.editFromBar && focus != "xl-editor" {
		if form.FormValue("xl-editor") != a.editBuf {
			form.SetFormValue("xl-editor", a.editBuf)
		}
	}
}

// formulaText is what the formula input shows: the live edit buffer while
// editing, otherwise the active cell's raw content.
func (a *app) formulaText() string {
	if a.editing {
		return a.editBuf
	}
	return a.sheet().Raw(a.active.Row, a.active.Col)
}

func (a *app) nameBoxText() string {
	if a.drag == dragCells && a.dragMoved {
		rows := a.sel.Max.Row - a.sel.Min.Row + 1
		cols := a.sel.Max.Col - a.sel.Min.Col + 1
		return itoa(rows) + "R x " + itoa(cols) + "C"
	}
	if a.sel.Min == a.sel.Max {
		return a.sel.Min.A1()
	}
	return a.sel.String()
}

// clickCellBox selects the cell under a click that arrived without a
// preceding Press, which is how the web preview delivers pointer input. The
// window path claims the press first, so it never reaches here twice.
func (a *app) clickCellBox(box ownframe.Box) bool {
	if a.overlay != "" || time.Since(a.pressSeen) < 300*time.Millisecond {
		return false
	}
	ref, ok := cellIDRef(box.ID)
	if !ok {
		return false
	}
	a.selectCell(a.mergeAnchor(ref))
	return true
}

// mergeAnchor maps a cell inside a merged range to the range's top-left cell.
func (a *app) mergeAnchor(c CellRef) CellRef {
	if m, ok := a.sheet().MergeAt(c.Row, c.Col); ok {
		return m.Min
	}
	return c
}

// rightClick is the ContextMenuFor hook. It moves the selection under the
// pointer first, then returns the menu rows for the element kind.
func (a *app) rightClick(id string) []ownframe.MenuItem {
	if a.overlay != "" || a.editing {
		return nil
	}
	rows := a.contextMenuFor(id)
	if len(rows) > 0 {
		a.prepareContext(id)
	}
	return rows
}

// handleLongPress opens the in-page context menu on a touch hold.
func (a *app) handleLongPress(_ context.Context, box ownframe.Box) error {
	if a.overlay != "" || !isGridID(box.ID) && !strings.HasPrefix(box.ID, "xl-tab-") {
		return nil
	}
	a.prepareContext(box.ID)
	a.openContextMenu(ctxKindFor(box.ID), box.X+box.W/2, box.Y+box.H/2)
	return nil
}

// prepareContext moves the selection under a right click the way Excel does:
// a cell outside the selection becomes the selection, and a header selects
// its column or row unless it is already inside the selection.
func (a *app) prepareContext(id string) {
	switch {
	case strings.HasPrefix(id, "xl-gch-"):
		if col, err := strconv.Atoi(strings.TrimPrefix(id, "xl-gch-")); err == nil &&
			(col < a.sel.Min.Col || col > a.sel.Max.Col || a.sel.Min.Row != 0) {
			a.selectRange(Range{Min: CellRef{Col: col}, Max: CellRef{Row: allCellsMaxRow, Col: col}})
		}
	case strings.HasPrefix(id, "xl-grh-"):
		if row, err := strconv.Atoi(strings.TrimPrefix(id, "xl-grh-")); err == nil &&
			(row < a.sel.Min.Row || row > a.sel.Max.Row || a.sel.Min.Col != 0) {
			a.selectRange(Range{Min: CellRef{Row: row}, Max: CellRef{Row: row, Col: allCellsMaxCol}})
		}
	default:
		if ref, ok := cellIDRef(id); ok {
			ref = a.mergeAnchor(ref)
			if !a.sel.Contains(ref) {
				a.selectCell(ref)
			}
		}
	}
}
