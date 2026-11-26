package ui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/chinmay-sawant/ownframe/ui"

	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/driver"
	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/monitor"
)

const monitorHistory = 120

// sample is one derived point of the live charts.
type sample struct {
	At                    time.Time
	TPS                   float64
	Active, Idle, Waiting float64
	HitPct                float64
	SizeBytes             int64
	P50ms, P95ms          float64
	JournalWAL            bool
	WaitingKnown          bool
	Snapshot              driver.MetricSnapshot
	HasSnapshot           bool
}

// monitorState is the live telemetry state behind the Monitor tab.
type monitorState struct {
	coll      *monitor.Collector[driver.MetricSnapshot]
	ring      *monitor.MetricRing[sample]
	rate      monitor.RateTracker
	pollEvery time.Duration // zero means paused
	frozen    bool
	filter    string // all | active | waiting | idletx
	err       string
	seen      map[string]bool
	polls     int

	// Chart hover: the column under the pointer, applied on the next tick.
	hoverChart string
	hoverIdx   int
	hoverDirty bool
}

func (m *monitorState) stopOnClose() {
	if m.coll != nil {
		m.coll.Stop()
		m.coll = nil
	}
}

func (a *App) hasMonitorTab() bool { return a.findTab(tabMonitor, "", "") >= 0 }

// startMonitor (re)starts the collector for the active driver. It does nothing
// without a connection.
func (a *App) startMonitor() {
	if a.drv == nil {
		return
	}
	a.stopMonitor()
	if a.mon.ring == nil {
		a.mon.ring = monitor.NewRing[sample](monitorHistory)
	}
	a.mon.ring.Reset()
	a.mon.rate.Reset()
	a.mon.seen = map[string]bool{}
	if a.mon.filter == "" {
		a.mon.filter = "all"
	}
	drv := a.drv
	c := monitor.NewCollector(func(ctx context.Context) (driver.MetricSnapshot, error) {
		return drv.CollectMetrics(ctx)
	}, a.mon.pollEvery)
	a.mon.coll = c
	c.Start(a.ctx)
}

func (a *App) stopMonitor() { a.mon.stopOnClose() }

// monitorVisible reports whether the Monitor tab is the active tab.
func (a *App) monitorVisible() bool {
	t := a.activeTabPtr()
	return t != nil && t.kind == tabMonitor
}

// drainMonitor moves collected snapshots into the ring. It reports whether the
// visible Monitor tab needs a repaint; with nothing pending it returns false so
// an idle or paused monitor costs no frames.
func (a *App) drainMonitor() bool {
	c := a.mon.coll
	if c == nil || !c.Pending() {
		return false
	}
	results := c.Drain()
	if len(results) == 0 {
		return false
	}
	changed := false
	for _, r := range results {
		if r.Err != nil {
			a.mon.err = r.Err.Error()
			changed = true
			continue
		}
		a.mon.err = ""
		a.mon.polls++
		a.observeServerSlow(r.Sample)
		if a.mon.frozen {
			continue
		}
		a.mon.ring.Push(a.deriveSample(r))
		changed = true
	}
	return changed && a.monitorVisible()
}

func (a *App) deriveSample(r monitor.Result[driver.MetricSnapshot]) sample {
	s := r.Sample
	smp := sample{
		At: r.At, Active: float64(s.ActiveConns), Idle: float64(s.IdleConns), Waiting: float64(s.WaitingLocks),
		HitPct: max(0, s.CacheHitRatio) * chartHundred, SizeBytes: s.DBSizeBytes, Snapshot: s, HasSnapshot: true,
		P50ms: float64(a.lat.P50().Microseconds()) / msPerSecond, P95ms: float64(a.lat.P95().Microseconds()) / msPerSecond,
		TPS: s.TPS,
	}
	if total := s.CommitTotal + s.RollbackTotal; total > 0 {
		if rate, ok := a.mon.rate.Update(r.At, uint64(total)); ok {
			smp.TPS = rate
		}
	}
	_, smp.JournalWAL = s.Extra["wal_pages"]
	return smp
}

// observeServerSlow copies long server-side statements into the slow log once.
func (a *App) observeServerSlow(s driver.MetricSnapshot) {
	for _, q := range s.ActiveQueries {
		if q.State != "active" || q.Query == "" {
			continue
		}
		key := fmt.Sprintf("%d|%s", q.PID, q.Query)
		if a.mon.seen[key] {
			continue
		}
		if a.slow.Observe(monitor.SlowQuery{At: s.CapturedAt, Duration: q.Duration, SQL: q.Query, Source: "server"}) {
			a.mon.seen[key] = true
		}
	}
}

func (a *App) setPollEvery(ctx context.Context, d time.Duration) error {
	a.mon.pollEvery = d
	if a.mon.coll != nil {
		a.mon.coll.SetInterval(d)
	}
	return a.redraw(ctx)
}

// ---- view ----

func (a *App) monitorView(t *workTab) *ui.Element {
	view := ui.Div().ID("monitor-tab").Class("qr-tab-body qr-monitor")
	if a.drv == nil {
		return view.Children(ui.Div().ID("monitor-empty").Class("qr-empty").Text("Connect to a database to monitor it."))
	}
	// The content scrolls by an app-managed offset (see onWheel): nested
	// overflow scrolling is outside the engine's verified contract.
	content := []*ui.Element{a.monitorHeader(), a.monitorKPIs()}
	if a.mon.err != "" {
		content = append(content, ui.Div().ID("monitor-error").Class("qr-note qr-note-err qr-pad").Text("Collector error: "+a.mon.err))
	}
	content = append(content, a.monitorCharts(), a.storagePanel(), a.activeQueriesPanel(), a.slowLogPanel())
	return view.Children(a.pane("monitor", "Monitor scroll", "", content...))
}

func (a *App) monitorHeader() *ui.Element {
	seg := ui.Div().Class("qr-seg").Attr("role", ariaRole("group")).Attr("aria-label", "Poll interval")
	for _, o := range []struct {
		id, label string
		d         time.Duration
	}{{"1s", "1s", time.Second}, {"2s", "2s", 2 * time.Second}, {"5s", "5s", 5 * time.Second}, {"pause", "Pause", 0}} {
		o := o
		cls := "qr-seg-item"
		if a.mon.pollEvery == o.d {
			cls += " qr-seg-on"
		}
		seg.Children(ui.Div().ID("mon-poll-"+o.id).Class(cls).Text(o.label).Attr("role", "button").
			OnClick(func(ctx context.Context) error { return a.setPollEvery(ctx, o.d) }))
	}
	freeze := "Freeze Snapshot"
	if a.mon.frozen {
		freeze = "Resume"
	}
	return ui.Div().ID("monitor-head").Class("qr-toolbar").Children(
		ui.Span().Class("qr-sidebar-title").Text("Live Monitor"),
		ui.Span().Class("qr-muted").Text("Poll"), seg,
		a.tbButton("mon-refresh", "Refresh Now", "qr-btn", false, func(ctx context.Context) error {
			if a.mon.coll == nil {
				a.startMonitor()
			}
			a.mon.coll.RefreshNow()
			return a.redraw(ctx)
		}),
		a.tbButton("mon-freeze", freeze, "qr-btn", false, func(ctx context.Context) error {
			a.mon.frozen = !a.mon.frozen
			return a.redraw(ctx)
		}),
	)
}

func (a *App) latestSample() (sample, bool) {
	if a.mon.ring == nil {
		return sample{}, false
	}
	return a.mon.ring.Last()
}

func kpiCard(id, title, value, sub, tone string) *ui.Element {
	card := ""
	switch tone {
	case "qr-ok":
		card = " qr-kpi-ok-card"
	case "qr-warn":
		card = " qr-kpi-warn-card"
	case "qr-err":
		card = " qr-kpi-err-card"
	}
	return ui.Div().ID(id).Class("qr-kpi"+card).Attr("role", ariaRole("group")).Attr("aria-label", title+" "+value).Children(
		ui.Div().Class("qr-kpi-title qr-muted").Text(title),
		ui.Div().Class("qr-kpi-value "+tone).Text(value),
		ui.Div().Class("qr-kpi-sub qr-muted").Text(sub),
	)
}

func (a *App) monitorKPIs() *ui.Element {
	s, ok := a.latestSample()
	row := ui.Div().ID("monitor-kpis").Class("qr-kpis")
	if !ok {
		return row.Children(
			kpiCard("kpi-conns", "Active Connections", "-", "waiting for first sample", ""),
			kpiCard("kpi-tps", "Transactions / sec", "-", "", ""),
			kpiCard("kpi-cache", "Cache Hit Ratio", "-", "", ""),
			kpiCard("kpi-size", "Database Size & Latency p95", "-", "", ""),
		)
	}
	first := kpiCard("kpi-conns", "Active Connections", fmt.Sprintf("%.0f", s.Active),
		fmt.Sprintf("%.0f idle, %.0f waiting", s.Idle, s.Waiting), "qr-ok")
	if a.cfg.Engine == driver.EngineSQLite {
		mode := "DELETE"
		if s.JournalWAL {
			mode = "WAL"
		}
		first = kpiCard("kpi-conns", "Journal Mode", mode, fmt.Sprintf("%.0f open connections", s.Active), "qr-ok")
	}
	hit := "qr-ok"
	switch {
	case s.HasSnapshot && s.Snapshot.CacheHitRatio < 0:
		hit = "" // no cache ratio on this engine: neutral, not an alarm
	case s.HitPct < hitBadBelow:
		hit = "qr-err"
	case s.HitPct < hitWarnBelow:
		hit = "qr-warn"
	}
	lat := "qr-ok"
	if s.P95ms > latBadMs {
		lat = "qr-err"
	} else if s.P95ms > latWarnMs {
		lat = "qr-warn"
	}
	return row.Children(
		first,
		kpiCard("kpi-tps", "Transactions / sec", fmtNum(s.TPS), "committed + rolled back", "qr-ok"),
		kpiCard("kpi-cache", "Cache Hit Ratio", cacheLabel(s), "blocks served from cache", hit),
		kpiCard("kpi-size", "Database Size & Latency p95", humanBytes(int(s.SizeBytes)), fmt.Sprintf("p95 %.1f ms, p50 %.1f ms", s.P95ms, s.P50ms), lat),
	)
}

func (a *App) monitorCharts() *ui.Element {
	wrap := ui.Div().ID("monitor-charts").Class("qr-charts")
	if a.mon.ring == nil {
		return wrap
	}
	ser := func(f func(sample) float64) []float64 { return monitor.Series(a.mon.ring, f) }
	times := make([]time.Time, a.mon.ring.Len())
	for i := range times {
		times[i] = a.mon.ring.At(i).At
	}
	tps := chartCard(chartSpec{ID: "chart-tps", Times: times, Hover: a.chartHoverFor("chart-tps"), Title: "Throughput (TPS / QPS)", Unit: "/s",
		Series: []chartSeries{{Name: "TPS", Values: ser(func(s sample) float64 { return s.TPS }), Class: "qr-s1"}}})
	conns := chartCard(chartSpec{ID: "chart-conns", Times: times, Hover: a.chartHoverFor("chart-conns"), Title: "Connections (Active vs Idle vs Waiting)",
		Series: []chartSeries{
			{Name: "Active", Values: ser(func(s sample) float64 { return s.Active }), Class: "qr-s1"},
			{Name: "Idle", Values: ser(func(s sample) float64 { return s.Idle }), Class: "qr-s2"},
			{Name: "Waiting", Values: ser(func(s sample) float64 { return s.Waiting }), Class: "qr-s3"},
		}})
	cache := a.cacheOrSizeCard(times, ser)
	lat := chartCard(chartSpec{ID: "chart-latency", Times: times, Hover: a.chartHoverFor("chart-latency"), Title: "Query Latency (p50 / p95 ms)", Unit: "ms",
		Series: []chartSeries{
			{Name: "p50", Values: ser(func(s sample) float64 { return s.P50ms }), Class: "qr-s1"},
			{Name: "p95", Values: ser(func(s sample) float64 { return s.P95ms }), Class: "qr-s3"},
		}})
	// Two explicit rows of two equal cards: they span exactly the width of
	// the KPI row and the panels, at any window size.
	return wrap.Children(
		ui.Div().Class("qr-chart-row").Children(tps, conns),
		ui.Div().Class("qr-chart-row").Children(cache, lat),
	)
}

// cacheOrSizeCard is the cache-hit card where the engine reports a ratio. On
// SQLite, which cannot, the slot shows database size over time instead of a
// flat 0% series.
func (a *App) cacheOrSizeCard(times []time.Time, ser func(func(sample) float64) []float64) *ui.Element {
	if s, ok := a.latestSample(); ok && s.HasSnapshot && s.Snapshot.CacheHitRatio < 0 {
		return chartCard(chartSpec{ID: "chart-cache", Times: times, Hover: a.chartHoverFor("chart-cache"), Title: "Database Size (MB)", Unit: " MB",
			Series: []chartSeries{{Name: "Size", Values: ser(func(s sample) float64 { return float64(s.SizeBytes) / (bytesPerKB * bytesPerKB) }), Class: "qr-s1"}}})
	}
	return chartCard(chartSpec{ID: "chart-cache", Times: times, Hover: a.chartHoverFor("chart-cache"), Title: "Cache Hit Ratio (%)", Unit: "%", Max: chartHundred,
		Series: []chartSeries{{Name: "Hit", Values: ser(func(s sample) float64 { return s.HitPct }), Class: "qr-s1"}}})
}

// storagePanel lists the largest tables with proportion bars.
func (a *App) storagePanel() *ui.Element {
	panel := ui.Div().ID("storage-panel").Class("qr-panel").Children(ui.Div().Class("qr-panel-title").Text("Storage and table breakdown"))
	s, ok := a.latestSample()
	if !ok {
		return panel.Children(ui.Div().Class("qr-note qr-pad").Text("Waiting for the first sample."))
	}
	items := append([]driver.StorageItem(nil), s.Snapshot.Storage...)
	sort.SliceStable(items, func(i, j int) bool { return items[i].Bytes > items[j].Bytes })
	var top int64 = 1
	if len(items) > 0 {
		top = max(1, items[0].Bytes)
	}
	for i, it := range items {
		if i >= storageRowsShown {
			break
		}
		it := it
		pct := float64(it.Bytes) / float64(top) * chartHundred
		panel.Children(ui.Div().ID(fmt.Sprintf("storage-row-%d", i)).Class("qr-storage-row").Children(
			ui.Span().Class("qr-storage-name qr-mono").Text(it.Name),
			ui.Div().Class("qr-bar-track").Children(ui.Div().ID(fmt.Sprintf("storage-bar-%d", i)).Class("qr-bar").Width(ui.Percent(max(1, pct)))),
			ui.Span().Class("qr-muted qr-storage-size").Text(humanBytes(int(it.Bytes))),
			ui.Button("Data").ID(fmt.Sprintf("storage-data-%d", i)).Class("qr-btn qr-btn-ghost").Attr("type", "button").Attr("aria-label", "Open data for "+it.Name).
				OnClick(func(ctx context.Context) error { return a.openByName(ctx, it.Name, false) }),
			ui.Button("Structure").ID(fmt.Sprintf("storage-struct-%d", i)).Class("qr-btn qr-btn-ghost").Attr("type", "button").Attr("aria-label", "Open structure for "+it.Name).
				OnClick(func(ctx context.Context) error { return a.openByName(ctx, it.Name, true) }),
		))
	}
	return panel
}

// openByName finds a table node by name and opens its data or structure tab.
func (a *App) openByName(ctx context.Context, name string, structure bool) error {
	for _, s := range a.schema {
		for _, g := range s.Children {
			for _, n := range g.Children {
				if (n.Kind == driver.NodeTable || n.Kind == driver.NodeView) && (n.Name == name || n.QualifiedName() == name) {
					if structure {
						return a.openStructure(ctx, n)
					}
					return a.openTable(ctx, n)
				}
			}
		}
	}
	a.notify("No table named " + name)
	return a.redraw(ctx)
}

func queryMatches(filter string, q driver.ActiveQuery) bool {
	switch filter {
	case "active":
		return q.State == "active"
	case "waiting":
		return q.Blocked || q.WaitEvent != ""
	case "idletx":
		return strings.Contains(q.State, "idle in transaction")
	}
	return true
}

func (a *App) activeQueriesPanel() *ui.Element {
	panel := ui.Div().ID("active-panel").Class("qr-panel").Children(ui.Div().Class("qr-panel-title").Text("Active queries and locks"))
	seg := ui.Div().Class("qr-seg").Attr("role", ariaRole("group")).Attr("aria-label", "State filter")
	for _, o := range [][2]string{{"all", "All"}, {"active", "Active"}, {"waiting", "Waiting"}, {"idletx", "Idle in Tx"}} {
		o := o
		cls := "qr-seg-item"
		if a.mon.filter == o[0] {
			cls += " qr-seg-on"
		}
		seg.Children(ui.Div().ID("aq-filter-"+o[0]).Class(cls).Text(o[1]).Attr("role", "button").
			OnClick(func(ctx context.Context) error { a.mon.filter = o[0]; return a.redraw(ctx) }))
	}
	panel.Children(ui.Div().Class("qr-toolbar").Children(seg))
	s, ok := a.latestSample()
	if !ok {
		return panel.Children(ui.Div().Class("qr-note qr-pad").Text("Waiting for the first sample."))
	}
	n := 0
	for _, q := range s.Snapshot.ActiveQueries {
		if !queryMatches(a.mon.filter, q) {
			continue
		}
		q := q
		row := ui.Div().ID(fmt.Sprintf("aq-row-%d", n)).Class("qr-aq-row").Children(
			ui.Span().Class("qr-aq-pid qr-mono").Text(fmt.Sprint(q.PID)),
			ui.Span().Class("qr-aq-user").Text(q.User),
			ui.Span().Class("qr-badge").Text(q.State),
			ui.Span().Class("qr-badge "+durClass(q.Duration)).Text(fmtDur(q.Duration)),
			ui.Span().Class("qr-muted qr-aq-wait").Text(q.WaitEvent),
			ui.Span().Class("qr-aq-query qr-mono").Text(oneLine(q.Query)),
		)
		off := a.cfg.ReadOnly
		row.Children(
			a.tbButton(fmt.Sprintf("aq-cancel-%d", n), "Cancel", "qr-btn", off, func(ctx context.Context) error { return a.confirmBackend(ctx, q, false) }),
			a.tbButton(fmt.Sprintf("aq-term-%d", n), "Terminate", "qr-btn qr-btn-danger", off, func(ctx context.Context) error { return a.confirmBackend(ctx, q, true) }),
		)
		panel.Children(row)
		n++
	}
	if n == 0 {
		panel.Children(ui.Div().ID("aq-empty").Class("qr-note qr-pad").Text("No queries match this filter."))
	}
	return panel
}

func durClass(d time.Duration) string {
	switch {
	case d > time.Second:
		return "qr-badge-err"
	case d > 100*time.Millisecond:
		return "qr-badge-warn"
	}
	return "qr-badge-ok"
}

type terminator interface {
	TerminateQuery(ctx context.Context, pid int) error
}

// confirmBackend asks before cancelling or terminating a backend.
func (a *App) confirmBackend(ctx context.Context, q driver.ActiveQuery, terminate bool) error {
	verb := "Cancel"
	if terminate {
		verb = "Terminate"
	}
	return a.askConfirm(ctx, &confirmState{
		Title: fmt.Sprintf("%s backend %d?", verb, q.PID), Label: verb,
		Body: oneLine(q.Query),
		OK: func(ctx context.Context) error {
			var err error
			if terminate {
				if tm, ok := a.drv.(terminator); ok {
					err = tm.TerminateQuery(ctx, q.PID)
				} else {
					err = driver.ErrUnsupported
				}
			} else {
				err = a.drv.CancelQuery(ctx, q.PID)
			}
			if err != nil {
				a.notify(err.Error())
			} else {
				a.notify(fmt.Sprintf("%s requested for backend %d", verb, q.PID))
			}
			if a.mon.coll != nil {
				a.mon.coll.RefreshNow()
			}
			return a.redraw(ctx)
		},
	})
}

func (a *App) slowLogPanel() *ui.Element {
	panel := ui.Div().ID("slow-panel").Class("qr-panel").Children(
		ui.Div().Class("qr-panel-title").Text(fmt.Sprintf("Slow query log (>= %s)", fmtDur(a.slow.Threshold()))))
	entries := a.slow.Entries()
	if len(entries) == 0 {
		return panel.Children(ui.Div().ID("slow-empty").Class("qr-note qr-pad").Text("No slow queries captured yet."))
	}
	for i, e := range entries {
		if i >= slowRowsShown {
			break
		}
		e := e
		panel.Children(ui.Div().ID(fmt.Sprintf("slow-row-%d", i)).Class("qr-aq-row").Children(
			ui.Span().Class("qr-muted").Text(e.At.Format("15:04:05")),
			ui.Span().Class("qr-badge "+durClass(e.Duration)).Text(fmtDur(e.Duration)),
			ui.Span().Class("qr-muted").Text(fmt.Sprintf("%d rows", e.Rows)),
			ui.Span().Class("qr-badge").Text(e.Source),
			ui.Span().Class("qr-aq-query qr-mono").Text(oneLine(e.SQL)),
			a.tbButton(fmt.Sprintf("slow-open-%d", i), "Open in Editor & Explain", "qr-btn", false, func(ctx context.Context) error {
				t := a.newQueryTab(e.SQL)
				_ = t
				return a.startExplain(ctx, false)
			}),
		))
	}
	return panel
}

// cacheLabel shows n/a for engines that report no cache ratio (negative).
func cacheLabel(s sample) string {
	if s.HasSnapshot && s.Snapshot.CacheHitRatio < 0 {
		return "n/a"
	}
	return fmt.Sprintf("%.1f%%", s.HitPct)
}
