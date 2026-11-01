package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/chinmay-sawant/ownframe"
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
	page     *ownframe.Page
	logger   *log.Logger
	tab      string
	notice   string
	widgets  widgetsState
	settings settingsState
	layout   layoutState
	todos    todoState
	files    listState
	theme    themeState
	compare  compareState
	dialog   dialogState
}

// setLogger installs the diagnostic logger. A nil logger disables logging.
func (a *builderApp) setLogger(logger *log.Logger) {
	a.logger = logger
}

func (a *builderApp) logf(format string, args ...any) {
	if a.logger == nil {
		return
	}
	a.logger.Printf(format, args...)
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
	p, err := ui.NewPage(ownframe.Config{
		Title:     "OwnFrame · Builder Gallery",
		Width:     demoWidth,
		Height:    demoHeight,
		MinWidth:  demoMinWidth,
		MinHeight: demoMinHeight,
		DevTools:  true,
	}, a.screen())
	if err != nil {
		return nil, fmt.Errorf("new builder app: %w", err)
	}
	a.page = p
	p.Handle(ownframe.Handlers{Click: a.handleClick})
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
		return nil
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
	case "widgets", "settings", "layout", "todo", "list", "theme", "compare", "dialog":
		a.tab = name
		a.notice = map[string]string{
			"widgets":  "Widgets: buttons, badges, toggles, and cards from the Go builder.",
			"settings": "Settings: form state, inline validation, and save in Go.",
			"layout":   "Layout playground: direction, gap, padding, and box count.",
			"todo":     "Todo and counter: slice and integer state with re-render.",
			"list":     "File list: rows built with Children(rows...) in Go.",
			"theme":    "Theme: accent, radius, and spacing restyle the preview.",
			"compare":  "Builder vs HTML: the same card as Go code and live output.",
			"dialog":   "Dialog: a backdrop overlay with a centered panel.",
		}[name]
	}
}

func (a *builderApp) refresh(ctx context.Context) error {
	start := time.Now()
	fragment, err := ui.Render(a.screen())
	if err != nil {
		a.logf("refresh render err=%v dur=%s", err, time.Since(start))
		return fmt.Errorf("render builder gallery: %w", err)
	}
	if err := a.page.Load(ctx, renderBuilderDocument(fragment)); err != nil {
		a.logf("refresh load err=%v dur=%s", err, time.Since(start))
		return fmt.Errorf("reload builder gallery: %w", err)
	}
	a.logf("refresh ok bytes=%d dur=%s", len(fragment), time.Since(start))
	return nil
}

func renderBuilderDocument(fragment string) string {
	return "<!doctype html><html><head><meta charset=\"utf-8\"><style>" +
		"html,body{margin:0;padding:0;background-color:#111827}" +
		"@keyframes todo-done-pop{0%{transform:scale(0.92);opacity:0;}" +
		"60%{transform:scale(1.03);opacity:1;}}" +
		"100%{transform:scale(1);opacity:1;}}" +
		".todo-row:hover,.todo-toggle:hover{transform:translateY(-1px);}" +
		"</style></head><body>" + fragment + "</body></html>"
}
