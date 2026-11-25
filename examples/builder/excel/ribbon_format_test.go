package main

import (
	"context"
	"math"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/ui"
)

func TestFontFamilyControl(t *testing.T) {
	a := rbApp(t)
	rbSelect(a, "A2:B3")
	if html := rbRender(t, a.ribbon()); !strings.Contains(html, `aria-label="Font family"`) || !strings.Contains(html, "Default v") {
		t.Fatalf("Home ribbon has no font family control")
	}
	a.runAction("xl:family:mono")
	for _, ref := range []string{"A2", "B3"} {
		r, _ := ParseA1(ref)
		if got := a.wb.Sheet().StyleAt(r.Row, r.Col).FontFamily; got != "mono" {
			t.Errorf("%s family = %q", ref, got)
		}
	}
	rbSelect(a, "A2")
	if html := rbRender(t, a.ribbon()); !strings.Contains(html, "Mono v") {
		t.Errorf("ribbon does not reflect the Mono family")
	}
	if html := rbRender(t, a.fontFamilyMenu()); !strings.Contains(html, "xl:family:serif") {
		t.Errorf("family menu lacks Serif")
	}
	el := ui.Div()
	a.applyStyleExtras(el, a.wb.Sheet().StyleAt(1, 0), 96, 22, "x")
	if html := rbRender(t, el); !strings.Contains(html, "font-family") || !strings.Contains(html, "monospace") {
		t.Errorf("cell css lacks the mono stack: %s", html)
	}
	a.runAction("xl:family:")
	if a.wb.Sheet().StyleAt(1, 0).FontFamily != "" {
		t.Errorf("Default did not reset the family")
	}
}

func TestVerticalAlignAndIndent(t *testing.T) {
	a := rbApp(t)
	rbSelect(a, "A2")
	for _, v := range []string{"top", "middle", "bottom"} {
		a.runAction("xl:valign:" + v)
		if got := a.activeStyle().VAlign; got != v {
			t.Errorf("valign %q stored as %q", v, got)
		}
		html := rbRender(t, a.ribbon())
		if !strings.Contains(html, `aria-label="Align `+strings.ToUpper(v[:1])+v[1:]+`" aria-pressed="true"`) &&
			!strings.Contains(html, "rb-rb-on") {
			t.Errorf("ribbon does not show %q pressed", v)
		}
	}
	a.runAction("xl:indent:+")
	a.runAction("xl:indent:+")
	if a.activeStyle().Indent != 2 {
		t.Errorf("indent = %d", a.activeStyle().Indent)
	}
	for i := 0; i < 20; i++ {
		a.runAction("xl:indent:+")
	}
	if a.activeStyle().Indent != maxIndent {
		t.Errorf("indent should clamp at %d, got %d", maxIndent, a.activeStyle().Indent)
	}
	for i := 0; i < 20; i++ {
		a.runAction("xl:indent:-")
	}
	if a.activeStyle().Indent != 0 {
		t.Errorf("indent should clamp at 0, got %d", a.activeStyle().Indent)
	}
	// CSS: indent pads the left, middle and bottom push text down
	st := Style{Indent: 2, VAlign: "middle"}
	el := ui.Div()
	a.applyStyleExtras(el, st, 96, 40, "x")
	html := rbRender(t, el)
	if !strings.Contains(html, "padding-left:24px") || !strings.Contains(html, "padding-top:9px") {
		t.Errorf("css = %s", html)
	}
	el = ui.Div()
	a.applyStyleExtras(el, Style{VAlign: "bottom"}, 96, 40, "x")
	if html := rbRender(t, el); !strings.Contains(html, "padding-top:18px") {
		t.Errorf("bottom css = %s", html)
	}
	// default General alignment is untouched by the new fields
	if a.wb.Sheet().EffectiveAlign(1, 0) == "" {
		t.Errorf("EffectiveAlign lost the default")
	}
}

func TestBorderCSSPerEdge(t *testing.T) {
	st := Style{Edges: Edges{Left: EdgeThin, Right: EdgeThick, Top: EdgeNone, Bottom: EdgeDouble}}
	a := rbApp(t)
	el := ui.Div()
	a.applyStyleExtras(el, st, 96, 22, "x")
	html := rbRender(t, el)
	for _, want := range []string{"border-left:1px solid", "border-right:2px solid", "border-bottom:3px double"} {
		if !strings.Contains(html, want) {
			t.Errorf("css lacks %q: %s", want, html)
		}
	}
	if strings.Contains(html, "border-top") {
		t.Errorf("top edge should not be drawn: %s", html)
	}
	legacy := ui.Div()
	a.applyStyleExtras(legacy, Style{Border: "bottom"}, 96, 22, "x")
	if html := rbRender(t, legacy); !strings.Contains(html, "border-bottom:1px solid") {
		t.Errorf("legacy border not painted: %s", html)
	}
}

func TestFormatCellsDialogNewFields(t *testing.T) {
	a := rbApp(t)
	rbSelect(a, "B2:C3")
	a.runAction("xl:dlg:format")
	a.runAction("xl:fc:tab:Alignment")
	a.runAction("xl:fc:valign:middle")
	a.runAction("xl:fc:indent:+")
	a.runAction("xl:fc:tab:Font")
	a.runAction("xl:fc:family:serif")
	a.runAction("xl:fc:tab:Border")
	a.runAction("xl:fc:border:outside")
	html := rbRender(t, a.dialog())
	if !strings.Contains(html, "Outside Borders") || !strings.Contains(html, "Thick Outside Borders") {
		t.Errorf("Border tab lacks the preset list")
	}
	a.runAction("xl:fc:ok")
	sh := a.wb.Sheet()
	if st := sh.StyleAt(1, 1); st.VAlign != "middle" || st.Indent != 1 || st.FontFamily != "serif" {
		t.Errorf("B2 style = %+v", st)
	}
	if e := sh.StyleAt(1, 1).BorderEdges(); e != (Edges{Left: EdgeThin, Top: EdgeThin}) {
		t.Errorf("B2 outside edges = %+v", e)
	}
	if sh.StyleAt(2, 2).BorderEdges() != (Edges{Right: EdgeThin, Bottom: EdgeThin}) {
		t.Errorf("C3 outside edges = %+v", sh.StyleAt(2, 2).BorderEdges())
	}
	// cancel leaves cells alone
	a.runAction("xl:dlg:format")
	a.runAction("xl:fc:family:mono")
	a.runAction("xl:close")
	if sh.StyleAt(1, 1).FontFamily != "serif" {
		t.Errorf("cancel changed the family")
	}
}

func TestNumberFormatMenuHasDateAndTime(t *testing.T) {
	a := rbApp(t)
	html := rbRender(t, a.numberFormatMenu())
	for _, w := range []string{"Short Date", "Long Date", "Time", "xl:nf:LongDate", "xl:nf:Time"} {
		if !strings.Contains(html, w) {
			t.Errorf("number format menu lacks %q", w)
		}
	}
	rbSelect(a, "J2")
	a.wb.Edit(a.wb.Sheet(), 1, 9, "46351.5")
	a.runAction("xl:nf:Date")
	if got := a.wb.Sheet().Display(1, 9); got != "2026-11-25" {
		t.Errorf("Short Date = %q", got)
	}
	a.runAction("xl:nf:LongDate")
	if got := a.wb.Sheet().Display(1, 9); got != "Wed, Nov 25, 2026" {
		t.Errorf("Long Date = %q", got)
	}
	a.runAction("xl:nf:Time")
	if got := a.wb.Sheet().Display(1, 9); got != "12:00:00" {
		t.Errorf("Time = %q", got)
	}
}

func TestCondFormatNewRuleMenuItems(t *testing.T) {
	a := rbApp(t)
	html := rbRender(t, a.condFormatMenu())
	for _, w := range []string{"xl:cf:bottom", "xl:cf:above", "xl:cf:below"} {
		if !strings.Contains(html, w) {
			t.Errorf("conditional formatting menu lacks %q", w)
		}
	}
	rbSelect(a, "C2:C9")
	a.runAction("xl:cf:bottom")
	a.runAction("xl:cf:above")
	rules := a.wb.Sheet().CondRules()
	if len(rules) != 2 || rules[0].Rule.Kind != "bottom" || rules[1].Rule.Kind != "above" {
		t.Fatalf("rules = %+v", rules)
	}
	if dlg := rbRender(t, func() *ui.Element { a.openDialogByID("condfmt"); return a.dialog() }()); !strings.Contains(dlg, "Bottom N") ||
		!strings.Contains(dlg, "Above average") {
		t.Errorf("Manage Rules dialog lacks the new kinds")
	}
}

func TestRibbonCollapseExpand(t *testing.T) {
	a := rbApp(t)
	if a.ribbonHeight() != ribbonHeightFull || ribbonHeightFull != 88 {
		t.Fatalf("expanded height = %d", a.ribbonHeight())
	}
	if html := rbRender(t, a.ribbon()); !strings.Contains(html, "rb-ribbon-body") || !strings.Contains(html, `aria-expanded="true"`) {
		t.Fatalf("expanded ribbon lacks its body or the expanded state")
	}
	a.runAction("xl:ribbon:toggle")
	if !a.ribbonCollapsed || a.ribbonHeight() != 32 || ribbonHeightMin != 32 {
		t.Fatalf("collapsed = %v height = %d", a.ribbonCollapsed, a.ribbonHeight())
	}
	html := rbRender(t, a.ribbon())
	if strings.Contains(html, "rb-ribbon-body") || !strings.Contains(html, `aria-expanded="false"`) {
		t.Errorf("collapsed ribbon still has a body or reports expanded")
	}
	if !strings.Contains(html, "xl:tab:Home") {
		t.Errorf("collapsed ribbon lost its tabs")
	}
	// the grid gets the freed space
	cw, ch := a.canvasSize()
	a.runAction("xl:ribbon:toggle")
	cw2, ch2 := a.canvasSize()
	if cw != cw2 || ch <= ch2 {
		t.Errorf("canvas %dx%d collapsed vs %dx%d expanded", cw, ch, cw2, ch2)
	}
	// choosing a tab while collapsed expands it
	a.runAction("xl:ribbon:toggle")
	a.runAction("xl:tab:Data")
	if a.ribbonCollapsed || a.ribbonTab != "Data" {
		t.Errorf("tab click did not expand: collapsed=%v tab=%q", a.ribbonCollapsed, a.ribbonTab)
	}
}

func TestFilterPopoverSearchAndSort(t *testing.T) {
	a := rbApp(t)
	sh := a.wb.Sheet()
	a.runAction("xl:filter:toggle")
	a.openFilter(0) // the entry point the header chevron calls
	if a.overlay != "filter" || a.x.filterCol != 0 {
		t.Fatalf("openFilter: overlay=%q col=%d", a.overlay, a.x.filterCol)
	}
	html := rbRender(t, a.dialog())
	for _, w := range []string{"rb-fsearch", "Sort A to Z", "Sort Z to A", "Select All", "Clear Filter"} {
		if !strings.Contains(html, w) {
			t.Errorf("filter popover lacks %q", w)
		}
	}
	// typing in the search box narrows the list and checks only matches
	a.page.Form().SetFormValue("rb-fsearch", "nor")
	if _, err := a.overlayChange(context.Background(), ownframe.Box{ID: "rb-fsearch"}); err != nil {
		t.Fatal(err)
	}
	html = rbRender(t, a.dialog())
	if !strings.Contains(html, "North") || strings.Contains(html, "South") {
		t.Errorf("search did not narrow the list")
	}
	a.runAction("xl:filter:apply")
	visible := 0
	for r := 1; r <= 8; r++ {
		if !sh.HiddenRow(r) {
			visible++
			if sh.Raw(r, 0) != "North" {
				t.Errorf("row %d (%s) survived the filter", r+1, sh.Raw(r, 0))
			}
		}
	}
	if visible != 2 {
		t.Errorf("visible data rows = %d, want 2", visible)
	}
	// clear, then sort the table from the popover
	a.runAction("xl:filter:clear")
	a.wb.DeleteRows(sh, 9, 1) // drop the Total row so only data rows sort
	a.openFilter(2)
	a.runAction("xl:filter:sort:desc")
	if a.overlay != "" {
		t.Errorf("sort should close the popover")
	}
	if sh.Raw(0, 0) != "Region" {
		t.Errorf("header moved: %q", sh.Raw(0, 0))
	}
	prev := math.Inf(1)
	for r := 1; r <= 8; r++ {
		v := sh.Value(r, 2).Num
		if v > prev {
			t.Errorf("column C not descending at row %d: %v after %v", r+1, v, prev)
		}
		prev = v
	}
	a.runAction("xl:filter:open:1")
	a.runAction("xl:filter:sort:asc")
	if sh.Raw(1, 1) != "Gadget" {
		t.Errorf("sort A to Z from the popover: B2 = %q", sh.Raw(1, 1))
	}
}

func TestContextMenuThroughHandlers(t *testing.T) {
	a := rbApp(t)
	rows := a.page.ContextMenuFor("xl-gc-2-1") // B3
	if len(rows) < 8 {
		t.Fatalf("cell menu has %d rows", len(rows))
	}
	want := map[string]bool{"Cut": false, "Copy": false, "Paste": false, "Paste Special...": false, "Insert Row": false,
		"Delete Row": false, "Clear Contents": false, "Format Cells...": false, "Insert Comment": false}
	for _, r := range rows {
		if _, ok := want[r.Label]; ok {
			want[r.Label] = true
		}
	}
	for label, seen := range want {
		if !seen {
			t.Errorf("cell menu lacks %q", label)
		}
	}
	if a.active != (CellRef{2, 1}) {
		t.Errorf("right click did not select B3: %v", a.active)
	}
	if err := a.page.MenuAction(context.Background(), "xl:ctx:clear"); err != nil {
		t.Fatal(err)
	}
	if a.wb.Sheet().Raw(2, 1) != "" {
		t.Errorf("MenuAction did not clear B3")
	}
	if err := a.page.MenuAction(context.Background(), "xl:dlg:comment"); err != nil {
		t.Fatal(err)
	}
	if a.overlay != "comment" {
		t.Errorf("Insert Comment did not open the dialog: %q", a.overlay)
	}
	a.runAction("xl:close")

	rows = a.page.ContextMenuFor("xl-gch-3")
	labels := ""
	for _, r := range rows {
		labels += r.Label + "|"
	}
	for _, w := range []string{"Insert", "Delete", "Hide", "Unhide"} {
		if !strings.Contains(labels, w) {
			t.Errorf("column header menu lacks %q: %s", w, labels)
		}
	}
	for _, r := range a.page.ContextMenuFor("xl-formula") {
		if strings.HasPrefix(r.ID, "xl:") {
			t.Errorf("a non-grid element got the custom row %q", r.ID)
		}
	}
	if rows := a.page.ContextMenuFor("xl-tab-0"); len(rows) == 0 {
		t.Errorf("sheet tab has no context menu")
	}
}

func TestZoomControlsSetZoom(t *testing.T) {
	a := rbApp(t)
	rbSelect(a, "A1:D3")
	a.runAction("xl:tab:View")
	html := rbRender(t, a.ribbon())
	for _, w := range []string{"xl:zoom:in", "xl:zoom:out", "xl:zoom:reset", "xl:zoom:sel"} {
		if !strings.Contains(html, w) {
			t.Errorf("View ribbon lacks %q", w)
		}
	}
	a.runAction("xl:zoom:in")
	if a.x.zoom != 110 || a.zoomPct() != 110 {
		t.Errorf("zoom in = %d", a.x.zoom)
	}
	a.runAction("xl:zoom:out")
	a.runAction("xl:zoom:out")
	if a.x.zoom != 90 {
		t.Errorf("zoom out = %d", a.x.zoom)
	}
	a.runAction("xl:zoom:reset")
	if a.x.zoom != 100 {
		t.Errorf("reset = %d", a.x.zoom)
	}
	a.runAction("xl:zoom:sel")
	if a.x.zoom != 200 && a.x.zoom != 250 {
		// four columns fill the sheet area at 250 percent, capped at 200
		t.Errorf("zoom to selection = %d", a.x.zoom)
	}
	for i := 0; i < 30; i++ {
		a.runAction("xl:zoom:in")
	}
	if a.x.zoom != 200 {
		t.Errorf("zoom should cap at 200, got %d", a.x.zoom)
	}
	for i := 0; i < 40; i++ {
		a.runAction("xl:zoom:out")
	}
	if a.x.zoom != 50 {
		t.Errorf("zoom should floor at 50, got %d", a.x.zoom)
	}
}

// opColor returns the first fill op color inside a box as 0-255 channels.
func opColor(t *testing.T, a *app, id string) [3]int {
	t.Helper()
	d := a.page.Probe().Display()
	if d == nil {
		t.Fatal("no display list")
	}
	for _, box := range a.page.Boxes() {
		if box.ID != id {
			continue
		}
		owned := map[uint64]bool{}
		for _, op := range box.OperationIDs {
			owned[op] = true
		}
		for _, op := range d.Ops {
			if owned[op.ID] && op.Kind == ownframe.DisplayOpFillRect {
				return [3]int{int(math.Round(op.R * 255)), int(math.Round(op.G * 255)), int(math.Round(op.B * 255))}
			}
		}
		t.Fatalf("%s has no fill op", id)
	}
	t.Fatalf("no box %q", id)
	return [3]int{}
}

func TestConditionalFormattingVisualOverlay(t *testing.T) {
	a := rbApp(t)
	sh := a.wb.Sheet()
	rbSelect(a, "C2:C9")
	a.runAction("xl:cf:scale3")
	rbSelect(a, "D2:D9")
	a.runAction("xl:cf:databar")
	a.redraw()
	lo, hi := sh.StyleAt(7, 2).BG, sh.StyleAt(5, 2).BG // C8 (7600) and C6 (14800)
	if lo == "" || hi == "" || lo == hi {
		t.Fatalf("scale colors lo=%q hi=%q", lo, hi)
	}
	// C8 holds the smallest Q1 value (6900 in row 9), C6 the largest (14800)
	if got := opColor(t, a, "xl-gc-8-2"); got != [3]int{0xF8, 0x69, 0x6B} {
		t.Errorf("C9 (lowest) paints %v, want #F8696B", got)
	}
	if got := opColor(t, a, "xl-gc-5-2"); got != [3]int{0x63, 0xBE, 0x7B} {
		t.Errorf("C6 (highest) paints %v, want #63BE7B", got)
	}
	// the data bar is a child fill whose width follows the value: D6 holds the
	// column maximum (full bar) and D3 a lower value (a shorter bar).
	barWidth := func(id string) float64 {
		d := a.page.Probe().Display()
		w := 0.0
		for _, b := range a.page.Boxes() {
			if b.ID != id {
				continue
			}
			owned := map[uint64]bool{}
			for _, op := range b.OperationIDs {
				owned[op] = true
			}
			for _, op := range d.Ops {
				if owned[op.ID] && op.Kind == ownframe.DisplayOpFillRect && op.W > 0 {
					w = math.Max(w, op.W/d.PointsPerPixel)
				}
			}
		}
		return w
	}
	full, part := barWidth("xl-gc-5-3"), barWidth("xl-gc-2-3") // D6 = 15200, D3 = 8900
	f1, _, _ := sh.DataBar(5, 3)
	f2, _, _ := sh.DataBar(2, 3)
	if full == 0 || part == 0 || f1 != 1 || f2 >= f1 {
		t.Fatalf("data bars: full=%v part=%v fractions %v %v", full, part, f1, f2)
	}
	if got, want := part/full, f2/f1; math.Abs(got-want) > 0.06 {
		t.Errorf("bar width ratio %.3f, want %.3f (the value ratio)", got, want)
	}
}

func TestWrappedRowGrowsInThePicture(t *testing.T) {
	a := rbApp(t)
	rbSelect(a, "A2")
	a.wb.Edit(a.wb.Sheet(), 1, 0, "alpha beta gamma delta epsilon zeta eta theta iota kappa lambda")
	a.runAction("xl:wrap")
	a.redraw()
	var h2, h3 float64
	for _, b := range a.page.Boxes() {
		switch b.ID {
		case "xl-gc-1-0":
			h2 = b.H
		case "xl-gc-2-0":
			h3 = b.H
		}
	}
	if h2 <= h3+10 {
		t.Errorf("wrapped row height %v is not taller than the next row %v", h2, h3)
	}
	if got := a.wb.Sheet().RowHeight(1); got <= defaultRowHeight {
		t.Errorf("RowHeight = %d", got)
	}
}

func TestRibbonCutPasteMovesDependents(t *testing.T) {
	a := rbApp(t)
	sh := a.wb.Sheet()
	rbSelect(a, "C2")
	a.runAction("xl:cut")
	rbSelect(a, "K2")
	a.runAction("xl:paste")
	if sh.Raw(1, 2) != "" || sh.Raw(1, 10) != "12500" {
		t.Fatalf("cut did not move C2 to K2: C2=%q K2=%q", sh.Raw(1, 2), sh.Raw(1, 10))
	}
	// single-cell references follow the cell; a range that only partly lies in
	// the cut block stays where it is, as in Excel
	if got := sh.Raw(1, 7); got != "=IFERROR((F2-K2)/K2,0)" {
		t.Errorf("H2 = %q, want =IFERROR((F2-K2)/K2,0)", got)
	}
	if got := sh.Raw(1, 6); got != "=SUM(C2:F2)" {
		t.Errorf("G2 = %q, a partly cut range should not change", got)
	}
	if !a.wb.Undo() || sh.Raw(1, 2) != "12500" || sh.Raw(1, 7) != "=IFERROR((F2-C2)/C2,0)" {
		t.Errorf("undo did not restore C2 and H2: %q %q", sh.Raw(1, 2), sh.Raw(1, 7))
	}
}
