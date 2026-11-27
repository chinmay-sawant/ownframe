package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/chinmay-sawant/ownframe/ui"
)

// tbItem is one toolbar action that can collapse into the overflow menu.
type tbItem struct {
	id, label, class string
	disabled         bool
	check, on        bool // a toggle row that shows a check mark when on
	fn               func(context.Context) error
}

// Rough toolbar geometry, in CSS pixels, used to decide what fits.
const (
	tbButtonChrome  = 36 // icon, gaps and padding around a button label
	tbCharWidth     = 7
	tbGap           = 8
	tbRunWidth      = 70
	tbIconBtnWidth  = 40
	tbMoreWidth     = 52
	edToolbarFixed  = 350
	overflowMenuW   = 190
	titleBarH       = 36  // the menu scrim starts below the title bar
	tblToolbarFixed = 520 // Rows, page size, pager and spacer
)

// mainWidth is the width of the work area right of the rail and sidebar.
func (a *App) mainWidth() int {
	w := max(a.width, MinWidth) - railWidth
	if a.sidebarOpen {
		w -= a.sidebarW
	}
	return w
}

func itemWidth(label string) int { return tbButtonChrome + tbCharWidth*len(label) + tbGap }

// overflowGroup returns the buttons of items that fit in budget pixels, then a
// "..." button that opens a menu holding the rest plus the always-menu extras.
// Items are in priority order, so the last ones collapse first.
func (a *App) overflowGroup(prefix string, budget int, items, extras []tbItem) []*ui.Element {
	if len(extras) > 0 || len(items) > 0 {
		budget -= tbMoreWidth
	}
	var out []*ui.Element
	var menu []tbItem
	used := 0
	for _, it := range items {
		used += itemWidth(it.label)
		if used > budget || len(menu) > 0 {
			menu = append(menu, it)
			continue
		}
		out = append(out, a.tbButton(it.id, it.label, it.class, it.disabled, it.fn))
	}
	menu = append(menu, extras...)
	if len(menu) == 0 {
		return out
	}
	more := ui.Div().ID("ov-more-"+prefix).Class("qr-btn qr-ov-btn").Attr("role", "button").
		Attr("aria-label", "More actions").Attr("aria-haspopup", "menu").
		Attr("aria-expanded", fmt.Sprint(a.menuOpen == "ov-"+prefix)).Text("...").
		OnClick(func(ctx context.Context) error {
			a.ac.open = false
			if a.menuOpen == "ov-"+prefix {
				a.menuOpen = ""
			} else {
				a.menuOpen = "ov-" + prefix
			}
			return a.redraw(ctx)
		})
	if a.menuOpen == "ov-"+prefix {
		a.ov = overflowOpen{prefix: prefix, items: menu}
	}
	return append(out, ui.Div().Class("qr-ov-anchor").Children(more))
}

// overflowOpen is the menu the current frame's toolbar asked to show.
type overflowOpen struct {
	prefix string
	items  []tbItem
}

// overflowMenu is the "..." dropdown. Like the connection menu it lives in a
// full-window scrim so presses outside dismiss it, anchored under its button.
func (a *App) overflowMenu() *ui.Element {
	if !strings.HasPrefix(a.menuOpen, "ov-") || a.ov.prefix == "" || "ov-"+a.ov.prefix != a.menuOpen {
		return nil
	}
	left, top := 0.0, 0.0
	for _, b := range a.page.Boxes() {
		if b.ID == "ov-more-"+a.ov.prefix {
			left, top = b.X+b.W-overflowMenuW, b.Y+b.H+edBodyPad-titleBarH
		}
	}
	pop := ui.Div().ID("ov-menu-"+a.ov.prefix).Class("qr-menu qr-ov-menu").Attr("role", ariaRole("menu")).Attr("aria-label", "More actions").
		Position(ui.PositionAbsolute).Left(ui.Px(max(0, left))).Top(ui.Px(max(0, top)))
	for _, it := range a.ov.items {
		it := it
		row := ui.Div().ID("ov-item-"+it.id).Class("qr-list-row").Attr("role", ariaRole("menuitem"))
		if it.check {
			mark := ui.Span().Class("qr-ov-check")
			if it.on {
				mark = mark.Children(a.icon("check", toneBrand, iconSmall))
			}
			row.Children(mark)
		}
		row.Children(ui.Span().Text(it.label))
		if it.disabled {
			row.Class("qr-list-row qr-btn-off").Attr("aria-disabled", "true")
		} else {
			row.OnClick(func(ctx context.Context) error {
				a.menuOpen, a.ac.open = "", false
				if err := it.fn(ctx); err != nil {
					return err
				}
				return a.redraw(ctx)
			})
		}
		pop.Children(row)
	}
	return ui.Div().ID("ov-scrim").Class("qr-menu-scrim").Children(pop).
		OnClick(func(ctx context.Context) error { a.menuOpen = ""; return a.redraw(ctx) })
}

// toggleWrap flips editor word wrap and remembers it.
func (a *App) toggleWrap(ctx context.Context) error {
	a.wrap = !a.wrap
	a.savePref("editor.wrap", map[bool]string{true: "1", false: "0"}[a.wrap])
	a.syncWrap(a.activeTabPtr())
	if t := a.activeTabPtr(); t != nil && t.ed != nil {
		t.ed.ensureVisible(a.editorRows())
	}
	return a.redraw(ctx)
}
