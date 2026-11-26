package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/ui"

	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/driver"
)

func nodeGlyph(k driver.SchemaNodeKind) string {
	switch k {
	case driver.NodeTable:
		return "table"
	case driver.NodeView:
		return "eye"
	case driver.NodeMatView:
		return "layers"
	case driver.NodeIndex:
		return "hash"
	case driver.NodeTrigger:
		return "zap"
	case driver.NodeFunction:
		return "code"
	}
	return ""
}

func nodeKey(n driver.SchemaNode) string {
	if n.Kind == driver.NodeIndex || n.Kind == driver.NodeTrigger {
		return string(n.Kind) + ":" + n.Schema + "." + n.Table + "." + n.Name
	}
	return string(n.Kind) + ":" + n.Schema + "." + n.Name
}

// sidebar renders the 248 px schema explorer.
func (a *App) sidebar() *ui.Element {
	head := ui.Div().Class("qr-sidebar-head").Children(
		ui.Span().ID("sb-title").Class("qr-sidebar-title").Text(a.connLabel()),
		ui.Button("").ID("sb-refresh").Class("qr-iconbtn").Attr("type", "button").Attr("aria-label", "Refresh schema").Attr("title", "Refresh schema").
			Children(a.icon("refresh", toneMuted, iconMedium)).
			OnClick(func(ctx context.Context) error { a.refreshSchema(ctx); return a.redraw(ctx) }),
		ui.Button("").ID("sb-toggle").Class("qr-iconbtn").Attr("type", "button").Attr("aria-label", "Collapse sidebar").Attr("title", "Collapse sidebar").
			Children(a.icon("chevrons-left", toneMuted, iconMedium)).
			OnClick(func(ctx context.Context) error { a.sidebarOpen = false; return a.redraw(ctx) }),
	)
	filter := ui.Input().ID("sb-filter").Class("qr-input").Attr("type", "text").
		Attr("placeholder", "Filter tables").Attr("aria-label", "Filter schema objects").Attr("value", a.filter)
	tree := ui.Div().ID("sb-tree").Class("qr-tree").Attr("role", ariaRole("tree")).Attr("aria-label", "Schema tree")
	a.nodeByID = map[string]driver.SchemaNode{}
	q := strings.ToLower(strings.TrimSpace(a.filter))
	for _, s := range a.schema {
		a.renderSchemaNode(tree, s, q)
	}
	if a.connErr != "" {
		tree.Children(ui.Div().ID("sb-error").Class("qr-note qr-note-err qr-pad").Text(a.connErr))
	}
	return ui.Tag("aside").ID("sidebar").Class("qr-sidebar").Width(ui.Px(float64(a.sidebarW))).Children(head, ui.Div().Class("qr-filter").Children(filter), a.pane("tree", "Schema tree scroll", "qr-tree-pane", tree))
}

func matches(n driver.SchemaNode, q string) bool {
	if q == "" {
		return true
	}
	return strings.Contains(strings.ToLower(n.Name), q)
}

func (a *App) groupMatches(g driver.SchemaNode, q string) bool {
	for _, c := range g.Children {
		if matches(c, q) {
			return true
		}
	}
	return false
}

func (a *App) renderSchemaNode(tree *ui.Element, s driver.SchemaNode, q string) {
	any := false
	for _, g := range s.Children {
		if a.groupMatches(g, q) {
			any = true
		}
	}
	if q != "" && !any {
		return
	}
	skey := "schema:" + s.Name
	open := a.expanded[skey] || q != ""
	tree.Children(a.treeRow("tree-"+skey, 0, open, "", s.Name, -1, false, func(ctx context.Context) error {
		a.expanded[skey] = !a.expanded[skey]
		return a.redraw(ctx)
	}))
	if !open {
		return
	}
	for _, g := range s.Children {
		if !a.groupMatches(g, q) {
			continue
		}
		g := g
		gkey := "group:" + s.Name + ":" + g.Name
		gopen := a.expanded[gkey] || q != ""
		tree.Children(a.treeRow("tree-"+gkey, 1, gopen, "", g.Name, int64(len(g.Children)), false, func(ctx context.Context) error {
			a.expanded[gkey] = !a.expanded[gkey]
			return a.redraw(ctx)
		}))
		if !gopen {
			continue
		}
		for _, leaf := range g.Children {
			if !matches(leaf, q) {
				continue
			}
			leaf := leaf
			id := "tree-" + nodeKey(leaf)
			a.nodeByID[id] = leaf
			row := a.treeRow(id, leafDepth, false, nodeGlyph(leaf.Kind), leaf.Name, leaf.RowCount, a.selNode == id, func(ctx context.Context) error {
				a.selNode = id
				if leaf.Kind == driver.NodeTable || leaf.Kind == driver.NodeView || leaf.Kind == driver.NodeMatView {
					return a.openTable(ctx, leaf)
				}
				return a.redraw(ctx)
			})
			tree.Children(row)
		}
	}
}

func (a *App) treeRow(id string, depth int, open bool, glyph, label string, count int64, active bool, fn func(context.Context) error) *ui.Element {
	cls := "qr-tree-row"
	if active {
		cls += " qr-tree-row-active"
	}
	var caret, kind *ui.Element
	switch {
	case glyph != "":
		kind = a.icon(glyph, toneMuted, iconSmall)
	case depth == 0:
		caret = a.icon(map[bool]string{true: "chevron-down", false: "chevron-right"}[open], toneMuted, iconSmall)
		kind = a.icon("database", toneBrand, iconSmall)
	default:
		caret = a.icon(map[bool]string{true: "chevron-down", false: "chevron-right"}[open], toneMuted, iconSmall)
		kind = a.icon("folder", toneMuted, iconSmall)
	}
	row := ui.Div().ID(id).Class(cls).Attr("role", ariaRole("treeitem")).Attr("aria-label", label).
		Attr("aria-level", fmt.Sprint(depth+1)).PaddingLeft(ui.Px(float64(treeBasePad+depth*treeIndent))).Children(
		ui.Span().Class("qr-tree-caret").Children(caret),
		ui.Span().Class("qr-tree-kind").Children(kind),
		ui.Span().Class("qr-tree-label").Text(label),
	)
	if glyph == "" {
		row.Attr("aria-expanded", fmt.Sprint(open))
	}
	if count >= 0 {
		row.Children(ui.Span().Class("qr-count").Text(fmt.Sprint(count)))
	}
	return row.OnClick(fn)
}

// ---- context menus ----

type contextMenu struct {
	items []ownframe.MenuItem
	node  driver.SchemaNode
	row   int
	col   int
	grid  bool
}

func (a *App) contextMenuFor(id string) []ownframe.MenuItem {
	if n, ok := a.nodeByID[id]; ok {
		items := []ownframe.MenuItem{{ID: "refresh", Label: "Refresh Schema", Enabled: true}, {ID: "copyname", Label: "Copy Qualified Name", Enabled: true}}
		if n.Kind == driver.NodeTable || n.Kind == driver.NodeView || n.Kind == driver.NodeMatView {
			items = []ownframe.MenuItem{
				{ID: "open", Label: "Open Table Data", Enabled: true},
				{ID: "structure", Label: "Inspect Structure / DDL", Enabled: true},
				{ID: "newquery", Label: "New Query for Table", Enabled: true},
				{ID: "copyname", Label: "Copy Qualified Name", Enabled: true},
				{ID: "refresh", Label: "Refresh Schema", Enabled: true},
			}
		}
		a.ctxMenu = &contextMenu{items: items, node: n}
		return items
	}
	var r, c int
	if _, err := fmt.Sscanf(id, "cell-%d-%d", &r, &c); err == nil {
		if t := a.activeTabPtr(); t != nil && t.grid != nil {
			w := t.writable
			a.ctxMenu = &contextMenu{grid: true, row: r, col: c}
			items := []ownframe.MenuItem{
				{ID: "g-copy", Label: "Copy Value", Enabled: true},
				{ID: "g-json", Label: "Copy Row as JSON", Enabled: true},
				{ID: "g-csv", Label: "Copy Row as CSV", Enabled: true},
				{ID: "g-insert", Label: "Copy Row as INSERT", Enabled: true},
				{ID: "g-md", Label: "Copy Selection as Markdown Table", Enabled: true},
				{ID: "g-filter", Label: "Filter by Value", Enabled: true},
				{ID: "g-inspect", Label: "Open in Inspector", Enabled: true},
				{ID: "g-null", Label: "Set NULL", Enabled: w},
				{ID: "g-dup", Label: "Duplicate Row", Enabled: w},
				{ID: "g-del", Label: "Delete Row", Enabled: w},
			}
			a.ctxMenu.items = items
			return items
		}
	}
	return nil
}

func (a *App) menuAction(ctx context.Context, id string) error {
	cm := a.ctxMenu
	a.ctxMenu = nil
	if cm == nil {
		return nil
	}
	if cm.grid {
		return a.gridMenuAction(ctx, id, cm.row, cm.col)
	}
	n := cm.node
	switch id {
	case "open":
		return a.openTable(ctx, n)
	case "structure":
		return a.openStructure(ctx, n)
	case "newquery":
		t := a.newQueryTab("SELECT * FROM " + driver.QualifiedTable(qualSchema(a.cfg, n.Schema), n.Name) + " LIMIT 100;")
		t.dirty = true
		return a.redraw(ctx)
	case "copyname":
		a.copyClipboard(n.QualifiedName())
		a.notify("Copied " + n.QualifiedName())
		return a.redraw(ctx)
	case "refresh":
		a.refreshSchema(ctx)
		return a.redraw(ctx)
	}
	return nil
}

// qualSchema omits the schema for SQLite's "main" in generated SQL.
func qualSchema(cfg driver.ConnectionConfig, schema string) string {
	if cfg.Engine == driver.EngineSQLite {
		return ""
	}
	return schema
}
