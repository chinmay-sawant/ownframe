// Probe-patch tests for the menubar, navigation menu, command palette, and
// search panel (TestProbePatchBars*).
//
// What patches and what stays FULL by design (see ui/diff.go: same tag,
// same child count, same order at every node; every changed node needs a
// stable id; a removed attribute forces a rebuild):
//
//   - Command Active-row move on an identical list patches: rows already
//     carry stable ID+"-item-<n>" ids, so only aria-selected flips. No code
//     change was needed.
//   - Command query keystroke with the host list untouched patches: only the
//     input value attribute changes. The probe's command-query case stays
//     FULL because the host narrows Groups (2 rows -> 1), which is a
//     host-driven structural list change, not a component defect.
//   - Search selection (Value) change on an identical matched list patches:
//     searchOption now keeps the check indicator mounted (empty when
//     unselected) with a stable ID+"-option-<n>-check" id, so selection is
//     an aria-selected edit plus a text edit. Previously the check span
//     mounted/unmounted and changed the row's child count.
//   - Search empty-state rewrite (same empty matched set, new query) patches:
//     the empty div now carries ID+"-empty", so "No results" vs
//     `No results for "q"` is a text edit plus the input value edit.
//     A query that changes the matched set stays FULL: panel rows mount or
//     unmount and row ids (source-index based) shift.
//   - Search field label (placeholder/aria-label) edit on a closed panel
//     patches: attribute edits on the id-bearing input. No change needed.
//   - Menubar open-switch stays FULL even when both menus have the same row
//     shape: the open pane mounts inside the newly opened menu's slot and
//     unmounts from the old one (both slots change child count), the old
//     trigger drops aria-controls (attribute removal), and pane/row ids
//     derive from the menu id (file-menu-* -> edit-menu-* renames). Closed
//     menus must render no pane (menus_test.go) and a closed trigger must
//     carry no aria-controls, so no id scheme can paper over the move.
//   - NavigationMenu open-switch stays FULL for the same reasons: closed->open
//     adds the panel child to the root, and a switch renames the panel id
//     (docs-content -> blog-content) with its aria-controls/aria-labelledby
//     wiring and entry ids. Both shapes are pinned by navigation_test.go and
//     menu_hover_fix_test.go.
package chadcn_test

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

func barsCommandGroups() []chadcn.CommandGroup {
	return []chadcn.CommandGroup{{Label: "File", Items: []chadcn.CommandItem{
		{Value: "save", Label: "Save"},
		{Value: "open", Label: "Open"},
	}}}
}

func barsSearchItems() []chadcn.SearchItem {
	return []chadcn.SearchItem{
		{Value: "repo", Label: "Repo"},
		{Value: "docs", Label: "Docs"},
	}
}

func barsMenubar(open string) *ui.Element {
	items := []chadcn.MenuItem{{Label: "Open"}, {Label: "Rename"}, {Label: "Delete", Danger: true}}
	return ui.Div().ID("probe-root").Children(chadcn.Menubar(chadcn.MenubarOptions{
		ID:    "mb-1",
		Open:  open,
		Menus: []chadcn.MenubarMenu{{ID: "file", Label: "File", Items: items}, {ID: "edit", Label: "Edit", Items: items}},
	}))
}

// TestProbePatchBarsCommandActiveMove proves an Active-row move on an
// identical item list takes the patch path (aria-selected edits only).
func TestProbePatchBarsCommandActiveMove(t *testing.T) {
	groups := barsCommandGroups()
	v, ctx := mountProbeView(t, ui.Div().ID("probe-root").Children(
		chadcn.Command(chadcn.CommandOptions{ID: "cmd-1", Groups: groups, Active: "save"})))
	expectProbePatch(t, v, ctx, ui.Div().ID("probe-root").Children(
		chadcn.Command(chadcn.CommandOptions{ID: "cmd-1", Groups: groups, Active: "open"})))
}

// TestProbePatchBarsCommandQueryKeystroke proves a query keystroke patches
// while the host list is untouched (input value edit only). A host-filtered
// list change stays FULL by design; see the package comment.
func TestProbePatchBarsCommandQueryKeystroke(t *testing.T) {
	groups := barsCommandGroups()
	v, ctx := mountProbeView(t, ui.Div().ID("probe-root").Children(
		chadcn.Command(chadcn.CommandOptions{ID: "cmd-1", Query: "", Groups: groups, Active: "save"})))
	expectProbePatch(t, v, ctx, ui.Div().ID("probe-root").Children(
		chadcn.Command(chadcn.CommandOptions{ID: "cmd-1", Query: "s", Groups: groups, Active: "save"})))
}

// TestProbePatchBarsSearchSelectValue proves a selection change on an
// identical matched list patches (aria-selected + check text edits).
func TestProbePatchBarsSearchSelectValue(t *testing.T) {
	items := barsSearchItems()
	v, ctx := mountProbeView(t, ui.Div().ID("probe-root").Children(
		chadcn.Search(chadcn.SearchOptions{ID: "s-1", Open: true, Query: "", Items: items})))
	expectProbePatch(t, v, ctx, ui.Div().ID("probe-root").Children(
		chadcn.Search(chadcn.SearchOptions{ID: "s-1", Open: true, Query: "", Items: items, Value: "repo"})))
}

// TestProbePatchBarsSearchEmptyText proves an empty-state rewrite with the
// same (empty) matched set patches (input value + empty text edits).
func TestProbePatchBarsSearchEmptyText(t *testing.T) {
	items := []chadcn.SearchItem{{Value: "dashboard", Label: "Dashboard"}}
	v, ctx := mountProbeView(t, ui.Div().ID("probe-root").Children(
		chadcn.Search(chadcn.SearchOptions{ID: "s-1", Open: true, Query: "zzz", Items: items})))
	expectProbePatch(t, v, ctx, ui.Div().ID("probe-root").Children(
		chadcn.Search(chadcn.SearchOptions{ID: "s-1", Open: true, Query: "yyy", Items: items})))
}

// TestProbePatchBarsSearchFieldLabel proves a closed-panel label edit patches
// (placeholder/aria-label edits on the id-bearing input).
func TestProbePatchBarsSearchFieldLabel(t *testing.T) {
	items := barsSearchItems()
	v, ctx := mountProbeView(t, ui.Div().ID("probe-root").Children(
		chadcn.Search(chadcn.SearchOptions{ID: "s-1", Query: "", Placeholder: "Search site", Items: items})))
	expectProbePatch(t, v, ctx, ui.Div().ID("probe-root").Children(
		chadcn.Search(chadcn.SearchOptions{ID: "s-1", Query: "", Placeholder: "Find in site", Items: items})))
}

// TestProbePatchBarsMenubarIdenticalNoop proves an identical menubar rebuild
// does no page work.
func TestProbePatchBarsMenubarIdenticalNoop(t *testing.T) {
	v, ctx := mountProbeView(t, barsMenubar("file"))
	expectProbeNoop(t, v, ctx, barsMenubar("file"))
}
