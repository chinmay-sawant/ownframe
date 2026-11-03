package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

const (
	demoWidth         = 1120
	demoHeight        = 760
	demoMinWidth      = 900
	demoMinHeight     = 610
	builderDragLimitY = 68
	builderPlacement  = "ui-builder"
)

// builderApp is the single tabbed ui-builder host. Each tab lives in its
// own screen_*.go file and owns a small state struct below.
type builderApp struct {
	modalFocus chadcn.DialogFocus
	view       *ui.View
	themeDirty bool
	page       *ownframe.Page
	logger     *log.Logger
	tab        string
	notice     string
	widgets    widgetsState
	settings   settingsState
	layout     layoutState
	todos      todoState
	files      listState
	theme      themeState
	compare    compareState
	dialog     dialogState
	custom     customState

	center               chadcn.ScrollState
	centerMeasurePending bool
	layoutWidth          int
	layoutHeight         int
}

// setLogger installs the diagnostic logger. A nil logger disables logging.
func (a *builderApp) setLogger(logger *log.Logger) {
	a.logger = logger
}

func (a *builderApp) logf(format string, args ...any) {
	builderLogf(a.logger, format, args...)
}

func newBuilderApp() (*builderApp, error) {
	a := &builderApp{
		tab:      "widgets",
		notice:   "Pick a tab on the left — every control below is built in Go.",
		widgets:  defaultWidgetsState(),
		settings: defaultSettingsState(),
		layout:   defaultLayoutState(),
		todos:    defaultTodoState(),
		files:    defaultListState(),
		theme:    defaultThemeState(),
		compare:  defaultCompareState(),
		dialog:   defaultDialogState(),
	}
	view, err := ui.Mount(ownframe.Config{
		Theme:     chadcn.Dark + chadcn.Stylesheet + builderMotionCSS,
		Title:     "OwnFrame Components",
		Width:     demoWidth,
		Height:    demoHeight,
		MinWidth:  demoMinWidth,
		MinHeight: demoMinHeight,
		DevTools:  false,
		Scrollbar: ownframe.ScrollbarHidden,
	}, a.screen())
	if err != nil {
		return nil, fmt.Errorf("new builder app: %w", err)
	}
	a.view = view
	p := view.Page
	a.page = p
	p.Scroll().SetAllowScroll(false)
	view.Handle(ownframe.Handlers{
		Click:     a.handleClick,
		KeyDown:   a.handleKey,
		Wheel:     a.handleWheel,
		DragStart: a.startCenterScrollDrag,
		DragMove:  a.moveCenterScrollDrag,
		DragEnd:   a.endCenterScrollDrag,
	})
	a.layoutWidth, a.layoutHeight = p.Size()
	p.SetTick(a.refreshOnResize)
	return a, nil
}

func (a *builderApp) handleClick(ctx context.Context, box ownframe.Box) error {
	action := box.Action
	start := time.Now()
	a.logf("click action=%q tab=%s", action, a.tab)
	if name, ok := strings.CutPrefix(action, "tab:"); ok {
		a.selectTab(name)
		if err := a.refresh(ctx); err != nil {
			a.logf("click action=%q tab=%s err=%v dur=%s", action, a.tab, err, time.Since(start))
			return err
		}
		a.logf("click action=%q tab=%s ok dur=%s", action, a.tab, time.Since(start))
		return nil
	}
	switch action {
	case "win-min":
		a.logf("window minimize")
		a.page.Window().Minimize()
		return nil
	case "win-dev":
		dev := !a.page.Window().DevTools()
		a.logf("window devtools on=%v", dev)
		a.page.Window().SetDevTools(dev)
		return a.refresh(ctx)
	case "win-max":
		a.logf("window toggle-maximize")
		a.page.Window().ToggleMaximize()
		return nil
	case "win-close":
		a.logf("window close requested")
		a.page.Window().Close()
		return nil
	}
	if a.handleWidgets(action) {
		return a.refreshLogged(ctx, action, start)
	}
	if a.handleSettings(action) {
		return a.refreshLogged(ctx, action, start)
	}
	if a.handleLayout(action) {
		return a.refreshLogged(ctx, action, start)
	}
	if a.handleTodo(action) {
		return a.refreshLogged(ctx, action, start)
	}
	if a.handleList(action) {
		return a.refreshLogged(ctx, action, start)
	}
	if a.handleTheme(action) {
		a.themeDirty = true
		return a.refreshLogged(ctx, action, start)
	}
	if a.handleCompare(action) {
		return a.refreshLogged(ctx, action, start)
	}
	if a.handleDialog(action) {
		return a.refreshLogged(ctx, action, start)
	}
	a.logf("click action=%q tab=%s ignored dur=%s", action, a.tab, time.Since(start))
	return nil
}

// refreshLogged re-renders after a handled action and logs the outcome with
// the handling duration, so a stuck re-render shows up as a click without a
// matching ok line.
func (a *builderApp) refreshLogged(ctx context.Context, action string, start time.Time) error {
	if err := a.refresh(ctx); err != nil {
		a.logf("click action=%q tab=%s err=%v dur=%s", action, a.tab, err, time.Since(start))
		return err
	}
	a.logf("click action=%q tab=%s ok dur=%s", action, a.tab, time.Since(start))
	return nil
}

func (a *builderApp) selectTab(name string) {
	switch name {
	case "widgets", "settings", "layout", "todo", "list", "theme", "compare", "dialog", "custom", "catalog":
		a.tab = name
		a.center = chadcn.ScrollState{}
		a.notice = map[string]string{
			"widgets":  "Widgets: buttons, badges, toggles, and cards from the Go builder.",
			"settings": "Settings: form state, inline validation, and save in Go.",
			"layout":   "Layout playground: direction, gap, padding, and box count.",
			"todo":     "Todo and counter: slice and integer state with re-render.",
			"list":     "File list: rows built with Children(rows...) in Go.",
			"theme":    "Global palettes and design tokens, shared by every control.",
			"custom":   "Custom components inherit the app theme and accept local properties.",
			"compare":  "Builder vs HTML: the same card as Go code and live output.",
			"dialog":   "Dialog: a backdrop overlay with a centered panel.",
			"catalog":  "ChadCN catalog: every constructor in a representative state.",
		}[name]
	}
}

func (a *builderApp) refresh(ctx context.Context) error {
	switch {
	case a.tab == "compare" && a.compare.Preview:
		a.modalFocus.Activate(a.page, "compare-preview-x", "compare-preview-open")
	case a.tab == "dialog" && a.dialog.open:
		a.modalFocus.Activate(a.page, "dialog-x", "dialog-confirm", "dialog-cancel")
	default:
		a.modalFocus.Deactivate(a.page)
	}

	start := time.Now()
	if a.themeDirty {
		sheet, err := a.theme.stylesheet()
		if err != nil {
			return err
		}
		if err := a.page.SetTheme(sheet); err != nil {
			return err
		}
		a.themeDirty = false
	}
	if err := a.view.Update(ctx, a.screen()); err != nil {
		return fmt.Errorf("update builder gallery: %w", err)
	}
	a.centerMeasurePending = true
	a.logf("refresh ok dur=%s", time.Since(start))
	return nil
}

const builderMotionCSS = `
@keyframes todo-done-pop{0%{transform:scale(0.92);opacity:0;}60%{transform:scale(1.03);opacity:1;}100%{transform:scale(1);opacity:1;}}
.todo-row:hover,.todo-toggle:hover{transform:translateY(-1px);}
`

func (a *builderApp) handleKey(ctx context.Context, key string) error {
	if !a.scrollInputBlocked() && a.page.Form().FocusID() == builderScrollThumbID && a.center.Key(key, int(builderScrollStep)) {
		return a.refresh(ctx)
	}
	if key == "escape" && a.compare.Preview {
		a.handleCompare("compare:preview-close")
		return a.refresh(ctx)
	}
	if key == "escape" && a.dialog.open {
		a.handleDialog("dialog:close")
		return a.refresh(ctx)
	}
	return nil
}
