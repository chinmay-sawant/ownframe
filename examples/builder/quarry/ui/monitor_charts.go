package ui

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/ui"
)

// chartSeries is one plotted line of a live chart. The first series is drawn
// as a filled area under a 2 px cap line; later series are cap lines only, so
// overlapping series stay readable. The builder has no SVG tag, so every
// sample is one div column whose height tracks the value.
type chartSeries struct {
	Name   string
	Values []float64
	Class  string // CSS class naming the series colour token
}

// chartSpec describes one live chart card.
type chartSpec struct {
	ID     string
	Title  string
	Unit   string
	Series []chartSeries
	Times  []time.Time // one per sample; may be shorter than the series
	// Fixed axis maximum; zero means auto from the data.
	Max float64
	// Hover is the sample index under the pointer, or -1.
	Hover int
}

const (
	chartHeight  = 96 // plot height in px, the same for every card
	chartHitSep  = "-h"
	chartTimeFmt = "15:04:05"
)

// chartStats returns min, max, and current of a series.
func chartStats(v []float64) (lo, hi, cur float64) {
	if len(v) == 0 {
		return 0, 0, 0
	}
	lo, hi = v[0], v[0]
	for _, x := range v {
		lo = min(lo, x)
		hi = max(hi, x)
	}
	return lo, hi, v[len(v)-1]
}

func chartAvg(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	var sum float64
	for _, x := range v {
		sum += x
	}
	return sum / float64(len(v))
}

func fmtNum(v float64) string {
	switch {
	case v >= chartHundred:
		return strconv.FormatFloat(v, 'f', 0, 64)
	case v >= chartTen:
		return strconv.FormatFloat(v, 'f', 1, 64)
	default:
		return strconv.FormatFloat(v, 'f', 2, 64)
	}
}

// axisNum formats a y-axis tick compactly (1.2k, 45, 0.5).
func axisNum(v float64) string {
	switch {
	case v >= megaUnit:
		return strings.TrimSuffix(strconv.FormatFloat(v/megaUnit, 'f', 1, 64), ".0") + "M"
	case v >= kiloUnit:
		s := strconv.FormatFloat(v/kiloUnit, 'f', 1, 64)
		return strings.TrimSuffix(s, ".0") + "k"
	case v >= chartTen:
		return strconv.FormatFloat(v, 'f', 0, 64)
	default:
		return strings.TrimRight(strings.TrimRight(strconv.FormatFloat(v, 'f', 2, 64), "0"), ".")
	}
}

// niceCeil rounds v up to 1, 2, 2.5, 5 or 10 times a power of ten so the axis
// maximum and its midpoint are readable numbers.
func niceCeil(v float64) float64 {
	if v <= 0 {
		return 1
	}
	exp := math.Pow(chartTen, math.Floor(math.Log10(v)))
	for _, f := range []float64{1, 2, 2.5, 5, chartTen} {
		if v <= f*exp+1e-9 {
			return f * exp
		}
	}
	return chartTen * exp
}

// chartTop is the y-axis maximum of spec.
func chartTop(spec chartSpec) float64 {
	top := spec.Max
	if top <= 0 {
		for _, s := range spec.Series {
			_, hi, _ := chartStats(s.Values)
			top = max(top, hi)
		}
		top = niceCeil(top)
	}
	if top <= 0 {
		top = 1
	}
	return top
}

// chartCard renders one live chart: y-axis labels with min/mid/max gridlines,
// an area plus line plot, a time axis, a legend, and a hover column whose
// value and time replace the "now" readout in the legend.
func chartCard(spec chartSpec) *ui.Element {
	top := chartTop(spec)
	n := 0
	for _, s := range spec.Series {
		n = max(n, len(s.Values))
	}
	hover := spec.Hover
	if hover < 0 || hover >= n {
		hover = -1
	}
	plot := ui.Div().ID(spec.ID+"-plot").Class("qr-chart-plot").Children(
		ui.Div().Class("qr-chart-grid qr-chart-grid-0"),
		ui.Div().Class("qr-chart-grid qr-chart-grid-1"),
		ui.Div().Class("qr-chart-grid qr-chart-grid-2"),
	)
	for si, s := range spec.Series {
		bars := make([]*ui.Element, 0, len(s.Values))
		for i, v := range s.Values {
			h := max(chartMinBar, min(chartHeight, int(math.Round(v/top*chartHeight))))
			cls := "qr-chart-bar " + s.Class
			if si == 0 {
				cls += " qr-chart-area"
			}
			bars = append(bars, ui.Div().Class(cls).Height(ui.Px(float64(h))).ID(fmt.Sprintf("%s-b%d-%d", spec.ID, si, i)))
		}
		plot.Children(ui.Div().Class("qr-chart-layer").Children(bars...))
	}
	hits := make([]*ui.Element, 0, n)
	for i := 0; i < n; i++ {
		cls := "qr-chart-hit"
		if i == hover {
			cls += " qr-chart-hit-on"
		}
		hits = append(hits, ui.Div().Class(cls).ID(fmt.Sprintf("%s%s%d", spec.ID, chartHitSep, i)))
	}
	plot.Children(ui.Div().Class("qr-chart-layer qr-chart-hits").Children(hits...))

	yaxis := ui.Div().ID(spec.ID+"-y").Class("qr-chart-y").Attr("aria-hidden", "true").Children(
		ui.Span().Text(axisNum(top)), ui.Span().Text(axisNum(top/chartHalf)), ui.Span().Text("0"))
	body := ui.Div().Class("qr-chart-body").Children(yaxis, plot)

	xaxis := ui.Div().ID(spec.ID+"-x").Class("qr-chart-x").Attr("aria-hidden", "true")
	if len(spec.Times) > 1 {
		first, last := spec.Times[0], spec.Times[len(spec.Times)-1]
		mid := first.Add(last.Sub(first) / chartHalf)
		xaxis.Children(ui.Span().Text(first.Format(chartTimeFmt)), ui.Span().Text(mid.Format(chartTimeFmt)), ui.Span().Text(last.Format(chartTimeFmt)))
	} else {
		xaxis.Children(ui.Span().Text("waiting for samples"))
	}

	legend := ui.Div().ID(spec.ID + "-legend").Class("qr-chart-legend")
	tag := "now"
	if hover >= 0 && hover < len(spec.Times) {
		tag = spec.Times[hover].Format(chartTimeFmt)
	}
	for _, s := range spec.Series {
		lo, hi, cur := chartStats(s.Values)
		if hover >= 0 && hover < len(s.Values) {
			cur = s.Values[hover]
		}
		key := ui.Div().Class("qr-chart-key").Children(
			ui.Span().Class("qr-dot "+s.Class+"-dot"),
			ui.Text(s.Name).Class("qr-muted"),
			ui.Text(fmtNum(cur)+spec.Unit),
		)
		if len(spec.Series) == 1 {
			key.Children(
				ui.Text("min "+fmtNum(lo)).Class("qr-muted"),
				ui.Text("max "+fmtNum(hi)).Class("qr-muted"),
				ui.Text("avg "+fmtNum(chartAvg(s.Values))).Class("qr-muted"),
			)
		}
		legend.Children(key)
	}
	legend.Children(ui.Text(tag).ID(spec.ID + "-readout").Class("qr-chart-tag"))
	return ui.Div().ID(spec.ID).Class("qr-chart").
		Attr("role", "img").Attr("aria-label", spec.Title+" chart").
		Children(ui.Div().Class("qr-chart-title").Text(spec.Title), body, xaxis, legend)
}

// chartHoverIndex reports which chart column box id names, if any.
func chartHoverIndex(id string) (chart string, idx int, ok bool) {
	if !strings.HasPrefix(id, "chart-") {
		return "", 0, false
	}
	i := strings.LastIndex(id, chartHitSep)
	if i < 0 {
		return "", 0, false
	}
	n, err := strconv.Atoi(id[i+len(chartHitSep):])
	if err != nil {
		return "", 0, false
	}
	return id[:i], n, true
}

// trackChartHover records the chart column under the pointer. It only marks
// state; the next tick repaints, so a pointer move never rebuilds the tree
// from inside the hover callback.
func (a *App) trackChartHover(id string) {
	chart, idx, ok := chartHoverIndex(id)
	if !ok {
		chart, idx = "", -1
	}
	if chart == a.mon.hoverChart && (chart == "" || idx == a.mon.hoverIdx) {
		return
	}
	a.mon.hoverChart, a.mon.hoverIdx, a.mon.hoverDirty = chart, idx, true
}

// chartHoverFor is the hover index for chart id, or -1.
func (a *App) chartHoverFor(id string) int {
	if a.mon.hoverChart == id {
		return a.mon.hoverIdx
	}
	return -1
}

// onHoverAll feeds chart hover tracking, then the shared hover handler.
func (a *App) onHoverAll(ctx context.Context, previous, next ownframe.Box) (bool, error) {
	a.trackChartHover(next.ID)
	return a.onHover(ctx, previous, next)
}

// takeChartHover reports whether a hover change needs a repaint of the
// visible Monitor tab.
func (a *App) takeChartHover() bool {
	d := a.mon.hoverDirty
	a.mon.hoverDirty = false
	return d && a.monitorVisible()
}
