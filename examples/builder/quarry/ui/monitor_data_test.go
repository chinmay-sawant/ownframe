package ui

import (
	"strings"
	"testing"
)

// kpiTexts returns the painted text of the monitor KPI cards by id.
func (r *rig) kpiValue(id string) string {
	r.t.Helper()
	b := r.box(id)
	var out []string
	for _, op := range r.a.Page().Probe().Display().Ops {
		x, y := op.X*4/3, op.Y*4/3
		if op.Text != "" && x >= b.X && x <= b.X+b.W && y >= b.Y && y <= b.Y+b.H {
			out = append(out, strings.TrimSpace(op.Text))
		}
	}
	return strings.Join(out, " | ")
}

// TestMonitorFirstSampleHasRealSizeAndBreakdown settles whether "0 B" and an
// empty breakdown are a bug: they are not. The first real sample of the SQLite
// demo database and of the mock PostgreSQL carries a non-zero size and a
// per-object breakdown; the empty cards in earlier screenshots came from
// synthetic samples that had no snapshot.
func TestMonitorFirstSampleHasRealSizeAndBreakdown(t *testing.T) {
	for _, db := range []string{"", "mock:postgres"} {
		r := newRig(t, 1440, 960, db)
		r.openMonitor()
		s, ok := r.a.latestSample()
		if !ok || !s.HasSnapshot {
			t.Fatalf("%q: no real sample", db)
		}
		if s.SizeBytes <= 0 {
			t.Fatalf("%q: size %d", db, s.SizeBytes)
		}
		if len(s.Snapshot.Storage) < 3 {
			t.Fatalf("%q: storage breakdown %v", db, s.Snapshot.Storage)
		}
		for _, it := range s.Snapshot.Storage {
			if it.Bytes <= 0 || it.Name == "" {
				t.Fatalf("%q: bad storage item %+v", db, it)
			}
		}
		if !r.has("storage-row-0") || !r.has("storage-bar-2") {
			t.Fatalf("%q: breakdown rows not painted", db)
		}
		size := r.kpiValue("kpi-size")
		if strings.Contains(size, "| 0 B") || strings.HasPrefix(size, "Database Size & Latency p95 | 0 B") {
			t.Fatalf("%q: size KPI shows zero: %s", db, size)
		}
		if db == "" {
			// SQLite: real file size, real tables, journal mode from the pragma,
			// no cache ratio.
			var want int64
			res, err := r.a.drv.Query(r.ctx, "SELECT page_count * page_size FROM pragma_page_count, pragma_page_size", 1)
			r.must(err)
			want = res.Rows[0][0].I
			if s.SizeBytes != want {
				t.Fatalf("sqlite size %d, pragma says %d", s.SizeBytes, want)
			}
			names := map[string]bool{}
			for _, it := range s.Snapshot.Storage {
				names[it.Name] = true
			}
			if !names["orders"] && !names["audit_events"] {
				t.Fatalf("breakdown lacks demo tables: %v", s.Snapshot.Storage)
			}
			jm, err := r.a.drv.Query(r.ctx, "PRAGMA journal_mode", 1)
			r.must(err)
			wantMode := "DELETE"
			if strings.EqualFold(jm.Rows[0][0].S, "wal") {
				wantMode = "WAL"
			}
			if got := r.kpiValue("kpi-conns"); !strings.Contains(got, "Journal Mode") || !strings.Contains(got, wantMode) {
				t.Fatalf("journal KPI %q, want mode %s", got, wantMode)
			}
			if got := r.kpiValue("kpi-cache"); !strings.Contains(got, "n/a") {
				t.Fatalf("cache KPI should be n/a for SQLite: %q", got)
			}
			if got := r.kpiValue("kpi-tps"); !strings.Contains(got, "Transactions / sec") {
				t.Fatalf("tps KPI: %q", got)
			}
		}
	}
}

// TestMonitorChartsSpanFullWidth: the chart cards fill exactly the width of the
// KPI row and the panels below, at three window sizes in both themes.
func TestMonitorChartsSpanFullWidth(t *testing.T) {
	for _, sz := range [][2]int{{1280, 800}, {1440, 960}, {1920, 1080}} {
		for _, dark := range []bool{true, false} {
			r := newRig(t, sz[0], sz[1], "")
			if !dark {
				r.tap("tb-theme")
			}
			r.openMonitor()
			for i := 0; i < 10; i++ {
				r.a.PushSyntheticSample(i)
			}
			r.must(r.a.redraw(r.ctx))
			kpis, charts := r.box("monitor-kpis"), r.box("monitor-charts")
			right := func(id string) float64 { b := r.box(id); return b.X + b.W }
			wantRight := kpis.X + kpis.W - 12 // KPI row has 12 px side padding
			for _, row := range [][2]string{{"chart-tps", "chart-conns"}, {"chart-cache", "chart-latency"}} {
				if got := right(row[1]); got < wantRight-1 || got > wantRight+1 {
					t.Errorf("%dx%d dark=%v: %s right edge %.1f, want %.1f", sz[0], sz[1], dark, row[1], got, wantRight)
				}
				l, rr := r.box(row[0]), r.box(row[1])
				if d := l.W - rr.W; d < -1 || d > 1 {
					t.Errorf("%dx%d: cards in a row differ in width: %.1f vs %.1f", sz[0], sz[1], l.W, rr.W)
				}
			}
			if got, want := r.box("storage-panel").X+r.box("storage-panel").W, wantRight; got < want-1 || got > want+1 {
				t.Errorf("%dx%d: storage panel right %.1f vs charts %.1f", sz[0], sz[1], got, want)
			}
			_ = charts
		}
	}
}

// TestGridCellsAndTabsInAccessibilityTree: the engine's tree supports a fixed
// role set (no grid, row, cell, tab); Quarry maps to list, listitem, group and
// button with explicit names so every row, cell, header and tab is reachable.
func TestGridCellsAndTabsInAccessibilityTree(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.openTableByName("users")
	tree, err := r.a.Page().Access().AccessibilityTree()
	r.must(err)
	byID := map[string]string{}
	for _, n := range tree.Nodes {
		byID[n.ElementID] = n.Name
	}
	tab := r.a.activeTabPtr()
	want := map[string]string{
		"grid-body":                   "Result grid",
		"grid-h-0":                    "Column id",
		"grid-row-0":                  "Row 1, id 1",
		"cell-0-0":                    "id: 1",
		"cell-0-1":                    "email: ",
		"tab-" + itoa(tab.id):         "Table tab users",
		"tab-" + itoa(r.a.tabs[0].id): "Query tab Query 1",
		"tree-table:main.users":       "users",
		"tab-strip":                   "Workspace tabs",
		"tab-close-" + itoa(tab.id):   "Close users",
		"tree-group:main:Tables":      "Tables",
	}
	for id, prefix := range want {
		name, ok := byID[id]
		if !ok {
			t.Errorf("%s missing from the accessibility tree", id)
		} else if !strings.HasPrefix(name, prefix) {
			t.Errorf("%s name %q, want prefix %q", id, name, prefix)
		}
	}
	// Query-result grids and monitor panels are reachable too.
	r.a.activeTab = 0
	r.must(r.a.RunSQL(r.ctx, "SELECT id, email FROM users"))
	tree, _ = r.a.Page().Access().AccessibilityTree()
	rows := 0
	for _, n := range tree.Nodes {
		if strings.HasPrefix(n.ElementID, "grid-row-") && strings.HasPrefix(n.Name, "Row ") {
			rows++
		}
	}
	if rows < 10 {
		t.Errorf("only %d named grid rows in the tree", rows)
	}
}
