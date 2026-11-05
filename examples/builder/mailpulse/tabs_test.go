package main

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/ui"
)

func clickBox(action string) ownframe.Box { return ownframe.Box{Action: action} }

func TestMailpulseAllTabsRender(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatalf("newApp: %v", err)
	}
	tabs := []string{"overview", "audience", "subscribers", "campaigns", "content", "engagement", "revenue", "automations", "segments", "integrations", "analytics", "reports", "learn", "help", "settings"}
	for _, tab := range tabs {
		a.tab = tab
		body := a.contentBody()
		if body == nil {
			t.Fatalf("%s: contentBody nil", tab)
		}
		if _, err := ui.Render(body); err != nil {
			t.Fatalf("ui.Render(%s): %v", tab, err)
		}
		if _, err := ui.Render(a.screen()); err != nil {
			t.Fatalf("ui.Render(screen %s): %v", tab, err)
		}
		title, sub := a.pageTitle()
		if title == "" || sub == "" {
			t.Fatalf("%s: empty pageTitle", tab)
		}
	}
}

func TestMailpulseNavAndSelect(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatalf("newApp: %v", err)
	}
	ctx := context.Background()
	if err := a.handleClick(ctx, clickBox("nav:overview")); err != nil {
		t.Fatalf("nav overview: %v", err)
	}
	if a.tab != "overview" {
		t.Fatalf("tab = %q, want overview", a.tab)
	}
	if err := a.handleClick(ctx, clickBox("nav:help center")); err != nil {
		t.Fatalf("nav help center: %v", err)
	}
	if a.tab != "help" {
		t.Fatalf("tab = %q, want help", a.tab)
	}
	if err := a.handleClick(ctx, clickBox("nav:system settings")); err != nil {
		t.Fatalf("nav settings: %v", err)
	}
	if a.tab != "settings" {
		t.Fatalf("tab = %q, want settings", a.tab)
	}
	if err := a.handleClick(ctx, clickBox("select:campaign:c1:Weekly Brief")); err != nil {
		t.Fatalf("select: %v", err)
	}
	if !a.isSelected("campaign", "c1") {
		t.Fatalf("selection not set")
	}
	if !strings.Contains(a.toast, "Weekly Brief") {
		t.Fatalf("toast = %q, want selection label", a.toast)
	}
	if err := a.handleClick(ctx, clickBox("select:campaign:c1:Weekly Brief")); err != nil {
		t.Fatalf("toggle off: %v", err)
	}
	if a.isSelected("campaign", "c1") {
		t.Fatalf("selection should clear on second click")
	}
}

func TestMailpulseNotifDismissFlow(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatalf("newApp: %v", err)
	}
	if len(a.notifs) == 0 {
		t.Fatalf("expected seeded notifications")
	}
	ctx := context.Background()
	a.notifOpen = true
	if _, err := ui.Render(a.notifTrigger()); err != nil {
		t.Fatalf("render bell trigger: %v", err)
	}
	for _, n := range append([]notifItem(nil), a.notifs...) {
		if err := a.handleClick(ctx, clickBox("notif:dismiss:"+n.id)); err != nil {
			t.Fatalf("dismiss %s: %v", n.id, err)
		}
	}
	if len(a.notifs) != 0 {
		t.Fatalf("notifs = %d, want 0", len(a.notifs))
	}
	out, err := ui.Render(a.screen())
	if err != nil {
		t.Fatalf("render screen: %v", err)
	}
	if !strings.Contains(out, "No new notifications") {
		t.Fatalf("empty state missing")
	}
}

func TestMailpulseHoverCSSWired(t *testing.T) {
	if !strings.Contains(hoverCSS, ".mp-selectable") {
		t.Fatalf("hoverCSS missing mp-selectable")
	}
	sheet := (&app{}).themeSheet()
	if !strings.Contains(sheet, ".mp-selectable") {
		t.Fatalf("themeSheet missing hoverCSS")
	}
}

func TestMailpulseExportAndRangeEveryTab(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatalf("newApp: %v", err)
	}
	ctx := context.Background()
	tabs := []string{"overview", "audience", "subscribers", "campaigns", "content", "engagement", "revenue", "automations", "segments", "integrations", "analytics", "reports", "learn", "help", "settings"}
	for _, tab := range tabs {
		a.tab = tab
		if err := a.handleClick(ctx, clickBox("export")); err != nil {
			t.Fatalf("export on %s: %v", tab, err)
		}
		if !a.toastOpen || !strings.Contains(a.toast, "Export queued") {
			t.Fatalf("export toast missing on %s: %q", tab, a.toast)
		}
		a.toastOpen = false
		before := a.rangeLabel
		if err := a.handleClick(ctx, clickBox("range:cycle")); err != nil {
			t.Fatalf("range on %s: %v", tab, err)
		}
		if a.rangeLabel == before {
			t.Fatalf("range label did not cycle on %s", tab)
		}
	}
}

func TestMailpulseHeatCellLabels(t *testing.T) {
	if got := heatCellLabel(1, 3); !strings.Contains(got, "Tue") || !strings.Contains(got, "9a") {
		t.Fatalf("best cell label = %q, want Tue 9a", got)
	}
	if d, c, ok := parseHeatCellID("mp-heat-2-5"); !ok || d != 2 || c != 5 {
		t.Fatalf("parseHeatCellID failed: %d %d %v", d, c, ok)
	}
	if _, _, ok := parseHeatCellID("mp-sidebar"); ok {
		t.Fatalf("non-cell id parsed as cell")
	}
}

func TestMailpulseGrowthBarLabels(t *testing.T) {
	if len(overviewBarHeights()) != 12 || len(overviewBarValues()) != 12 {
		t.Fatalf("growth bars = %d heights %d values, want 12 and 12", len(overviewBarHeights()), len(overviewBarValues()))
	}
	total := 0
	for _, v := range overviewBarValues() {
		total += v
	}
	if total != 428 {
		t.Fatalf("growth values total = %d, want 428", total)
	}
	if got := growthBarLabel(6); !strings.Contains(got, "W7") || !strings.Contains(got, "+51") {
		t.Fatalf("peak bar label = %q, want W7 +51", got)
	}
	if i, ok := parseGrowthBarID("mp-growth-6"); !ok || i != 6 {
		t.Fatalf("parseGrowthBarID failed: %d %v", i, ok)
	}
	if _, ok := parseGrowthBarID("mp-heat-2-5"); ok {
		t.Fatalf("heat id parsed as growth bar")
	}
	if _, ok := parseGrowthBarID("mp-growth-12"); ok {
		t.Fatalf("out of range bar parsed")
	}
	a, err := newApp()
	if err != nil {
		t.Fatalf("newApp: %v", err)
	}
	a.tab = "overview"
	html, err := ui.Render(a.overviewGrowthCard())
	if err != nil {
		t.Fatalf("render growth card: %v", err)
	}
	if !strings.Contains(html, "mp-growth-6") || !strings.Contains(html, "mp-growth-readout") {
		t.Fatalf("growth card missing bar ids or readout")
	}
}

func TestMailpulseSubscriberSearch(t *testing.T) {
	if got := len(filterSubscribers("")); got != 8 {
		t.Fatalf("empty query = %d profiles, want 8", got)
	}
	ada := filterSubscribers("ada")
	if len(ada) != 1 || ada[0].name != "Ada Kim" {
		t.Fatalf("ada query = %+v", ada)
	}
	paid := filterSubscribers("PAID")
	if len(paid) != 4 {
		t.Fatalf("paid query = %d profiles, want 4", len(paid))
	}
	if len(filterSubscribers("zzz-no-match")) != 0 {
		t.Fatalf("nonsense query should match nothing")
	}
}

func TestMailpulseHeatHoverOnly(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatalf("newApp: %v", err)
	}
	html, err := ui.Render(a.heatmapCard())
	if err != nil {
		t.Fatalf("render heatmap: %v", err)
	}
	if strings.Contains(html, "select:heat") {
		t.Fatalf("heat cells must not carry a select action")
	}
	if !strings.Contains(html, "mp-heat-readout") {
		t.Fatalf("heatmap missing hover readout")
	}
	a.subQuery = "emma"
	if _, err := ui.Render(a.subscribersPage()); err != nil {
		t.Fatalf("render filtered subscribers: %v", err)
	}
}

// TestMailpulseRangeUpdatesData checks that the Overview, Engagement, and
// Audience tabs all respond to the range pill. Overview assertions cover
// overviewStatsFor/overviewSendsFor and the growth description; engagement
// and audience assertions only read the other owners' range-aware cards
// (kpisFor footnotes, rangeNewLabel, rangeWindow) without touching them.
func TestMailpulseRangeUpdatesData(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatalf("newApp: %v", err)
	}
	render := func(el *ui.Element) string {
		html, err := ui.Render(el)
		if err != nil {
			t.Fatalf("ui.Render: %v", err)
		}
		return html
	}

	if got := len(overviewSendsFor(7)); got != 2 {
		t.Fatalf("overviewSendsFor(7) = %d rows, want 2", got)
	}
	if got := len(overviewSendsFor(30)); got != 4 {
		t.Fatalf("overviewSendsFor(30) = %d rows, want 4", got)
	}
	if got := len(overviewSendsFor(90)); got != 6 {
		t.Fatalf("overviewSendsFor(90) = %d rows, want 6", got)
	}

	overview30 := render(a.overviewPage())
	if !strings.Contains(overview30, "+428 this month") {
		t.Fatalf("overview 30d missing +428 this month stat")
	}
	if !strings.Contains(overview30, "$12,480") {
		t.Fatalf("overview 30d missing $12,480 revenue stat")
	}
	if !strings.Contains(overview30, "last 30 days") {
		t.Fatalf("overview 30d missing last 30 days window")
	}

	a.rangeLabel = "Last 7 Days"
	overview7 := render(a.overviewPage())
	for _, want := range []string{"+96 this week", "$2,940", "last 7 days"} {
		if !strings.Contains(overview7, want) {
			t.Fatalf("overview 7d missing %q", want)
		}
	}
	for _, stale := range []string{"last 30 days", "+428 this month"} {
		if strings.Contains(overview7, stale) {
			t.Fatalf("overview 7d still contains %q", stale)
		}
	}
	if got := render(a.overviewSendsCard()); strings.Count(got, "select:send:") != 2 {
		t.Fatalf("overview 7d sends card = %d rows, want 2", strings.Count(got, "select:send:"))
	}
	if engagement7 := render(a.engagementPage()); !strings.Contains(engagement7, "102 replies this week") {
		t.Fatalf("engagement 7d missing weekly replies footnote")
	}
	audience7 := render(a.audiencePage())
	for _, want := range []string{"last 7 days", "New (7d)"} {
		if !strings.Contains(audience7, want) {
			t.Fatalf("audience 7d missing %q", want)
		}
	}
	if strings.Contains(audience7, "last 30 days") {
		t.Fatalf("audience 7d still contains last 30 days")
	}

	a.rangeLabel = "Last 90 Days"
	overview90 := render(a.overviewPage())
	for _, want := range []string{"$36,900", "+1,204 this quarter", "+$3,240 this quarter", "last 90 days"} {
		if !strings.Contains(overview90, want) {
			t.Fatalf("overview 90d missing %q", want)
		}
	}
	if strings.Contains(overview90, "last 30 days") {
		t.Fatalf("overview 90d still contains last 30 days")
	}
	if got := render(a.overviewSendsCard()); strings.Count(got, "select:send:") != 6 {
		t.Fatalf("overview 90d sends card = %d rows, want 6", strings.Count(got, "select:send:"))
	}
	if engagement90 := render(a.engagementPage()); !strings.Contains(engagement90, "1,180 replies this quarter") {
		t.Fatalf("engagement 90d missing quarterly replies footnote")
	}
	audience90 := render(a.audiencePage())
	for _, want := range []string{"last 90 days", "New (90d)"} {
		if !strings.Contains(audience90, want) {
			t.Fatalf("audience 90d missing %q", want)
		}
	}
}

func TestMailpulseOverviewRangeRedraws(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatalf("newApp: %v", err)
	}
	render := func(el *ui.Element) string {
		html, err := ui.Render(el)
		if err != nil {
			t.Fatalf("ui.Render: %v", err)
		}
		return html
	}

	// Bar values in each window sum to the window total in the readout.
	if total := sumInts(barValues(7)); total != 96 {
		t.Fatalf("barValues(7) total = %d, want 96", total)
	}
	if total := sumInts(barValues(90)); total != 1204 {
		t.Fatalf("barValues(90) total = %d, want 1204", total)
	}
	if total := sumInts(barValues(30)); total != 428 {
		t.Fatalf("barValues(30) total = %d, want 428", total)
	}

	// 7-day window: 7 daily bars, daily description, cooled heatmap.
	a.rangeLabel = "Last 7 Days"
	overview7 := render(a.overviewPage())
	for _, want := range []string{"per day, last 7 days", "+96 net new", "26%", "Last 2 campaigns"} {
		if !strings.Contains(overview7, want) {
			t.Fatalf("overview 7d missing %q", want)
		}
	}
	if strings.Contains(overview7, "mp-growth-7") {
		t.Fatalf("overview 7d renders an 8th bar, want 7 daily bars")
	}
	if !strings.Contains(overview7, "scaleY(") {
		t.Fatalf("overview 7d bars missing scaleY transform")
	}
	if got := heatValueFor(1, 3, 7); got != 4 {
		t.Fatalf("heatValueFor(Tue 9a, 7d) = %d, want pinned peak 4", got)
	}
	if got := heatValueFor(0, 2, 7); got != heatValue(0, 2)-1 {
		t.Fatalf("heatValueFor(Mon 8a, 7d) = %d, want dampened %d", got, heatValue(0, 2)-1)
	}

	// 90-day window: 12 weekly bars, quarterly totals, full heat.
	a.rangeLabel = "Last 90 Days"
	overview90 := render(a.overviewPage())
	for _, want := range []string{"per week, last 90 days", "+1,204 net new", "28%", "Last 6 campaigns"} {
		if !strings.Contains(overview90, want) {
			t.Fatalf("overview 90d missing %q", want)
		}
	}
	if !strings.Contains(overview90, "mp-growth-11") {
		t.Fatalf("overview 90d missing 12th bar")
	}

	// Default window keeps the hand-tuned 30-day track.
	a.rangeLabel = "Last 30 Days"
	if heights := barHeightsFor(30); len(heights) != len(overviewBarHeights()) || heights[0] != overviewBarHeights()[0] {
		t.Fatalf("barHeightsFor(30) changed the 30-day track")
	}
}

func sumInts(values []int) int {
	total := 0
	for _, v := range values {
		total += v
	}
	return total
}

// TestMailpulseRangeSwitchUpdatesDataHandsFree clicks the range pill twice
// and asserts every range-driven readout follows with no hover and no ticks:
// the committed display must already carry the new window. Motion sampling
// (the 3s entrance rise and the 0.9s toggle glide) is covered by
// TestMailpulseEntranceGrowsFromZero and live-window observation; this test
// pins the hands-free data contract behind the animation.
func TestMailpulseRangeSwitchUpdatesDataHandsFree(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatalf("newApp: %v", err)
	}
	ctx := context.Background()
	a.tab = "overview"
	if err := a.refresh(ctx); err != nil {
		t.Fatalf("refresh overview: %v", err)
	}
	barScaleY := func(id string) float64 {
		t.Helper()
		var box ownframe.Box
		for _, b := range a.page.Boxes() {
			if b.ID == id {
				box = b
				break
			}
		}
		if box.ID == "" {
			t.Fatalf("element %q has no hit-test box", id)
		}
		owned := make(map[uint64]struct{}, len(box.OperationIDs))
		for _, oid := range box.OperationIDs {
			owned[oid] = struct{}{}
		}
		for i := range a.page.Probe().Display().Ops {
			op := &a.page.Probe().Display().Ops[i]
			if _, ok := owned[op.ID]; ok && op.Kind == ownframe.DisplayOpFillRect {
				return op.Transform().D
			}
		}
		t.Fatalf("element %q has no fill operation", id)
		return 0
	}
	var rangeBtn ownframe.Box
	for _, b := range a.page.Boxes() {
		if b.Action == "range:cycle" {
			rangeBtn = b
			break
		}
	}
	if rangeBtn.Action == "" {
		t.Fatalf("range:cycle control has no hit-test box")
	}

	from := barScaleY("mp-growth-0")
	clickRange := func() {
		t.Helper()
		if err := a.page.Click(ctx, rangeBtn.X+rangeBtn.W/2, rangeBtn.Y+rangeBtn.H/2); err != nil {
			t.Fatalf("click range:cycle: %v", err)
		}
	}
	clickRange()
	clickRange()
	if a.rangeLabel != "Last 7 Days" {
		t.Fatalf("rangeLabel = %q, want Last 7 Days", a.rangeLabel)
	}
	to := barScaleY("mp-growth-0")
	if from == to {
		t.Fatalf("bar scale did not change: %v", from)
	}
	if want := barHeightsFor(7)[0] / 100; to != want {
		t.Fatalf("7d bar scale = %v, want %v", to, want)
	}
	if got := a.growthReadout(); got != "+96 net new subscribers over the last 7 days." {
		t.Fatalf("7d growth readout = %q", got)
	}
	overview7, err := ui.Render(a.overviewPage())
	if err != nil {
		t.Fatalf("render overview 7d: %v", err)
	}
	if !strings.Contains(overview7, "per day, last 7 days") {
		t.Fatalf("7d overview missing daily growth description")
	}
	if stats := a.page.Probe().Motion(); stats.TransitionError != "" {
		t.Fatalf("transition bind failed: %s", stats.TransitionError)
	}
}

// TestMailpulseEntranceGrowsFromZero pins the transition-based entrance:
// fresh mounts commit bars at scaleY(0) under the mp-bar-slow 3s transition,
// the escort tick flips them to their real values, and the 3s run then eases
// them up hands-free with the base 0.9s transition reserved for toggles. All
// three assertions read committed or sampled paint deterministically with
// zero sleeps.
func TestMailpulseEntranceGrowsFromZero(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatalf("newApp: %v", err)
	}
	ctx := context.Background()
	a.tab = "overview"
	if err := a.refresh(ctx); err != nil {
		t.Fatalf("refresh overview: %v", err)
	}
	barScaleY := func(id string) float64 {
		t.Helper()
		var box ownframe.Box
		for _, b := range a.page.Boxes() {
			if b.ID == id {
				box = b
				break
			}
		}
		if box.ID == "" {
			t.Fatalf("element %q has no hit-test box", id)
		}
		owned := make(map[uint64]struct{}, len(box.OperationIDs))
		for _, oid := range box.OperationIDs {
			owned[oid] = struct{}{}
		}
		for i := range a.page.Probe().Display().Ops {
			op := &a.page.Probe().Display().Ops[i]
			if _, ok := owned[op.ID]; ok && op.Kind == ownframe.DisplayOpFillRect {
				return op.Transform().D
			}
		}
		t.Fatalf("element %q has no fill operation", id)
		return 0
	}

	target := overviewBarHeights()[0] / 100
	if got := barScaleY("mp-growth-0"); got != 0 {
		t.Fatalf("fresh mount bar scale = %v, want 0 (zeros committed)", got)
	}

	// First tick only arms the flip (lets the zero commit present on its own
	// frame; the window batches one frame's refreshes into a single commit).
	if err := a.page.Testing().Tick(ctx); err != nil {
		t.Fatalf("escort arm tick failed: %v", err)
	}
	if got := barScaleY("mp-growth-0"); got != 0 {
		t.Fatalf("pre-flip bar scale = %v, want 0 (still zeros)", got)
	}

	if err := a.page.Testing().Tick(ctx); err != nil {
		t.Fatalf("escort flip tick failed: %v", err)
	}
	if got := barScaleY("mp-growth-0"); got != target {
		t.Fatalf("flipped bar scale = %v, want committed target %v", got, target)
	}

	if err := a.page.Testing().Tick(ctx); err != nil {
		t.Fatalf("entrance sample tick failed: %v", err)
	}
	if got := barScaleY("mp-growth-0"); !(got >= 0 && got < target) {
		t.Fatalf("entrance scale = %v, want 0 <= D < target %v", got, target)
	}
}

// TestMailpulseDotBarsHaveStableIDs checks that the dot share bars carry
// stable per-dot element ids so transform transitions and keyframes can
// target them instead of snapping (TransitionsNoID).
func TestMailpulseDotBarsHaveStableIDs(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatalf("newApp: %v", err)
	}
	audience, err := ui.Render(a.audiencePage())
	if err != nil {
		t.Fatalf("render audiencePage: %v", err)
	}
	if !strings.Contains(audience, "mp-src-dot-0") {
		t.Fatalf("audiencePage missing mp-src-dot-0")
	}
	if got := strings.Count(audience, "mp-src-dot-"); got < audienceDotTotal {
		t.Fatalf("audiencePage has %d dot ids, want at least %d", got, audienceDotTotal)
	}
	engage, err := ui.Render(a.engageCard())
	if err != nil {
		t.Fatalf("render engageCard: %v", err)
	}
	if !strings.Contains(engage, "mp-eng-dot-0") {
		t.Fatalf("engageCard missing mp-eng-dot-0")
	}
	if got := strings.Count(engage, "mp-eng-dot-"); got < engageDotCount {
		t.Fatalf("engageCard has %d dot ids, want at least %d", got, engageDotCount)
	}
}

// TestMailpulseEntranceReplaysOnRevisit proves the Overview entrance replays
// on every visit hands-free: each nav:overview re-arms the zero-scale mount
// and the escort flip, so leaving and returning replays the 0 -> target rise
// instead of playing once. All assertions drive explicit ticks with zero
// sleeps.
func TestMailpulseEntranceReplaysOnRevisit(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatalf("newApp: %v", err)
	}
	ctx := context.Background()
	barScaleY := func(id string) float64 {
		t.Helper()
		var box ownframe.Box
		for _, b := range a.page.Boxes() {
			if b.ID == id {
				box = b
				break
			}
		}
		if box.ID == "" {
			t.Fatalf("element %q has no hit-test box", id)
		}
		owned := make(map[uint64]struct{}, len(box.OperationIDs))
		for _, oid := range box.OperationIDs {
			owned[oid] = struct{}{}
		}
		for i := range a.page.Probe().Display().Ops {
			op := &a.page.Probe().Display().Ops[i]
			if _, ok := owned[op.ID]; ok && op.Kind == ownframe.DisplayOpFillRect {
				return op.Transform().D
			}
		}
		t.Fatalf("element %q has no fill operation", id)
		return 0
	}

	if err := a.handleClick(ctx, clickBox("nav:overview")); err != nil {
		t.Fatalf("nav overview: %v", err)
	}
	if !a.enterBars {
		t.Fatalf("enterBars = false after nav:overview, want true")
	}
	if _, err := ui.Render(a.overviewPage()); err != nil {
		t.Fatalf("render overviewPage: %v", err)
	}
	target := overviewBarHeights()[0] / 100
	if err := a.page.Testing().Tick(ctx); err != nil {
		t.Fatalf("seen tick failed: %v", err)
	}
	if got := barScaleY("mp-growth-0"); got != 0 {
		t.Fatalf("seen tick bar scale = %v, want 0", got)
	}
	if err := a.page.Testing().Tick(ctx); err != nil {
		t.Fatalf("flip tick failed: %v", err)
	}
	if got := barScaleY("mp-growth-0"); got != target {
		t.Fatalf("flipped bar scale = %v, want committed target %v", got, target)
	}
	if err := a.page.Testing().Tick(ctx); err != nil {
		t.Fatalf("glide tick failed: %v", err)
	}
	if got := barScaleY("mp-growth-0"); !(got > 0 && got < target) {
		t.Fatalf("glide bar scale = %v, want 0 < D < %v", got, target)
	}

	if err := a.handleClick(ctx, clickBox("nav:audience")); err != nil {
		t.Fatalf("nav audience: %v", err)
	}
	if err := a.handleClick(ctx, clickBox("nav:overview")); err != nil {
		t.Fatalf("revisit nav overview: %v", err)
	}
	if !a.enterBars {
		t.Fatalf("enterBars = false after revisit nav:overview, want true (re-armed)")
	}
	if err := a.page.Testing().Tick(ctx); err != nil {
		t.Fatalf("revisit seen tick failed: %v", err)
	}
	if got := barScaleY("mp-growth-0"); got != 0 {
		t.Fatalf("revisit seen tick bar scale = %v, want 0", got)
	}
	if err := a.page.Testing().Tick(ctx); err != nil {
		t.Fatalf("revisit flip tick failed: %v", err)
	}
	if got := barScaleY("mp-growth-0"); got != target {
		t.Fatalf("revisit flipped bar scale = %v, want committed target %v", got, target)
	}
}
