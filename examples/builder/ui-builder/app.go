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
	modalFocus                 chadcn.DialogFocus
	view                       *ui.View
	themeDirty                 bool
	page                       *ownframe.Page
	logger                     *log.Logger
	tab                        string
	notice                     string
	widgets                    widgetsState
	settings                   settingsState
	layout                     layoutState
	todos                      todoState
	files                      listState
	theme                      themeState
	compare                    compareState
	dialog                     dialogState
	custom                     customState
	catalogTabSel              string
	catalogSlider              float64
	catalogCheck               bool
	catalogSwitch              bool
	catalogSelect              string
	catalogSelectOpen          bool
	catalogAccordion           []string
	catalogAlertDialogOpen     bool
	catalogCalendarMonth       time.Time
	catalogCalendarSelected    time.Time
	catalogCarouselIndex       int
	catalogCollapsibleOpen     bool
	catalogComboboxOpen        bool
	catalogComboboxValue       string
	catalogComboboxQuery       string
	catalogCommandActive       string
	catalogCommandQuery        string
	catalogContextOpen         bool
	catalogDataSelected        map[int]bool
	catalogDateOpen            bool
	catalogDateSelected        time.Time
	catalogDateMonth           time.Time
	catalogDialogOpen          bool
	catalogDrawerOpen          bool
	catalogDropdownOpen        bool
	catalogHoverOpen           bool
	catalogMenubarOpen         string
	catalogNavOpen             string
	catalogPopoverOpen         bool
	catalogTooltipOpen         bool
	catalogSheetOpen           bool
	catalogPage                int
	catalogQuestionAnswer      string
	catalogRadioValue          string
	catalogTogglePressed       bool
	catalogToggleGroupSelected []string
	catalogToastOpen           bool
	catalogProgress            int
	catalogFilter              string
	catalogSidebarOpen         bool
	catalogChoiceSelected      bool
	catalogSearchOpen          bool
	catalogSearchQuery         string
	catalogSearchValue         string

	center               chadcn.ScrollState
	centerMeasurePending bool
	scrollDirty          bool
	// catalogGeo pins the last full-render catalog geometry. Windowed
	// renders reuse it instead of re-measuring estimates off spacers,
	// which would compound the pitch error frame over frame.
	catalogGeoMeasured bool
	catalogGeoHeaderPx int
	catalogGeoPitchPx  int
	catalogGeoRowsH    int
	layoutWidth        int
	layoutHeight       int
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
		tab:                    "widgets",
		notice:                 "Pick a tab on the left — every control below is built in Go.",
		widgets:                defaultWidgetsState(),
		settings:               defaultSettingsState(),
		layout:                 defaultLayoutState(),
		todos:                  defaultTodoState(),
		files:                  defaultListState(),
		theme:                  defaultThemeState(),
		compare:                defaultCompareState(),
		dialog:                 defaultDialogState(),
		catalogAccordion:       []string{"one"},
		catalogCollapsibleOpen: true,
		catalogCommandActive:   "calendar",
		catalogDataSelected:    map[int]bool{},
		catalogPage:            1,
		catalogRadioValue:      "medium",
		catalogTogglePressed:   true,
		catalogToastOpen:       true,
		catalogProgress:        catalogProgress,
		catalogFilter:          "weekly",
		catalogSidebarOpen:     true,
		catalogChoiceSelected:  true,
	}
	sheet, err := defaultThemeState().stylesheet()
	if err != nil {
		return nil, err
	}
	view, err := ui.Mount(ownframe.Config{
		Theme:       sheet,
		Title:       "OwnFrame Components",
		Width:       demoWidth,
		Height:      demoHeight,
		MinWidth:    demoMinWidth,
		MinHeight:   demoMinHeight,
		DevTools:    false,
		RetainState: true,
		Scrollbar:   ownframe.ScrollbarHidden,
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
		Change:    a.handleChange,
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
	if a.handleCatalog(action) {
		return a.refreshLogged(ctx, action, start)
	}
	a.logf("click action=%q tab=%s ignored dur=%s", action, a.tab, time.Since(start))
	return nil
}

// refreshLogged re-renders after a handled action and logs the outcome with
// the handling duration, so a stuck re-render shows up as a click without a
// matching ok line. Pointer clicks leave focus on the button, which keeps the
// :focus-visible ring painted forever, so the focus is cleared before the
// re-render. Text inputs never reach this path (they have no data-action and
// type through handleChange, which preserves focus), so typing is unaffected.
func (a *builderApp) refreshLogged(ctx context.Context, action string, start time.Time) error {
	a.page.Form().RequestFocus("")
	if err := a.refresh(ctx); err != nil {
		a.logf("click action=%q tab=%s err=%v dur=%s", action, a.tab, err, time.Since(start))
		return err
	}
	a.logf("click action=%q tab=%s ok dur=%s", action, a.tab, time.Since(start))
	return nil
}

// refreshBlur clears pointer focus before re-rendering so clicked buttons do
// not keep the :focus-visible outline. Typing paths must call refresh, not
// this, so text inputs keep their caret.
func (a *builderApp) refreshBlur(ctx context.Context) error {
	a.page.Form().RequestFocus("")
	return a.refresh(ctx)
}

// handleChange carries live input text into host-owned query state for the
// catalog Command, Combobox, and Search palettes plus the widgets Search
// card, then re-renders so the lists filter. Unlike refreshLogged it preserves
// focus so typing can continue. When a Search query matches no prefilled row,
// the notice names the background fetch the host would run for remote rows.
func (a *builderApp) handleChange(ctx context.Context, box ownframe.Box) error {
	switch box.ID {
	case "catalog-command-input":
		a.catalogCommandQuery = a.page.Form().FormValue(box.ID)
		a.notice = "Catalog command query: " + a.catalogCommandQuery + "."
		return a.refresh(ctx)
	case "catalog-combobox-input":
		a.catalogComboboxQuery = a.page.Form().FormValue(box.ID)
		return a.refresh(ctx)
	case "catalog-search-input":
		a.catalogSearchQuery = a.page.Form().FormValue(box.ID)
		a.catalogSearchOpen = true
		if len(chadcn.FilterSearchItems(catalogSearchItems(), a.catalogSearchQuery)) == 0 && a.catalogSearchQuery != "" {
			a.notice = "Catalog search: no local match for \"" + a.catalogSearchQuery + "\" — background fetch would run."
		} else {
			a.notice = "Catalog search query: " + a.catalogSearchQuery + "."
		}
		return a.refresh(ctx)
	case "widgets-search-input":
		a.widgets.search = a.page.Form().FormValue(box.ID)
		a.widgets.searchOpen = true
		if len(chadcn.FilterSearchItems(widgetSearchItems(), a.widgets.search)) == 0 && a.widgets.search != "" {
			a.notice = "Search: no local match for \"" + a.widgets.search + "\" — background fetch would run."
		} else {
			a.notice = "Search query: " + a.widgets.search + "."
		}
		return a.refresh(ctx)
	default:
		return nil
	}
}

func (a *builderApp) selectTab(name string) {
	switch name {
	case "widgets", "settings", "layout", "todo", "list", "theme", "compare", "dialog", "custom", "catalog",
		"chartjs", "shadcnchart", "recharts", "tremor", "echarts", "apexcharts", "nivo":
		a.tab = name
		a.center = chadcn.ScrollState{}
		a.catalogGeoMeasured = false
		a.notice = map[string]string{
			"widgets":     "Widgets: buttons, badges, toggles, and cards from the Go builder.",
			"settings":    "Settings: form state, inline validation, and save in Go.",
			"layout":      "Layout playground: direction, gap, padding, and box count.",
			"todo":        "Todo and counter: slice and integer state with re-render.",
			"list":        "File list: rows built with Children(rows...) in Go.",
			"theme":       "Global palettes and design tokens, shared by every control.",
			"custom":      "Custom components inherit the app theme and accept local properties.",
			"compare":     "Builder vs HTML: the same card as Go code and live output.",
			"dialog":      "Dialog: a backdrop overlay with a centered panel.",
			"catalog":     "ChadCN catalog: every constructor in a representative state.",
			"chartjs":     "Chart.js: bar, line, pie, doughnut, radar, polar area, bubble, and scatter.",
			"shadcnchart": "shadcn charts: area, bar, line, pie, radar, and radial.",
			"recharts": "Recharts: area, bar, line, composed, pie, radar, radial bar, " +
				"scatter, funnel, sankey, sunburst, and treemap.",
			"tremor":     "Tremor: area, bar, line, combo, donut, sparks, lists, tracker, and progress.",
			"echarts":    "ECharts: cartesian, statistical, map, tree, graph, and gauge charts.",
			"apexcharts": "ApexCharts: line, bar, range, pie, and hierarchical charts.",
			"nivo":       "Nivo: bar, line, pie, calendar, hierarchy, and flow charts.",
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
	a.scrollDirty = false
	a.logf("refresh ok dur=%s", time.Since(start))
	return nil
}

const builderMotionCSS = `
@keyframes todo-done-pop{0%{transform:scale(0.92);opacity:0;}60%{transform:scale(1.03);opacity:1;}100%{transform:scale(1);opacity:1;}}
.todo-row:hover,.todo-toggle:hover{transform:translateY(-1px);}
`

func (a *builderApp) handleKey(ctx context.Context, key string) error {
	if a.scrollCenterKey(key) {
		return nil
	}
	if key == "escape" && a.compare.Preview {
		a.handleCompare("compare:preview-close")
		return a.refresh(ctx)
	}
	if key == "escape" && a.dialog.open {
		a.handleDialog("dialog:close")
		return a.refresh(ctx)
	}
	if key == "escape" && a.widgets.searchOpen {
		a.widgets.searchOpen = false
		a.notice = "Search dismissed."
		return a.refresh(ctx)
	}
	if key == "escape" && a.catalogSearchOpen {
		a.catalogSearchOpen = false
		a.notice = "Catalog search dismissed."
		return a.refresh(ctx)
	}
	return nil
}
