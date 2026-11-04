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
	if _, _, ok := parseHeatKey("1-3"); !ok {
		t.Fatalf("parseHeatKey failed")
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

func TestMailpulseHeatSelectAndChange(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatalf("newApp: %v", err)
	}
	ctx := context.Background()
	if err := a.handleClick(ctx, clickBox("select:heat:1-3:Tue 9a")); err != nil {
		t.Fatalf("heat select: %v", err)
	}
	if a.heatDay != 1 || a.heatCol != 3 {
		t.Fatalf("heat cell not pinned: %d %d", a.heatDay, a.heatCol)
	}
	if !a.isSelected("heat", "1-3") {
		t.Fatalf("heat selection not set")
	}
	a.subQuery = "emma"
	if _, err := ui.Render(a.subscribersPage()); err != nil {
		t.Fatalf("render filtered subscribers: %v", err)
	}
}
