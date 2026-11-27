package ui

import (
	"strconv"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/driver"
)

func TestSidebarTreeAndFilter(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	if !r.has("tree-schema:main") || !r.has("tree-table:main.users") {
		t.Fatal("tree should start with main expanded and users listed")
	}
	r.tap("tree-schema:main") // collapse
	if r.has("tree-table:main.users") {
		t.Fatal("collapse did not hide children")
	}
	r.tap("tree-schema:main")
	r.tap("tree-group:main:Views")
	if !r.has("tree-view:main.v_revenue_by_day") {
		t.Fatal("expanding Views should list v_revenue_by_day")
	}
	r.a.page.Form().SetFormValue("sb-filter", "ord")
	r.must(r.a.onChange(r.ctx, r.box("sb-filter")))
	if r.has("tree-table:main.users") || !r.has("tree-table:main.orders") || !r.has("tree-table:main.order_items") {
		t.Fatal("filter 'ord' should keep orders and order_items only")
	}
	r.tap("tree-table:main.orders")
	if t0 := r.a.activeTabPtr(); t0.kind != tabTable || t0.table != "orders" {
		t.Fatalf("clicking a table should open it, got %+v", t0)
	}
	r.settle()
	if r.a.activeTabPtr().grid == nil || r.a.activeTabPtr().totalRows != 2500 {
		t.Fatalf("orders not loaded: total=%d", r.a.activeTabPtr().totalRows)
	}
}

func TestSidebarContextMenu(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	id := "tree-table:main.orders"
	items := r.a.contextMenuFor(id)
	want := []string{"open", "structure", "newquery", "copyname", "refresh"}
	if len(items) != len(want) {
		t.Fatalf("items = %+v", items)
	}
	for i, it := range items {
		if it.ID != want[i] || !it.Enabled {
			t.Errorf("item %d = %+v", i, it)
		}
	}
	act := func(id, node string) {
		r.a.contextMenuFor(node)
		r.must(r.a.menuAction(r.ctx, id))
	}
	act("open", id)
	r.settle()
	if tb := r.a.activeTabPtr(); tb.kind != tabTable || tb.table != "orders" {
		t.Fatalf("open: %+v", tb)
	}
	act("structure", id)
	if tb := r.a.activeTabPtr(); tb.kind != tabStructure {
		t.Fatalf("structure: %+v", tb)
	}
	act("newquery", id)
	if tb := r.a.activeTabPtr(); tb.kind != tabQuery || !strings.Contains(tb.ed.text, `FROM "orders"`) {
		t.Fatalf("new query text = %q", tb.ed.text)
	}
	act("copyname", id)
	if got, _ := ownframe.Read(); got != "main.orders" {
		t.Fatalf("clipboard = %q", got)
	}
	n := len(r.a.schema)
	act("refresh", id)
	if len(r.a.schema) != n {
		t.Fatal("refresh changed schema count")
	}
	if r.a.contextMenuFor("no-such-id") != nil {
		t.Fatal("unknown id must fall back to the default menu")
	}
}

func TestStructureInspectorTab(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.must(r.a.menuActionFor("tree-table:main.orders", "structure"))
	if !r.has("columns-table") {
		t.Fatal("columns sub-tab expected")
	}
	if !r.has("col-row-0") || !strings.Contains(r.texts(), "user_id") {
		t.Fatalf("columns not painted: %q", r.texts())
	}
	r.tap("st-sub-indexes")
	if !r.has("indexes-table") || !strings.Contains(r.texts(), "idx_orders_user") {
		t.Fatalf("indexes not painted: %q", r.texts())
	}
	if !r.has("fk-row-0") {
		t.Fatal("foreign key row expected")
	}
	r.tap("st-sub-ddl")
	if !strings.Contains(r.texts(), "CREATE") {
		t.Fatalf("DDL not painted: %q", r.texts())
	}
	r.tap("ddl-copy")
	if got, _ := ownframe.Read(); !strings.Contains(strings.ToUpper(got), "CREATE TABLE") {
		t.Fatalf("copied DDL = %q", got)
	}
	// Syntax highlighting reaches the DDL block.
	kw := false
	for _, op := range r.a.Page().Probe().Display().Ops {
		if op.Kind == ownframe.DisplayOpText && strings.TrimSpace(op.Text) == "CREATE" {
			kw = true
		}
	}
	if !kw {
		t.Fatal("CREATE keyword should be its own coloured run")
	}
}

// menuActionFor invokes a sidebar context-menu action for a node id.
func (a *App) menuActionFor(node, action string) error {
	a.contextMenuFor(node)
	return a.menuAction(a.ctx, action)
}

func TestWorkspaceTabLifecycle(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	if len(r.a.tabs) != 1 {
		t.Fatal("one query tab at start")
	}
	r.tap("tab-new")
	r.chord("n", false)
	if len(r.a.tabs) != 3 {
		t.Fatalf("tabs = %d", len(r.a.tabs))
	}
	r.openTableByName("users")
	r.must(r.a.menuActionFor("tree-table:main.users", "structure"))
	r.a.openMonitorTab()
	r.must(r.a.redraw(r.ctx))
	kinds := map[tabKind]bool{}
	for _, tb := range r.a.tabs {
		kinds[tb.kind] = true
	}
	if len(kinds) != 4 {
		t.Fatalf("expected all four tab kinds, got %v", kinds)
	}
	r.a.activeTab = 0
	r.chord("tab", false)
	if r.a.activeTab != 1 {
		t.Fatalf("Ctrl+Tab -> %d", r.a.activeTab)
	}
	first := r.a.tabs[0]
	first.dirty = true
	r.must(r.a.redraw(r.ctx))
	if !r.has("tab-dirty-" + itoa(first.id)) {
		t.Fatal("dirty indicator missing")
	}
	r.tap("tab-" + itoa(r.a.tabs[2].id))
	if r.a.activeTab != 2 {
		t.Fatal("tab click did not switch")
	}
	n := len(r.a.tabs)
	r.chord("w", false)
	if len(r.a.tabs) != n-1 {
		t.Fatal("Ctrl+W did not close the tab")
	}
	r.tap("tab-close-" + itoa(first.id))
	for len(r.a.tabs) > 0 && r.has("tab-close-"+itoa(r.a.tabs[0].id)) && len(r.a.tabs) > 1 {
		r.tap("tab-close-" + itoa(r.a.tabs[0].id))
	}
	if len(r.a.tabs) < 1 || r.a.activeTabPtr() == nil {
		t.Fatal("closing every tab must leave a usable query tab")
	}
}

func itoa(n int) string { return strconv.Itoa(n) }

var _ = driver.EngineSQLite
