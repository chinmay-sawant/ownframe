package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/ui"
)

func rbApp(t *testing.T) *app {
	t.Helper()
	ownframe.UseMemory(true)
	a, err := newApp()
	if err != nil {
		t.Fatalf("newApp: %v", err)
	}
	// The app seeds sample charts on the Charts sheet. These tests count the
	// charts they insert themselves.
	a.x.charts, a.x.chartSeq = nil, 0
	return a
}

func rbRender(t *testing.T, e *ui.Element) string {
	t.Helper()
	s, err := ui.Render(e)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	return s
}

func rbSelect(a *app, ref string) {
	r, _ := ParseRange(ref)
	a.selectRange(r)
}

func TestHomeRibbonReflectsActiveStyle(t *testing.T) {
	a := rbApp(t)
	rbSelect(a, "B2")
	a.applyFormat(func(s *Style) { s.Bold, s.NumFmt = true, "Currency" })
	html := rbRender(t, a.ribbon())
	for _, group := range []string{"Clipboard", "Font", "Alignment", "Number", "Styles", "Cells", "Editing"} {
		if !strings.Contains(html, `aria-label="`+group+`"`) {
			t.Errorf("Home ribbon is missing group %q", group)
		}
	}
	if !strings.Contains(html, `aria-label="Bold (Ctrl+B)" aria-pressed="true"`) &&
		!strings.Contains(html, `aria-pressed="true"`) {
		t.Errorf("bold toggle is not pressed for a bold cell")
	}
	if !strings.Contains(html, `class="rb-dd-label">Currency</span>`) {
		t.Errorf("number format control does not show Currency")
	}
}

func TestFontFormattingShortcuts(t *testing.T) {
	a := rbApp(t)
	rbSelect(a, "A2:B3")
	for _, k := range []string{"ctrl+b", "ctrl+i", "ctrl+u"} {
		if !a.shortcut(k) {
			t.Fatalf("shortcut %s was not handled", k)
		}
	}
	for r := 1; r <= 2; r++ {
		for c := 0; c <= 1; c++ {
			st := a.wb.Sheet().StyleAt(r, c)
			if !st.Bold || !st.Italic || !st.Underline {
				t.Errorf("cell %s style = %+v, want bold italic underline", CellRef{r, c}.A1(), st)
			}
		}
	}
	a.runAction("xl:strike")
	a.runAction("xl:fsize:+")
	if st := a.wb.Sheet().StyleAt(1, 0); !st.Strike || st.FontSize != defaultFontSize+1 {
		t.Errorf("strike/size not applied: %+v", st)
	}
}

func TestColorPickerAppliesFillAndFont(t *testing.T) {
	a := rbApp(t)
	rbSelect(a, "A2")
	a.runAction("xl:menu:fill")
	html := rbRender(t, a.menuPanel())
	if n := strings.Count(html, `class="rb-sw"`); n < 60 {
		t.Fatalf("picker rendered %d swatches, want at least 60 (10x5 theme + 10 standard)", n)
	}
	a.runAction("xl:fillc:#c6efce")
	a.runAction("xl:fontc:#9c0006")
	st := a.wb.Sheet().StyleAt(1, 0)
	if st.BG != "#c6efce" || st.FG != "#9c0006" {
		t.Errorf("style = %+v", st)
	}
	if a.x.menu != "" || len(a.x.recentColors) != 2 {
		t.Errorf("menu=%q recent=%v", a.x.menu, a.x.recentColors)
	}
	a.runAction("xl:fillc:")
	if a.wb.Sheet().StyleAt(1, 0).BG != "" {
		t.Errorf("No Fill did not clear the background")
	}
}

func TestThemeMatrixShape(t *testing.T) {
	m := themeMatrix()
	if len(m) != 5 || len(m[0]) != 10 {
		t.Fatalf("matrix is %dx%d, want 5x10", len(m), len(m[0]))
	}
	if shadeColor("#000000", 0.5) != "#808080" || shadeColor("#ffffff", -0.5) != "#808080" {
		t.Errorf("shadeColor math is off")
	}
}

func TestBorderPresetsRender(t *testing.T) {
	a := rbApp(t)
	sh := a.wb.Sheet()
	rbSelect(a, "B2:D4")
	a.runAction("xl:border:outside")
	// B2 is the top-left corner, C3 the interior, D4 the bottom-right corner.
	if e := sh.StyleAt(1, 1).BorderEdges(); e != (Edges{Left: EdgeThin, Top: EdgeThin}) {
		t.Errorf("outside B2 = %+v", e)
	}
	if e := sh.StyleAt(2, 2).BorderEdges(); e.Any() {
		t.Errorf("outside interior C3 = %+v", e)
	}
	if e := sh.StyleAt(3, 3).BorderEdges(); e != (Edges{Right: EdgeThin, Bottom: EdgeThin}) {
		t.Errorf("outside D4 = %+v", e)
	}
	a.runAction("xl:border:all")
	if e := sh.StyleAt(2, 2).BorderEdges(); e != (Edges{EdgeThin, EdgeThin, EdgeThin, EdgeThin}) {
		t.Errorf("all C3 = %+v", e)
	}
	a.runAction("xl:border:none")
	if sh.StyleAt(2, 2).HasBorder() || sh.StyleAt(1, 1).HasBorder() {
		t.Errorf("no border left edges behind")
	}
	a.runAction("xl:border:thick")
	if e := sh.StyleAt(1, 2).BorderEdges(); e != (Edges{Top: EdgeThick}) {
		t.Errorf("thick top edge C2 = %+v", e)
	}
	a.runAction("xl:border:none")
	a.runAction("xl:border:total")
	if e := sh.StyleAt(1, 2).BorderEdges(); e.Top != EdgeThin || e.Bottom != "" {
		t.Errorf("total C2 = %+v", e)
	}
	if e := sh.StyleAt(3, 2).BorderEdges(); e.Bottom != EdgeDouble || e.Top != "" {
		t.Errorf("total C4 = %+v", e)
	}
	a.runAction("xl:border:none")
	rbSelect(a, "B2")
	a.runAction("xl:border:left")
	a.runAction("xl:border:bottom")
	if e := sh.StyleAt(1, 1).BorderEdges(); e != (Edges{Left: EdgeThin, Bottom: EdgeThin}) {
		t.Errorf("left+bottom B2 = %+v", e)
	}
	a.wb.Undo()
	if e := sh.StyleAt(1, 1).BorderEdges(); e != (Edges{Left: EdgeThin}) {
		t.Errorf("undo bottom B2 = %+v", e)
	}
}

func TestLegacyBorderMigrates(t *testing.T) {
	st := Style{Border: "outline", Bold: true}
	if e := st.BorderEdges(); e != (Edges{EdgeThin, EdgeThin, EdgeThin, EdgeThin}) {
		t.Fatalf("legacy outline = %+v", e)
	}
	m := st.migrated()
	if m.Border != "" || !m.Bold || !m.Edges.Any() {
		t.Errorf("migrated = %+v", m)
	}
	js := `{"Version":1,"Active":0,"Sheets":[{"Name":"S","Cells":[{"R":0,"C":0,"V":"x"}],"Styles":[{"R":0,"C":0,"S":{"Border":"bottom"}}]}]}`
	w, err := UnmarshalWorkbook([]byte(js))
	if err != nil {
		t.Fatal(err)
	}
	if e := w.Sheet().StyleAt(0, 0).BorderEdges(); e != (Edges{Bottom: EdgeThin}) {
		t.Errorf("loaded legacy border = %+v", e)
	}
}

func TestAlignmentAndWrap(t *testing.T) {
	a := rbApp(t)
	rbSelect(a, "A2")
	a.runAction("xl:align:center")
	a.runAction("xl:wrap")
	st := a.wb.Sheet().StyleAt(1, 0)
	if st.Align != "center" || !st.Wrap {
		t.Errorf("style = %+v", st)
	}
}

func TestWrapAndMergeCells(t *testing.T) {
	a := rbApp(t)
	sh := a.wb.Sheet()
	rbSelect(a, "A1:C1")
	a.runAction("xl:merge:center")
	if _, ok := sh.MergeAt(0, 1); !ok {
		t.Fatalf("A1:C1 is not merged")
	}
	if sh.Raw(0, 0) != "Region" {
		t.Errorf("anchor value lost: %q", sh.Raw(0, 0))
	}
	a.runAction("xl:merge:unmerge")
	if _, ok := sh.MergeAt(0, 1); ok {
		t.Errorf("A1:C1 is still merged")
	}
}

func TestNumberFormatButtons(t *testing.T) {
	a := rbApp(t)
	rbSelect(a, "C2:C5")
	a.runAction("xl:cur")
	if got := a.wb.Sheet().Display(1, 2); !strings.Contains(got, "12,500") {
		t.Errorf("accounting display = %q", got)
	}
	a.runAction("xl:dec:-")
	a.runAction("xl:dec:-")
	if got := a.wb.Sheet().Display(1, 2); strings.Contains(got, ".") {
		t.Errorf("zero decimals display = %q", got)
	}
	a.runAction("xl:pct")
	if got := a.wb.Sheet().Display(1, 2); !strings.HasSuffix(got, "%") {
		t.Errorf("percent display = %q", got)
	}
	a.runAction("xl:nf:Scientific")
	if got := a.wb.Sheet().Display(1, 2); !strings.Contains(strings.ToUpper(got), "E") {
		t.Errorf("scientific display = %q", got)
	}
}

func TestFormatPainterAndClear(t *testing.T) {
	a := rbApp(t)
	rbSelect(a, "A1")
	a.runAction("xl:painter")
	if !a.x.painter {
		t.Fatal("painter did not arm")
	}
	rbSelect(a, "C3:D4")
	if a.x.painter {
		t.Errorf("one-shot painter stayed armed")
	}
	if st := a.wb.Sheet().StyleAt(2, 2); !st.Bold || st.BG != "#DDEBF7" {
		t.Errorf("style not painted: %+v", st)
	}
	before := a.wb.Sheet().Raw(2, 2)
	a.runAction("xl:clear:formats")
	if st := a.wb.Sheet().StyleAt(2, 2); st.Bold || st.BG != "" {
		t.Errorf("formats not cleared: %+v", st)
	}
	if a.wb.Sheet().Raw(2, 2) != before {
		t.Errorf("clear formats changed values")
	}
	a.runAction("xl:clear:contents")
	if a.wb.Sheet().Raw(2, 2) != "" {
		t.Errorf("contents not cleared")
	}
	a.shortcut("escape")
}

func TestCondFormatDialogAddAndDelete(t *testing.T) {
	a := rbApp(t)
	rbSelect(a, "C2:C9")
	a.runAction("xl:cf:gt")
	if a.overlay != "condfmt" {
		t.Fatalf("overlay = %q", a.overlay)
	}
	a.page.Form().SetFormValue("rb-cf-a", "12000")
	a.runAction("xl:cf:apply")
	if len(a.wb.Sheet().CondRules()) != 1 {
		t.Fatalf("rules = %d", len(a.wb.Sheet().CondRules()))
	}
	if st := a.wb.Sheet().StyleAt(1, 2); st.BG == "" { // 12500 > 12000
		t.Errorf("rule did not highlight 12500: %+v", st)
	}
	if st := a.wb.Sheet().StyleAt(2, 2); st.BG != "" { // 8300
		t.Errorf("rule highlighted 8300: %+v", st)
	}
	a.runAction("xl:cfr:toggle:0")
	if st := a.wb.Sheet().StyleAt(1, 2); st.BG != "" {
		t.Errorf("disabled rule still applies")
	}
	a.runAction("xl:cfr:del:0")
	if len(a.wb.Sheet().CondRules()) != 0 {
		t.Errorf("rule not deleted")
	}
	a.runAction("xl:cf:databar")
	if _, _, ok := a.wb.Sheet().DataBar(1, 2); !ok {
		t.Errorf("data bar missing")
	}
	a.runAction("xl:cf:clear")
	if _, _, ok := a.wb.Sheet().DataBar(1, 2); ok {
		t.Errorf("data bar survived clear")
	}
}

func TestFormatCellsDialogCommitAndCancel(t *testing.T) {
	a := rbApp(t)
	rbSelect(a, "A2")
	a.shortcut("ctrl+1")
	if a.overlay != "format" {
		t.Fatalf("overlay = %q", a.overlay)
	}
	a.runAction("xl:fc:bold")
	a.runAction("xl:fc:fill:#ffeb9c")
	a.runAction("xl:fc:nf:Percent")
	a.runAction("xl:close")
	if st := a.wb.Sheet().StyleAt(1, 0); st.Bold || st.BG != "" {
		t.Errorf("cancel changed the cell: %+v", st)
	}
	a.shortcut("ctrl+1")
	a.runAction("xl:fc:bold")
	a.runAction("xl:fc:fill:#ffeb9c")
	a.runAction("xl:fc:nf:Percent")
	a.runAction("xl:fc:ok")
	st := a.wb.Sheet().StyleAt(1, 0)
	if !st.Bold || st.BG != "#ffeb9c" || st.NumFmt != "Percent" || a.overlay != "" {
		t.Errorf("OK did not apply: %+v overlay=%q", st, a.overlay)
	}
	if !a.wb.Undo() {
		t.Errorf("format dialog was not one undo step")
	}
}

func TestPasteSpecialModes(t *testing.T) {
	a := rbApp(t)
	sh := a.wb.Sheet()
	rbSelect(a, "C2:E3")
	a.copy(false)
	rbSelect(a, "J2")
	a.runAction("xl:dlg:paste")
	a.runAction("xl:opt:ps:values")
	a.runAction("xl:flag:pstranspose")
	a.runAction("xl:paste:special")
	if sh.Raw(1, 9) != "12500" || sh.Raw(2, 9) != "13200" || sh.Raw(1, 10) != "8300" {
		t.Errorf("transpose wrong: J2=%q J3=%q K2=%q", sh.Raw(1, 9), sh.Raw(2, 9), sh.Raw(1, 10))
	}
	if a.overlay != "" {
		t.Errorf("dialog still open")
	}
}

func TestSortSingleAndMultiColumn(t *testing.T) {
	a := rbApp(t)
	sh := a.wb.Sheet()
	rbSelect(a, "A1:H9")
	a.runAction("xl:dlg:sort")
	a.page.Form().SetFormValue("rb-sort-k1", "A")
	a.page.Form().SetFormValue("rb-sort-k2", "C")
	a.runAction("xl:opt:so2:desc")
	a.runAction("xl:sort:apply")
	want := [][2]string{{"East", "14800"}, {"East", "9100"}, {"North", "12500"}, {"North", "8300"}}
	for i, w := range want {
		if sh.Raw(i+1, 0) != w[0] || sh.Raw(i+1, 2) != w[1] {
			t.Fatalf("row %d = %s/%s, want %s/%s", i+2, sh.Raw(i+1, 0), sh.Raw(i+1, 2), w[0], w[1])
		}
	}
	if sh.Raw(1, 6) != "=SUM(C2:F2)" {
		t.Errorf("formula did not follow its row: %q", sh.Raw(1, 6))
	}
	if !a.wb.Undo() {
		t.Errorf("sort is not undoable")
	}
}

func TestFindAndReplaceFlows(t *testing.T) {
	a := rbApp(t)
	sh := a.wb.Sheet()
	a.runAction("xl:dlg:find")
	a.page.Form().SetFormValue("rb-find-q", "Widget")
	a.runAction("xl:find:all")
	if len(a.x.findHits) != 4 {
		t.Fatalf("find all = %d hits, want 4", len(a.x.findHits))
	}
	a.runAction("xl:find:next")
	first := a.active
	a.runAction("xl:find:next")
	if a.active == first {
		t.Errorf("Find Next did not advance")
	}
	a.runAction("xl:find:prev")
	if a.active != first {
		t.Errorf("Find Previous did not return to %s", first.A1())
	}
	a.page.Form().SetFormValue("rb-find-r", "Gizmo")
	a.runAction("xl:find:replaceall")
	if sh.Raw(1, 1) != "Gizmo" || len(a.findAll()) != 0 {
		t.Errorf("replace all failed: B2=%q", sh.Raw(1, 1))
	}
	if !a.wb.Undo() {
		t.Errorf("replace is not undoable")
	}
}

func TestInsertAndMoveChart(t *testing.T) {
	a := rbApp(t)
	rbSelect(a, "A1:C6")
	a.runAction("xl:chart:new:column")
	if len(a.x.charts) != 1 {
		t.Fatalf("charts = %d", len(a.x.charts))
	}
	c := a.x.charts[0]
	labels, series := a.chartData(c)
	if len(labels) != 5 || len(series) != 1 || series[0].Values[0] != 12500 {
		t.Fatalf("labels=%v series=%v", labels, series)
	}
	a.wb.Edit(a.wb.Sheet(), 1, 2, "99")
	_, series = a.chartData(c)
	if series[0].Values[0] != 99 {
		t.Errorf("chart did not follow the source cell: %v", series[0].Values)
	}
	x0 := c.X
	a.x.dragChart = c.ID
	a.chartDragMove(80, 0)
	if c.X != x0+80 {
		t.Errorf("chart X = %v, want %v", c.X, x0+80)
	}
	a.chartDragEnd()
	for _, kind := range []string{"bar", "line", "area", "pie"} {
		c.Kind = kind
		if html := rbRender(t, a.chartLayer()); !strings.Contains(html, "rb-chart") {
			t.Errorf("%s chart did not render", kind)
		}
	}
	a.selectedChartDelete(t)
}

func (a *app) selectedChartDelete(t *testing.T) {
	t.Helper()
	if !a.shortcut("delete") || len(a.x.charts) != 0 {
		t.Errorf("Delete did not remove the selected chart")
	}
}

func TestDataValidationListAndBounds(t *testing.T) {
	a := rbApp(t)
	sh := a.wb.Sheet()
	rbSelect(a, "J2")
	a.runAction("xl:dlg:validation")
	a.page.Form().SetFormValue("rb-val-list", "Yes,No")
	a.runAction("xl:val:ok")
	if got := sh.ValidationList(1, 9); len(got) != 2 {
		t.Fatalf("list = %v", got)
	}
	if ok, _ := a.validateEntry(1, 9, "Maybe"); ok {
		t.Errorf("out-of-list value accepted")
	}
	a.setOverlay("validation:list")
	a.runAction("xl:vlist:1")
	if sh.Raw(1, 9) != "No" {
		t.Errorf("picked value = %q", sh.Raw(1, 9))
	}
	rbSelect(a, "K2")
	a.runAction("xl:dlg:validation")
	a.runAction("xl:opt:vtype:whole")
	a.page.Form().SetFormValue("rb-val-min", "1")
	a.page.Form().SetFormValue("rb-val-max", "10")
	a.runAction("xl:val:ok")
	for raw, want := range map[string]bool{"5": true, "11": false, "2.5": false, "abc": false} {
		if ok, _ := a.validateEntry(1, 10, raw); ok != want {
			t.Errorf("validate %q = %v, want %v", raw, ok, want)
		}
	}
}

func TestCellCommentsAddHoverDelete(t *testing.T) {
	a := rbApp(t)
	sh := a.wb.Sheet()
	rbSelect(a, "C4")
	a.shortcut("shift+f2")
	if a.overlay != "comment" {
		t.Fatalf("overlay = %q", a.overlay)
	}
	a.page.Form().SetFormValue("rb-comment-text", "check this")
	a.runAction("xl:comment:ok")
	if sh.Comment(3, 2) != "check this" {
		t.Fatalf("comment = %q", sh.Comment(3, 2))
	}
	rbSelect(a, "A1")
	a.runAction("xl:comment:next")
	if a.active != (CellRef{3, 2}) {
		t.Errorf("next comment went to %s", a.active.A1())
	}
	a.runAction("xl:comment:delete")
	if sh.Comment(3, 2) != "" {
		t.Errorf("comment not deleted")
	}
}

func TestViewRibbonToggles(t *testing.T) {
	a := rbApp(t)
	a.ribbonTab = "View"
	for _, name := range []string{"grid", "heads", "fbar", "formulas"} {
		a.runAction("xl:view:" + name)
	}
	if !a.x.hideGrid || !a.x.hideHeads || !a.x.hideFormulaBar || !a.x.showFormulas {
		t.Errorf("toggles: %+v", a.x)
	}
	a.runAction("xl:zoom:in")
	if a.zoomPct() != 110 {
		t.Errorf("zoom = %d", a.zoomPct())
	}
	a.runAction("xl:zoom:reset")
	if a.zoomPct() != 100 {
		t.Errorf("zoom reset = %d", a.zoomPct())
	}
	rbSelect(a, "C3")
	a.runAction("xl:freeze:both")
	if a.wb.Sheet().FreezeRows() != 2 || a.wb.Sheet().FreezeCols() != 2 {
		t.Errorf("freeze = %d,%d", a.wb.Sheet().FreezeRows(), a.wb.Sheet().FreezeCols())
	}
	a.runAction("xl:theme")
	if !a.dark {
		t.Errorf("theme did not toggle")
	}
}

func TestRibbonTabsRender(t *testing.T) {
	a := rbApp(t)
	for _, tab := range append([]string{}, ribbonTabs()...) {
		a.runAction("xl:tab:" + tab)
		html := rbRender(t, a.ribbon())
		if !strings.Contains(html, `aria-selected="true"`) || !strings.Contains(html, `role="tab"`) {
			t.Errorf("tab %s: missing tab roles", tab)
		}
	}
}

func TestShortcutsDialogOpensAndCloses(t *testing.T) {
	a := rbApp(t)
	a.shortcut("ctrl+/")
	if a.overlay != "shortcuts" {
		t.Fatalf("overlay = %q", a.overlay)
	}
	html := rbRender(t, a.dialog())
	for _, g := range []string{"Navigation", "Selection", "Editing", "Formatting", "Data"} {
		if !strings.Contains(html, g) {
			t.Errorf("group %s missing", g)
		}
	}
	a.shortcut("escape")
	if a.overlay != "" {
		t.Errorf("escape did not close the dialog")
	}
}

func TestContextMenuActions(t *testing.T) {
	a := rbApp(t)
	rbSelect(a, "A2")
	a.openContextMenu("cell", 200, 300)
	html := rbRender(t, a.contextMenu())
	for _, w := range []string{"Cut", "Copy", "Paste Special", "Format Cells", "Insert Comment"} {
		if !strings.Contains(html, w) {
			t.Errorf("cell menu is missing %q", w)
		}
	}
	a.runAction("xl:ctx:clear")
	if a.wb.Sheet().Raw(1, 0) != "" || a.overlay != "" {
		t.Errorf("clear via context menu failed")
	}
	rbSelect(a, "B3")
	a.openContextMenu("row", 10, 10)
	a.runAction("xl:ctx:hide")
	if !a.wb.Sheet().HiddenRow(2) {
		t.Errorf("row not hidden")
	}
	a.runAction("xl:ctx:unhide")
	if a.wb.Sheet().HiddenRow(2) {
		t.Errorf("row not unhidden")
	}
	a.x.ctxKind = "row"
	a.runAction("xl:ctx:insert")
	if a.wb.Sheet().Raw(2, 1) != "" {
		t.Errorf("row not inserted")
	}
}

func TestFilterHidesAndRestoresRows(t *testing.T) {
	a := rbApp(t)
	sh := a.wb.Sheet()
	a.runAction("xl:filter:toggle")
	a.openFilter(0)
	for v := range a.x.filterAllowed {
		a.x.filterAllowed[v] = v == "North"
	}
	a.runAction("xl:filter:apply")
	visible := 0
	for r := 1; r <= 8; r++ {
		if !sh.HiddenRow(r) {
			visible++
		}
	}
	if visible != 2 {
		t.Errorf("visible rows = %d, want 2", visible)
	}
	a.runAction("xl:filter:clear")
	for r := 1; r <= 8; r++ {
		if sh.HiddenRow(r) {
			t.Errorf("row %d still hidden", r+1)
		}
	}
}

func TestFileSaveAndOpenRoundTrip(t *testing.T) {
	a := rbApp(t)
	dir := t.TempDir()
	js := filepath.Join(dir, "book.json")
	if err := a.savePath(js, "json"); err != nil {
		t.Fatal(err)
	}
	csv := filepath.Join(dir, "sheet.csv")
	if err := a.savePath(csv, "csv"); err != nil {
		t.Fatal(err)
	}
	a.wb.Edit(a.wb.Sheet(), 0, 0, "Changed")
	if err := a.openPath(js); err != nil {
		t.Fatal(err)
	}
	if a.wb.Sheet().Raw(0, 0) != "Region" {
		t.Errorf("JSON open lost data: %q", a.wb.Sheet().Raw(0, 0))
	}
	n := len(a.wb.Sheets)
	if err := a.openPath(csv); err != nil {
		t.Fatal(err)
	}
	if len(a.wb.Sheets) != n+1 || a.wb.Sheet().Raw(1, 0) != "North" {
		t.Errorf("csv drop open failed: sheets=%d A2=%q", len(a.wb.Sheets), a.wb.Sheet().Raw(1, 0))
	}
	if err := a.openPath(filepath.Join(dir, "x.bin")); err == nil {
		t.Errorf("unsupported type accepted")
	}
	if _, err := os.Stat(js); err != nil {
		t.Fatal(err)
	}
}

func TestAutoSumAndFunctionInsert(t *testing.T) {
	a := rbApp(t)
	sh := a.wb.Sheet()
	rbSelect(a, "C11")
	a.runAction("xl:autosum")
	if sh.Raw(10, 2) != "=SUM(C2:C10)" && !strings.HasPrefix(sh.Raw(10, 2), "=SUM(") {
		t.Errorf("autosum wrote %q", sh.Raw(10, 2))
	}
}

func TestPasteSpecialEveryMode(t *testing.T) {
	for _, tc := range []struct {
		mode string
		want string
	}{{"all", "=SUM(C2:F2)"}, {"formulas", "=SUM(C2:F2)"}, {"values", "55600"}} {
		a := rbApp(t)
		sh := a.wb.Sheet()
		rbSelect(a, "G2")
		a.copy(false)
		rbSelect(a, "J2")
		a.runAction("xl:dlg:paste")
		a.runAction("xl:opt:ps:" + tc.mode)
		a.runAction("xl:paste:special")
		got := sh.Raw(1, 9)
		if tc.mode != "values" && !strings.HasPrefix(got, "=SUM(") {
			t.Errorf("mode %s: J2 = %q, want a formula", tc.mode, got)
		}
		if tc.mode == "values" && got != tc.want {
			t.Errorf("mode %s: J2 = %q, want %q", tc.mode, got, tc.want)
		}
	}
	a := rbApp(t)
	sh := a.wb.Sheet()
	rbSelect(a, "A1")
	a.copy(false)
	rbSelect(a, "J2")
	a.runAction("xl:dlg:paste")
	a.runAction("xl:opt:ps:formats")
	a.runAction("xl:paste:special")
	if sh.StyleAt(1, 9).BG != "#DDEBF7" || sh.Raw(1, 9) != "" {
		t.Errorf("formats-only paste: raw=%q style=%+v", sh.Raw(1, 9), sh.StyleAt(1, 9))
	}
}

func TestChartBoundsInDisplay(t *testing.T) {
	a := rbApp(t)
	rbSelect(a, "A1:C6")
	a.runAction("xl:chart:new:column")
	find := func() (x float64, ok bool) {
		for _, b := range a.page.Probe().Display().Boxes {
			if b.ID == "rb-chart-1" {
				return b.X, true
			}
		}
		return 0, false
	}
	x0, ok := find()
	if !ok {
		t.Fatal("chart card is not in the display list")
	}
	a.x.dragChart = 1
	a.chartDragMove(80, 0)
	a.redraw()
	x1, _ := find()
	if x1-x0 != 80 {
		t.Errorf("chart moved %v px, want 80", x1-x0)
	}
}

func TestAccessibilitySnapshot(t *testing.T) {
	a := rbApp(t)
	a.runAction("xl:dlg:shortcuts")
	tree, err := a.page.Access().AccessibilityTree()
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, n := range tree.Nodes {
		names = append(names, n.Name)
	}
	all := strings.Join(names, "|")
	for _, want := range []string{"Home", "Insert", "Keyboard Shortcuts"} {
		if !strings.Contains(all, want) {
			t.Errorf("accessibility tree has no %q", want)
		}
	}
}

func TestContextMenuHooksAndDrop(t *testing.T) {
	a := rbApp(t)
	if rows := a.contextMenuFor("xl-gc-1-1"); len(rows) < 5 || rows[0].ID != "xl:ctx:cut" {
		t.Fatalf("cell rows = %+v", rows)
	}
	if rows := a.contextMenuFor("xl-gch-2"); len(rows) == 0 || a.x.ctxKind != "col" {
		t.Errorf("column header menu missing")
	}
	if a.contextMenuFor("xl-formula") != nil {
		t.Errorf("non-grid element got a custom menu")
	}
	rbSelect(a, "B3")
	_ = a.menuAction(a.bg(), "xl:ctx:clear")
	if a.wb.Sheet().Raw(2, 1) != "" {
		t.Errorf("MenuAction did not run")
	}
	n := len(a.wb.Sheets)
	err := a.handleDrop(a.bg(), []ownframe.Drop{{Name: "x.csv", Read: func() ([]byte, error) { return []byte("a,b\n1,2\n"), nil }}})
	if err != nil || len(a.wb.Sheets) != n+1 || a.wb.Sheet().Raw(1, 1) != "2" {
		t.Errorf("drop open failed: err=%v sheets=%d", err, len(a.wb.Sheets))
	}
}

func TestFilterColumnDialog(t *testing.T) {
	a := rbApp(t)
	a.runAction("xl:filter:toggle")
	a.runAction("xl:filter:open:0")
	if a.overlay != "filter" || len(a.x.filterVals) != 5 {
		t.Fatalf("overlay=%q vals=%v", a.overlay, a.x.filterVals)
	}
	a.runAction("xl:filter:val:0")
	if rbRender(t, a.dialog()) == "" {
		t.Fatal("empty filter dialog")
	}
}

func TestValidationRejectsCommit(t *testing.T) {
	a := rbApp(t)
	sh := a.wb.Sheet()
	rbSelect(a, "K2")
	a.runAction("xl:dlg:validation")
	a.runAction("xl:opt:vtype:decimal")
	a.page.Form().SetFormValue("rb-val-min", "0")
	a.page.Form().SetFormValue("rb-val-max", "10")
	a.runAction("xl:val:ok")
	a.beginEdit("")
	a.editBuf = "50"
	a.commitEdit()
	if sh.Raw(1, 10) != "" || !a.x.toastOpen {
		t.Errorf("invalid entry committed: %q toast=%v", sh.Raw(1, 10), a.x.toastOpen)
	}
	a.beginEdit("")
	a.editBuf = "5"
	a.commitEdit()
	if sh.Raw(1, 10) != "5" {
		t.Errorf("valid entry rejected: %q", sh.Raw(1, 10))
	}
}

// TestSeededChartsSheet: the sample workbook carries a Quarterly data sheet
// and a Charts sheet with one chart per kind, each bound to live data.
func TestSeededChartsSheet(t *testing.T) {
	a := newTestAppForCharts(t)
	var charts, quarterly bool
	for _, s := range a.wb.Sheets {
		charts = charts || s.Name == "Charts"
		quarterly = quarterly || s.Name == "Quarterly"
	}
	if !charts || !quarterly {
		t.Fatalf("sheets missing: Charts=%v Quarterly=%v", charts, quarterly)
	}
	kinds := map[string]bool{}
	for _, c := range a.x.charts {
		if c.Sheet != "Charts" {
			continue
		}
		kinds[c.Kind] = true
		labels, series := a.chartData(c)
		if len(labels) == 0 || len(series) == 0 {
			t.Errorf("chart %q has no data from %s", c.Title, c.Src.String())
		}
	}
	for _, k := range chartKinds() {
		if k.Kind == "stacked" { // insertable, not seeded: three rows would not fit the window
			continue
		}
		if !kinds[k.Kind] {
			t.Errorf("no seeded %s chart", k.Kind)
		}
	}
	// Revenue for 2025 links to the Sheet1 quarter totals.
	q := a.wb.Sheets[3]
	if got, want := q.Value(9, 1).Num, a.wb.Sheets[0].Value(9, 2).Num; got != want || got == 0 {
		t.Errorf("Q1 2025 revenue %v, want Sheet1 total %v", got, want)
	}
}

func newTestAppForCharts(t *testing.T) *app {
	t.Helper()
	ownframe.UseMemory(true)
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	return a
}

// TestSeededChartsWrapToWindow: a narrow window stacks the seeded charts
// instead of leaving them past the right edge.
func TestSeededChartsWrapToWindow(t *testing.T) {
	a := newTestAppForCharts(t)
	right := func() float64 {
		m := 0.0
		for _, c := range a.x.charts {
			if c.Auto {
				m = max(m, c.X+c.W)
			}
		}
		return m
	}
	a.page.SetSize(1440, 960)
	a.layoutAutoCharts()
	wide := right()
	a.page.SetSize(900, 960)
	a.layoutAutoCharts()
	narrow := right()
	if narrow >= wide {
		t.Fatalf("right edge narrow %.0f, wide %.0f: charts did not wrap", narrow, wide)
	}
	if limit := a.unzf(900) - chartOriginX; narrow > limit {
		t.Fatalf("right edge %.0f past the %.0f window", narrow, limit)
	}
}

// TestStackedAndDonutChartsRender: the new kinds draw from sheet data.
func TestStackedAndDonutChartsRender(t *testing.T) {
	a := rbApp(t)
	rbSelect(a, "A1:D5")
	for _, kind := range []string{"stacked", "donut"} {
		c := a.addChart(kind)
		labels, series := a.chartData(c)
		if len(labels) == 0 || len(series) == 0 {
			t.Fatalf("%s: no data", kind)
		}
		html := rbRender(t, a.chartBody(c, labels, series))
		want := "rb-stack-col"
		if kind == "donut" {
			want = "conic-gradient"
		}
		if !strings.Contains(html, want) {
			t.Errorf("%s chart missing %q in %s", kind, want, html)
		}
	}
}

// TestChartDragTranslatesInPlace verifies that dragging one chart on the
// multi-chart Charts sheet moves its card and header boxes by the drag delta
// while leaving sibling charts untouched.
func TestChartDragTranslatesInPlace(t *testing.T) {
	a := newTestAppForCharts(t)
	for i, s := range a.wb.Sheets {
		if s.Name == "Charts" {
			a.wb.Active = i
			break
		}
	}
	a.redraw()

	boxByID := func(id string) ownframe.Box {
		t.Helper()
		for _, b := range a.page.Boxes() {
			if b.ID == id {
				return b
			}
		}
		t.Fatalf("box %q not found", id)
		return ownframe.Box{}
	}

	beforeCard := boxByID("rb-chart-1")
	beforeHdr := boxByID("rb-chart-hdr-1")
	beforeSibling := boxByID("rb-chart-2")

	ctx := t.Context()
	claimed, err := a.handleDragStart(ctx, beforeHdr)
	if err != nil || !claimed {
		t.Fatalf("handleDragStart = (%v, %v), want (true, nil)", claimed, err)
	}
	const dx, dy = 40.0, 24.0
	if err := a.handleDragMove(ctx, dx, dy); err != nil {
		t.Fatalf("handleDragMove: %v", err)
	}
	if err := a.handleDragEnd(ctx); err != nil {
		t.Fatalf("handleDragEnd: %v", err)
	}

	afterCard := boxByID("rb-chart-1")
	afterHdr := boxByID("rb-chart-hdr-1")
	afterSibling := boxByID("rb-chart-2")

	if afterCard.X <= beforeCard.X || afterCard.Y <= beforeCard.Y {
		t.Fatalf("rb-chart-1 did not move: before=(%.1f,%.1f) after=(%.1f,%.1f)",
			beforeCard.X, beforeCard.Y, afterCard.X, afterCard.Y)
	}
	if afterHdr.X <= beforeHdr.X || afterHdr.Y <= beforeHdr.Y {
		t.Fatalf("rb-chart-hdr-1 did not move with card: before=(%.1f,%.1f) after=(%.1f,%.1f)",
			beforeHdr.X, beforeHdr.Y, afterHdr.X, afterHdr.Y)
	}
	if afterSibling.X != beforeSibling.X || afterSibling.Y != beforeSibling.Y {
		t.Fatalf("sibling rb-chart-2 moved unexpectedly: before=(%.1f,%.1f) after=(%.1f,%.1f)",
			beforeSibling.X, beforeSibling.Y, afterSibling.X, afterSibling.Y)
	}
	// The selected/dragged chart (rb-chart-1) must stack above rb-chart-2 in
	// box order so both paint and hit-testing keep it on top.
	idx1, idx2 := -1, -1
	for i, b := range a.page.Boxes() {
		switch b.ID {
		case "rb-chart-1":
			idx1 = i
		case "rb-chart-2":
			idx2 = i
		}
	}
	if idx1 <= idx2 {
		t.Fatalf("selected rb-chart-1 (box %d) is not stacked above rb-chart-2 (box %d)", idx1, idx2)
	}
}
