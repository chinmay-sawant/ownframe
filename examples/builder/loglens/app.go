package main

import (
	"context"
	"strings"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

const (
	windowWidth  = 1440
	windowHeight = 960
	minWidth     = 1100
	minHeight    = 720

	logPageSize = 8

	defaultSampleRate  = 100
	defaultSplitRatio  = 0.55
	defaultThreshold   = 85
	defaultSliderValue = 30
)

// app owns the dashboard state. Panels are pure functions of this state and
// are rebuilt through View.Update after every action.
type app struct {
	view *ui.View
	page *ownframe.Page
	dark bool
	tab  string

	// Shell.
	sidebarCollapsed bool
	navOpen          string
	menubarOpen      string
	userMenuOpen     bool
	notifOpen        bool
	commandOpen      bool
	commandQuery     string
	commandActive    string

	// Overview.
	loading   bool
	rangeMode string

	// Logs and filters.
	timeRange     string
	service       string
	severity      []string
	sortValue     string
	fromDate      time.Time
	calendarMonth time.Time
	query         string
	pageIndex     int
	selectedRows  map[int]bool
	selectOpen    bool
	dateOpen      bool
	comboboxOpen  bool
	contextOpen   bool
	contextX      int
	contextY      int
	drawerOpen    bool
	drawerIndex   int
	drawerTab     string
	liveTail      bool
	autoRefresh   bool
	sampleRate    int
	showContext   bool
	density       string

	// Traces.
	traceIndex  int
	splitRatio  float64
	traceSpan   string
	traceScroll chadcn.ScrollState

	// Incident room.
	accordionOpen   []string
	collapsibleOpen bool
	carouselIndex   int
	questionAnswer  string
	rtl             bool

	// Settings sheet.
	settingsOpen  bool
	settingsTab   string
	email         string
	webhook       string
	notes         string
	channel       string
	severityValue string
	otp           string
	mute          bool
	alertsEnabled bool
	threshold     int
	sliderValue   int
	radioValue    string
	checkValue    bool

	// Overlays.
	dialogOpen  bool
	alertOpen   bool
	toastOpen   bool
	toastTitle  string
	toastBody   string
	tooltipOpen bool
	hoverOpen   bool

	// extra holds panel-local values that do not deserve a field yet. Panels
	// may use it, but shared state above is preferred.
	extra map[string]any
}

func newApp() (*app, error) {
	a := &app{
		dark:          true,
		tab:           "overview",
		rangeMode:     "24h",
		timeRange:     "24 hours",
		service:       "All services",
		severity:      []string{"error", "warn"},
		sortValue:     "Newest first",
		fromDate:      time.Now(),
		calendarMonth: time.Now(),
		selectedRows:  map[int]bool{},
		sampleRate:    defaultSampleRate,
		density:       "comfortable",
		settingsTab:   "routing",
		drawerTab:     "payload",
		splitRatio:    defaultSplitRatio,
		threshold:     defaultThreshold,
		sliderValue:   defaultSliderValue,
		radioValue:    "warning",
		channel:       "Slack",
		severityValue: "warning",
		extra:         map[string]any{},
	}
	view, err := ui.Mount(ownframe.Config{
		Title:     "Loglens · Production",
		Width:     windowWidth,
		Height:    windowHeight,
		MinWidth:  minWidth,
		MinHeight: minHeight,
		Theme:     a.themeSheet(),
		DevTools:  false,
		Scrollbar: ownframe.ScrollbarHidden,
	}, a.screen())
	if err != nil {
		return nil, err
	}
	a.view = view
	a.page = view.Page
	view.Handle(ownframe.Handlers{
		Click:   a.handleClick,
		Hover:   a.handleHover,
		KeyDown: a.handleKey,
	})
	return a, nil
}

func (a *app) themeSheet() string {
	if a.dark {
		return chadcn.Dark + chadcn.Stylesheet + appCSS
	}
	return chadcn.Light + chadcn.Stylesheet + appCSS
}

func (a *app) refresh(ctx context.Context) error {
	if a.view == nil {
		return nil
	}
	return a.view.Update(ctx, a.screen())
}

func (a *app) setTab(ctx context.Context, tab string) error {
	a.tab = tab
	return a.refresh(ctx)
}

func (a *app) toggleTheme(ctx context.Context) error {
	a.dark = !a.dark
	if err := a.page.SetTheme(a.themeSheet()); err != nil {
		return err
	}
	return a.refresh(ctx)
}

func (a *app) showToast(ctx context.Context, title, body string) error {
	a.toastOpen = true
	a.toastTitle = title
	a.toastBody = body
	return a.refresh(ctx)
}

func (a *app) closeOverlays(ctx context.Context) error {
	changed := a.commandOpen || a.dialogOpen || a.alertOpen || a.settingsOpen ||
		a.drawerOpen || a.contextOpen || a.notifOpen || a.userMenuOpen ||
		a.navOpen != "" || a.menubarOpen != "" || a.selectOpen || a.comboboxOpen || a.dateOpen
	a.commandOpen = false
	a.dialogOpen = false
	a.alertOpen = false
	a.settingsOpen = false
	a.drawerOpen = false
	a.contextOpen = false
	a.notifOpen = false
	a.userMenuOpen = false
	a.navOpen = ""
	a.menubarOpen = ""
	a.selectOpen = false
	a.comboboxOpen = false
	a.dateOpen = false
	if !changed {
		return nil
	}
	return a.refresh(ctx)
}

func (a *app) handleClick(ctx context.Context, box ownframe.Box) error {
	switch {
	case strings.HasPrefix(box.Action, "nav:"):
		a.navOpen = ""
		return a.setTab(ctx, strings.TrimPrefix(box.Action, "nav:"))
	case box.Action == "theme:toggle":
		return a.toggleTheme(ctx)
	case box.Action == "settings:open":
		a.settingsOpen = true
		return a.refresh(ctx)
	case box.Action == "settings:close":
		a.settingsOpen = false
		return a.refresh(ctx)
	case box.Action == "toast:close":
		a.toastOpen = false
		return a.refresh(ctx)
	case box.Action == "command:open":
		a.commandOpen = true
		return a.refresh(ctx)
	case box.Action == "dialog:open":
		a.dialogOpen = true
		return a.refresh(ctx)
	case box.Action == "alert:open":
		a.alertOpen = true
		return a.refresh(ctx)
	}
	return nil
}

func (a *app) handleHover(ctx context.Context, previous, next ownframe.Box) (bool, error) {
	_ = previous
	tooltip := next.ID == "tooltip-trigger"
	hover := next.ID == "hover-trigger"
	if a.tooltipOpen == tooltip && a.hoverOpen == hover {
		return false, nil
	}
	a.tooltipOpen = tooltip
	a.hoverOpen = hover
	return true, a.refresh(ctx)
}

func (a *app) handleKey(ctx context.Context, key string) error {
	if key != "escape" {
		return nil
	}
	return a.closeOverlays(ctx)
}

func (a *app) formValue(id string) string {
	return a.page.Form().FormValue(id)
}

func (a *app) setFormValue(id, value string) {
	a.page.Form().SetFormValue(id, value)
}
