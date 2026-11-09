package main

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// loglensLayoutBox finds one box in the current frame by id, text, or tag.
func loglensLayoutBox(t *testing.T, p *ownframe.Page, match func(ownframe.Box) bool) ownframe.Box {
	t.Helper()
	for _, box := range p.Boxes() {
		if match(box) {
			return box
		}
	}
	t.Fatal("expected box is missing from the frame")
	return ownframe.Box{}
}

// TestOverviewRegionCardsFit checks the lower overview row stays in two
// separate columns and the on-call calendar receives usable width.
func TestOverviewRegionCardsFit(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	a.calendarMonth = time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC)
	a.onCallDate = time.Date(2026, time.November, 10, 0, 0, 0, 0, time.UTC)
	a.page.SetSize(windowWidth, 2400)
	if err := a.page.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	var onCall, regions ownframe.Box
	var calendar ownframe.Box
	var regionMap ownframe.Box
	for _, box := range a.page.Boxes() {
		if box.Tag == "section" && strings.HasPrefix(box.Text, "On-call rotation") {
			onCall = box
		}
		if box.Tag == "section" && strings.HasPrefix(box.Text, "Edge regions") {
			regions = box
		}
		if box.Text == "November 2026" {
			calendar = box
		}
		if box.ID == "overview-region-map" {
			regionMap = box
		}
	}
	if onCall.W <= 0 || regions.W <= 0 {
		t.Fatalf("overview cards have no layout width: on-call=%+v regions=%+v", onCall, regions)
	}
	if onCall.X+onCall.W > regions.X {
		t.Fatalf("overview cards overlap: on-call=%+v regions=%+v", onCall, regions)
	}
	if calendar.W <= 0 || calendar.X < onCall.X || calendar.X+calendar.W > onCall.X+onCall.W {
		t.Fatalf("calendar does not fit inside the on-call card: calendar=%+v card=%+v", calendar, onCall)
	}
	if regionMap.W <= 0 || regionMap.H <= 0 || regionMap.W/regionMap.H < 3.2 || regionMap.W/regionMap.H > 3.6 {
		t.Fatalf("region map does not keep its wide ratio: %+v", regionMap)
	}
}

func TestSidebarBrandFitsHeader(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	a.page.SetSize(minWidth, minHeight)
	if err := a.page.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	sidebar := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool { return box.ID == "app-sidebar" })
	if sidebar.W < 260 {
		t.Fatalf("sidebar width was not increased for the project label: %+v", sidebar)
	}
	for _, label := range []string{"ACME", "Observability", "Production"} {
		if !strings.Contains(a.page.HTML(), label) {
			t.Fatalf("sidebar project label %q is missing from the page", label)
		}
	}
	if !strings.Contains(appCSS, ".app-sidebar-project-name { display: flex; flex-direction: column;") {
		t.Fatal("sidebar project name no longer uses explicit stacked lines")
	}
}

func TestRegionErrorsShowPointDetailsOnHover(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	a.page.SetSize(windowWidth, 2400)
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	cell1 := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool { return box.ID == "region-map-1" })
	if err := a.page.Edit().Hover(ctx, cell1.X+cell1.W/2, cell1.Y+cell1.H/2); err != nil {
		t.Fatal(err)
	}
	cellTip := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool {
		return box.Tag == "span" && strings.Contains(box.Text, "us-west-1") && strings.Contains(box.Text, "errors")
	})
	if cellTip.W <= 0 || !strings.Contains(cellTip.Text, "us-west-1") {
		t.Fatalf("cell hover detail is missing point data: %+v", cellTip)
	}
	// Row 1 tooltip should be positioned horizontally near the cell and above it
	if cellTip.X < cell1.X-30 || cellTip.X > cell1.X+cell1.W {
		t.Fatalf("cell hover tooltip is not positioned over the cell: cell.X=%.1f tip.X=%.1f", cell1.X, cellTip.X)
	}
	if cellTip.Y >= cell1.Y {
		t.Fatalf("row 1 tooltip is not positioned above the cell: cell.Y=%.1f tip.Y=%.1f", cell1.Y, cellTip.Y)
	}

	// Test hovering Row 2 (e.g. region-map-8):
	// 1. Tooltip must be below row 2 (not above, so it never overlaps row 1)
	// 2. Row 1 position must NOT move or shift!
	cell8 := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool { return box.ID == "region-map-8" })
	if err := a.page.Edit().Hover(ctx, cell8.X+cell8.W/2, cell8.Y+cell8.H/2); err != nil {
		t.Fatal(err)
	}
	cell8Tip := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool {
		return box.Tag == "span" && strings.Contains(box.Text, "eu-central-1") && strings.Contains(box.Text, "errors")
	})
	if cell8Tip.W <= 0 || !strings.Contains(cell8Tip.Text, "eu-central-1") {
		t.Fatalf("row 2 hover detail is missing point data: %+v", cell8Tip)
	}
	if cell8Tip.Y <= cell8.Y {
		t.Fatalf("row 2 tooltip must be positioned below row 2: cell8.Y=%.1f tip.Y=%.1f", cell8.Y, cell8Tip.Y)
	}
	// Check row 1 didn't move
	cell1After := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool { return box.ID == "region-map-1" })
	if cell1After.X != cell1.X || cell1After.W != cell1.W {
		t.Fatalf("row 1 shifted when row 2 hovered: X: %.1f -> %.1f, W: %.1f -> %.1f", cell1.X, cell1After.X, cell1.W, cell1After.W)
	}

	bar := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool { return box.ID == "region-heat-13" })
	before := a.page.Probe().Stats()
	if err := a.page.Edit().Hover(ctx, bar.X+bar.W/2, bar.Y+bar.H/2); err != nil {
		t.Fatal(err)
	}
	tip := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool {
		return box.Tag == "span" && strings.Contains(box.Text, "us-east-1") && strings.Contains(box.Text, "92 errors")
	})
	if tip.W <= 0 || !strings.Contains(tip.Text, "us-east-1") || !strings.Contains(tip.Text, "92 errors") {
		t.Fatalf("hover detail is missing point data: %+v", tip)
	}
	after := a.page.Probe().Stats()
	if after.Parses != before.Parses || after.Cascades != before.Cascades {
		t.Fatalf("hover rebuilt the page: parses %d -> %d, cascades %d -> %d", before.Parses, after.Parses, before.Cascades, after.Cascades)
	}
}

func TestButtonHoverUsesScopedLayout(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	a.page.SetSize(windowWidth, 2400)
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	before := a.page.Probe().Stats()
	btn := loglensLayoutBox(t, a.page, func(b ownframe.Box) bool { return b.ID == "on-call-swap-btn" })
	if err := a.page.Edit().Hover(ctx, btn.X+btn.W/2, btn.Y+btn.H/2); err != nil {
		t.Fatal(err)
	}
	after := a.page.Probe().Stats()
	if after.HoverPatches != before.HoverPatches+1 || after.FullLayouts != before.FullLayouts {
		t.Fatalf("hover used full layout: before=%+v after=%+v", before, after)
	}
	if after.HoverPatchErrors != before.HoverPatchErrors || after.HoverFallbacks != before.HoverFallbacks {
		t.Fatalf("hover retried or fell back: before=%+v after=%+v", before, after)
	}
	if after.Parses != before.Parses || after.Cascades != before.Cascades {
		t.Fatal("button hover parsed or cascaded the document")
	}
}
func TestOnCallCalendarSequentialClicks(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	a.calendarMonth = time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC)
	a.onCallDate = time.Date(2026, time.November, 10, 0, 0, 0, 0, time.UTC)
	ctx := context.Background()
	a.page.SetSize(windowWidth, 2400)
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	for _, dayCell := range []string{
		"on-call-calendar-day-15",
		"on-call-calendar-day-16",
		"on-call-calendar-day-17",
		"on-call-calendar-day-18",
		"on-call-calendar-day-19",
	} {
		box := loglensLayoutBox(t, a.page, func(b ownframe.Box) bool { return b.ID == dayCell })
		t.Logf("Clicking %s at (%.1f, %.1f) Action=%s", dayCell, box.X+box.W/2, box.Y+box.H/2, box.Action)
		if err := a.page.Click(ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
			t.Fatalf("failed to click %s: %v", dayCell, err)
		}
		if dayCell == "on-call-calendar-day-16" {
			t.Log("Simulating a.refresh(ctx) (e.g. data refresh / live tail)")
			if err := a.refresh(ctx); err != nil {
				t.Fatalf("refresh failed: %v", err)
			}
		}
		expectedResponder := a.onCallFor(a.onCallDate)
		t.Logf("Selected onCallDate: %v, responder: %s", a.onCallDate, expectedResponder.Name)
	}

	// Test clicking an outside day (e.g. cell 30 which is Dec 1 in Nov 2026 grid)
	outsideCell := loglensLayoutBox(t, a.page, func(b ownframe.Box) bool { return b.ID == "on-call-calendar-day-30" })
	if outsideCell.Action == "" {
		t.Fatal("outside day cell 30 has no click action registered")
	}
	if err := a.page.Click(ctx, outsideCell.X+outsideCell.W/2, outsideCell.Y+outsideCell.H/2); err != nil {
		t.Fatalf("failed to click outside day cell 30: %v", err)
	}
	if a.onCallDate.Month() != time.December || a.onCallDate.Day() != 1 {
		t.Fatalf("clicking outside day did not select Dec 1: %v", a.onCallDate)
	}
	if a.calendarMonth.Month() != time.December {
		t.Fatalf("clicking outside day did not advance calendarMonth to December: %v", a.calendarMonth)
	}
}

// TestSettingsTabsKeepHiddenValues moves through the settings sections and
// saves, proving controls on an inactive section still reach Page.Form.
func TestSettingsTabsKeepHiddenValues(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	a.settingsOpen = true
	if err := a.refresh(ctx); err != nil {
		t.Fatal(err)
	}
	a.setFormValue("settings-email", "oncall@example.com")
	a.setFormValue("settings-notes", "handover")
	for index := range settingsOTPLength {
		a.setFormValue("settings-otp-"+string(rune('0'+index)), string(rune('1'+index)))
	}
	for _, tab := range []string{"noise", "verification"} {
		a.settingsTab = tab
		if err := a.refresh(ctx); err != nil {
			t.Fatal(err)
		}
	}
	save := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool {
		return box.Tag == "button" && box.Text == "Save"
	})
	if err := a.page.Click(ctx, save.X+save.W/2, save.Y+save.H/2); err != nil {
		t.Fatal(err)
	}
	if a.settingsOpen {
		t.Fatal("Save left the settings sheet open")
	}
	if a.email != "oncall@example.com" || a.notes != "handover" || a.otp != "123456" {
		t.Fatalf("hidden values lost: email=%q notes=%q otp=%q", a.email, a.notes, a.otp)
	}
}

// TestDialogFitsMinimumWindow checks the create-alert controls stay above the
// footer at the 1100x720 minimum instead of scrolling under it.
func TestDialogFitsMinimumWindow(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	a.page.SetSize(minWidth, minHeight)
	a.dialogOpen = true
	if err := a.refresh(ctx); err != nil {
		t.Fatal(err)
	}
	footer := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool {
		return box.Tag == "footer" && box.Text == "CancelSave"
	})
	for _, id := range []string{"alert-enabled", "alert-page"} {
		box := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool { return box.ID == id })
		if box.Y+box.H > footer.Y {
			t.Fatalf("%s ends at %.0f, under the footer at %.0f", id, box.Y+box.H, footer.Y)
		}
	}
}

// TestDrawerTabsFitMinimumWindow checks both drawer tabs end above the footer
// at the 1100x720 minimum.
func TestDrawerTabsFitMinimumWindow(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	a.page.SetSize(minWidth, minHeight)
	a.drawerOpen = true
	a.drawerIndex = 0
	for _, tab := range []string{"payload", "trace"} {
		a.drawerTab = tab
		if err := a.refresh(ctx); err != nil {
			t.Fatal(err)
		}
		footer := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool {
			return box.Tag == "footer" && box.Text == "Close"
		})
		note := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool {
			return box.Text == "Retained for 30 days. Tokens and passwords are redacted in exports."
		})
		if note.Y+note.H > footer.Y {
			t.Fatalf("%s content ends at %.0f, under the footer at %.0f", tab, note.Y+note.H, footer.Y)
		}
	}
}

func loglensSameBox(a, b ownframe.Box) bool {
	return a.X == b.X && a.Y == b.Y && a.W == b.W && a.H == b.H
}

// TestTopbarChromeStaysClear keeps the brand, nav, and menubar from
// overlapping and proves an open nav panel does not move them.
func TestTopbarChromeStaysClear(t *testing.T) {
	for _, width := range []int{minWidth, windowWidth} {
		a, err := newApp()
		if err != nil {
			t.Fatal(err)
		}
		ctx := context.Background()
		a.page.SetSize(width, minHeight)
		if err := a.refresh(ctx); err != nil {
			t.Fatal(err)
		}
		brand := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool { return box.Text == "Loglens" })
		nav := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool { return box.ID == "top-nav" })
		menubar := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool { return box.ID == "top-menubar" })
		if brand.X+brand.W > nav.X {
			t.Fatalf("%d: brand ends at %.0f under the nav at %.0f", width, brand.X+brand.W, nav.X)
		}
		if nav.X+nav.W > menubar.X {
			t.Fatalf("%d: nav ends at %.0f under the menubar at %.0f", width, nav.X+nav.W, menubar.X)
		}
		closeBox := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool {
			return box.Tag == "button" && box.Text == "×"
		})
		if closeBox.X+closeBox.W > float64(width) {
			t.Fatalf("%d: window controls end at %.0f", width, closeBox.X+closeBox.W)
		}
		a.navOpen = "product-platform"
		if err := a.refresh(ctx); err != nil {
			t.Fatal(err)
		}
		movedNav := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool { return box.ID == "top-nav" })
		movedMenubar := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool { return box.ID == "top-menubar" })
		if !loglensSameBox(nav, movedNav) || !loglensSameBox(menubar, movedMenubar) {
			t.Fatalf("%d: open panel moved the chrome: nav %+v -> %+v menubar %+v -> %+v", width, nav, movedNav, menubar, movedMenubar)
		}
		panel := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool { return box.ID == "product-platform-content" })
		if panel.Y < nav.Y+nav.H {
			t.Fatalf("%d: nav panel starts at %.0f inside the trigger row ending at %.0f", width, panel.Y, nav.Y+nav.H)
		}
	}
}

// TestSettingsTabsFitMinimumWindow checks each settings section keeps its
// fields above the Save footer at the 1100x720 minimum.
func TestSettingsTabsFitMinimumWindow(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	a.page.SetSize(minWidth, minHeight)
	a.settingsOpen = true
	checks := map[string][]string{
		"routing":      {"settings-email", "settings-channel", "settings-webhook"},
		"noise":        {"settings-threshold"},
		"verification": {"settings-otp-5", "settings-notes"},
	}
	for _, tab := range []string{"routing", "noise", "verification"} {
		a.settingsTab = tab
		if err := a.refresh(ctx); err != nil {
			t.Fatal(err)
		}
		footer := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool {
			return box.Tag == "button" && box.Text == "Save"
		})
		for _, id := range checks[tab] {
			box := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool { return box.ID == id })
			if box.Y+box.H > footer.Y {
				t.Fatalf("%s: %s ends at %.0f under Save at %.0f", tab, id, box.Y+box.H, footer.Y)
			}
		}
	}
}

// TestOnCallResponderSelectionAndSwap verifies the on-call calendar day selection
// and responder swapping updates the card header dynamically.
func TestOnCallResponderSelectionAndSwap(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	a.page.SetSize(windowWidth, 2400)
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	// Verify Swap button is present and clickable
	swapBtn := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool {
		return box.ID == "on-call-swap-btn"
	})
	if err := a.page.Click(ctx, swapBtn.X+swapBtn.W/2, swapBtn.Y+swapBtn.H/2); err != nil {
		t.Fatal(err)
	}
	if !a.onCallDialogOpen {
		t.Fatal("clicking Swap did not open the on-call dialog")
	}

	// Swap responder to Liam Taylor
	liam := onCallResponder{Name: "Liam Taylor", Role: "Primary", Initials: "LT", Region: "Americas"}
	targetDate := time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC)
	if err := a.setOnCall(ctx, targetDate, liam); err != nil {
		t.Fatal(err)
	}
	a.onCallDialogOpen = false
	if err := a.refresh(ctx); err != nil {
		t.Fatal(err)
	}

	// Verify the card header now reflects Liam Taylor
	if !strings.Contains(a.page.HTML(), "Liam Taylor") {
		t.Fatal("on-call card header does not display Liam Taylor after assignment")
	}
	if !strings.Contains(a.page.HTML(), "Americas") {
		t.Fatal("on-call card header does not display Americas after assignment")
	}
}

// TestOverviewRegionCellsFillWidth checks that the region cell grid adapts and fills
// wider than the old 216px fixed box in standard view.
func TestOverviewRegionCellsFillWidth(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	a.page.SetSize(windowWidth, 2400)
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	cell1 := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool { return box.ID == "region-map-1" })
	cell7 := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool { return box.ID == "region-map-7" })
	totalRowSpan := (cell7.X + cell7.W) - cell1.X
	if totalRowSpan < 300 {
		t.Fatalf("region cells row span %.1f is not filling width responsively (expected >= 300px)", totalRowSpan)
	}
	if !strings.Contains(appCSS, "pointer-events: none") {
		t.Fatal("app-heat-tooltip is missing pointer-events: none")
	}
}

// TestFineGrainedStateUpdatesUseElementPatching verifies that micro-interactions
// like selecting an on-call calendar day, toggling filters, and auto-refresh use
// fine-grained element patching rather than full page reparsing/cascading.
func TestFineGrainedStateUpdatesUseElementPatching(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	a.page.SetSize(windowWidth, 2400)
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	// 1. On-call calendar selection test
	targetDayBox := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool {
		return box.ID == "on-call-calendar-day-18"
	})

	statsBefore := a.page.Probe().Stats()
	if err := a.page.Click(ctx, targetDayBox.X+targetDayBox.W/2, targetDayBox.Y+targetDayBox.H/2); err != nil {
		t.Fatal(err)
	}

	statsAfter := a.page.Probe().Stats()
	if statsAfter.Parses != statsBefore.Parses || statsAfter.Cascades != statsBefore.Cascades {
		t.Fatalf("calendar day selection triggered full reparse: parses %d -> %d, cascades %d -> %d",
			statsBefore.Parses, statsAfter.Parses, statsBefore.Cascades, statsAfter.Cascades)
	}

	// Verify responder details updated in rendered display boxes
	expectedResponder := a.onCallFor(a.onCallDate)
	nameFound := false
	for _, box := range a.page.Boxes() {
		if strings.Contains(box.Text, expectedResponder.Name) {
			nameFound = true
			break
		}
	}
	if !nameFound {
		t.Fatalf("Rendered boxes missing expected responder name %q after patch", expectedResponder.Name)
	}

	// 2. Switch to logs tab and test filter micro-interactions
	if err := a.setTab(ctx, "logs"); err != nil {
		t.Fatal(err)
	}
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	// Test Live tail toggle patching
	liveTailBox := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool {
		return box.ID == "log-live-tail"
	})
	statsBefore = a.page.Probe().Stats()
	if err := a.page.Click(ctx, liveTailBox.X+liveTailBox.W/2, liveTailBox.Y+liveTailBox.H/2); err != nil {
		t.Fatal(err)
	}
	statsAfter = a.page.Probe().Stats()
	if statsAfter.Parses != statsBefore.Parses || statsAfter.Cascades != statsBefore.Cascades {
		t.Fatalf("live tail toggle triggered full reparse: parses %d -> %d, cascades %d -> %d",
			statsBefore.Parses, statsAfter.Parses, statsBefore.Cascades, statsAfter.Cascades)
	}
	if !a.liveTail {
		t.Fatal("expected liveTail to be true")
	}

	// Test Show context checkbox patching
	contextBox := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool {
		return box.ID == "log-context"
	})
	statsBefore = a.page.Probe().Stats()
	if err := a.page.Click(ctx, contextBox.X+contextBox.W/2, contextBox.Y+contextBox.H/2); err != nil {
		t.Fatal(err)
	}
	statsAfter = a.page.Probe().Stats()
	if statsAfter.Parses != statsBefore.Parses || statsAfter.Cascades != statsBefore.Cascades {
		t.Fatalf("show context toggle triggered full reparse: parses %d -> %d, cascades %d -> %d",
			statsBefore.Parses, statsAfter.Parses, statsBefore.Cascades, statsAfter.Cascades)
	}
	if !a.showContext {
		t.Fatal("expected showContext to be true")
	}

	// Test Auto refresh switch patching
	autoRefreshBox := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool {
		return box.ID == "log-auto-refresh"
	})
	statsBefore = a.page.Probe().Stats()
	if err := a.page.Click(ctx, autoRefreshBox.X+autoRefreshBox.W/2, autoRefreshBox.Y+autoRefreshBox.H/2); err != nil {
		t.Fatal(err)
	}
	statsAfter = a.page.Probe().Stats()
	if statsAfter.Parses != statsBefore.Parses || statsAfter.Cascades != statsBefore.Cascades {
		t.Fatalf("auto refresh toggle triggered full reparse: parses %d -> %d, cascades %d -> %d",
			statsBefore.Parses, statsAfter.Parses, statsBefore.Cascades, statsAfter.Cascades)
	}
	if !a.autoRefresh {
		t.Fatal("expected autoRefresh to be true")
	}

	// Test Sample rate slider cycling (from 100% it cycles to 10%)
	sampleRateBox := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool {
		return box.ID == "log-sample-rate"
	})
	if err := a.page.Click(ctx, sampleRateBox.X+sampleRateBox.W/2, sampleRateBox.Y+sampleRateBox.H/2); err != nil {
		t.Fatal(err)
	}
	if a.sampleRate != 10 {
		t.Fatalf("expected sampleRate 10, got %d", a.sampleRate)
	}
}
