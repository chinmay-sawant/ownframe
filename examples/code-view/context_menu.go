package main

import (
	"context"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

// contextMenuRow mirrors the window's context-menu row. The host row type
// lives in an internal package, so the page bridge converts this shape at the
// boundary.
type contextMenuRow struct {
	ID      string
	Label   string
	Enabled bool
}

const (
	menuChromeWidth = 196
	menuChromeRow   = 26
	menuChromePad   = 8

	contextMenuIDPrefix     = "code-view:"
	contextMenuCloseID      = contextMenuIDPrefix + "tab-close:"
	contextMenuCloseOtherID = contextMenuIDPrefix + "tab-close-others:"
	contextMenuCloseAllID   = contextMenuIDPrefix + "tab-close-all:"
	contextMenuCloseLeftID  = contextMenuIDPrefix + "tab-close-left:"
	contextMenuCloseRightID = contextMenuIDPrefix + "tab-close-right:"
)

// presentMenu paints the right-click rows in the page instead of the shell menu.
func (a *App) presentMenu(x, y int, items []ownframe.MenuItem) bool {
	if a == nil || a.page == nil || len(items) == 0 {
		return false
	}

	rows := make([]menuViewRow, 0, len(items))
	for i, item := range items {
		rows = append(rows, menuViewRow{
			Index:   i,
			ID:      item.ID,
			Label:   item.Label,
			Enabled: item.Enabled,
		})
	}

	w, h := a.page.Size()
	menuH := a.zx(menuChromePad) + len(rows)*a.zx(menuChromeRow)
	if x+a.zx(menuChromeWidth) > w {
		x = w - a.zx(menuChromeWidth)
	}

	if y+menuH > h {
		y = h - menuH
	}

	if x < 0 {
		x = 0
	}

	if y < 0 {
		y = 0
	}

	a.menuOpen = true
	a.menuX = x
	a.menuY = y
	a.menuItems = rows
	a.refresh()

	if err := a.page.Redraw(context.Background()); err != nil {
		return false
	}

	return true
}

// closeMenu hides the page menu. The caller redraws.
func (a *App) closeMenu() {
	a.menuOpen = false
	a.menuItems = nil
}

// dismissOpenMenu closes the menu when the pointer is outside its rows.
// A row still runs. The caller redraws.
func (a *App) dismissOpenMenu(id, action string) bool {
	if a == nil || !a.menuOpen {
		return false
	}

	if strings.HasPrefix(id, "ctx-") || strings.HasPrefix(action, "menu-item:") {
		return false
	}

	a.closeMenu()
	a.refresh()

	return true
}

// runMenuItem closes the menu and runs one row. Built-in edit rows use the
// same buffer methods the shell menu used.
func (a *App) runMenuItem(ctx context.Context, id string) error {
	a.closeMenu()

	switch id {
	case "cut":
		text, ok, err := a.onCut(ctx)
		if err != nil || !ok {
			return err
		}

		ownframe.Write(text)
		a.refresh()

		return a.page.Redraw(ctx)
	case "copy":
		text, ok, err := a.onCopy(ctx)
		if err != nil || !ok {
			return err
		}

		ownframe.Write(text)
		a.refresh()

		return a.page.Redraw(ctx)
	case "paste":
		text, err := ownframe.Read()
		if err != nil {
			a.refresh()

			return a.page.Redraw(ctx)
		}

		if err := a.onPaste(ctx, text); err != nil {
			return err
		}

		return a.page.Redraw(ctx)
	case "select-all":
		if err := a.onSelectAll(ctx); err != nil {
			return err
		}

		return a.page.Redraw(ctx)
	case "undo":
		if err := a.onUndo(ctx); err != nil {
			return err
		}

		return a.page.Redraw(ctx)
	case "redo":
		if err := a.onRedo(ctx); err != nil {
			return err
		}

		return a.page.Redraw(ctx)
	default:
		return a.MenuAction(ctx, id)
	}
}

// ContextMenuFor returns actions for the editor or the tab under the pointer.
func (a *App) ContextMenuFor(id string) []contextMenuRow {
	if id == "code-body" && a.editing && a.edit != nil {
		selected := a.editCaret != a.editAnchor
		return []contextMenuRow{
			{ID: "cut", Label: "Cut", Enabled: selected},
			{ID: "copy", Label: "Copy", Enabled: selected},
			{ID: "paste", Label: "Paste", Enabled: true},
			{ID: "select-all", Label: "Select All", Enabled: true},
			{ID: "undo", Label: "Undo", Enabled: len(a.editUndo) > 0},
			{ID: "redo", Label: "Redo", Enabled: len(a.editRedo) > 0},
		}
	}
	group, index, ok := contextMenuTabID(id)
	if !ok {
		return nil
	}

	tabs := a.tabs
	if group == 1 {
		tabs = a.tabs2
	}

	if index < 0 || index >= len(tabs) {
		return nil
	}

	return []contextMenuRow{
		{ID: contextMenuCloseID + contextMenuTabTarget(group, index), Label: "Close", Enabled: true},
		{ID: contextMenuCloseOtherID + contextMenuTabTarget(group, index), Label: "Close Others", Enabled: len(tabs) > 1},
		{ID: contextMenuCloseAllID + strconv.Itoa(group), Label: "Close All", Enabled: len(tabs) > 0},
		{ID: contextMenuCloseLeftID + contextMenuTabTarget(group, index), Label: "Close All to the Left", Enabled: index > 0},
		{
			ID:      contextMenuCloseRightID + contextMenuTabTarget(group, index),
			Label:   "Close All to the Right",
			Enabled: index < len(tabs)-1,
		},
	}
}

// MenuAction handles tab-closing actions in the group that supplied the row.
func (a *App) MenuAction(ctx context.Context, id string) error {
	switch {
	case strings.HasPrefix(id, contextMenuCloseLeftID):
		if group, index, ok := contextMenuTarget(strings.TrimPrefix(id, contextMenuCloseLeftID)); ok {
			return a.contextMenuCloseSide(ctx, group, index, true)
		}
	case strings.HasPrefix(id, contextMenuCloseRightID):
		if group, index, ok := contextMenuTarget(strings.TrimPrefix(id, contextMenuCloseRightID)); ok {
			return a.contextMenuCloseSide(ctx, group, index, false)
		}
	case strings.HasPrefix(id, contextMenuCloseOtherID):
		if group, index, ok := contextMenuTarget(strings.TrimPrefix(id, contextMenuCloseOtherID)); ok {
			return a.contextMenuCloseOtherTabs(ctx, group, index)
		}
	case strings.HasPrefix(id, contextMenuCloseAllID):
		if group, ok := contextMenuGroup(strings.TrimPrefix(id, contextMenuCloseAllID)); ok {
			return a.contextMenuCloseAllTabs(ctx, group)
		}
	case strings.HasPrefix(id, contextMenuCloseID):
		if group, index, ok := contextMenuTarget(strings.TrimPrefix(id, contextMenuCloseID)); ok {
			return a.contextMenuCloseTab(ctx, group, index)
		}
	}

	return nil
}

func (a *App) contextMenuCloseSide(ctx context.Context, group, index int, left bool) error {
	tabs, active := &a.tabs, &a.active
	if group == 1 {
		tabs, active = &a.tabs2, &a.active2
	}
	if index < 0 || index >= len(*tabs) {
		return nil
	}
	a.finishEdit()
	if left {
		*tabs = append([]string(nil), (*tabs)[index:]...)
		*active = max(0, *active-index)
	} else {
		*tabs = append([]string(nil), (*tabs)[:index+1]...)
		*active = min(*active, index)
	}
	return a.contextMenuFinish(ctx)
}

// contextMenuTabID reads tab-N, tab-close-N, tab2-N, or tab2-close-N. Other
// synthetic ids, such as tab-working or tab-details, do not match.
func contextMenuTabID(id string) (int, int, bool) {
	for _, candidate := range []struct {
		prefix string
		group  int
	}{
		{"tab2-close-", 1},
		{"tab-close-", 0},
		{"tab2-", 1},
		{"tab-", 0},
	} {
		if !strings.HasPrefix(id, candidate.prefix) {
			continue
		}

		index, ok := contextMenuIndex(strings.TrimPrefix(id, candidate.prefix))
		if ok {
			return candidate.group, index, true
		}
	}

	return 0, 0, false
}

// contextMenuTarget reads group:index from a custom close row.
func contextMenuTarget(rest string) (int, int, bool) {
	groupText, indexText, ok := strings.Cut(rest, ":")
	if !ok {
		return 0, 0, false
	}

	group, ok := contextMenuGroup(groupText)
	if !ok {
		return 0, 0, false
	}

	index, ok := contextMenuIndex(indexText)
	if !ok {
		return 0, 0, false
	}

	return group, index, true
}

// contextMenuGroup accepts only the two editor groups.
func contextMenuGroup(text string) (int, bool) {
	group, err := strconv.Atoi(text)
	if err != nil || (group != 0 && group != 1) {
		return 0, false
	}

	return group, true
}

// contextMenuIndex accepts a non-negative decimal index.
func contextMenuIndex(text string) (int, bool) {
	if text == "" {
		return 0, false
	}

	for _, r := range text {
		if r < '0' || r > '9' {
			return 0, false
		}
	}

	index, err := strconv.Atoi(text)
	if err != nil {
		return 0, false
	}

	return index, true
}

// contextMenuTabTarget joins the group and index into the custom row suffix.
func contextMenuTabTarget(group, index int) string {
	return strconv.Itoa(group) + ":" + strconv.Itoa(index)
}

// contextMenuCloseTab drops one tab and keeps the group's active index on a
// real tab.
func (a *App) contextMenuCloseTab(ctx context.Context, group, index int) error {
	a.finishEdit()

	if group == 1 {
		if index < 0 || index >= len(a.tabs2) {
			return nil
		}

		a.tabs2 = append(append([]string(nil), a.tabs2[:index]...), a.tabs2[index+1:]...)
		if index < a.active2 {
			a.active2--
		}

		a.active2 = contextMenuClampIndex(a.active2, len(a.tabs2))
		if len(a.tabs2) == 0 {
			a.split = false
			a.focusGroup = 0
		}
	} else {
		if index < 0 || index >= len(a.tabs) {
			return nil
		}

		a.tabs = append(append([]string(nil), a.tabs[:index]...), a.tabs[index+1:]...)
		if index < a.active {
			a.active--
		}

		a.active = contextMenuClampIndex(a.active, len(a.tabs))
	}

	return a.contextMenuFinish(ctx)
}

// contextMenuCloseOtherTabs keeps one tab in the group and drops the rest.
func (a *App) contextMenuCloseOtherTabs(ctx context.Context, group, index int) error {
	a.finishEdit()

	if group == 1 {
		if index < 0 || index >= len(a.tabs2) {
			return nil
		}

		a.tabs2 = []string{a.tabs2[index]}
		a.active2 = 0
	} else {
		if index < 0 || index >= len(a.tabs) {
			return nil
		}

		a.tabs = []string{a.tabs[index]}
		a.active = 0
	}

	return a.contextMenuFinish(ctx)
}

// contextMenuCloseAllTabs clears one group. An empty second group ends the
// split.
func (a *App) contextMenuCloseAllTabs(ctx context.Context, group int) error {
	a.finishEdit()

	if group == 1 {
		a.tabs2 = nil
		a.active2 = 0
		a.split = false
		a.focusGroup = 0
	} else {
		a.tabs = nil
		a.active = 0
	}

	return a.contextMenuFinish(ctx)
}

// contextMenuFinish settles and paints the page after a tab action.
func (a *App) contextMenuFinish(ctx context.Context) error {
	a.preview = false

	if a.page == nil {
		return nil
	}

	a.resetScroll()
	a.ensureTabVisible(0)
	a.ensureTabVisible(1)
	a.refresh()

	return a.page.Redraw(ctx)
}

// contextMenuClampIndex keeps an active index on a real tab, with 0 for an
// empty group.
func contextMenuClampIndex(index, count int) int {
	if index < 0 || count == 0 {
		return 0
	}

	if index >= count {
		return count - 1
	}

	return index
}
