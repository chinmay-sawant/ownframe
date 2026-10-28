package main

import "strings"

// editorGroup is the tab set one editor pane holds. Tabs is the open order,
// Active indexes it, and Path mirrors the active tab so a template reads one
// field. The zero value is an empty group.
type editorGroup struct {
	Tabs   []string
	Active int
	Path   string
}

// Open shows path and focuses it, appending when it is not open yet. A path
// already in the group is only focused. When the group is full the oldest tab
// is dropped and returned as evicted. A max below 1 means maxTabs.
func (g *editorGroup) Open(path string, max int) (int, string) {
	if path == "" {
		return -1, ""
	}

	if index := g.index(path); index >= 0 {
		g.Active = index
		g.Path = path

		return index, ""
	}

	limit := max
	if limit < 1 {
		limit = maxTabs
	}

	g.Tabs = append(g.Tabs, path)
	evicted := ""

	for len(g.Tabs) > limit {
		evicted = g.Tabs[0]
		g.Tabs = append(g.Tabs[:0], g.Tabs[1:]...)
	}

	g.Active = len(g.Tabs) - 1
	g.Path = path

	return g.Active, evicted
}

// Select focuses one tab. An out-of-range index leaves the group unchanged.
func (g *editorGroup) Select(index int) bool {
	if index < 0 || index >= len(g.Tabs) {
		return false
	}

	g.Active = index
	g.Path = g.Tabs[index]

	return true
}

// SelectPath focuses the tab holding path.
func (g *editorGroup) SelectPath(path string) bool {
	index := g.index(path)
	if index < 0 {
		return false
	}

	return g.Select(index)
}

// Close drops one tab and keeps Active on a real tab. It reports the closed
// path; an out-of-range index changes nothing and reports false.
func (g *editorGroup) Close(index int) (string, bool) {
	if index < 0 || index >= len(g.Tabs) {
		return "", false
	}

	closed := g.Tabs[index]
	g.Tabs = append(g.Tabs[:index], g.Tabs[index+1:]...)

	if index < g.Active {
		g.Active--
	}

	if g.Active >= len(g.Tabs) {
		g.Active = len(g.Tabs) - 1
	}

	if g.Active < 0 {
		g.Active = 0
	}

	g.Path = g.ActivePath()

	return closed, true
}

// ClosePath closes the tab holding path.
func (g *editorGroup) ClosePath(path string) bool {
	index := g.index(path)
	if index < 0 {
		return false
	}

	_, ok := g.Close(index)

	return ok
}

// ActivePath returns the active tab, or "" for an empty group. Path holds the
// same value after every method call.
func (g *editorGroup) ActivePath() string {
	if g.Active < 0 || g.Active >= len(g.Tabs) {
		return ""
	}

	return g.Tabs[g.Active]
}

// Len reports how many tabs the group holds.
func (g *editorGroup) Len() int { return len(g.Tabs) }

// Paths copies the open tabs in strip order, so a caller can pass the result
// to another group without aliasing this one.
func (g *editorGroup) Paths() []string {
	if len(g.Tabs) == 0 {
		return nil
	}

	paths := make([]string, len(g.Tabs))
	copy(paths, g.Tabs)

	return paths
}

// Reset replaces the group with paths, trimming each entry, dropping blanks,
// and keeping the first of each duplicate. The list is capped at maxTabs and
// active is clamped to a real tab, so a restored session cannot leave Active
// past the end or Path stale.
func (g *editorGroup) Reset(paths []string, active int) {
	tabs := make([]string, 0, maxTabs)

	seen := make(map[string]bool, len(paths))

	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" || seen[path] {
			continue
		}

		seen[path] = true

		tabs = append(tabs, path)

		if len(tabs) == maxTabs {
			break
		}
	}

	g.Tabs = tabs

	g.Active = active
	if g.Active >= len(g.Tabs) {
		g.Active = len(g.Tabs) - 1
	}

	if g.Active < 0 {
		g.Active = 0
	}

	g.Path = g.ActivePath()
}

// index returns the tab holding path, or -1.
func (g *editorGroup) index(path string) int {
	for i, tab := range g.Tabs {
		if tab == path {
			return i
		}
	}

	return -1
}
