package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// Search palette tuning. The engine has no CSS scrolling viewport, so the
// panel shows a finite window of the matches and the host scrolls it with
// the wheel (scroll.go). Raise the caps to show more per query.
const (
	// inventorySearchMaxSKUs caps the SKU section of one query.
	inventorySearchMaxSKUs = 5
	// inventorySearchMaxPOs caps the purchase-order section of one query.
	inventorySearchMaxPOs = 5
	// inventorySearchVisible is the number of flattened rows the panel
	// shows at once; the wheel moves the window over the capped matches.
	inventorySearchVisible = 6
)

// searchHit is one flattened palette row with its section, so the wheel
// can window across groups and searchGroups can regroup the slice.
type searchHit struct {
	group string
	item  chadcn.CommandItem
}

// searchBox is the topbar search trigger. The row emits the search:open
// action; while a.searchOpen the command panel below it filters SKUs and
// purchase orders through chadcn.Command.
func (a *app) searchBox() *ui.Element {
	wrap := ui.Div().Class("inv-search-wrap").Children(
		ui.Div().Class("inv-search").Attr("data-action", "search:open").Children(
			ui.Text("Search"),
			chadcn.Kbd("⌘K"),
		),
	)
	if !a.searchOpen {
		if a.searchPicked != "" {
			wrap.Children(ui.Text("Picked: " + a.searchPicked).Class("inv-search-picked"))
		}
		return wrap
	}
	return wrap.Children(
		ui.Div().ID("inv-search-panel").Class("inv-search-panel").Children(
			chadcn.Command(chadcn.CommandOptions{
				ID:     "inv-search",
				Query:  a.searchQuery,
				Groups: a.searchGroups(),
				OnSelect: func(ctx context.Context, _ string) error {
					a.searchOpen = false
					return a.refresh(ctx)
				},
			}),
			a.searchMore(),
		),
	)
}

// searchMore is the clipping hint below the listbox: how many of the
// capped matches the window hides, inviting a wheel scroll.
func (a *app) searchMore() *ui.Element {
	total := len(a.searchHits())
	if total <= inventorySearchVisible {
		return nil
	}
	return ui.Div().Class("inv-search-more").
		Text(fmt.Sprintf("Showing %d of %d — scroll for more", inventorySearchVisible, total))
}

// searchBackdrop is the click-away layer behind the open palette: any click
// outside the panel lands here and closes it, mirroring the mygo Popover
// and the loglens command overlay backdrop. It paints below the panel
// (z-index 50 vs 60) and above the page chrome.
func (a *app) searchBackdrop() *ui.Element {
	if !a.searchOpen {
		return nil
	}
	return ui.Div().Class("inv-search-backdrop").Attr("data-action", "search:close")
}

// handleChange carries live palette typing into a.searchQuery, then
// re-renders so the groups filter. Focus is preserved so typing continues.
// The command input ID derives from the Command ID: "inv-search-input".
func (a *app) handleChange(ctx context.Context, box ownframe.Box) error {
	if box.ID != "inv-search-input" {
		return nil
	}
	a.searchQuery = a.page.Form().FormValue(box.ID)
	a.searchOffset = 0
	return a.refresh(ctx)
}

// searchHits flattens every match for the current query, SKUs first,
// capped to the finite per-section budgets.
func (a *app) searchHits() []searchHit {
	q := strings.ToLower(strings.TrimSpace(a.searchQuery))

	hits := make([]searchHit, 0, inventorySearchMaxSKUs+inventorySearchMaxPOs)
	for _, s := range sampleSKUs {
		if len(hits) >= inventorySearchMaxSKUs {
			break
		}
		if q != "" && !strings.Contains(strings.ToLower(s.Name+" "+s.Code+" "+s.Category), q) {
			continue
		}
		hits = append(hits, searchHit{"SKUs", chadcn.CommandItem{
			Value:    "search:pick:" + s.Code,
			Label:    s.Name,
			Shortcut: s.Status,
		}})
	}
	poCount := 0
	for _, po := range samplePOs {
		if poCount >= inventorySearchMaxPOs {
			break
		}
		if q != "" && !strings.Contains(strings.ToLower(po.Vendor+" "+po.PONo), q) {
			continue
		}
		poCount++
		hits = append(hits, searchHit{"Purchase orders", chadcn.CommandItem{
			Value:    "search:pick:" + po.PONo,
			Label:    po.PONo + " · " + po.Vendor,
			Shortcut: po.Badge,
		}})
	}
	return hits
}

// searchWindow clamps the visible offset into the capped matches.
func (a *app) searchWindow() (start, total int) {
	total = len(a.searchHits())
	maxStart := total - inventorySearchVisible
	if maxStart < 0 {
		maxStart = 0
	}
	start = min(max(a.searchOffset, 0), maxStart)
	return start, total
}

// searchGroups builds the two palette sections for the visible window of
// the capped matches. An empty query shows the first SKUs and all purchase
// orders. Matching is a case-insensitive substring over the SKU name, code,
// and category, or the order vendor and number.
func (a *app) searchGroups() []chadcn.CommandGroup {
	start, _ := a.searchWindow()
	window := a.searchHits()[start:min(start+inventorySearchVisible, len(a.searchHits()))]

	groups := make([]chadcn.CommandGroup, 0, 2)
	for _, name := range []string{"SKUs", "Purchase orders"} {
		items := make([]chadcn.CommandItem, 0)
		for _, hit := range window {
			if hit.group == name {
				items = append(items, hit.item)
			}
		}
		if len(items) > 0 {
			groups = append(groups, chadcn.CommandGroup{Label: name, Items: items})
		}
	}
	return groups
}
