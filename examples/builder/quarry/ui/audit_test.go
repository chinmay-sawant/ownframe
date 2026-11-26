package ui

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/store"
)

func TestFreshStoreStartsDark(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	a, err := NewApp(ctx, LaunchOptions{DataDir: dir, DemoDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if !a.Dark() {
		t.Fatal("a fresh store must start in QuarryDark")
	}
	if err := a.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, op := range a.Page().Probe().Display().Ops {
		if strings.TrimSpace(op.Text) == "QuarryDark" {
			found = true
		}
	}
	if !found {
		t.Fatal("status bar should read QuarryDark")
	}
	// A saved light choice survives a restart; clearing it returns to dark.
	if err := a.toggleTheme(ctx); err != nil {
		t.Fatal(err)
	}
	a.Close()
	b, err := NewApp(ctx, LaunchOptions{DataDir: dir, DemoDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	if b.Dark() {
		t.Fatal("the saved light theme should be restored")
	}
	st, err := store.Open(store.Memory)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()
	if st.Theme(ctx) != "dark" {
		t.Fatal("store default theme must be dark")
	}
}

func TestSettingsApplyAndPersist(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	open := func() *rig {
		r := newRig(t, 1440, 960, "")
		return r
	}
	_ = open
	a, err := NewApp(ctx, LaunchOptions{DataDir: dir, DemoDir: t.TempDir(), Width: 1440, Height: 960})
	if err != nil {
		t.Fatal(err)
	}
	r := &rig{t: t, a: a, ctx: ctx}
	r.must(a.Page().Redraw(ctx))
	r.tap("rail-settings")
	if !r.has("settings") || !r.has("set-limit-10,000") {
		t.Fatal("settings view missing its controls")
	}
	r.tap("set-limit-10,000")
	r.tap("set-poll-5s")
	r.tap("set-confirm") // the switch starts on
	r.tap("set-theme-light")
	r.a.page.Form().SetFormValue("set-slow", "120")
	r.tap("set-apply")
	if a.defLimit != tenThousandRows || a.mon.pollEvery.Seconds() != 5 || a.confirmDestructive || a.Dark() {
		t.Fatalf("settings not applied: limit=%d poll=%v confirm=%v dark=%v", a.defLimit, a.mon.pollEvery, a.confirmDestructive, a.Dark())
	}
	if a.slow.Threshold().Milliseconds() != 120 {
		t.Fatal("slow threshold")
	}
	a.newQueryTab("")
	if a.activeTabPtr().rowLimit != tenThousandRows {
		t.Fatal("new tabs should use the default row limit")
	}
	a.Close()
	b, err := NewApp(ctx, LaunchOptions{DataDir: dir, DemoDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	if b.defLimit != tenThousandRows || b.mon.pollEvery.Seconds() != 5 || b.confirmDestructive || b.Dark() || b.slow.Threshold().Milliseconds() != 120 {
		t.Fatalf("settings not restored: %d %v %v %v %v", b.defLimit, b.mon.pollEvery, b.confirmDestructive, b.Dark(), b.slow.Threshold())
	}
}

func TestNoTextGlyphsStandInForIcons(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.openTableByName("users")
	r.must(r.a.menuActionFor("tree-table:main.orders", "structure"))
	r.must(r.a.redraw(r.ctx))
	banned := map[string]bool{"DB": true, "Ex": true, "Mn": true, "Hi": true, "St": true, "v": true, ">": true, "x": true, "[]": true, "T": true, "-": true, "<<": true, "Q": true}
	for _, op := range r.a.Page().Probe().Display().Ops {
		if banned[strings.TrimSpace(op.Text)] {
			t.Errorf("text glyph %q is standing in for an icon", op.Text)
		}
	}
	imgs := 0
	for _, b := range r.a.Page().Boxes() {
		if b.Tag == "img" {
			imgs++
		}
	}
	if imgs < 25 {
		t.Errorf("only %d icons painted; rail, tree, tabs and toolbar should all use them", imgs)
	}
	for _, id := range []string{"rail-connections", "rail-explorer", "rail-monitor", "rail-history", "rail-settings", "sb-refresh", "sb-toggle", "tb-theme", "win-min", "win-max", "win-close", "tab-new"} {
		b := r.box(id)
		hasImg := false
		for _, c := range r.a.Page().Boxes() {
			if c.Tag == "img" && c.X >= b.X && c.Y >= b.Y && c.X+c.W <= b.X+b.W+0.5 && c.Y+c.H <= b.Y+b.H+0.5 {
				hasImg = true
			}
		}
		if !hasImg {
			t.Errorf("%s has no icon", id)
		}
	}
	// The title bar and rail carry the Quarry logo.
	logo := false
	for _, b := range r.a.Page().Boxes() {
		if b.Tag == "img" && b.Y < titleBarHeight && b.X < 40 {
			logo = true
		}
	}
	if !logo {
		t.Error("Quarry logo missing from the title bar")
	}
}

func TestWindowControlsAndResize(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	for _, id := range []string{"win-min", "win-max", "win-close"} {
		b := r.box(id)
		if b.Action != id || b.W < 40 || b.H != titleBarHeight {
			t.Errorf("%s box %+v", id, b)
		}
	}
	label := func() string {
		tree, err := r.a.Page().Access().AccessibilityTree()
		r.must(err)
		for _, n := range tree.Nodes {
			if n.ElementID == "win-max" {
				return n.Name
			}
		}
		return ""
	}
	if label() != "Maximize" {
		t.Fatalf("label = %q", label())
	}
	r.tap("win-max") // must not panic without a desktop host
	maxed := true
	r.a.maximizedFn = func() bool { return maxed }
	r.tick()
	if label() != "Restore" {
		t.Fatalf("after maximize the control should offer Restore, got %q", label())
	}
	maxed = false
	r.tick()
	if label() != "Maximize" {
		t.Fatalf("after restore the control should offer Maximize, got %q", label())
	}
	// A resize (maximize changes the page size) re-lays out the shell and grid.
	r.openTableByName("orders")
	r.a.Page().SetSize(1920, 1080)
	r.tick()
	r.must(r.a.Page().Redraw(r.ctx))
	r.tick()
	if r.a.width != 1920 || r.a.height != 1080 {
		t.Fatalf("size not tracked: %dx%d", r.a.width, r.a.height)
	}
	st, body := r.box("statusbar"), r.box("grid-body")
	if int(st.Y+st.H) != 1080 || body.Y+body.H > st.Y+0.5 || int(r.box("titlebar").W) != 1920 {
		t.Fatalf("layout after resize: status %+v body %+v", st, body)
	}
	if !r.a.Draggable(1500, 10) || r.a.Draggable(1900, 10) {
		t.Fatal("drag guard must follow the new width")
	}
}

func TestPaletteHoverLayoutAndDismiss(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.chord("k", false)
	pal := r.box("palette")
	rows := r.countBoxes("pal-")
	rows -= r.countBoxes("pal-list") + r.countBoxes("pal-input")
	wantH := float64(paletteSides*paletteInset + paletteSearchH + paletteGap + min(rows, paletteRowsShown)*paletteRowH + panelBorder)
	if pal.H > wantH+1 || pal.H < wantH-1 {
		t.Fatalf("palette height %v, want content height %v for %d rows", pal.H, wantH, rows)
	}
	if pal.H > 520 {
		t.Fatalf("palette too tall: %v", pal.H)
	}
	snap := func() map[string][4]float64 {
		m := map[string][4]float64{}
		for _, b := range r.a.Page().Boxes() {
			if strings.HasPrefix(b.ID, "pal-") || b.ID == "palette" {
				m[b.ID] = [4]float64{b.X, b.Y, b.W, b.H}
			}
		}
		return m
	}
	before := snap()
	for i := 0; i < paletteRowsShown; i++ {
		b := r.box("pal-" + itoa(i))
		r.must(r.a.Page().Edit().Hover(r.ctx, b.X+b.W/2, b.Y+b.H/2))
		r.tick()
		after := snap()
		for id, g := range before {
			if after[id] != g {
				t.Fatalf("hovering row %d moved %s: %v -> %v", i, id, g, after[id])
			}
		}
	}
	// Badges share one width so labels line up between Action and Table rows.
	x0 := r.box("pal-0").X
	for i := 1; i < paletteRowsShown; i++ {
		if r.box("pal-"+itoa(i)).X != x0 {
			t.Fatal("row geometry differs")
		}
	}
	// Clicking inside the panel keeps it open; outside, and Escape, close it.
	r.tap("pal-input")
	if r.a.modal != "palette" {
		t.Fatal("click inside the palette must not close it")
	}
	pb := r.box("palette")
	r.must(r.a.Page().Click(r.ctx, pb.X+pb.W-3, pb.Y+pb.H-2)) // panel padding, no row
	if r.a.modal != "palette" {
		t.Fatal("click on the panel padding must not close it")
	}
	p := r.a.Page()
	r.must(p.Click(r.ctx, 20, 600))
	r.tick()
	if r.a.modal != "" {
		t.Fatal("clicking outside must close the palette")
	}
	r.chord("k", false)
	r.key("escape")
	if r.a.modal != "" {
		t.Fatal("Escape must close the palette")
	}
	// More rows than fit scroll inside the fixed-size panel.
	r.chord("k", false)
	h0 := r.box("palette").H
	_, err := r.a.onWheel(r.ctx, 700, 300, 0, -3)
	r.must(err)
	if r.a.pal.top == 0 || r.box("palette").H != h0 {
		t.Fatalf("wheel should scroll the list inside a fixed panel (top=%d h=%v->%v)", r.a.pal.top, h0, r.box("palette").H)
	}
	for i := 0; i < 14; i++ {
		r.key("arrowdown")
	}
	if r.a.pal.active < r.a.pal.top || r.a.pal.active >= r.a.pal.top+paletteRowsShown {
		t.Fatal("active row left the window")
	}
}

func TestInspectorButtonOpensAndCloses(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	tab := r.a.activeTabPtr()
	tab.ed = newEditor(`SELECT id, action, payload FROM audit_events ORDER BY id`)
	r.chord("enter", false)
	r.settle()
	if r.has("inspector") {
		t.Fatal("drawer should start closed")
	}
	r.tap("rs-inspector")
	if !r.has("inspector") {
		t.Fatal("the Inspector button must open the drawer")
	}
	ins := r.box("inspector")
	if ins.W < float64(minInspector) || ins.H < 100 {
		t.Fatalf("drawer geometry %+v", ins)
	}
	r.tap("cell-0-2")
	if !strings.Contains(r.texts(), "action") && !strings.Contains(r.texts(), "{") {
		t.Fatal("drawer should show the selected cell")
	}
	if r.box("insp-body").H < 40 {
		t.Fatal("JSON payload should pretty-print over several lines")
	}
	for _, op := range r.a.Page().Probe().Display().Ops {
		if strings.TrimSpace(op.Text) == `"ip"` || strings.Contains(op.Text, `"action"`) {
			goto ok
		}
	}
	t.Fatalf("pretty JSON keys not painted: %q", r.texts())
ok:
	r.tap("cell-1-1") // another cell updates the drawer
	r.tap("rs-inspector")
	if r.has("inspector") {
		t.Fatal("the Inspector button must also close the drawer")
	}
	r.tap("rs-inspector")
	r.tap("insp-close")
	if r.has("inspector") {
		t.Fatal("Close must hide the drawer")
	}
	// Same in a table tab, and Escape closes it.
	r.openTableByName("audit_events")
	r.tap("tbl-inspector")
	if !r.has("inspector") {
		t.Fatal("table inspector")
	}
	r.a.focus = "grid"
	r.key("escape")
	if r.has("inspector") {
		t.Fatal("Escape closes the drawer")
	}
	r.png("inspector-final")
}
