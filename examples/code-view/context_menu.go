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
	ID    string
	Label string
	// Hint is a right-aligned shortcut label such as Ctrl+C.
	Hint    string
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

	contextMenuDiscardFileID = contextMenuIDPrefix + "discard-file:"
	contextMenuStageFileID   = contextMenuIDPrefix + "stage-file:"
	contextMenuUnstageFileID = contextMenuIDPrefix + "unstage-file:"
	contextMenuDiscardAllID  = contextMenuIDPrefix + "discard-all"
	contextMenuStageAllID    = contextMenuIDPrefix + "stage-all"
	contextMenuUnstageAllID  = contextMenuIDPrefix + "unstage-all"
	contextMenuStageSelID    = contextMenuIDPrefix + "stage-selected"
	contextMenuUnstageSelID  = contextMenuIDPrefix + "unstage-selected"
	contextMenuDiscardSelID  = contextMenuIDPrefix + "discard-selected"
	contextMenuCopyMsgID     = contextMenuIDPrefix + "copy-message:"
	contextMenuCopyIDID      = contextMenuIDPrefix + "copy-id:"
	contextMenuCopyPathID    = contextMenuIDPrefix + "copy-path:"
	contextMenuCopyAbsID     = contextMenuIDPrefix + "copy-abs:"
	contextMenuCutID         = contextMenuIDPrefix + "cut:"
	contextMenuCopyID        = contextMenuIDPrefix + "copy:"
	contextMenuPasteID       = contextMenuIDPrefix + "paste:"
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
			Hint:    item.Hint,
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

// ContextMenuFor returns actions for the editor, the tab, the working tree,
// or the history rail under the pointer.
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
	if rows, ok := a.contextMenuForExplorer(id); ok {
		return rows
	}
	if rows, ok := a.scmContextMenu(id); ok {
		return rows
	}
	if rows, ok := a.historyContextMenu(id); ok {
		return rows
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

// scmContextMenu returns working-tree rows. Per-file ids offer stage or
// unstage based on the section, plus discard; multi-select adds bulk rows.
func (a *App) scmContextMenu(id string) ([]contextMenuRow, bool) {
	if a.mode != "changes" {
		return nil, false
	}

	for _, prefix := range []string{"discard-staged-", "wrap-staged-", "change-", "staged-", "stage-", "discard-", "unstage-", "wrap-"} {
		if strings.HasPrefix(id, prefix) {
			path := strings.TrimPrefix(id, prefix)
			if prefix == "discard-staged-" {
				// ID is discard-staged-<path>; path is already clean.
			}
			if path == "" {
				return nil, false
			}

			if _, ok := a.scmChangeFile(path); !ok {
				return nil, false
			}

			staged := a.isStaged(path)
			selCount := a.selectedCount()
			selActive := selCount > 1 && a.selected[path]

			var rows []contextMenuRow
			if staged {
				rows = append(rows, contextMenuRow{ID: contextMenuUnstageFileID + path, Label: "Unstage Changes", Enabled: true})
			} else {
				rows = append(rows, contextMenuRow{ID: contextMenuStageFileID + path, Label: "Stage Changes", Enabled: true})
			}

			rows = append(rows, contextMenuRow{ID: contextMenuDiscardFileID + path, Label: "Discard Changes", Enabled: true})

			if selActive {
				rows = append(rows,
					contextMenuRow{ID: contextMenuStageSelID, Label: "Stage Selected", Enabled: true},
					contextMenuRow{ID: contextMenuUnstageSelID, Label: "Unstage Selected", Enabled: true},
					contextMenuRow{ID: contextMenuDiscardSelID, Label: "Discard Selected", Enabled: true},
				)
			}

			hasChanges := len(a.changes.Files) > 0
			rows = append(rows,
				contextMenuRow{ID: contextMenuStageAllID, Label: "Stage All", Enabled: hasChanges},
			)

			if len(a.changes.Staged) > 0 {
				rows = append(rows, contextMenuRow{ID: contextMenuUnstageAllID, Label: "Unstage All", Enabled: true})
			}

			rows = append(rows, contextMenuRow{ID: contextMenuDiscardAllID, Label: "Discard All", Enabled: hasChanges})

			return rows, true
		}
	}

	switch id {
	case "changes-list", "changes-head", "changes-staged-head", "changes-unstaged-head", "history-working", "changes-stage-all", "changes-discard-all", "changes-unstage-all", "sidebar":
		hasChanges := len(a.changes.Files) > 0
		if !hasChanges {
			return nil, false
		}

		var rows []contextMenuRow
		if a.selectedCount() > 1 {
			rows = append(rows,
				contextMenuRow{ID: contextMenuStageSelID, Label: "Stage Selected", Enabled: true},
				contextMenuRow{ID: contextMenuUnstageSelID, Label: "Unstage Selected", Enabled: true},
				contextMenuRow{ID: contextMenuDiscardSelID, Label: "Discard Selected", Enabled: true},
			)
		}

		rows = append(rows, contextMenuRow{ID: contextMenuStageAllID, Label: "Stage All", Enabled: true})

		if len(a.changes.Staged) > 0 {
			rows = append(rows, contextMenuRow{ID: contextMenuUnstageAllID, Label: "Unstage All", Enabled: true})
		}

		rows = append(rows, contextMenuRow{ID: contextMenuDiscardAllID, Label: "Discard All", Enabled: true})

		return rows, true
	}

	return nil, false
}

// historyContextMenu returns commit rows. Each commit offers its message and
// ID on the clipboard. It never returns the working-tree rows.
func (a *App) historyContextMenu(id string) ([]contextMenuRow, bool) {
	if a.mode != "changes" {
		return nil, false
	}

	if !strings.HasPrefix(id, "commit-") ||
		strings.HasPrefix(id, "commit-file-") ||
		strings.HasPrefix(id, "commit-open-") {
		return nil, false
	}

	short := strings.TrimPrefix(id, "commit-")
	if short == "" {
		return nil, false
	}

	hash := ""
	for i := range a.commits.Commits {
		if a.commits.Commits[i].Short == short {
			hash = a.commits.Commits[i].Hash

			break
		}
	}

	if hash == "" {
		return nil, false
	}

	return []contextMenuRow{
		{ID: contextMenuCopyMsgID + hash, Label: "Copy Message", Enabled: true},
		{ID: contextMenuCopyIDID + hash, Label: "Copy Commit ID", Enabled: true},
	}, true
}

// MenuAction handles tab-closing, working-tree, and history actions.
func (a *App) MenuAction(ctx context.Context, id string) error {
	switch {
	case strings.HasPrefix(id, contextMenuDiscardFileID):
		path := strings.TrimPrefix(id, contextMenuDiscardFileID)
		a.openDiscardConfirm("discard-file", []string{path})

		if a.page == nil {
			return nil
		}

		return a.page.Redraw(ctx)
	case strings.HasPrefix(id, contextMenuStageFileID):
		path := strings.TrimPrefix(id, contextMenuStageFileID)
		if err := a.stageFile(ctx, path); err != nil {
			return err
		}

		if a.page == nil {
			return nil
		}

		return a.page.Redraw(ctx)
	case strings.HasPrefix(id, contextMenuUnstageFileID):
		path := strings.TrimPrefix(id, contextMenuUnstageFileID)
		if err := a.unstageFile(ctx, path); err != nil {
			return err
		}

		if a.page == nil {
			return nil
		}

		return a.page.Redraw(ctx)
	case id == contextMenuDiscardAllID:
		a.openDiscardConfirm("discard-all", a.workingAllPaths())

		if a.page == nil {
			return nil
		}

		return a.page.Redraw(ctx)
	case id == contextMenuStageAllID:
		if err := a.stageAll(ctx); err != nil {
			return err
		}

		if a.page == nil {
			return nil
		}

		return a.page.Redraw(ctx)
	case id == contextMenuUnstageAllID:
		if err := a.unstageAll(ctx); err != nil {
			return err
		}

		if a.page == nil {
			return nil
		}

		return a.page.Redraw(ctx)
	case id == contextMenuStageSelID:
		if err := a.stageSelected(ctx); err != nil {
			return err
		}

		if a.page == nil {
			return nil
		}

		return a.page.Redraw(ctx)
	case id == contextMenuUnstageSelID:
		if err := a.unstageSelected(ctx); err != nil {
			return err
		}

		if a.page == nil {
			return nil
		}

		return a.page.Redraw(ctx)
	case id == contextMenuDiscardSelID:
		a.openDiscardConfirm("discard-selected", a.selectedList())

		if a.page == nil {
			return nil
		}

		return a.page.Redraw(ctx)
	case strings.HasPrefix(id, contextMenuCopyMsgID):
		a.copyCommitMessage(strings.TrimPrefix(id, contextMenuCopyMsgID))
		if a.page == nil {
			return nil
		}

		return a.page.Redraw(ctx)
	case strings.HasPrefix(id, contextMenuCopyIDID):
		a.copyCommitID(strings.TrimPrefix(id, contextMenuCopyIDID))
		if a.page == nil {
			return nil
		}

		return a.page.Redraw(ctx)
	case strings.HasPrefix(id, contextMenuCopyPathID):
		a.copyTreePath(strings.TrimPrefix(id, contextMenuCopyPathID))

		return a.redrawTreeMenu(ctx)
	case strings.HasPrefix(id, contextMenuCopyAbsID):
		a.copyTreeAbsPath(strings.TrimPrefix(id, contextMenuCopyAbsID))

		return a.redrawTreeMenu(ctx)
	case strings.HasPrefix(id, contextMenuCutID):
		a.stageTreeCopy(a.copyTargets(strings.TrimPrefix(id, contextMenuCutID)), true)

		return a.redrawTreeMenu(ctx)
	case strings.HasPrefix(id, contextMenuCopyID):
		a.stageTreeCopy(a.copyTargets(strings.TrimPrefix(id, contextMenuCopyID)), false)

		return a.redrawTreeMenu(ctx)
	case strings.HasPrefix(id, contextMenuPasteID):
		return a.pasteTreeCopy(ctx, strings.TrimPrefix(id, contextMenuPasteID))
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
