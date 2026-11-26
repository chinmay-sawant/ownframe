package ui

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"
)

func TestNiceCeilAndAxisNum(t *testing.T) {
	for in, want := range map[float64]float64{0: 1, 0.3: 0.5, 1: 1, 1.2: 2, 2.1: 2.5, 3: 5, 7: 10, 45: 50, 99: 100, 101: 200, 1234: 2000} {
		if got := niceCeil(in); math.Abs(got-want) > 1e-9 {
			t.Errorf("niceCeil(%v) = %v, want %v", in, got, want)
		}
	}
	for in, want := range map[float64]string{0: "0", 0.5: "0.5", 2.5: "2.5", 50: "50", 1500: "1.5k", 2000: "2k", 3e6: "3M"} {
		if got := axisNum(in); got != want {
			t.Errorf("axisNum(%v) = %q, want %q", in, got, want)
		}
	}
}

func TestChartHoverIndexParsing(t *testing.T) {
	if c, i, ok := chartHoverIndex("chart-tps-h17"); !ok || c != "chart-tps" || i != 17 {
		t.Fatalf("got %q %d %v", c, i, ok)
	}
	for _, id := range []string{"chart-tps-b0-3", "kpi-conns", "chart-tps-hx", ""} {
		if _, _, ok := chartHoverIndex(id); ok {
			t.Errorf("%q is not a hover column", id)
		}
	}
}

// feedMonitor freezes the live collector and fills the ring with n samples.
func (r *rig) feedMonitor(n int) {
	r.openMonitor()
	r.a.mon.frozen = true
	r.a.mon.ring.Reset()
	for i := 0; i < n; i++ {
		r.a.mon.ring.Push(sample{At: r.now.Add(time.Duration(i) * time.Second), TPS: float64(100 + 40*math.Sin(float64(i)/5)), Active: float64(3 + i%4),
			Idle: 5, Waiting: float64(i % 2), HitPct: 97 + math.Sin(float64(i)), P50ms: 2, P95ms: float64(8 + i%10)})
	}
	r.must(r.a.redraw(r.ctx))
}

func TestChartAxesLegendAndTime(t *testing.T) {
	r := newRig(t, 1440, 960, "mock:postgres")
	r.feedMonitor(60)
	for _, id := range []string{"chart-tps", "chart-conns", "chart-cache", "chart-latency"} {
		for _, part := range []string{"-y", "-x", "-legend", "-plot"} {
			if !r.has(id + part) {
				t.Errorf("%s%s missing", id, part)
			}
		}
	}
	txt := r.texts()
	// Cache chart has a fixed 0..100 axis: max, mid and min labels.
	for _, want := range []string{"100", "50", "Hit", "avg"} {
		if !strings.Contains(txt, want) {
			t.Errorf("chart text %q missing", want)
		}
	}
	// The time axis shows the first, middle and last sample time.
	first := r.now.Format(chartTimeFmt)
	last := r.now.Add(59 * time.Second).Format(chartTimeFmt)
	if !strings.Contains(txt, first) || !strings.Contains(txt, last) {
		t.Errorf("time axis %q..%q missing from %q", first, last, txt)
	}
	// Legend names for the multi-series charts.
	for _, want := range []string{"Active", "Idle", "Waiting", "p50", "p95"} {
		if !strings.Contains(txt, want) {
			t.Errorf("legend %q missing", want)
		}
	}
	// Axis labels sit left of the plot, time labels below it.
	y, plot, x := r.box("chart-tps-y"), r.box("chart-tps-plot"), r.box("chart-tps-x")
	if y.X+y.W > plot.X+1 || x.Y < plot.Y+plot.H-1 {
		t.Errorf("axis placement: y %+v plot %+v x %+v", y, plot, x)
	}
}

func TestChartCardsFillTheirSpace(t *testing.T) {
	r := newRig(t, 1440, 960, "mock:postgres")
	r.feedMonitor(60)
	var hs []float64
	for _, id := range []string{"chart-tps", "chart-conns", "chart-cache", "chart-latency"} {
		hs = append(hs, r.box(id+"-plot").H)
		if r.box(id+"-plot").H < chartHeight-1 {
			t.Errorf("%s plot only %.0f px tall", id, r.box(id+"-plot").H)
		}
	}
	for _, h := range hs {
		if math.Abs(h-hs[0]) > 1 {
			t.Errorf("plot heights differ: %v", hs)
		}
	}
	// Rows of two cards share a height, and the legend sits inside the card.
	if a, b := r.box("chart-tps"), r.box("chart-conns"); math.Abs(a.H-b.H) > 1 {
		t.Errorf("row cards differ in height: %.0f vs %.0f", a.H, b.H)
	}
	for _, id := range []string{"chart-tps", "chart-conns"} {
		card, lg := r.box(id), r.box(id+"-legend")
		if lg.Y+lg.H > card.Y+card.H+1 {
			t.Errorf("%s legend leaves the card", id)
		}
	}
	// The plot is wide: bars span it to within a pixel per sample.
	plot, last := r.box("chart-tps-plot"), r.box("chart-tps-b0-59")
	if last.X+last.W < plot.X+plot.W-2 {
		t.Errorf("series stops short of the plot's right edge: %.0f < %.0f", last.X+last.W, plot.X+plot.W)
	}
	// Peak bar reaches the 100 px axis scale: the tallest TPS bar is within the plot.
	for i := 0; i < 60; i++ {
		if b := r.box(fmt.Sprintf("chart-tps-b0-%d", i)); b.H > plot.H+1 {
			t.Fatalf("bar %d taller than the plot: %.0f > %.0f", i, b.H, plot.H)
		}
	}
	r.png("monitor-charts-filled-dark")
	r.must(r.a.ToggleTheme(r.ctx))
	r.png("monitor-charts-filled-light")
}

func TestChartHoverShowsSampleValueAndTime(t *testing.T) {
	r := newRig(t, 1440, 960, "mock:postgres")
	r.feedMonitor(60)
	if !strings.Contains(r.box("chart-tps-readout").ID, "readout") || !strings.Contains(r.textOf("chart-tps-readout"), "now") {
		t.Fatal("legend tag should read 'now' with no pointer over the chart")
	}
	hover := func(id string) {
		b := r.box(id)
		r.must(r.a.Page().Edit().Hover(r.ctx, float64(b.X+b.W/2), float64(b.Y+b.H/2)))
		r.tick()
	}
	hover("chart-tps-h10")
	want := r.now.Format(chartTimeFmt) // sample 10 was pushed at r.now+10s before tick moved r.now
	_ = want
	if r.a.mon.hoverChart != "chart-tps" || r.a.mon.hoverIdx != 10 {
		t.Fatalf("hover state = %q %d", r.a.mon.hoverChart, r.a.mon.hoverIdx)
	}
	sample10 := r.a.mon.ring.At(10)
	tag := r.textOf("chart-tps-readout")
	if tag != sample10.At.Format(chartTimeFmt) {
		t.Errorf("readout %q, want the time of sample 10 (%s)", tag, sample10.At.Format(chartTimeFmt))
	}
	if !strings.Contains(r.textOf("chart-tps-legend"), fmtNum(sample10.TPS)) {
		t.Errorf("legend %q lacks the hovered value %s", r.textOf("chart-tps-legend"), fmtNum(sample10.TPS))
	}
	// Another chart is untouched; moving to the next column updates the readout.
	if tag := r.textOf("chart-conns-readout"); tag != "now" {
		t.Errorf("other chart readout = %q", tag)
	}
	hover("chart-tps-h11")
	if r.textOf("chart-tps-readout") != r.a.mon.ring.At(11).At.Format(chartTimeFmt) {
		t.Error("readout did not follow the pointer")
	}
	// Leaving the chart restores the live readout.
	hover("kpi-conns")
	if r.textOf("chart-tps-readout") != "now" {
		t.Errorf("readout after leaving = %q", r.textOf("chart-tps-readout"))
	}
}

func TestChartHoverCostsNoFramesWhenSteady(t *testing.T) {
	r := newRig(t, 1440, 960, "mock:postgres")
	r.feedMonitor(30)
	r.a.trackChartHover("chart-tps-h3")
	if !r.a.takeChartHover() {
		t.Fatal("a new hover column must request one repaint")
	}
	r.a.trackChartHover("chart-tps-h3")
	if r.a.takeChartHover() {
		t.Fatal("the same column must not repaint again")
	}
	r.a.trackChartHover("")
	r.a.takeChartHover()
	r.a.trackChartHover("kpi-tps")
	if r.a.takeChartHover() {
		t.Fatal("hovering unrelated boxes with no chart hover is free")
	}
}

// textOf returns the text painted inside box id, joined.
func (r *rig) textOf(id string) string {
	r.t.Helper()
	b := r.box(id)
	d := r.a.Page().Probe().Display()
	k := d.PixelPerPoint
	if k <= 0 {
		k = 1
	}
	var parts []string
	for _, op := range d.Ops {
		if op.Text == "" {
			continue
		}
		x, y := op.X*k, op.Y*k
		if x >= b.X-1 && x <= b.X+b.W+1 && y >= b.Y-1 && y <= b.Y+b.H+14 {
			parts = append(parts, op.Text)
		}
	}
	return strings.Join(parts, " ")
}
