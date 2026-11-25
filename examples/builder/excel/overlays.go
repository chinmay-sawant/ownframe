package main

import (
	"context"
	"strings"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// findHit is one Find All result.
type findHit struct {
	SheetIdx int
	Sheet    string
	Ref      CellRef
	Text     string
}

// overlayState is the state owned by the ribbon, dialogs, menus and charts.
// app embeds it as field x.
type overlayState struct {
	menu             string  // open dropdown id, "" when none
	anchorX, anchorY float64 // window position under the trigger that opened it

	recentColors []string

	painter, painterSticky bool
	painterStyle           Style
	painterSrc             Range

	opt map[string]string // dialog options (radio buttons, flags, drafts)

	findHits []findHit
	findIdx  int
	findMsg  string

	fmtDraft, fmtOrig Style
	fmtTab            string

	filterOn      bool
	filterCol     int
	filterVals    []string
	filterAllowed map[string]bool

	charts    []*ChartSpec
	chartSeq  int
	selChart  int
	dragChart int

	ctxKind    string
	cellsDim   string // "row" or "col" while a Cells menu action runs
	ctxX, ctxY float64

	msg, toastTitle string
	toastOpen       bool

	// View options, read by the grid and shell.
	hideGrid, hideHeads, hideFormulaBar, showFormulas bool
	zoom                                              int // percent, 0 means the default for the window width

	// The scaled theme sheet last built and the zoom the page was given.
	themeCSS      string
	themeDark     bool
	themeCachePct int
	themePct      int
}

// zoomDefaultPct is the zoom shown when none was set.
const zoomDefaultPct = 100

// zoomPct returns the zoom percentage: the user's choice, or the default for
// the window width until one is made.
func (a *app) zoomPct() int {
	if a.x.zoom != 0 {
		return a.x.zoom
	}
	if a.page == nil {
		return zoomDefaultPct
	}
	w, _ := a.page.Size()
	return autoZoomPct(w)
}

// overlays returns the elements mounted above the page: charts, the open
// dialog, the context menu and the toast.
func (a *app) overlays() []*ui.Element {
	var out []*ui.Element
	if p := a.menuPanel(); p != nil {
		out = append(out, p)
	}
	if l := a.chartLayer(); l != nil {
		out = append(out, l)
	}
	if c := a.contextMenu(); c != nil {
		out = append(out, c)
	}
	if d := a.dialog(); d != nil {
		out = append(out, d)
	}
	if a.x.toastOpen {
		out = append(out, chadcn.Toast(chadcn.ToastOptions{
			ID: "rb-toast", Title: a.x.toastTitle, Description: a.x.msg, CloseAction: "xl:toast:close", Open: true,
		}))
	}
	return out
}

func (a *app) notify(title, body string) {
	a.x.toastOpen, a.x.toastTitle, a.x.msg = true, title, body
}

// overlayClick closes an open dropdown when the click lands outside any
// ribbon action, then lets the caller continue with the click.
func (a *app) overlayClick(_ context.Context, box ownframe.Box) (bool, error) {
	if a.x.menu != "" && !strings.HasPrefix(box.Action, "xl:") {
		a.x.menu = ""
		a.redraw()
	}
	return false, nil
}

// ribbonClick handles every "xl:" action from the ribbon, dropdowns, dialogs,
// context menu and charts.
func (a *app) ribbonClick(ctx context.Context, box ownframe.Box) (bool, error) {
	if !strings.HasPrefix(box.Action, "xl:") {
		return false, nil
	}
	if strings.HasPrefix(box.Action, "xl:menu:") {
		a.x.anchorX, a.x.anchorY = box.X, box.Y+box.H
	}
	return a.runActionCtx(ctx, box.Action), nil
}

// overlayChange is the hook for dialog text fields. Most values are read from
// the page form when a dialog button runs; the filter search box filters live.
func (a *app) overlayChange(_ context.Context, box ownframe.Box) (bool, error) {
	if box.ID == "rb-fsearch" && a.overlay == "filter" {
		a.filterSearch(a.input("rb-fsearch"))
		a.redraw()
		return true, nil
	}
	return false, nil
}

// afterSelect applies an armed Format Painter after the selection moves. The
// selection code calls it once the new selection is in place.
func (a *app) afterSelect() {
	if !a.x.painter || a.sel == a.x.painterSrc {
		return
	}
	src := a.x.painterStyle
	a.applyFormat(func(s *Style) { *s = src })
	if !a.x.painterSticky {
		a.x.painter = false
	}
}

// shortcut handles the formatting and dialog shortcuts. key is the lowercase
// key name with modifiers joined by "+", for example "ctrl+b". It reports
// whether the key was consumed.
func (a *app) shortcut(key string) bool {
	return a.shortcutCtx(a.bg(), key)
}

// shortcutCtx is shortcut with the handler context threaded to the action.
func (a *app) shortcutCtx(ctx context.Context, key string) bool {
	switch key {
	case "ctrl+=", "ctrl++", "ctrl+equal", "ctrl+plus", "ctrl+kpadd", "ctrl+shift+=", "ctrl+shift+equal",
		"ctrl+shift+plus", "ctrl+-", "ctrl+minus", "ctrl+kpsubtract", "ctrl+shift+minus", "ctrl+0", "ctrl+kp0":
		return a.zoomKey(key)
	case "ctrl+b":
		return a.runActionCtx(ctx, "xl:bold")
	case "ctrl+i":
		return a.runActionCtx(ctx, "xl:italic")
	case "ctrl+u":
		return a.runActionCtx(ctx, "xl:underline")
	case "ctrl+1":
		return a.runActionCtx(ctx, "xl:dlg:format")
	case "ctrl+f", "ctrl+h":
		return a.runActionCtx(ctx, "xl:dlg:find")
	case "ctrl+alt+v":
		return a.runActionCtx(ctx, "xl:dlg:paste")
	case "ctrl+/", "ctrl+shift+/":
		return a.runActionCtx(ctx, "xl:dlg:shortcuts")
	case "shift+f2":
		return a.runActionCtx(ctx, "xl:dlg:comment")
	case "ctrl+shift+l":
		return a.runActionCtx(ctx, "xl:filter:toggle")
	case "ctrl+`", "ctrl+~":
		return a.runActionCtx(ctx, "xl:view:formulas")
	case "escape":
		if a.x.painter {
			a.x.painter, a.x.painterSticky = false, false
			a.redraw()
			return true
		}
		if a.x.menu != "" || a.overlay != "" || a.x.toastOpen {
			a.x.menu, a.x.toastOpen = "", false
			a.setOverlay("")
			return true
		}
	case "delete":
		if a.x.selChart != 0 && a.chartByID(a.x.selChart) != nil {
			a.deleteChart(a.x.selChart)
			a.redraw()
			return true
		}
	}
	return false
}

// readClipboard reads the system clipboard as text.
func readClipboard() (string, error) { return ownframe.Read() }
