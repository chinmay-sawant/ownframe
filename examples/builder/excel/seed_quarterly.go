package main

import (
	"fmt"
	"strconv"
)

// quarterlyRow is one quarter of the Quarterly sheet. Revenue for the last
// four quarters is not stored: it links to the Sheet1 quarter totals.
type quarterlyRow struct {
	label           string
	revenue         float64
	cogs, operating float64
}

// quarterlyHistory is twelve quarters, oldest first. The 2025 revenue cells
// are formulas on Sheet1, so editing a Sheet1 line moves the whole model.
func quarterlyHistory() []quarterlyRow {
	return []quarterlyRow{
		{"Q1 2023", 58000, 34800, 13900}, {"Q2 2023", 61500, 36600, 14400},
		{"Q3 2023", 64200, 38100, 14900}, {"Q4 2023", 68900, 40300, 15700},
		{"Q1 2024", 70100, 41200, 16100}, {"Q2 2024", 73800, 42900, 16800},
		{"Q3 2024", 77500, 44800, 17300}, {"Q4 2024", 79000, 45400, 17700},
		{"Q1 2025", 0, 46800, 18200}, {"Q2 2025", 0, 48900, 18900},
		{"Q3 2025", 0, 51200, 19600}, {"Q4 2025", 0, 54100, 20400},
	}
}

// chartsTableWidth is the column width of the Charts sheet tables.
const (
	chartsTableWidth = 100
	chartQuarters    = 8 // quarters shown on the Charts sheet
)

// seedQuarterlyModel appends the Quarterly data sheet and the Charts sheet.
// Both read Sheet1, so every table and chart follows the sample data.
func seedQuarterlyModel(w *Workbook) {
	q := newSheet(w, "Quarterly")
	c := newSheet(w, "Charts")
	w.Sheets = append(w.Sheets, q, c)
	seedQuarterly(q)
	seedChartsSheet(c)
	w.reloadAll()
}

func seedQuarterly(s *Sheet) {
	set := func(row, col int, raw string) { s.setRawQuiet(CellRef{row, col}, raw) }
	for i, h := range []string{"Quarter", "Revenue", "COGS", "Gross Profit", "Operating Expenses", "Net Income", "Margin"} {
		set(0, i, h)
	}
	sheet1Cols := []string{"C", "D", "E", "F"}
	hist := quarterlyHistory()
	for i, r := range hist {
		row := i + 1
		n := strconv.Itoa(row + 1)
		set(row, 0, r.label)
		if i >= len(hist)-len(sheet1Cols) {
			set(row, 1, "=Sheet1!"+sheet1Cols[i-(len(hist)-len(sheet1Cols))]+"10")
		} else {
			set(row, 1, fmt.Sprintf("%.0f", r.revenue))
		}
		set(row, 2, fmt.Sprintf("%.0f", r.cogs))
		set(row, 3, "=B"+n+"-C"+n)
		set(row, 4, fmt.Sprintf("%.0f", r.operating))
		set(row, 5, "=D"+n+"-E"+n)
		set(row, 6, "=IFERROR(F"+n+"/B"+n+",0)")
	}
	last := len(hist)
	set(last+1, 0, "Total")
	for col := 1; col <= 5; col++ {
		name := ColName(col)
		set(last+1, col, fmt.Sprintf("=SUM(%s2:%s%d)", name, name, last+1))
	}
	set(last+1, 6, fmt.Sprintf("=IFERROR(F%d/B%d,0)", last+2, last+2))
	s.SetStyle(Range{CellRef{0, 0}, CellRef{0, 6}}, func(st *Style) {
		st.Bold, st.BG, st.Align = true, "#DDEBF7", "center"
		st.Edges.Bottom = EdgeThin
	})
	s.SetStyle(Range{CellRef{1, 1}, CellRef{last + 1, 5}}, func(st *Style) { st.NumFmt, st.Decimals = "Currency", -1 })
	s.SetStyle(Range{CellRef{1, 6}, CellRef{last + 1, 6}}, func(st *Style) { st.NumFmt, st.Decimals = "Percent", 1 })
	s.SetStyle(Range{CellRef{last + 1, 0}, CellRef{last + 1, 6}}, func(st *Style) { st.Bold, st.Edges.Top = true, EdgeThin })
	for col := 0; col <= 6; col++ {
		s.SetColWidth(col, sampleColWidth+20)
	}
	s.SetFreeze(1, 0)
}

func seedChartsSheet(s *Sheet) {
	set := func(ref, raw string) {
		c, _ := ParseA1(ref)
		s.setRawQuiet(c, raw)
	}
	set("A1", "Charts")
	set("A3", "Region")
	set("B3", "Revenue")
	for i, name := range []string{"North", "South", "East", "West"} {
		n := strconv.Itoa(4 + i)
		set("A"+n, name)
		set("B"+n, `=SUMIF(Sheet1!A2:A9,A`+n+`,Sheet1!G2:G9)`)
	}
	set("A10", "Product")
	set("B10", "Revenue")
	for i, name := range []string{"Widget", "Gadget"} {
		n := strconv.Itoa(11 + i)
		set("A"+n, name)
		set("B"+n, `=SUMIF(Sheet1!B2:B9,A`+n+`,Sheet1!G2:G9)`)
	}
	set("A16", "Quarter")
	set("B16", "Revenue")
	set("C16", "Gross Profit")
	set("D16", "Net Income")
	// The tables chart the latest eight quarters, which stay readable at
	// chart width. The Quarterly sheet keeps all twelve.
	skip := len(quarterlyHistory()) - chartQuarters
	for i := 0; i < chartQuarters; i++ {
		n := strconv.Itoa(17 + i)
		src := strconv.Itoa(2 + skip + i)
		set("A"+n, "=Quarterly!A"+src)
		set("B"+n, "=Quarterly!B"+src)
		set("C"+n, "=Quarterly!D"+src)
		set("D"+n, "=Quarterly!F"+src)
	}
	bold := func(st *Style) {
		st.Bold, st.BG, st.Align = true, "#DDEBF7", "center"
		st.Edges.Bottom = EdgeThin
	}
	s.SetStyle(Range{CellRef{0, 0}, CellRef{0, 0}}, func(st *Style) { st.Bold, st.FontSize = true, 18 })
	s.SetStyle(Range{CellRef{2, 0}, CellRef{2, 1}}, bold)
	s.SetStyle(Range{CellRef{9, 0}, CellRef{9, 1}}, bold)
	s.SetStyle(Range{CellRef{15, 0}, CellRef{15, 3}}, bold)
	for _, r := range []Range{{CellRef{3, 1}, CellRef{6, 1}}, {CellRef{10, 1}, CellRef{11, 1}}, {CellRef{16, 1}, CellRef{15 + chartQuarters, 3}}} {
		s.SetStyle(r, func(st *Style) { st.NumFmt, st.Decimals = "Currency", -1 })
	}
	for col := 0; col <= 3; col++ {
		s.SetColWidth(col, chartsTableWidth)
	}
}

// Auto layout of the seeded charts: they start right of the tables and wrap
// to the window width.
const (
	autoChartX   = 420.0 // left edge, clear of the four table columns
	autoChartY   = 10.0
	autoChartGap = 10.0
	autoChartW   = 300.0
	autoChartH   = 290.0
	autoChartTop = 365.0 // height of the column chart, which carries totals
	autoPadRight = 24.0  // scrollbar and margin kept clear on the right
)

// seedCharts places one chart of each plotted kind on the Charts sheet, bound
// to the tables seeded beside them, then lays them out for the window.
func (a *app) seedCharts() {
	rng := func(ref string) Range {
		r, _ := ParseRange(ref)
		return r
	}
	specs := []struct {
		kind, title, src string
		h                float64
	}{
		{"column", "Revenue, gross profit and net income", "A16:D24", autoChartTop},
		{"line", "Revenue trend", "A16:B24", autoChartTop},
		{"area", "Revenue and gross profit", "A16:C24", autoChartTop},
		{"bar", "Revenue by region", "A3:B7", autoChartH},
		{"pie", "Revenue by product", "A10:B12", autoChartH},
		{"donut", "Revenue share by region", "A3:B7", autoChartH},
	}
	for _, sp := range specs {
		a.x.chartSeq++
		a.x.charts = append(a.x.charts, &ChartSpec{
			ID: a.x.chartSeq, Sheet: "Charts", Kind: sp.kind, Title: sp.title,
			Src: rng(sp.src), W: autoChartW, H: sp.h, Auto: true,
		})
	}
	a.layoutAutoCharts()
}

// layoutAutoCharts flows the Auto charts left to right and wraps them to the
// window width, so a narrow window stacks them instead of clipping them. The
// width is in unzoomed CSS px, the unit chart coordinates use.
func (a *app) layoutAutoCharts() {
	w, _ := a.viewportSize()
	avail := a.unzf(float64(w)) - chartOriginX - autoPadRight
	x, y, rowH := autoChartX, autoChartY, 0.0
	for _, c := range a.x.charts {
		if !c.Auto {
			continue
		}
		if x > autoChartX && x+c.W > avail {
			x, y, rowH = autoChartX, y+rowH+autoChartGap, 0
		}
		c.X, c.Y = x, y
		x += c.W + autoChartGap
		rowH = max(rowH, c.H)
	}
}
