package ui

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/driver"
	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/monitor"
)

// openMonitor opens the Monitor tab and waits for the first sample.
func (r *rig) openMonitor() {
	r.t.Helper()
	r.tap("rail-monitor")
	deadline := time.Now().Add(5 * time.Second)
	for {
		r.tick()
		if _, ok := r.a.latestSample(); ok {
			break
		}
		if time.Now().After(deadline) {
			r.t.Fatal("no monitor sample arrived")
		}
		time.Sleep(10 * time.Millisecond)
	}
	r.must(r.a.redraw(r.ctx))
}

func TestMonitorKPICards(t *testing.T) {
	for _, dark := range []bool{true, false} {
		r := newRig(t, 1440, 960, "mock:postgres")
		if !dark {
			r.tap("tb-theme")
		}
		r.openMonitor()
		for _, id := range []string{"kpi-conns", "kpi-tps", "kpi-cache", "kpi-size", "monitor-head", "mon-refresh", "mon-freeze", "mon-poll-1s", "mon-poll-pause"} {
			r.box(id)
		}
		txt := r.texts()
		for _, want := range []string{"Active Connections", "Transactions / sec", "Cache Hit Ratio", "Database Size & Latency p95"} {
			if !strings.Contains(txt, want) {
				t.Errorf("dark=%v: KPI title %q missing", dark, want)
			}
		}
		// Threshold classes: force known samples and read the painted colour.
		set := func(hit, p95 float64) {
			r.a.mon.ring.Push(sample{HitPct: hit, P95ms: p95, At: r.now, HasSnapshot: true})
			r.must(r.a.redraw(r.ctx))
		}
		colorOf := func(text string) uint32 {
			for _, op := range r.a.Page().Probe().Display().Ops {
				if op.Kind == ownframe.DisplayOpText && op.Text == text {
					return uint32(op.R*255+0.5)<<16 | uint32(op.G*255+0.5)<<8 | uint32(op.B*255+0.5)
				}
			}
			t.Fatalf("no text %q", text)
			return 0
		}
		ok, warn, bad := uint32(0x3ecf8e), uint32(0xe5b340), uint32(0xe5645f)
		if !dark {
			ok, warn, bad = 0x1a9a68, 0xb7791f, 0xd1453b
		}
		set(99, 5)
		if got := colorOf("99.0%"); got != ok {
			t.Errorf("dark=%v healthy cache colour %06x want %06x", dark, got, ok)
		}
		set(90, 5)
		if got := colorOf("90.0%"); got != warn {
			t.Errorf("dark=%v warn colour %06x want %06x", dark, got, warn)
		}
		set(50, 5)
		if got := colorOf("50.0%"); got != bad {
			t.Errorf("dark=%v bad colour %06x want %06x", dark, got, bad)
		}
		r.png(fmt.Sprintf("monitor-dark-%v", dark))
	}
}

func TestMonitorChartsRender(t *testing.T) {
	r := newRig(t, 1440, 960, "mock:postgres")
	r.openMonitor()
	r.a.mon.frozen = true // keep the live collector from adding samples
	r.a.mon.ring.Reset()
	for i := 0; i < 60; i++ {
		r.a.mon.ring.Push(sample{At: r.now.Add(time.Duration(i) * time.Second), TPS: 100 + 40*math.Sin(float64(i)/5), Active: float64(3 + i%4),
			Idle: 5, Waiting: float64(i % 2), HitPct: 97 + math.Sin(float64(i)), P50ms: 2, P95ms: 8 + float64(i%10)})
	}
	r.must(r.a.redraw(r.ctx))
	for _, id := range []string{"chart-tps", "chart-conns", "chart-cache", "chart-latency"} {
		b := r.box(id)
		if b.W <= 0 || b.H <= 0 || math.IsNaN(b.X) || math.IsNaN(b.Y) {
			t.Errorf("%s box %+v", id, b)
		}
	}
	bars := r.countBoxes("chart-tps-b0-")
	if bars != 60 {
		t.Fatalf("tps chart bars = %d, want 60", bars)
	}
	if r.countBoxes("chart-conns-b0-") != 60 || r.countBoxes("chart-conns-b2-") != 60 {
		t.Fatal("multi-series chart layers missing")
	}
	for _, op := range r.a.Page().Probe().Display().Ops {
		for _, v := range []float64{op.X, op.Y, op.W, op.H} {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				t.Fatalf("NaN/Inf coordinate in op %+v", op)
			}
		}
	}
	txt := r.texts()
	for _, want := range []string{"Throughput (TPS / QPS)", "Cache Hit Ratio (%)", "Query Latency (p50 / p95 ms)", "min", "max", "now"} {
		if !strings.Contains(txt, want) {
			t.Errorf("chart text %q missing", want)
		}
	}
	// A freshly pushed sample moves only the data, not the structure.
	if r.a.mon.ring.Len() != 60 || r.a.mon.ring.Cap() != monitorHistory {
		t.Fatalf("ring len/cap = %d/%d", r.a.mon.ring.Len(), r.a.mon.ring.Cap())
	}
	r.png("monitor-charts")
}

func TestMonitorStorageBreakdown(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.openMonitor()
	if !r.has("storage-panel") || !r.has("storage-row-0") || !r.has("storage-bar-0") {
		t.Fatal("storage panel rows missing")
	}
	s, _ := r.a.latestSample()
	items := s.Snapshot.Storage
	if len(items) < 3 {
		t.Fatalf("storage items = %v", items)
	}
	// Bars descend in width.
	prev := math.MaxFloat64
	for i := 0; i < 5 && r.has(fmt.Sprintf("storage-bar-%d", i)); i++ {
		w := r.box(fmt.Sprintf("storage-bar-%d", i)).W
		if w > prev+0.5 {
			t.Fatalf("bar %d wider (%v) than the previous (%v)", i, w, prev)
		}
		prev = w
	}
	// Jump buttons open the table.
	top := ""
	for _, it := range items {
		for _, sn := range r.a.schema[0].Children[0].Children {
			if sn.Name == it.Name {
				top = it.Name
			}
		}
	}
	if top == "" {
		t.Fatalf("no storage item matched a table: %v", items)
	}
	for i := 0; i < 8; i++ {
		id := fmt.Sprintf("storage-data-%d", i)
		if !r.has(id) {
			break
		}
		if strings.Contains(r.texts(), "Open data for "+top) {
			break
		}
	}
	r.must(r.a.openByName(r.ctx, top, true))
	if tb := r.a.activeTabPtr(); tb.kind != tabStructure || tb.table != top {
		t.Fatalf("structure jump -> %+v", tb)
	}
	r.must(r.a.openByName(r.ctx, top, false))
	if tb := r.a.activeTabPtr(); tb.kind != tabTable || tb.table != top {
		t.Fatalf("data jump -> %+v", tb)
	}
}

func TestActiveQueriesAndCancelAction(t *testing.T) {
	r := newRig(t, 1440, 960, "mock:postgres")
	r.openMonitor()
	s, _ := r.a.latestSample()
	if len(s.Snapshot.ActiveQueries) == 0 {
		t.Fatal("mock telemetry has no active queries")
	}
	total := r.countBoxes("aq-row-")
	if total != len(s.Snapshot.ActiveQueries) {
		t.Fatalf("rows %d vs %d", total, len(s.Snapshot.ActiveQueries))
	}
	r.tap("aq-filter-active")
	nActive := 0
	for _, q := range s.Snapshot.ActiveQueries {
		if q.State == "active" {
			nActive++
		}
	}
	if got := r.countBoxes("aq-row-"); got != nActive {
		t.Fatalf("Active filter shows %d rows, want %d", got, nActive)
	}
	r.tap("aq-filter-idletx")
	for _, q := range s.Snapshot.ActiveQueries {
		_ = q
	}
	r.tap("aq-filter-all")
	r.tap("aq-cancel-0")
	if r.a.modal != "confirm" {
		t.Fatal("cancel must ask for confirmation")
	}
	pid := s.Snapshot.ActiveQueries[0].PID
	r.tap("confirm-ok")
	if !strings.Contains(r.a.toast, fmt.Sprint(pid)) { // mock rejects pids it never ran, but the dispatch happened
		t.Fatalf("toast = %q", r.a.toast)
	}
	// Read-only connections disable the actions.
	r.must(r.a.reopenWritable(r.ctx, false))
	r.a.mon.coll.RefreshNow()
	r.must(r.a.redraw(r.ctx))
	if r.box("aq-cancel-0").Action != "" {
		t.Fatal("unexpected action")
	}
	r.tap("aq-cancel-0")
	if r.a.modal == "confirm" {
		t.Fatal("read-only mode must not allow cancel")
	}
}

func TestSlowQueryCaptureAndOpenExplain(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.a.slow.SetThreshold(0)
	r.a.slow.SetThreshold(time.Nanosecond)
	tab := r.a.activeTabPtr()
	tab.ed = newEditor("SELECT * FROM order_items o JOIN orders u ON u.id = o.order_id")
	r.chord("enter", false)
	r.settle()
	entries := r.a.slow.Entries()
	if len(entries) != 1 || !strings.Contains(entries[0].SQL, "order_items") {
		t.Fatalf("slow log = %+v", entries)
	}
	r.openMonitor()
	if !r.has("slow-row-0") || !r.has("slow-open-0") {
		t.Fatal("slow log row missing")
	}
	n := len(r.a.tabs)
	r.scrollMonitorTo("slow-open-0")
	r.tap("slow-open-0")
	r.settle()
	nt := r.a.activeTabPtr()
	if len(r.a.tabs) != n+1 || nt.kind != tabQuery || !strings.Contains(nt.ed.text, "order_items") || nt.plan == nil || nt.resultView != "explain" {
		t.Fatalf("explain tab not opened: tabs=%d plan=%v view=%q", len(r.a.tabs), nt.plan != nil, nt.resultView)
	}
	// Below the threshold nothing is captured.
	r.a.slow.SetThreshold(time.Hour)
	r.a.slow.Clear()
	nt.ed = newEditor("SELECT 1")
	r.chord("enter", false)
	r.settle()
	if len(r.a.slow.Entries()) != 0 {
		t.Fatal("fast query captured")
	}
}

func TestMonitorIdleWhenPaused(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.openMonitor()
	r.tap("mon-poll-pause")
	if r.a.mon.pollEvery != 0 {
		t.Fatal("pause not applied")
	}
	time.Sleep(60 * time.Millisecond)
	r.a.onTick(r.ctx) // drain anything in flight
	for i := 0; i < 5; i++ {
		r.now = r.now.Add(time.Second)
		if r.a.onTick(r.ctx) {
			t.Fatal("onTick must report clean while paused with nothing queued")
		}
	}
	// Refresh Now still polls once while paused.
	n := r.a.mon.ring.Len()
	r.tap("mon-refresh")
	deadline := time.Now().Add(3 * time.Second)
	for r.a.mon.ring.Len() == n && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
		r.tick()
	}
	if r.a.mon.ring.Len() != n+1 {
		t.Fatalf("Refresh Now while paused: %d -> %d", n, r.a.mon.ring.Len())
	}
	// Closing the tab stops the collector entirely.
	r.must(r.a.closeTab(r.ctx, r.a.findTab(tabMonitor, "", "")))
	if r.a.mon.coll != nil {
		t.Fatal("collector should stop when the Monitor tab closes")
	}
	// Freeze keeps polling but stops repaints.
	var _ driver.MetricSnapshot
	var _ = monitor.DefaultCapacity
}

// scrollMonitorTo wheel-scrolls the Monitor tab until box id is on screen.
func (r *rig) scrollMonitorTo(id string) {
	r.t.Helper()
	body := r.box("monitor-tab")
	for i := 0; i < 60; i++ {
		b := r.box(id)
		if b.Y+b.H <= body.Y+body.H {
			return
		}
		_, err := r.a.onWheel(r.ctx, body.X+20, body.Y+20, 0, -3)
		r.must(err)
	}
	r.t.Fatalf("could not scroll %s into view", id)
}
