package main

import (
	"context"
	"strconv"
	"strings"
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
	contextMenuIDPrefix     = "code-view:"
	contextMenuCloseID      = contextMenuIDPrefix + "tab-close:"
	contextMenuCloseOtherID = contextMenuIDPrefix + "tab-close-others:"
	contextMenuCloseAllID   = contextMenuIDPrefix + "tab-close-all:"
)

// ContextMenuFor returns the tab actions for the tab or close button under the
// pointer. Every other element keeps the page's built-in menu.
func (a *App) ContextMenuFor(id string) []contextMenuRow {
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
	}
}

// MenuAction closes one tab, the other tabs, or every tab in the group that
// supplied the row.
func (a *App) MenuAction(ctx context.Context, id string) error {
	switch {
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
