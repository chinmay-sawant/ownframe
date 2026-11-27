package ui

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/driver"
	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/store"
)

func TestCommandPaletteSearchAndRun(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.chord("k", false)
	if r.a.modal != "palette" || !r.has("pal-input") || !r.has("pal-0") {
		t.Fatal("Ctrl+K should open the palette with results")
	}
	r.a.page.Form().SetFormValue("pal-input", "ordit")
	r.must(r.a.onChange(r.ctx, r.box("pal-input")))
	items := r.a.paletteItems()
	if len(items) == 0 || items[0].label != "main.order_items" {
		t.Fatalf("fuzzy 'ordit' -> %+v", items)
	}
	r.key("enter")
	if tb := r.a.activeTabPtr(); r.a.modal != "" || tb.kind != tabTable || tb.table != "order_items" {
		t.Fatalf("enter should open the table, got %+v", tb)
	}
	r.settle()
	// Theme toggle by keyboard.
	r.chord("p", false)
	r.a.page.Form().SetFormValue("pal-input", "theme")
	r.must(r.a.onChange(r.ctx, r.box("pal-input")))
	r.key("enter")
	if r.a.Dark() {
		t.Fatal("palette theme action should switch to light")
	}
	// Arrow keys move the highlight; click runs.
	r.chord("k", false)
	r.key("arrowdown")
	if r.a.pal.active != 1 {
		t.Fatalf("active = %d", r.a.pal.active)
	}
	r.key("escape")
	if r.a.modal != "" {
		t.Fatal("escape closes")
	}
	r.tap("tb-palette")
	r.a.page.Form().SetFormValue("pal-input", "monitor")
	r.must(r.a.onChange(r.ctx, r.box("pal-input")))
	r.tap("pal-0")
	if r.a.activeTabPtr().kind != tabMonitor {
		t.Fatal("Open Live Monitor")
	}
	for _, want := range []string{"Export Current Results", "Toggle Read-Only Mode", "New SQL Query"} {
		found := false
		r.a.pal = paletteState{}
		for _, it := range r.a.paletteItems() {
			found = found || it.label == want
		}
		if !found {
			t.Errorf("palette lacks %q", want)
		}
	}
	if fuzzyScore("xyz", "users") >= 0 || fuzzyScore("usr", "users") < 0 {
		t.Fatal("fuzzy matching")
	}
}

func TestSplitPaneDragClampsAndPersists(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	drag := func(id string, dx, dy float64) {
		claimed, err := r.a.onDragStart(r.ctx, r.box(id))
		r.must(err)
		if !claimed {
			t.Fatalf("%s not claimed", id)
		}
		r.must(r.a.onDragMove(r.ctx, dx, dy))
		r.must(r.a.onDragEnd(r.ctx))
	}
	drag("split-sidebar", 60, 0)
	if r.a.sidebarW != 308 {
		t.Fatalf("sidebar %d", r.a.sidebarW)
	}
	drag("split-sidebar", 900, 0)
	if r.a.sidebarW != maxSidebarW {
		t.Fatalf("max clamp %d", r.a.sidebarW)
	}
	drag("split-sidebar", -900, 0)
	if r.a.sidebarW != minSidebarW {
		t.Fatalf("min clamp %d", r.a.sidebarW)
	}
	if int(r.box("sidebar").W) != minSidebarW {
		t.Fatal("layout did not follow the drag")
	}
	drag("split-editor", 0, 5000)
	if r.a.editorH != maxEditorH {
		t.Fatalf("editor %d", r.a.editorH)
	}
	drag("split-editor", 0, -5000)
	if r.a.editorH != minEditorH {
		t.Fatalf("editor min %d", r.a.editorH)
	}
	r.a.activeTabPtr().grid = newGridState(syntheticResult(10))
	r.a.activeTabPtr().inspectorOpen = true
	r.must(r.a.redraw(r.ctx))
	drag("split-inspector", -100, 0)
	if r.a.inspectorW != 320+100 {
		t.Fatalf("inspector %d", r.a.inspectorW)
	}
	drag("colresize-1", 40, 0)
	if r.a.activeTabPtr().grid.widths[1] <= 72 {
		t.Fatal("column width unchanged")
	}
	if got, _ := r.a.loadPref("split.sidebar"); got != "180" {
		t.Fatalf("sidebar pref %q", got)
	}
	// Persisted values restore into a new app sharing the store.
	r.a.restoreLayout()
	if r.a.sidebarW != minSidebarW || r.a.editorH != minEditorH || r.a.inspectorW != 420 {
		t.Fatalf("restore = %d %d %d", r.a.sidebarW, r.a.editorH, r.a.inspectorW)
	}
	r.a.savePref("split.sidebar", "9999")
	r.a.restoreLayout()
	if r.a.sidebarW != maxSidebarW {
		t.Fatal("stored value must clamp on restore")
	}
}

func TestProdReadOnlyGuardrail(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	cfg := r.a.cfg
	cfg.EnvTag, cfg.ID, cfg.Name = driver.EnvProd, "", "Prod DB"
	r.must(r.a.connect(r.ctx, cfg))
	r.must(r.a.redraw(r.ctx))
	if !r.has("prod-banner") || !strings.Contains(r.texts(), "Production connection - Read-Only Mode active") {
		t.Fatal("banner missing")
	}
	if !r.has("tb-lock") || !r.has("st-readonly") {
		t.Fatal("lock indicators missing")
	}
	tab := r.a.activeTabPtr()
	tab.ed = newEditor("UPDATE users SET plan = 'x' WHERE id = 1")
	r.chord("enter", false)
	r.settle()
	if tab.err == nil || !strings.Contains(tab.err.Message, "read-only") {
		t.Fatalf("write must be blocked, err=%+v", tab.err)
	}
	// Unlock requires an explicit confirmation.
	r.tap("prod-unlock")
	if r.a.modal != "confirm" || r.a.cfg.ReadOnly == false {
		t.Fatal("unlock must be confirmed first")
	}
	r.tap("confirm-ok")
	if r.a.cfg.ReadOnly || !strings.Contains(r.texts(), "UNLOCKED") {
		t.Fatal("writes should be unlocked after confirmation")
	}
	r.chord("enter", false)
	r.settle()
	if tab.err != nil {
		t.Fatalf("write after unlock: %+v", tab.err)
	}
	// DROP needs the table name typed.
	tab.ed = newEditor("DROP TABLE order_items")
	r.chord("enter", false)
	if r.a.modal != "confirm" || r.a.running != nil {
		t.Fatal("DROP must ask for confirmation")
	}
	r.a.page.Form().SetFormValue("confirm-typed", "wrong")
	r.tap("confirm-ok")
	if r.a.modal != "confirm" || r.a.modalErr == "" || r.a.running != nil {
		t.Fatal("wrong name must not run the statement")
	}
	r.a.page.Form().SetFormValue("confirm-typed", "order_items")
	r.tap("confirm-ok")
	r.settle()
	if _, err := r.a.drv.Query(r.ctx, "SELECT 1 FROM order_items", 1); err == nil {
		t.Fatal("table should be gone after the confirmed DROP")
	}
}

func TestAccessibilitySnapshot(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.tap("rail-monitor")
	r.tap("tab-new")
	tree, err := r.a.Page().Access().AccessibilityTree()
	r.must(err)
	names := map[string]bool{}
	for _, n := range tree.Nodes {
		{
			if n.HasBounds && (n.Bounds.W < 0 || n.Bounds.H < 0) {
				t.Errorf("invalid bounds %+v", n)
			}
		}
		names[n.ElementID] = n.Name != ""
	}
	for _, id := range []string{"tb-conn", "tb-palette", "tb-theme", "win-min", "win-max", "win-close", "rail-connections", "rail-explorer",
		"rail-monitor", "rail-history", "rail-settings", "sb-filter", "sb-refresh", "tab-new", "ed-run", "ed-explain", "editor-body"} {
		ok, present := names[id]
		if !present {
			t.Errorf("%s missing from accessibility tree", id)
		} else if !ok {
			t.Errorf("%s has no accessible name", id)
		}
	}
	// Grid and monitor KPI semantics.
	r.openTableByName("users")
	tree, _ = r.a.Page().Access().AccessibilityTree()
	got := map[string]bool{}
	for _, n := range tree.Nodes {
		if n.Name != "" {
			got[n.ElementID] = true
		}
	}
	if !got["tbl-apply"] || !got["tbl-filter"] {
		t.Errorf("table grid header / controls unnamed: %v", got)
	}
	r.openMonitor()
	tree, _ = r.a.Page().Access().AccessibilityTree()
	kpi := false
	for _, n := range tree.Nodes {
		if n.ElementID == "kpi-conns" && n.Name != "" {
			kpi = true
		}
	}
	if !kpi {
		t.Error("KPI card has no name")
	}
	// Password field in the connection modal stays redacted.
	r.tap("rail-connections")
	r.tap("cd-engine-postgresql")
	r.a.page.Form().SetFormValue("cd-pass", "hunter2")
	r.must(r.a.redraw(r.ctx))
	tree, _ = r.a.Page().Access().AccessibilityTree()
	seen := false
	for _, n := range tree.Nodes {
		if strings.Contains(n.Value, "hunter2") || strings.Contains(n.Name, "hunter2") {
			t.Errorf("password leaked: %+v", n)
		}
		if n.ElementID == "cd-pass" {
			seen = true
			if !n.Password || n.Value != "" {
				t.Errorf("password node not redacted: %+v", n)
			}
		}
	}
	if !seen {
		t.Error("password node missing")
	}
}

func TestKeyboardOnlyWorkflow(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.chord("k", false)
	r.a.page.Form().SetFormValue("pal-input", "orders")
	r.must(r.a.onChange(r.ctx, r.box("pal-input")))
	r.key("enter")
	r.settle()
	tab := r.a.activeTabPtr()
	if tab.kind != tabTable {
		t.Fatal("Ctrl+K jump to table")
	}
	r.tap("cell-0-0")
	r.key("arrowdown")
	r.key("arrowright")
	if tab.grid.fr != 1 || tab.grid.fc != 1 {
		t.Fatalf("grid nav %d,%d", tab.grid.fr, tab.grid.fc)
	}
	n := len(r.a.tabs)
	r.chord("n", false)
	if len(r.a.tabs) != n+1 || !r.a.editorFocused() {
		t.Fatal("Ctrl+N should open a focused query tab")
	}
	t.Logf("focus=%q form=%q modal=%q", r.a.focus, r.a.page.Form().FocusID(), r.a.modal)
	r.typ("SELECT count(*) AS total FROM orders")
	t.Logf("text=%q", r.a.activeTabPtr().ed.text)
	r.chord("enter", false)
	r.settle()
	q := r.a.activeTabPtr()
	if q.grid == nil || q.grid.rows[0][0].I != 2500 {
		t.Fatalf("query result %+v", q.err)
	}
	dark := r.a.Dark()
	r.chord("t", true)
	if r.a.Dark() == dark {
		t.Fatal("Ctrl+Shift+T toggles theme")
	}
	r.chord("w", false)
	if len(r.a.tabs) != n {
		t.Fatal("Ctrl+W closes the tab")
	}
	_ = context.Background
	_ = store.Memory
	_ = ownframe.Read
}
