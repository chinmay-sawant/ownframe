package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
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

	logPageSize = demoFeedPageSize

	defaultSampleRate  = 100
	defaultSplitRatio  = 0.55
	defaultThreshold   = 85
	defaultSliderValue = 30
	dataRefreshSeconds = 1

	secondsPerMinute = 60
	percentageScale  = 100
)

type onCallResponder struct {
	Name     string
	Role     string
	Initials string
	Region   string
}

func defaultResponders() [4]onCallResponder {
	return [4]onCallResponder{
		{Name: "Ada Kim", Role: "Primary", Initials: "AK", Region: "Europe"},
		{Name: "Liam Taylor", Role: "Primary", Initials: "LT", Region: "Americas"},
		{Name: "Yuki Ono", Role: "Primary", Initials: "YO", Region: "Asia-Pacific"},
		{Name: "Alex Chen", Role: "Secondary", Initials: "AC", Region: "Global"},
	}
}

// app owns the dashboard state. Panels are pure functions of this state and
// are rebuilt through View.Update after every action.
type app struct {
	interactions *chadcn.Interactions
	view         *ui.View
	page         *ownframe.Page
	dark         bool
	tab          string

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
	feed *telemetryFeed

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
	traceIndex     int
	tracePageIndex int
	splitRatio     float64
	traceSpan      string

	// Incident room.
	accordionOpen   []string
	collapsibleOpen bool
	carouselIndex   int
	questionAnswer  string
	rtl             bool

	// On-call rotation.
	onCallDate       time.Time
	onCallSchedule   map[string]onCallResponder
	onCallDialogOpen bool
	onCallDialogDate time.Time

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
	dialogOpen       bool
	alertOpen        bool
	toastOpen        bool
	toastTitle       string
	toastBody        string
	tooltipOpen      bool
	hoverOpen        bool
	lastUpdatedAt    time.Time
	nextLiveTick     time.Time
	nextDataRefresh  time.Time
	lastRefreshCount int
	liveLabel        string

	// extra holds panel-local values that do not deserve a field yet. Panels
	// may use it, but shared state above is preferred.
	extra map[string]any
}

func newApp() (*app, error) {
	started := time.Now()
	a := &app{
		dark:           true,
		tab:            "overview",
		timeRange:      "24 hours",
		service:        "All services",
		severity:       []string{"error", "warn"},
		sortValue:      "Newest first",
		fromDate:       time.Now(),
		calendarMonth:  time.Now(),
		selectedRows:   map[int]bool{},
		sampleRate:     defaultSampleRate,
		density:        "comfortable",
		settingsTab:    "routing",
		drawerTab:      "payload",
		splitRatio:     defaultSplitRatio,
		threshold:      defaultThreshold,
		sliderValue:    defaultSliderValue,
		radioValue:     "warning",
		channel:        "Slack",
		severityValue:  "warning",
		extra:          map[string]any{},
		lastUpdatedAt:  started,
		onCallDate:     started,
		onCallSchedule: map[string]onCallResponder{},
		feed:           newTelemetryFeed(time.Now, started.UnixNano()),
	}
	a.interactions = chadcn.NewInteractions()
	view, err := a.interactions.Mount(context.Background(), ownframe.Config{
		Title:       "Loglens · Production",
		Width:       windowWidth,
		Height:      windowHeight,
		MinWidth:    minWidth,
		MinHeight:   minHeight,
		Theme:       a.themeSheet(),
		DevTools:    false,
		RetainState: true,
		Scrollbar:   ownframe.ScrollbarHidden,
	}, a.screen)
	if err != nil {
		return nil, err
	}
	a.view = view
	a.page = view.Page
	a.interactions.Handle(ownframe.Handlers{
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
	a.configureLiveTick()
	return a.interactions.Update(ctx, a.screen)
}

func (a *app) setTab(ctx context.Context, tab string) error {
	a.tab = tab
	a.page.ScrollTo(0, 0)
	return a.refresh(ctx)
}

func (a *app) refreshTelemetry(ctx context.Context) error {
	if a.feed == nil {
		return nil
	}
	added := a.feed.refresh()
	now := time.Now()
	a.lastUpdatedAt = now
	a.lastRefreshCount = added
	a.pageIndex = 0
	a.tracePageIndex = 0
	a.traceIndex = 0
	return a.showToast(ctx, "Data refreshed", strconv.Itoa(added)+" new generated log and trace records are available.")
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
		a.navOpen != "" || a.menubarOpen != "" || a.selectOpen || a.comboboxOpen || a.dateOpen ||
		a.onCallDialogOpen
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
	a.onCallDialogOpen = false
	if !changed {
		return nil
	}
	return a.refresh(ctx)
}

func (a *app) closeHeaderPopovers() {
	a.tooltipOpen = false
	a.hoverOpen = false
	a.notifOpen = false
	a.userMenuOpen = false
}

func (a *app) handleClick(ctx context.Context, box ownframe.Box) error {
	if a.navOpen != "" && !strings.HasPrefix(box.ID, "product-") && box.ID != "top-nav" {
		a.navOpen = ""
	}
	if a.menubarOpen != "" && !strings.HasPrefix(box.ID, "menu-") && box.ID != "top-menubar" {
		a.menubarOpen = ""
	}
	if a.notifOpen && !strings.HasPrefix(box.ID, "notifications-popover") {
		a.notifOpen = false
	}
	if a.userMenuOpen && !strings.HasPrefix(box.ID, "user-menu") {
		a.userMenuOpen = false
	}
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
	case box.Action == "oncall:swap":
		a.onCallDialogDate = a.onCallDate
		a.onCallDialogOpen = true
		return a.refresh(ctx)
	}
	return nil
}

func (a *app) onCallFor(day time.Time) onCallResponder {
	if day.IsZero() {
		day = time.Now()
	}
	key := day.Format("2006-01-02")
	if r, ok := a.onCallSchedule[key]; ok {
		return r
	}
	responders := defaultResponders()
	idx := (day.Year()*366 + day.YearDay()) % len(responders)
	if idx < 0 {
		idx = -idx
	}
	return responders[idx]
}

func (a *app) setOnCall(ctx context.Context, day time.Time, responder onCallResponder) error {
	if a.onCallSchedule == nil {
		a.onCallSchedule = make(map[string]onCallResponder)
	}
	key := day.Format("2006-01-02")
	a.onCallSchedule[key] = responder
	a.onCallDate = day
	return a.showToast(ctx, "Coverage updated", responder.Name+" is on call for "+day.Format("Jan 2")+".")
}

func (a *app) handleHover(ctx context.Context, previous, next ownframe.Box) (bool, error) {
	_ = previous
	if a.commandOpen || a.dialogOpen || a.alertOpen || a.settingsOpen {
		return false, nil
	}
	tooltip := next.ID == "tooltip-trigger" || next.ID == "live-tooltip-content"
	hover := next.ID == "hover-trigger" || next.ID == "oncall-hover" || next.ID == "oncall-hover-content"
	if a.tooltipOpen == tooltip && a.hoverOpen == hover {
		return false, nil
	}
	a.tooltipOpen = tooltip
	a.hoverOpen = hover
	a.configureLiveTick()
	if (tooltip || hover) && (a.navOpen != "" || a.menubarOpen != "" || a.notifOpen || a.userMenuOpen) {
		a.navOpen = ""
		a.menubarOpen = ""
		a.notifOpen = false
		a.userMenuOpen = false
		return true, a.refresh(ctx)
	}
	patches := []ownframe.ElementPatch{
		{
			ID: "live-tooltip",
			Edits: []ownframe.ElementEdit{
				{Attribute: "class", Value: topTooltipClass(tooltip)},
			},
		},
		{
			ID: "oncall-hover",
			Edits: []ownframe.ElementEdit{
				{Attribute: "class", Value: topHoverCardClass(hover)},
			},
		},
	}
	if tooltip {
		label := a.liveUpdatedLabel(time.Now())
		patches = append(patches, ownframe.ElementPatch{
			ID: "live-tooltip-content",
			Edits: []ownframe.ElementEdit{
				{Text: &label},
			},
		})
	}
	if err := a.patchElements(ctx, patches); err != nil {
		return false, err
	}
	return false, nil
}

func (a *app) liveUpdatedLabel(now time.Time) string {
	if a.lastUpdatedAt.IsZero() {
		return "Updated just now"
	}
	seconds := int(now.Sub(a.lastUpdatedAt).Seconds())
	if seconds < 1 {
		return "Updated just now"
	}
	if seconds == 1 {
		return "Updated 1 second ago"
	}
	if seconds < secondsPerMinute {
		return "Updated " + strconv.Itoa(seconds) + " seconds ago"
	}
	minutes := seconds / secondsPerMinute
	if minutes == 1 {
		return "Updated 1 minute ago"
	}
	return "Updated " + strconv.Itoa(minutes) + " minutes ago"
}

func topTooltipClass(open bool) string {
	if open {
		return "of-tooltip of-tooltip--bottom of-tooltip--css-hover of-tooltip--hovered"
	}
	return "of-tooltip of-tooltip--bottom of-tooltip--css-hover"
}

func topHoverCardClass(open bool) string {
	if open {
		return "of-hover-card of-hover-card--bottom of-hover-card--css-hover of-hover-card--hovered"
	}
	return "of-hover-card of-hover-card--bottom of-hover-card--css-hover"
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

func (a *app) patchElement(ctx context.Context, id string, edits ...ownframe.ElementEdit) error {
	err := a.page.PatchElement(ctx, id, edits...)
	if errors.Is(err, ownframe.ErrPatchRefused) {
		return a.refresh(ctx)
	}
	return err
}

func (a *app) patchElements(ctx context.Context, patches []ownframe.ElementPatch) error {
	if len(patches) == 0 {
		return nil
	}
	err := a.page.PatchElements(ctx, patches)
	if errors.Is(err, ownframe.ErrPatchRefused) {
		return a.refresh(ctx)
	}
	return err
}

func (a *app) patchHover(ctx context.Context, id string, edits ...ownframe.ElementEdit) error {
	return a.patchElement(ctx, id, edits...)
}

func boolString(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func (a *app) patchToggle(ctx context.Context, id string, pressed bool) error {
	return a.patchElement(ctx, id, ownframe.ElementEdit{
		Attribute: "aria-pressed",
		Value:     boolString(pressed),
	})
}

func (a *app) patchSwitch(ctx context.Context, id string, checked bool) error {
	trackColor := "var(--of-control, #262626)"
	thumbColor := "var(--of-switch-thumb, #ffffff)"
	marginLeft := "2px"
	if checked {
		trackColor = "var(--of-primary, #171717)"
		thumbColor = "var(--of-on-primary, #fafafa)"
		marginLeft = "18px"
	}
	patches := []ownframe.ElementPatch{
		{
			ID: id,
			Edits: []ownframe.ElementEdit{
				{Attribute: "aria-checked", Value: boolString(checked)},
			},
		},
		{
			ID: id + "-track",
			Edits: []ownframe.ElementEdit{
				{Attribute: "style", Value: fmt.Sprintf("align-items:center;background-color:%s;border-radius:10px;display:flex;flex-shrink:0;height:20px;width:36px;", trackColor)},
			},
		},
		{
			ID: id + "-thumb",
			Edits: []ownframe.ElementEdit{
				{Attribute: "style", Value: fmt.Sprintf("background-color:%s;border-radius:8px;flex-shrink:0;height:16px;margin-left:%s;width:16px;", thumbColor, marginLeft)},
			},
		},
	}
	return a.patchElements(ctx, patches)
}

func (a *app) patchCheckbox(ctx context.Context, id string, checked bool) error {
	markText := ""
	class := "of-checkbox"
	if checked {
		markText = "✓"
		class = "of-checkbox of-checkbox--checked"
	}
	patches := []ownframe.ElementPatch{
		{
			ID: id,
			Edits: []ownframe.ElementEdit{
				{Attribute: "aria-checked", Value: boolString(checked)},
				{Attribute: "class", Value: class},
			},
		},
		{
			ID: id + "-mark",
			Edits: []ownframe.ElementEdit{
				{Text: &markText},
			},
		},
	}
	return a.patchElements(ctx, patches)
}

func (a *app) patchSlider(ctx context.Context, id string, value, min, max float64, label string) error {
	fill := (value - min) / (max - min) * percentageScale
	valStr := strconv.Itoa(int(value))
	patches := []ownframe.ElementPatch{
		{
			ID: id,
			Edits: []ownframe.ElementEdit{
				{Attribute: "aria-valuenow", Value: valStr},
			},
		},
		{
			ID: id + "-range",
			Edits: []ownframe.ElementEdit{
				{Attribute: "style", Value: fmt.Sprintf("width:%.1f%%;", fill)},
			},
		},
	}
	if label != "" {
		labelID := id + "-text"
		if id == "alert-threshold-slider" {
			labelID = "alert-threshold-badge"
		}
		patches = append(patches, ownframe.ElementPatch{
			ID: labelID,
			Edits: []ownframe.ElementEdit{
				{Text: &label},
			},
		})
	}
	if id == "alert-threshold-slider" {
		description := fmt.Sprintf("Five minutes above %d per 1,000 requests.", int(value))
		patches = append(patches, ownframe.ElementPatch{
			ID: "alert-threshold-slider-description",
			Edits: []ownframe.ElementEdit{
				{Text: &description},
			},
		})
	}
	return a.patchElements(ctx, patches)
}

func (a *app) configureLiveTick() {
	if a.page == nil {
		return
	}
	if !a.tooltipOpen && !a.autoRefresh {
		a.page.SetTick(nil)
		a.nextLiveTick = time.Time{}
		return
	}
	a.page.SetTick(func(ctx context.Context) error { return a.tickLiveAt(ctx, time.Now()) })
	if a.nextLiveTick.IsZero() {
		a.nextLiveTick = time.Now().Add(time.Second)
	}
	a.page.WantFrameAt(a.nextLiveTick)
}

func (a *app) tickLiveAt(ctx context.Context, now time.Time) error {
	if now.Before(a.nextLiveTick) {
		return nil
	}
	a.nextLiveTick = now.Add(time.Second)
	a.page.WantFrameAt(a.nextLiveTick)
	if a.autoRefresh {
		if a.nextDataRefresh.IsZero() {
			a.nextDataRefresh = now
		}
		if !now.Before(a.nextDataRefresh) {
			a.lastRefreshCount = a.feed.refresh()
			a.lastUpdatedAt = now
			a.nextDataRefresh = now.Add(time.Duration(dataRefreshSeconds) * time.Second)
			if a.liveTail {
				a.pageIndex = 0
				a.tracePageIndex = 0
			}
			return a.refresh(ctx)
		}
	}
	if !a.tooltipOpen {
		return nil
	}
	label := a.liveUpdatedLabel(now)
	if label == a.liveLabel {
		return nil
	}
	a.liveLabel = label
	return a.patchHover(ctx, "live-tooltip-content", ownframe.ElementEdit{Text: &label})
}
