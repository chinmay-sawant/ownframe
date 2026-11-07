// Open->open content edits on popups and menus must take the ui.View.Update
// patch path (parses+0), not the full-render fallback. Closed->open mounts or
// unmounts the panel, which is structural and stays FULL by design.
//
// What patches and what stays FULL by design (see ui/diff.go: same tag, same
// child count, same order at every node; every changed node needs a stable
// id; a removed attribute forces a rebuild; plus the engine refuses a patch
// whose element changed size):
//
//   - Popover/HoverCard body edits patch: the host gives the body a stable
//     id, and the fixed-width panel keeps the box stable. No code change was
//     needed. Popup bodies are host-built elements; anonymous children (such
//     as bare ui.Text) carry no id and never patch, so the probes below use
//     id-bearing bodies.
//   - Dropdown/ContextMenu simple-row relabels patch: a label-only row now
//     carries the label as the button's own text, and the menu min-width
//     keeps the button box stable, so the relabel is a text edit on the
//     id-bearing row. Previously the label sat in a span that resized with
//     its text, which the engine refuses.
//   - Menu row ids are positional (menu id + index): stable across relabels
//     while the host keeps item order. Keep the item list order stable
//     across updates; an insert or delete changes the child count and stays
//     FULL.
//   - Tooltip content edits stay FULL: the label is width max-content, so any
//     text change resizes it and the engine refuses the splice
//     ("element subtree changed size"). No id scheme can paper over the
//     resize without redesigning the label width, which would change visible
//     behavior.
//   - Submenu and shortcut rows keep the label in a span (the button also
//     holds the chevron or shortcut), so their relabels resize the span and
//     stay FULL. Group labels carry no stable id and stay FULL.
package chadcn_test

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

func probeMenusPopRoot(el *ui.Element) *ui.Element {
	return ui.Div().ID("probe-root").Children(el)
}

func probeMenusPopTrigger(label, id string) *ui.Element {
	return chadcn.Button(label, chadcn.ButtonOptions{}).ID(id)
}

func probeMenusPopItems() []chadcn.MenuItem {
	return []chadcn.MenuItem{
		{Label: "Open", Shortcut: "⌘O"},
		{Label: "Rename"},
		{Label: "Delete", Danger: true},
	}
}

func TestProbePatchMenusPop_PopoverBodyEdit(t *testing.T) {
	t.Parallel()
	first := probeMenusPopRoot(chadcn.Popover(chadcn.PopoverOptions{
		ID: "pop-1", Open: true,
		Trigger: probeMenusPopTrigger("Info", "pop-trig"),
	}, ui.P().ID("pop-body").Text("detail")))
	v, ctx := mountProbeView(t, first)
	next := probeMenusPopRoot(chadcn.Popover(chadcn.PopoverOptions{
		ID: "pop-1", Open: true,
		Trigger: probeMenusPopTrigger("Info", "pop-trig"),
	}, ui.P().ID("pop-body").Text("detail v2")))
	expectProbePatch(t, v, ctx, next)
}

func TestProbePatchMenusPop_HoverCardBodyEdit(t *testing.T) {
	t.Parallel()
	first := probeMenusPopRoot(chadcn.HoverCard(chadcn.HoverCardOptions{
		ID: "hc-1", Open: true,
	}, probeMenusPopTrigger("User", "hc-trig"), ui.P().ID("hc-body").Text("bio")))
	v, ctx := mountProbeView(t, first)
	next := probeMenusPopRoot(chadcn.HoverCard(chadcn.HoverCardOptions{
		ID: "hc-1", Open: true,
	}, probeMenusPopTrigger("User", "hc-trig"), ui.P().ID("hc-body").Text("bio v2")))
	expectProbePatch(t, v, ctx, next)
}

func TestProbePatchMenusPop_DropdownItemLabelEdit(t *testing.T) {
	t.Parallel()
	first := probeMenusPopRoot(chadcn.DropdownMenu(chadcn.DropdownMenuOptions{
		ID: "dd-1", Open: true, Items: probeMenusPopItems(),
		Trigger: probeMenusPopTrigger("Menu", "dd-trig"),
	}))
	v, ctx := mountProbeView(t, first)
	items := probeMenusPopItems()
	items[1].Label = "Rename v2"
	next := probeMenusPopRoot(chadcn.DropdownMenu(chadcn.DropdownMenuOptions{
		ID: "dd-1", Open: true, Items: items,
		Trigger: probeMenusPopTrigger("Menu", "dd-trig"),
	}))
	expectProbePatch(t, v, ctx, next)
}

func TestProbePatchMenusPop_ContextMenuItemLabelEdit(t *testing.T) {
	t.Parallel()
	first := probeMenusPopRoot(chadcn.ContextMenu(chadcn.ContextMenuOptions{
		ID: "cm-1", Open: true, X: 120, Y: 80, Items: probeMenusPopItems(),
	}))
	v, ctx := mountProbeView(t, first)
	items := probeMenusPopItems()
	items[1].Label = "Rename v2"
	next := probeMenusPopRoot(chadcn.ContextMenu(chadcn.ContextMenuOptions{
		ID: "cm-1", Open: true, X: 120, Y: 80, Items: items,
	}))
	expectProbePatch(t, v, ctx, next)
}

func TestProbePatchMenusPop_OpenNoop(t *testing.T) {
	t.Parallel()
	first := probeMenusPopRoot(chadcn.Popover(chadcn.PopoverOptions{
		ID: "pop-1", Open: true,
		Trigger: probeMenusPopTrigger("Info", "pop-trig"),
	}, ui.P().ID("pop-body").Text("detail")))
	v, ctx := mountProbeView(t, first)
	next := probeMenusPopRoot(chadcn.Popover(chadcn.PopoverOptions{
		ID: "pop-1", Open: true,
		Trigger: probeMenusPopTrigger("Info", "pop-trig"),
	}, ui.P().ID("pop-body").Text("detail")))
	expectProbeNoop(t, v, ctx, next)
}
