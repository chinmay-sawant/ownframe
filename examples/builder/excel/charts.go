package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// chartKinds lists the insertable chart types.
func chartKinds() []struct{ Kind, Label string } {
	return []struct{ Kind, Label string }{
		{"column", "Column"},
		{"bar", "Bar"},
		{"line", "Line"},
		{"area", "Area"},
		{"pie", "Pie"},
		{"donut", "Donut"},
		{"stacked", "Stacked Column"},
	}
}

// Chart placement and scaling constants.
const (
	chartStartX  = 360.0 // left edge of the first inserted chart
	chartStartY  = 40.0  // top edge of the first inserted chart
	percentScale = 100.0 // fraction-to-percent multiplier
)

const (
	// New charts cascade by chartCascadeStep px, wrapping every chartCascadeSteps.
	chartCascadeStep  = 24
	chartCascadeSteps = 5
	chartDefaultW     = 380.0
	chartDefaultH     = 290.0
	// chartOriginX/Y place chart coordinates (0,0) at the top-left of the
	// grid body in window coordinates (title 36 + tabs 28 + ribbon 88 +
	// formula bar 28 + column header 24, row header 40).
	chartOriginX = 40.0
	chartOriginY = 176.0
)

// ChartSpec describes one floating chart bound to a sheet range. Series are
// read from Src on every render, so edits and recalculation show up at once.
type ChartSpec struct {
	ID         int
	Sheet      string
	Kind       string
	Title      string
	Src        Range
	X, Y, W, H float64
	// Auto marks a seeded chart whose position follows the window width. A
	// drag by the user clears it.
	Auto bool
}

// chartData reads categories and series from the spec's source range. The
// first column holds category labels and the first row holds series names
// when it is not numeric. A single-column range becomes one series.
func (a *app) chartData(c *ChartSpec) (labels []string, series []chadcn.ChartSeries) {
	var sh *Sheet
	for _, s := range a.wb.Sheets {
		if s.Name == c.Sheet {
			sh = s
		}
	}
	if sh == nil {
		return nil, nil
	}
	r := c.Src
	labelCol := r.Min.Col
	firstData := r.Min.Col + 1
	if r.Max.Col == r.Min.Col {
		labelCol, firstData = -1, r.Min.Col
	}
	// Series are the columns that hold numbers below the first row.
	var dataCols []int
	for col := firstData; col <= r.Max.Col; col++ {
		for row := r.Min.Row + 1; row <= r.Max.Row; row++ {
			if sh.Value(row, col).Kind == KindNumber {
				dataCols = append(dataCols, col)
				break
			}
		}
	}
	if len(dataCols) == 0 {
		return nil, nil
	}
	header := sh.Value(r.Min.Row, dataCols[0]).Kind != KindNumber
	startRow := r.Min.Row
	if header {
		startRow++
	}
	for row := startRow; row <= r.Max.Row; row++ {
		if labelCol >= 0 {
			labels = append(labels, sh.Display(row, labelCol))
		} else {
			labels = append(labels, strconv.Itoa(row-startRow+1))
		}
	}
	for _, col := range dataCols {
		name := ColName(col)
		if header {
			name = sh.Display(r.Min.Row, col)
		}
		vals := make([]float64, 0, len(labels))
		for row := startRow; row <= r.Max.Row; row++ {
			v := sh.Value(row, col)
			if v.Kind == KindNumber {
				vals = append(vals, v.Num)
			} else {
				vals = append(vals, 0)
			}
		}
		series = append(series, chadcn.ChartSeries{Label: name, Values: vals})
	}
	return labels, series
}

// addChart inserts a chart for the current selection and selects it.
func (a *app) addChart(kind string) *ChartSpec {
	a.x.chartSeq++
	off := float64((a.x.chartSeq - 1) % chartCascadeSteps * chartCascadeStep)
	c := &ChartSpec{
		ID: a.x.chartSeq, Sheet: a.wb.Sheet().Name, Kind: kind,
		Title: a.wb.Sheet().Name + " " + kind + " chart", Src: a.sel,
		X: chartStartX + off, Y: chartStartY + off, W: chartDefaultW, H: chartDefaultH,
	}
	a.x.charts = append(a.x.charts, c)
	a.x.selChart = c.ID
	return c
}

func (a *app) chartByID(id int) *ChartSpec {
	for _, c := range a.x.charts {
		if c.ID == id {
			return c
		}
	}
	return nil
}

func (a *app) deleteChart(id int) {
	out := a.x.charts[:0]
	for _, c := range a.x.charts {
		if c.ID != id {
			out = append(out, c)
		}
	}
	a.x.charts = out
	if a.x.selChart == id {
		a.x.selChart = 0
	}
}

// chartLayer renders every chart of the active sheet as an absolutely
// positioned card over the grid.
func (a *app) chartLayer() *ui.Element {
	if len(a.x.charts) == 0 {
		return nil
	}
	layer := ui.Div().Class("rb-chart-layer").Position(ui.PositionFixed).
		Left(ui.Px(a.zf(chartOriginX))).Top(ui.Px(a.zf(chartOriginY))).Width(ui.Px(0)).Height(ui.Px(0))
	any := false
	for _, c := range a.x.charts {
		if c.Sheet != a.wb.Sheet().Name {
			continue
		}
		any = true
		layer.Children(a.chartCard(c))
	}
	if !any {
		return nil
	}
	return layer
}

func (a *app) chartCard(c *ChartSpec) *ui.Element {
	idn := strconv.Itoa(c.ID)
	cls := "rb-chart"
	if a.x.selChart == c.ID {
		cls += " rb-chart-selected"
	}
	labels, series := a.chartData(c)
	card := ui.Div().ID("rb-chart-"+idn).Class(cls).Attr("role", "group").
		Attr("aria-label", c.Title).Attr("data-action", "xl:chart:select:"+idn).
		Position(ui.PositionAbsolute).Left(ui.Px(a.zf(c.X))).Top(ui.Px(a.zf(c.Y))).
		Width(ui.Px(a.zf(c.W))).Height(ui.Px(a.zf(c.H)))
	head := ui.Div().ID("rb-chart-hdr-"+idn).Class("rb-chart-head").Children(
		ui.Span().Class("rb-chart-title").Text(c.Title),
		ui.Button("x").Attr("type", "button").Class("rb-chart-close").
			Attr("aria-label", "Delete chart").Attr("data-action", "xl:chart:del:"+idn),
	)
	card.Children(head, ui.Div().Class("rb-chart-body").Children(a.chartBody(c, labels, series)))
	return card
}

func (a *app) chartBody(c *ChartSpec, labels []string, series []chadcn.ChartSeries) *ui.Element {
	if len(series) == 0 || len(labels) == 0 {
		return ui.Div().Class("rb-chart-empty").Text("No data in " + c.Src.String())
	}
	idn := "xlc" + strconv.Itoa(c.ID)
	switch c.Kind {
	case "bar":
		return barChart(labels, series[0])
	case "pie":
		return pieChart(labels, series[0], false)
	case "donut":
		return pieChart(labels, series[0], true)
	case "stacked":
		return stackedChart(labels, series)
	case "line", "area":
		if len(series) == 1 {
			series = append(series, chadcn.ChartSeries{Label: " ", Values: make([]float64, len(labels))})
		}
		return chadcn.AreaChart(chadcn.AreaChartOptions{ID: idn, Series: series[:2], Labels: labels})
	}
	return chadcn.Chart(chadcn.ChartOptions{ID: idn, Series: series, Labels: labels})
}

func chartPalette() []string {
	return []string{"#4472c4", "#ed7d31", "#a5a5a5", "#ffc000", "#5b9bd5", "#70ad47", "#264478", "#9e480e"}
}

// barChart draws horizontal bars with div geometry.
func barChart(labels []string, s chadcn.ChartSeries) *ui.Element {
	maxV := 0.0
	for _, v := range s.Values {
		if v > maxV {
			maxV = v
		}
	}
	root := ui.Div().Class("rb-hbars")
	for i, l := range labels {
		pct := 0.0
		if maxV > 0 && i < len(s.Values) && s.Values[i] > 0 {
			pct = s.Values[i] / maxV * percentScale
		}
		root.Children(ui.Div().Class("rb-hbar-row").Children(
			ui.Span().Class("rb-hbar-label").Text(l),
			ui.Div().Class("rb-hbar-track").Children(
				ui.Div().Class("rb-hbar-fill").Width(ui.Percent(pct)).
					BackgroundColor(ui.Hex(chartPalette()[i%len(chartPalette())])),
			),
		))
	}
	return root
}

// pieChart is a conic-gradient disc with a legend. A donut punches a hole in
// the middle with a disc in the sheet colour.
func pieChart(labels []string, s chadcn.ChartSeries, donut bool) *ui.Element {
	total := 0.0
	for _, v := range s.Values {
		if v > 0 {
			total += v
		}
	}
	legend := ui.Div().Class("rb-pie-legend")
	var stops []string
	acc := 0.0
	for i, l := range labels {
		v := 0.0
		if i < len(s.Values) && s.Values[i] > 0 {
			v = s.Values[i]
		}
		pct := 0.0
		if total > 0 {
			pct = v / total * percentScale
		}
		col := chartPalette()[i%len(chartPalette())]
		if pct > 0 {
			stops = append(stops, fmt.Sprintf("%s %.2f%%, %s %.2f%%", col, acc, col, acc+pct))
			acc += pct
		}
		legend.Children(ui.Div().Class("rb-pie-item").Children(
			ui.Span().Class("rb-pie-dot").BackgroundColor(ui.Hex(col)),
			ui.Span().Text(fmt.Sprintf("%s  %.1f%%", l, pct)),
		))
	}
	disc := ui.Div().Class("rb-pie-disc")
	if len(stops) > 0 {
		disc.CSSRaw("background", "conic-gradient("+strings.Join(stops, ", ")+")")
	}
	if donut {
		disc.Children(ui.Div().Class("rb-pie-hole"))
	}
	return ui.Div().Class("rb-pie").Children(disc, legend)
}

// stackedChart stacks every series in one column per label. Column height is
// the label's total against the largest total, and each segment's share of the
// column follows its value.
func stackedChart(labels []string, series []chadcn.ChartSeries) *ui.Element {
	totals := make([]float64, len(labels))
	maxT := 0.0
	for i := range labels {
		for _, s := range series {
			if i < len(s.Values) && s.Values[i] > 0 {
				totals[i] += s.Values[i]
			}
		}
		maxT = max(maxT, totals[i])
	}
	plot := ui.Div().Class("rb-stack")
	names := ui.Div().Class("rb-stack-names")
	for i, l := range labels {
		col := ui.Div().Class("rb-stack-col")
		if maxT > 0 {
			col.CSSRaw("height", fmt.Sprintf("%.2f%%", totals[i]/maxT*percentScale))
		}
		for k, s := range series {
			if i < len(s.Values) && s.Values[i] > 0 {
				col.Children(ui.Div().Class("rb-stack-seg").CSSRaw("flex-grow", fmt.Sprintf("%.2f", s.Values[i])).
					BackgroundColor(ui.Hex(chartPalette()[k%len(chartPalette())])))
			}
		}
		plot.Children(col)
		names.Children(ui.Span().Class("rb-stack-name").Text(l))
	}
	legend := ui.Div().Class("rb-pie-legend")
	for k, s := range series {
		legend.Children(ui.Div().Class("rb-pie-item").Children(
			ui.Span().Class("rb-pie-dot").BackgroundColor(ui.Hex(chartPalette()[k%len(chartPalette())])),
			ui.Span().Text(s.Label),
		))
	}
	return ui.Div().Class("rb-stacked").Children(plot, names, legend)
}

// chartDragStart claims a press on a chart header. The host calls it from its
// DragStart handler with the pressed box.
func (a *app) chartDragStart(box ownframe.Box) bool {
	const prefix = "rb-chart-hdr-"
	if !strings.HasPrefix(box.ID, prefix) {
		return false
	}
	id, err := strconv.Atoi(strings.TrimPrefix(box.ID, prefix))
	if err != nil || a.chartByID(id) == nil {
		return false
	}
	a.x.selChart, a.x.dragChart = id, id
	return true
}

// chartDragMove moves the dragged chart by CSS deltas.
func (a *app) chartDragMove(dx, dy float64) bool {
	c := a.chartByID(a.x.dragChart)
	if c == nil {
		return false
	}
	c.Auto = false
	c.X += a.unzf(dx)
	c.Y += a.unzf(dy)
	if c.X < 0 {
		c.X = 0
	}
	if c.Y < 0 {
		c.Y = 0
	}
	return true
}

// chartDragEnd releases the drag. It reports whether a chart was dragged.
func (a *app) chartDragEnd() bool {
	was := a.x.dragChart != 0
	a.x.dragChart = 0
	return was
}
