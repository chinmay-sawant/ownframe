package main

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestNewPOCard(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.handleClick(ctx, ownframe.Box{Action: "po:new"}); err != nil {
		t.Fatal(err)
	}
	if !a.poOpen {
		t.Fatalf("poOpen = false after po:new, want true")
	}
	html := a.page.HTML()
	for _, want := range []string{"New purchase order", "po:create"} {
		if !strings.Contains(html, want) {
			t.Fatalf("page HTML after po:new misses %q", want)
		}
	}
	if err := a.handleClick(ctx, ownframe.Box{Action: "po:create"}); err != nil {
		t.Fatal(err)
	}
	if a.poOpen {
		t.Fatalf("poOpen = true after po:create, want false")
	}
	if err := a.handleClick(ctx, ownframe.Box{Action: "po:new"}); err != nil {
		t.Fatal(err)
	}
	if !a.poOpen {
		t.Fatalf("poOpen = false after second po:new, want true")
	}
	if err := a.handleClick(ctx, ownframe.Box{Action: "po:close"}); err != nil {
		t.Fatal(err)
	}
	if a.poOpen {
		t.Fatalf("poOpen = true after po:close, want false")
	}
}

func TestSearchPalette(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.handleClick(ctx, ownframe.Box{Action: "search:open"}); err != nil {
		t.Fatal(err)
	}
	if !a.searchOpen {
		t.Fatalf("searchOpen = false after search:open, want true")
	}
	if html := a.page.HTML(); !strings.Contains(html, "inv-search-panel") {
		t.Fatalf("page HTML after search:open misses the inv-search-panel")
	}
	a.page.Form().SetFormValue("inv-search-input", "matcha")
	if err := a.handleChange(ctx, ownframe.Box{ID: "inv-search-input"}); err != nil {
		t.Fatal(err)
	}
	if a.searchQuery != "matcha" {
		t.Fatalf("searchQuery = %q, want %q", a.searchQuery, "matcha")
	}
	html := a.page.HTML()
	if !strings.Contains(html, "Ceremonial matcha") {
		t.Fatalf("page HTML after filtering for matcha misses Ceremonial matcha")
	}
	// The stock table always lists every SKU, so scope the absence check to
	// the palette region between the panel and the body viewport.
	if region := searchPaletteRegion(t, html); strings.Contains(region, "Nitrile gloves") {
		t.Fatalf("search palette still lists Nitrile gloves for query matcha")
	}
	if err := a.handleClick(ctx, ownframe.Box{Action: "search:pick:NW-CF-3302"}); err != nil {
		t.Fatal(err)
	}
	if a.searchOpen {
		t.Fatalf("searchOpen = true after search:pick:NW-CF-3302, want false")
	}
}

// searchPaletteRegion slices the rendered page to the command palette: from
// the panel element to the body viewport that follows the topbar. The id="
// prefixes keep the markers off the stylesheet's class selectors.
func searchPaletteRegion(t *testing.T, html string) string {
	t.Helper()
	start := strings.Index(html, `id="inv-search-panel"`)
	if start < 0 {
		t.Fatalf("page HTML misses id=%q", "inv-search-panel")
	}
	end := strings.Index(html[start:], `id="inv-body-viewport"`)
	if end < 0 {
		t.Fatalf("page HTML misses id=%q after the search panel", "inv-body-viewport")
	}
	return html[start : start+end]
}

func TestBinHoverData(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	handled, err := a.handleHover(ctx, ownframe.Box{}, ownframe.Box{ID: "bin-C-L3"})
	if err != nil {
		t.Fatal(err)
	}
	if handled {
		t.Fatalf("handleHover(bin-C-L3) handled = true, want false (direct patch paints)")
	}
	if a.hoverBin != "bin-C-L3" {
		t.Fatalf("hoverBin = %q, want %q", a.hoverBin, "bin-C-L3")
	}
	// Probe().Element returns the retained registry node, which carries no
	// text field, so read the painted text off the layout box instead.
	if el := a.page.Probe().Element("bin-detail"); el == nil {
		t.Fatalf("Probe().Element(%q) = nil, want the retained bin-detail node", "bin-detail")
	}
	text := ""
	for _, box := range a.page.Boxes() {
		if box.ID == "bin-detail" {
			text = box.Text
			break
		}
	}
	if !strings.Contains(text, "92") {
		t.Fatalf("bin-detail text = %q, want it to contain %q", text, "92")
	}
	if _, err := a.handleHover(ctx, ownframe.Box{ID: "bin-C-L3"}, ownframe.Box{}); err != nil {
		t.Fatal(err)
	}
	if a.hoverBin != "" {
		t.Fatalf("hoverBin = %q after hovering away, want empty", a.hoverBin)
	}
}

func TestEscapeClosesOverlays(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.handleClick(ctx, ownframe.Box{Action: "po:new"}); err != nil {
		t.Fatal(err)
	}
	if err := a.handleClick(ctx, ownframe.Box{Action: "search:open"}); err != nil {
		t.Fatal(err)
	}
	if !a.poOpen || !a.searchOpen {
		t.Fatalf("poOpen = %v, searchOpen = %v, want both true before escape", a.poOpen, a.searchOpen)
	}
	if err := a.handleKey(ctx, "escape"); err != nil {
		t.Fatal(err)
	}
	if a.poOpen || a.searchOpen {
		t.Fatalf("poOpen = %v, searchOpen = %v after escape, want both false", a.poOpen, a.searchOpen)
	}
}

func TestStatusFilterPills(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if html := a.page.HTML(); !strings.Contains(html, `data-action="status:Stockout"`) {
		t.Fatalf("stock pills miss the status:Stockout action")
	}
	if err := a.handleClick(ctx, ownframe.Box{Action: "status:Stockout"}); err != nil {
		t.Fatal(err)
	}
	if a.statusFilter != "Stockout" {
		t.Fatalf("statusFilter = %q, want %q", a.statusFilter, "Stockout")
	}
	html := a.page.HTML()
	for _, want := range []string{"Cold brew filter packs", "Ceremonial matcha"} {
		if !strings.Contains(html, want) {
			t.Fatalf("filtered table misses stockout SKU %q", want)
		}
	}
	for _, gone := range []string{"House espresso blend", "Vanilla syrup", "Nitrile gloves"} {
		if strings.Contains(html, gone) {
			t.Fatalf("filtered table still lists %q under Stockout", gone)
		}
	}
	if err := a.handleClick(ctx, ownframe.Box{Action: "status:Reorder"}); err != nil {
		t.Fatal(err)
	}
	if html := a.page.HTML(); !strings.Contains(html, "Oat milk barista") || strings.Contains(html, "Ceremonial matcha") {
		t.Fatalf("Reorder filter shows wrong rows")
	}
	if err := a.handleClick(ctx, ownframe.Box{Action: "status:"}); err != nil {
		t.Fatal(err)
	}
	if a.statusFilter != "" {
		t.Fatalf("statusFilter = %q after All, want empty", a.statusFilter)
	}
	if html := a.page.HTML(); !strings.Contains(html, "Vanilla syrup") {
		t.Fatalf("All filter hides SKUs, want the full table back")
	}
}

func TestSearchBackdrop(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if html := a.page.HTML(); strings.Contains(html, "inv-search-backdrop") {
		t.Fatalf("backdrop renders while search is closed")
	}
	if err := a.handleClick(ctx, ownframe.Box{Action: "search:open"}); err != nil {
		t.Fatal(err)
	}
	if html := a.page.HTML(); !strings.Contains(html, `data-action="search:close"`) {
		t.Fatalf("open search misses the click-away backdrop")
	}
	if err := a.handleClick(ctx, ownframe.Box{Action: "search:close"}); err != nil {
		t.Fatal(err)
	}
	if a.searchOpen {
		t.Fatalf("searchOpen = true after search:close, want false")
	}
	if html := a.page.HTML(); strings.Contains(html, "inv-search-backdrop") {
		t.Fatalf("backdrop still renders after search:close")
	}
}

func TestSearchPickedLine(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.handleClick(ctx, ownframe.Box{Action: "search:pick:NW-CF-3302"}); err != nil {
		t.Fatal(err)
	}
	if a.searchPicked != "NW-CF-3302" {
		t.Fatalf("searchPicked = %q, want %q", a.searchPicked, "NW-CF-3302")
	}
	if html := a.page.HTML(); !strings.Contains(html, "Picked: NW-CF-3302") {
		t.Fatalf("topbar misses the picked line after search:pick")
	}
}

func TestInboundSelectAndViewAll(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.handleClick(ctx, ownframe.Box{Action: "po:view:PO-10477"}); err != nil {
		t.Fatal(err)
	}
	if a.selectedPO != "PO-10477" {
		t.Fatalf("selectedPO = %q, want %q", a.selectedPO, "PO-10477")
	}
	if html := a.page.HTML(); !strings.Contains(html, "Oat Co. Foodservice") {
		t.Fatalf("inbound header misses the selected order detail")
	}
	if err := a.handleClick(ctx, ownframe.Box{Action: "po:viewall"}); err != nil {
		t.Fatal(err)
	}
	if a.selectedPO != "" {
		t.Fatalf("selectedPO = %q after po:viewall, want empty", a.selectedPO)
	}
	if html := a.page.HTML(); !strings.Contains(html, "4 open · $186,420") {
		t.Fatalf("inbound header misses the totals after po:viewall")
	}
}

func TestExportToast(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.handleClick(ctx, ownframe.Box{Action: "export:run"}); err != nil {
		t.Fatal(err)
	}
	if !a.toastOpen {
		t.Fatalf("toastOpen = false after export:run, want true")
	}
	html := a.page.HTML()
	for _, want := range []string{"Export queued", "toast:close"} {
		if !strings.Contains(html, want) {
			t.Fatalf("page HTML after export:run misses %q", want)
		}
	}
	if err := a.handleClick(ctx, ownframe.Box{Action: "toast:close"}); err != nil {
		t.Fatal(err)
	}
	if a.toastOpen {
		t.Fatalf("toastOpen = true after toast:close, want false")
	}
}

func TestSearchPanelOnScreen(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.handleClick(ctx, ownframe.Box{Action: "search:open"}); err != nil {
		t.Fatal(err)
	}
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	_, winH := a.page.Size()
	found := false
	for _, box := range a.page.Boxes() {
		if box.ID != "inv-search-panel" {
			continue
		}
		found = true
		if box.Y < 0 || box.Y+box.H > float64(winH)+2 {
			t.Fatalf("search panel at y=%v h=%v outside viewport h=%d", box.Y, box.H, winH)
		}
	}
	if !found {
		t.Fatalf("inv-search-panel box missing from layout")
	}
}

func TestAlertsPopoverKeepsTopbarStill(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	geom := func() map[string]ownframe.Box {
		out := map[string]ownframe.Box{}
		for _, b := range a.page.Boxes() {
			if b.ID == "notifications-popover" || b.ID == "user-menu" || b.ID == "inv-search-wrap" {
				out[b.ID] = b
			}
		}
		return out
	}
	// The search wrap has no ID; locate the search row via its action box.
	_ = geom
	before := map[string][4]float64{}
	for _, b := range a.page.Boxes() {
		if b.ID == "notifications-popover" || b.ID == "user-menu" {
			before[b.ID] = [4]float64{b.X, b.Y, b.W, b.H}
		}
		if b.Action == "search:open" && b.W > 200 {
			before["search-row"] = [4]float64{b.X, b.Y, b.W, b.H}
		}
	}
	a.notifOpen = true
	if err := a.refresh(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	for _, b := range a.page.Boxes() {
		var key string
		switch {
		case b.ID == "notifications-popover" || b.ID == "user-menu":
			key = b.ID
		case b.Action == "search:open" && b.W > 200:
			key = "search-row"
		default:
			continue
		}
		want := before[key]
		got := [4]float64{b.X, b.Y, b.W, b.H}
		if key == "notifications-popover" {
			if got[2] != 64 {
				t.Fatalf("popover root width = %v, want pinned 64", got[2])
			}
			want[2] = 64 // width is the pinned slot; position must not move
			got[2] = 64
		}
		if want != got {
			t.Fatalf("%s moved: was %v, now %v", key, want, got)
		}
	}
	if _, ok := before["search-row"]; !ok {
		t.Fatalf("search row action box missing from layout")
	}
}

func TestClickAwayClosesPopovers(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	a.notifOpen = true
	a.userMenuOpen = true
	if err := a.handleClick(ctx, ownframe.Box{ID: "some-plain-text"}); err != nil {
		t.Fatal(err)
	}
	if a.notifOpen || a.userMenuOpen {
		t.Fatalf("notifOpen=%v userMenuOpen=%v after outside click, want both false",
			a.notifOpen, a.userMenuOpen)
	}
	// Clicks inside the popover subtree keep it open.
	a.notifOpen = true
	if err := a.handleClick(ctx, ownframe.Box{ID: "notifications-popover-content"}); err != nil {
		t.Fatal(err)
	}
	if !a.notifOpen {
		t.Fatalf("notifOpen = false after inside click, want true")
	}
}

func TestSearchScrollWindow(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.handleClick(ctx, ownframe.Box{Action: "search:open"}); err != nil {
		t.Fatal(err)
	}
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	total := len(a.searchHits())
	if total <= inventorySearchVisible {
		t.Fatalf("capped hits = %d, want more than visible %d for a scroll test",
			total, inventorySearchVisible)
	}
	var px, py float64
	for _, b := range a.page.Boxes() {
		if b.ID == "inv-search-panel" {
			px, py = b.X+b.W/2, b.Y+b.H/2
		}
	}
	if px == 0 {
		t.Fatalf("inv-search-panel box missing")
	}
	for i := 0; i < total; i++ {
		consumed, changed := a.scrollSearchPanel(px, py, -1)
		if !consumed {
			t.Fatalf("wheel over panel not consumed at step %d", i)
		}
		_ = changed
	}
	if a.searchOffset != total-inventorySearchVisible {
		t.Fatalf("searchOffset = %d, want clamped max %d", a.searchOffset, total-inventorySearchVisible)
	}
	for i := 0; i < total; i++ {
		a.scrollSearchPanel(px, py, 1)
	}
	if a.searchOffset != 0 {
		t.Fatalf("searchOffset = %d after scrolling up, want 0", a.searchOffset)
	}
	// Off-panel wheel does not touch the palette offset.
	a.searchOffset = 2
	consumed, changed := a.scrollSearchPanel(10, 500, -1)
	if consumed || changed || a.searchOffset != 2 {
		t.Fatalf("off-panel wheel leaked: consumed=%v changed=%v offset=%d", consumed, changed, a.searchOffset)
	}
}

func TestSearchFiniteCapAndHint(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.handleClick(ctx, ownframe.Box{Action: "search:open"}); err != nil {
		t.Fatal(err)
	}
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if n := len(a.searchHits()); n > inventorySearchMaxSKUs+inventorySearchMaxPOs {
		t.Fatalf("searchHits = %d, want at most %d finite rows", n, inventorySearchMaxSKUs+inventorySearchMaxPOs)
	}
	if html := a.page.HTML(); !strings.Contains(html, "scroll for more") {
		t.Fatalf("open palette misses the scroll hint with %d hits", len(a.searchHits()))
	}
}
