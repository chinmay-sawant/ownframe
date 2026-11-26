package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/virtual"

	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/driver"
	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/sqlkit"
)

// gridOverscan is zero: rows are whole and snapped, and rows outside the
// viewport must not exist because filled spans are not clipped by overflow.
const gridOverscan = 0

// gridState holds one result set and its viewport, selection, and widths.
type gridState struct {
	cols      []driver.ColumnMeta
	rows      [][]driver.CellValue
	widths    []int
	scrollY   float64
	scrollX   float64 // unused horizontal pixel offset; columns scroll by hcol
	hcol      int     // first scrolled column
	all       [][]driver.CellValue
	truncated bool
	// viewH is the measured pixel height of the body; zero until the first layout.
	viewH int

	// Selection: anchor and focus cell (row, col); -1 means none. A row
	// selection spans every column.
	ar, ac, fr, fc int
	rowSel         bool
}

func newGridState(res driver.QueryResult) *gridState {
	g := &gridState{cols: res.Columns, rows: res.Rows, all: res.Rows, truncated: res.Truncated, ar: -1, ac: -1, fr: -1, fc: -1}
	g.widths = make([]int, len(g.cols))
	for i, c := range g.cols {
		w := gridHeaderPad + gridNameCharW*len(c.Name)
		if len(c.Type) > 0 {
			w = max(w, gridHeaderPad+gridTypeCharW*len(c.Type))
		}
		for r := 0; r < len(g.rows) && r < 50; r++ {
			if i < len(g.rows[r]) {
				txt, _ := cellDisplay(g.rows[r][i])
				w = max(w, gridCellPad+gridCellCharW*min(len(txt), gridCellChars))
			}
		}
		g.widths[i] = max(gridMinColWidth, min(w, gridMaxColWidth))
	}
	return g
}

func (g *gridState) totalWidth() int {
	w := gutterWidth
	for _, x := range g.widths {
		w += x
	}
	return w
}

// clampScroll keeps the offsets inside the content for a viewW x viewH view.
func (g *gridState) clampScroll(viewW, viewH int, extraRows int) {
	maxY := max(0, float64((len(g.rows)+extraRows)*gridRowHeight-viewH))
	maxX := max(0, float64(g.totalWidth()-viewW))
	g.scrollY = max(0, min(g.scrollY, maxY))
	g.scrollX = max(0, min(g.scrollX, maxX))
}

func (g *gridState) selected(r, c int) bool {
	if g.fr < 0 || g.ar < 0 {
		return false
	}
	r0, r1 := min(g.ar, g.fr), max(g.ar, g.fr)
	c0, c1 := min(g.ac, g.fc), max(g.ac, g.fc)
	if g.rowSel {
		return r >= r0 && r <= r1
	}
	return r >= r0 && r <= r1 && c >= c0 && c <= c1
}

// ---- typed cell rendering ----

func humanBytes(n int) string {
	switch {
	case n < bytesPerKB:
		return fmt.Sprintf("%d B", n)
	case n < bytesPerKB*bytesPerKB:
		return fmt.Sprintf("%.1f KB", float64(n)/bytesPerKB)
	}
	return fmt.Sprintf("%.1f MB", float64(n)/bytesPerKB/bytesPerKB)
}

// cellDisplay returns the visible text and style class for a cell.
func cellDisplay(c driver.CellValue) (string, string) {
	switch c.Kind {
	case driver.KindNull:
		return "NULL", "qr-null"
	case driver.KindInt, driver.KindFloat:
		return c.String(), "qr-num"
	case driver.KindBool:
		return c.String(), map[bool]string{true: "qr-bool qr-bool-true", false: "qr-bool qr-bool-false"}[c.B]
	case driver.KindTime:
		return c.String(), "qr-time"
	case driver.KindJSON:
		return jsonPreview(c.S), "qr-json"
	case driver.KindBlob:
		return fmt.Sprintf("<BLOB %s>", humanBytes(len(c.Blob))), "qr-blob"
	}
	return strings.ReplaceAll(strings.ReplaceAll(c.S, "\n", " "), "\r", ""), ""
}

func jsonPreview(s string) string {
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return strings.ReplaceAll(s, "\n", " ")
	}
	switch x := v.(type) {
	case map[string]any:
		return fmt.Sprintf("{...} %d keys", len(x))
	case []any:
		return fmt.Sprintf("[...] %d items", len(x))
	}
	return s
}

// prettyJSON indents valid JSON; ok is false for other text.
func prettyJSON(s string) (string, bool) {
	var buf bytes.Buffer
	if err := json.Indent(&buf, []byte(s), "", "  "); err != nil {
		return s, false
	}
	return buf.String(), true
}

func cellAny(c driver.CellValue) any {
	switch c.Kind {
	case driver.KindNull:
		return nil
	case driver.KindInt:
		return c.I
	case driver.KindFloat:
		return c.F
	case driver.KindBool:
		return c.B
	case driver.KindTime:
		return c.T
	case driver.KindBlob:
		return c.Blob
	}
	return c.S
}

// parseCell converts edited text back to a cell of the same kind as old.
func parseCell(text string, old driver.CellValue) driver.CellValue {
	switch old.Kind {
	case driver.KindInt:
		var n int64
		if _, err := fmt.Sscan(text, &n); err == nil {
			return driver.Int(n)
		}
	case driver.KindFloat:
		var f float64
		if _, err := fmt.Sscan(text, &f); err == nil {
			return driver.Float(f)
		}
	case driver.KindBool:
		switch strings.ToLower(strings.TrimSpace(text)) {
		case "true", "1", "t":
			return driver.Bool(true)
		case "false", "0", "f":
			return driver.Bool(false)
		}
	case driver.KindJSON:
		return driver.JSON(text)
	case driver.KindTime:
		if tm, err := time.Parse("2006-01-02 15:04:05", text); err == nil {
			return driver.Time(tm)
		}
	}
	if old.Kind == driver.KindNull {
		return driver.Text(text)
	}
	return driver.Text(text)
}

// ---- layout ----

// gridViewH is the pixel height available for grid rows in tab t.
func (a *App) gridViewH(t *workTab) int {
	if t.grid != nil && t.grid.viewH > 0 {
		return t.grid.viewH
	}
	h := a.height - titleBarHeight - statusHeight - tabBarHeight - gridHeadHeight
	if a.prodBanner() != nil {
		h -= 26
	}
	switch t.kind {
	case tabQuery:
		h -= a.editorH + splitHandleH + resultsHeadH
		if t.err != nil {
			h -= 40
		}
	case tabTable:
		h -= tableBarCount*tableBarH + changesetH
		if t.cs != nil && t.cs.Pending() > 0 {
			h -= 40
		}
	}
	return max(gridRowHeight*gridMinViewRows, h)
}

func (a *App) gridViewW(t *workTab) int {
	w := a.width - railWidth
	if a.sidebarOpen {
		w -= a.sidebarW
	}
	if t.inspectorOpen {
		w -= a.inspectorW
	}
	return w
}

// extraRows is the count of staged inserts drawn below the data rows.
func extraRows(t *workTab) int {
	if t.cs == nil {
		return 0
	}
	return len(t.cs.Inserts())
}

// renderGrid draws header, gutter, and the virtualized body of t's grid.
func (a *App) renderGrid(t *workTab) *ui.Element {
	g := t.grid
	if g == nil {
		if t.kind == tabQuery {
			return a.emptyState("grid-empty", "table", "No result set", "Write a query above and run it. Ctrl+Enter runs the statement at the caret, F5 too.",
				"empty-run", "Run query", func(ctx context.Context) error { return a.runFromKey(ctx, false) })
		}
		return a.emptyState("grid-empty", "table", "No rows", "This table has no rows for the current filter.", "", "", nil)
	}
	viewH, viewW := a.gridViewH(t), a.gridViewW(t)
	g.clampScroll(viewW, viewH, extraRows(t))
	nExtra := extraRows(t)
	cols, fixedW, scrollOff, scrollContent, viewport := a.gridLayout(t, g)
	win := virtual.Window(g.scrollY, float64(viewH), gridRowHeight, len(g.rows)+nExtra, gridOverscan)

	head := ui.Div().ID("grid-header").Class("qr-grid-head").Attr("role", ariaRole("row"))
	head.Children(ui.Div().Class("qr-grid-gutter qr-grid-hcell").Text("#"))
	for _, i := range cols {
		i, c := i, g.cols[i]
		sub := c.Type
		if c.PrimaryKey {
			sub = "PK " + sub
		} else if c.ForeignKey {
			sub = "FK " + sub
		}
		var arrow *ui.Element
		if t.sortCol == i+1 && t.sortDir != 0 {
			arrow = a.icon(map[int]string{1: "chevron-up", 2: "chevron-down"}[t.sortDir], toneBrand, iconSmall)
		}
		cell := ui.Div().ID(fmt.Sprintf("grid-h-%d", i)).Class("qr-grid-hcell").Width(ui.Px(float64(g.widths[i]))).
			Attr("role", ariaRole("columnheader")).Attr("aria-label", "Column "+c.Name).Children(
			ui.Div().Class("qr-grid-hline").Children(ui.Span().Class("qr-grid-hname").Text(c.Name), arrow),
			ui.Span().Class("qr-grid-htype qr-muted").Text(strings.TrimSpace(sub)),
		)
		if c.PrimaryKey || c.ForeignKey {
			cell.Class("qr-grid-hcell qr-grid-hkey")
		}
		cell.OnClick(func(ctx context.Context) error { return a.cycleSort(ctx, t, i+1) })
		head.Children(cell)
		head.Children(ui.Div().ID(fmt.Sprintf("colresize-%d", i)).Class("qr-colresize"))
	}

	body := ui.Div().ID("grid-body").Class("qr-grid-body").Height(ui.Px(float64(viewH))).Attr("role", ariaRole("grid")).
		Attr("aria-label", "Result grid").Attr("aria-rowcount", fmt.Sprint(len(g.rows)))
	inner := ui.Div().Class("qr-grid-inner").MarginTop(ui.Px(-(g.scrollY - win.TopSpacer)))
	for r := win.Start; r < win.End; r++ {
		inner.Children(a.gridRow(t, g, r, cols))
	}
	body.Children(inner)
	if hb := a.hbar("grid", "Result grid horizontal scroll", scrollContent, viewport, scrollOff,
		func(o int) { g.hcol = a.colAtOffset(t, g, o) }); hb != nil {
		body.Children(hb.Position(ui.PositionAbsolute).Left(ui.Px(float64(fixedW))))
	}
	body.Children(a.scrollbarFor("grid", "Result grid scroll", (len(g.rows)+nExtra)*gridRowHeight, viewH, int(g.scrollY),
		func(o int) { g.scrollY = float64(o / gridRowHeight * gridRowHeight) }))
	return ui.Div().ID("grid").Class("qr-grid").Children(head, body)
}

func (a *App) rowID(t *workTab, g *gridState, r int) (sqlkit.RowID, bool) {
	if len(t.pk) == 0 || r >= len(g.rows) {
		return sqlkit.RowID{}, false
	}
	var id sqlkit.RowID
	for _, name := range t.pk {
		for ci, c := range g.cols {
			if c.Name == name {
				id.Cols = append(id.Cols, name)
				id.Vals = append(id.Vals, cellAny(g.rows[r][ci]))
			}
		}
	}
	return id, len(id.Cols) == len(t.pk)
}

func (a *App) gridRow(t *workTab, g *gridState, r int, cols []int) *ui.Element {
	cls := "qr-grid-row"
	if r%2 == 1 {
		cls += " qr-grid-alt"
	}
	var cells []driver.CellValue
	var id sqlkit.RowID
	hasID := false
	insertRow := r >= len(g.rows)
	if insertRow {
		ins := t.cs.Inserts()[r-len(g.rows)]
		cells = make([]driver.CellValue, len(g.cols))
		for i, c := range g.cols {
			for j, name := range ins.Cols {
				if name == c.Name {
					cells[i] = anyCell(ins.Vals[j])
				}
			}
		}
		cls += " qr-diff-insert"
	} else {
		cells = g.rows[r]
		id, hasID = a.rowID(t, g, r)
		if hasID && t.cs != nil && t.cs.IsDeleted(id) {
			cls += " qr-diff-delete"
		}
	}
	row := ui.Div().ID(fmt.Sprintf("grid-row-%d", r)).Class(cls).Attr("role", ariaRole("row")).Attr("aria-rowindex", fmt.Sprint(r+1)).
		Attr("aria-label", rowLabel(g, cells, r))
	row.Children(ui.Div().ID(fmt.Sprintf("grid-gutter-%d", r)).Class("qr-grid-gutter").Text(fmt.Sprint(r + 1)).
		OnClick(func(ctx context.Context) error {
			g.ar, g.ac, g.fr, g.fc, g.rowSel = r, 0, r, max(0, len(g.cols)-1), true
			a.focus = "grid"
			return a.redraw(ctx)
		}))
	for _, ci := range cols {
		ci := ci
		var c driver.CellValue
		if ci < len(cells) {
			c = cells[ci]
		}
		text, tcls := cellDisplay(c)
		ccls := "qr-cell " + cellAlign(tcls)
		edited := false
		if hasID && t.cs != nil {
			if v, ok := t.cs.EditedValue(id, g.cols[ci].Name); ok {
				c = anyCell(v)
				text, tcls = cellDisplay(c)
				ccls = "qr-cell qr-diff-edit " + cellAlign(tcls)
				edited = true
			}
		}
		_ = edited
		if g.selected(r, ci) {
			ccls += " qr-cell-sel"
		}
		cell := ui.Div().ID(fmt.Sprintf("cell-%d-%d", r, ci)).Class(ccls).Width(ui.Px(float64(g.widths[ci]))).
			Attr("role", ariaRole("gridcell")).Attr("aria-label", g.cols[ci].Name+": "+text)
		switch {
		case t.editing != nil && t.editing.row == r && t.editing.col == ci:
			cell.Children(a.inlineEditor(t))
		case tcls == "qr-null" || strings.HasPrefix(tcls, "qr-bool") || tcls == "qr-blob" || tcls == "qr-json":
			cell.Children(ui.Span().Class(tcls + " qr-cellpill").Text(text))
		default:
			cell.Text(text)
		}
		cell.OnClick(func(ctx context.Context) error { return a.clickCell(ctx, t, r, ci) })
		row.Children(cell)
	}
	return row
}

func anyCell(v any) driver.CellValue {
	switch x := v.(type) {
	case nil:
		return driver.Null()
	case driver.CellValue:
		return x
	case int64:
		return driver.Int(x)
	case int:
		return driver.Int(int64(x))
	case float64:
		return driver.Float(x)
	case bool:
		return driver.Bool(x)
	case time.Time:
		return driver.Time(x)
	case []byte:
		return driver.Blob(x)
	case string:
		return driver.Text(x)
	}
	return driver.Text(fmt.Sprint(v))
}

// clickCell selects a cell and detects a double click for edit or inspect.
func (a *App) clickCell(ctx context.Context, t *workTab, r, c int) error {
	g := t.grid
	id := fmt.Sprintf("cell-%d-%d", r, c)
	now := a.clock()
	dbl := a.lastClickID == id && now.Sub(a.lastClickAt) < 450*time.Millisecond
	a.lastClickID, a.lastClickAt = id, now
	shift, _ := a.readMods()
	a.focus = "grid"
	if shift && g.ar >= 0 {
		g.fr, g.fc, g.rowSel = r, c, false
	} else {
		g.ar, g.ac, g.fr, g.fc, g.rowSel = r, c, r, c, false
	}
	if dbl {
		if t.writable && t.kind == tabTable && r < len(g.rows) {
			return a.beginInlineEdit(ctx, t, r, c)
		}
		t.inspectorOpen = true
	}
	return a.redraw(ctx)
}

func (a *App) cycleSort(ctx context.Context, t *workTab, col int) error {
	if t.sortCol != col {
		t.sortCol, t.sortDir = col, 1
	} else {
		t.sortDir = (t.sortDir + 1) % sortStates
		if t.sortDir == 0 {
			t.sortCol = 0
		}
	}
	if t.kind == tabQuery { // a result set is sorted in memory
		if t.grid != nil {
			t.grid.rebuild(t.quick, t.sortCol, t.sortDir)
		}
		return a.redraw(ctx)
	}
	t.page = 0
	return a.loadTablePage(ctx, t)
}

// rebuild recomputes rows from the original result: keep rows with a cell
// containing filter (case-insensitive), then order by column sortCol-1.
func (g *gridState) rebuild(filter string, sortCol, dir int) {
	f := strings.ToLower(strings.TrimSpace(filter))
	rows := make([][]driver.CellValue, 0, len(g.all))
	for _, r := range g.all {
		if f == "" || rowContains(r, f) {
			rows = append(rows, r)
		}
	}
	if dir != 0 && sortCol > 0 && sortCol-1 < len(g.cols) {
		c := sortCol - 1
		sort.SliceStable(rows, func(i, j int) bool {
			x, y := rows[i][c], rows[j][c]
			if x.IsNull() || y.IsNull() {
				return !x.IsNull() && y.IsNull() // NULLs sort last
			}
			if dir == sortDesc {
				return compareCells(x, y) > 0
			}
			return compareCells(x, y) < 0
		})
	}
	g.rows = rows
	g.scrollY = 0
	g.ar, g.ac, g.fr, g.fc, g.rowSel = -1, -1, -1, -1, false
}

func rowContains(r []driver.CellValue, f string) bool {
	for _, c := range r {
		if strings.Contains(strings.ToLower(c.String()), f) {
			return true
		}
	}
	return false
}

// compareCells orders two non-NULL cells: numbers and times by value, the
// rest by text.
func compareCells(a, b driver.CellValue) int {
	switch {
	case a.IsNumeric() && b.IsNumeric():
		x, y := cellFloat(a), cellFloat(b)
		switch {
		case x < y:
			return -1
		case x > y:
			return 1
		}
		return 0
	case a.Kind == driver.KindTime && b.Kind == driver.KindTime:
		switch {
		case a.T.Before(b.T):
			return -1
		case a.T.After(b.T):
			return 1
		}
		return 0
	}
	return strings.Compare(strings.ToLower(a.String()), strings.ToLower(b.String()))
}

func cellFloat(c driver.CellValue) float64 {
	if c.Kind == driver.KindInt {
		return float64(c.I)
	}
	return c.F
}

// ---- keyboard, scroll, copy ----

func (a *App) onWheel(ctx context.Context, x, y, dx, dy float64) (bool, error) {
	if key, ok := a.paneAt(x, y); ok {
		unit := float64(wheelPixels)
		if u := a.regions[key].unit; u > 0 {
			unit = float64(u)
		}
		if a.regions[key].st.Wheel(dy, unit) {
			return true, a.redraw(ctx)
		}
		return true, nil
	}
	if a.modal == "palette" {
		a.paletteScroll(-int(dy * wheelEditorLines))
		return true, a.redraw(ctx)
	}
	if a.modal != "" {
		return true, nil
	}
	t := a.activeTabPtr()
	if t == nil {
		return false, nil
	}
	inside := func(id string) bool {
		for _, b := range a.page.Boxes() {
			if b.ID == id {
				return x >= float64(b.X) && x <= float64(b.X+b.W) && y >= float64(b.Y) && y <= float64(b.Y+b.H)
			}
		}
		return false
	}
	if t.kind == tabQuery && t.ed != nil && inside("editor-wrap") {
		t.ed.scroll = max(0, min(t.ed.scroll-int(dy*wheelEditorLines), max(0, t.ed.lineCount()-a.editorRows())))
		if dx != 0 {
			t.ed.hcol = max(0, min(t.ed.hcol-int(dx*wheelEditorLines), max(0, t.ed.longestLine()-a.editorViewCols()+edHScrollLead)))
		}
		return true, a.redraw(ctx)
	}
	if t.grid != nil && (t.kind == tabQuery || t.kind == tabTable) && inside("grid-body") {
		t.grid.scrollY -= dy * wheelPixels
		if dx != 0 {
			step := 1
			if dx > 0 {
				step = -1
			}
			t.grid.hcol = max(0, min(t.grid.hcol+step, a.maxHcol(t, t.grid)))
		}
		t.grid.clampScroll(a.gridViewW(t), a.gridViewH(t), extraRows(t))
		return true, a.redraw(ctx)
	}
	return false, nil
}

// gridKey handles arrows, paging, and Home/End on the active grid.
func (a *App) gridKey(ctx context.Context, t *workTab, key string, shift bool) error {
	g := t.grid
	if g == nil || len(g.rows) == 0 {
		return nil
	}
	if a.focus != "grid" {
		return nil
	}
	name := strings.TrimPrefix(key, "shift+")
	if strings.HasPrefix(key, "shift+") {
		shift = true
	}
	r, c := g.fr, g.fc
	if r < 0 {
		r, c = 0, 0
	}
	last := len(g.rows) - 1
	pageRows := max(1, a.gridViewH(t)/gridRowHeight-1)
	switch name {
	case "arrowup":
		r--
	case "arrowdown":
		r++
	case "arrowleft":
		c--
	case "arrowright":
		c++
	case "pageup":
		r -= pageRows
	case "pagedown":
		r += pageRows
	case "home":
		c = 0
	case "end":
		c = len(g.cols) - 1
	case "delete":
		if t.writable && g.fr >= 0 {
			return a.stageDeleteSelected(ctx, t)
		}
		return nil
	case "enter", "f2":
		if g.fr < 0 {
			return nil
		}
		a.suppressSubmit = key == "enter" // the Enter key also sends a Submit
		if t.writable && t.kind == tabTable && g.fr < len(g.rows) {
			return a.beginInlineEdit(ctx, t, g.fr, g.fc)
		}
		t.inspectorOpen = true
		return a.redraw(ctx)
	default:
		return nil
	}
	r, c = max(0, min(r, last)), max(0, min(c, len(g.cols)-1))
	g.fr, g.fc = r, c
	if !shift || g.ar < 0 {
		g.ar, g.ac = r, c
	}
	g.rowSel = false
	a.revealCol(t, g, c)
	// Keep the focus row inside the window.
	top := float64(r * gridRowHeight)
	if top < g.scrollY {
		g.scrollY = top
	} else if top+gridRowHeight > g.scrollY+float64(a.gridViewH(t)) {
		g.scrollY = top + gridRowHeight - float64(a.gridViewH(t))
	}
	g.clampScroll(a.gridViewW(t), a.gridViewH(t), extraRows(t))
	return a.redraw(ctx)
}

// gridCopyText formats the selection: one cell as plain text, a block as TSV.
func (a *App) gridCopyText(t *workTab) string {
	g := t.grid
	if g == nil || g.ar < 0 || g.fr < 0 {
		return ""
	}
	r0, r1 := min(g.ar, g.fr), max(g.ar, g.fr)
	c0, c1 := min(g.ac, g.fc), max(g.ac, g.fc)
	if g.rowSel {
		c0, c1 = 0, len(g.cols)-1
	}
	if r0 == r1 && c0 == c1 {
		return cellPlain(g.rows[r0][c0])
	}
	var lines []string
	for r := r0; r <= r1 && r < len(g.rows); r++ {
		var f []string
		for c := c0; c <= c1; c++ {
			f = append(f, cellPlain(g.rows[r][c]))
		}
		lines = append(lines, strings.Join(f, "\t"))
	}
	return strings.Join(lines, "\n")
}

func cellPlain(c driver.CellValue) string {
	if c.IsNull() {
		return ""
	}
	return c.String()
}

// ---- results pane (query tab) ----

func (a *App) resultsPane(t *workTab) *ui.Element {
	sub := func(id, label string) *ui.Element {
		cls := "qr-seg-item"
		if t.resultView == id {
			cls += " qr-seg-on"
		}
		return ui.Div().ID("rs-"+id).Class(cls).Text(label).Attr("role", ariaRole("tab")).
			OnClick(func(ctx context.Context) error { t.resultView = id; return a.redraw(ctx) })
	}
	head := ui.Div().ID("results-head").Class("qr-results-head").Children(
		ui.Div().Class("qr-seg").Children(sub("results", "Results"), sub("explain", "Explain"), sub("messages", "Messages")),
	)
	if t.grid != nil {
		info := fmt.Sprintf("%d rows", len(t.grid.rows))
		if len(t.grid.rows) != len(t.grid.all) {
			info = fmt.Sprintf("%d of %d rows", len(t.grid.rows), len(t.grid.all))
		}
		if t.grid.truncated {
			info += " (truncated)"
		}
		head.Children(ui.Span().ID("results-info").Class("qr-muted qr-nowrap").Text(info))
	}
	head.Children(ui.Div().Class("qr-spacer"))
	if t.grid != nil && t.resultView == "results" {
		head.Children(
			ui.Input().ID("grid-filter").Class("qr-input qr-quick").Attr("type", "text").Attr("placeholder", "Filter rows").
				Attr("aria-label", "Filter result rows").Attr("value", t.quick),
			a.tbButton("rs-freeze", freezeLabel(t), "qr-btn qr-btn-ghost", false, func(ctx context.Context) error {
				t.freeze = !t.freeze
				return a.redraw(ctx)
			}),
		)
	}
	if t.grid != nil {
		head.Children(
			a.tbButton("rs-inspector", "Inspector", "qr-btn", false, func(ctx context.Context) error {
				t.inspectorOpen = !t.inspectorOpen
				return a.redraw(ctx)
			}),
			a.tbButton("rs-export", "Export", "qr-btn", false, func(ctx context.Context) error { return a.openExport(ctx, false) }),
		)
	}
	pane := ui.Div().ID("results-pane").Class("qr-results").Children(head)
	if t.err != nil {
		pane.Children(a.errorBanner(t))
	}
	var content *ui.Element
	switch t.resultView {
	case "explain":
		content = a.explainView(t.plan, t.rawPlan, func(ctx context.Context) error { t.rawPlan = !t.rawPlan; return a.redraw(ctx) })
	case "messages":
		list := ui.Div().ID("messages").Class("qr-messages")
		if len(t.messages) == 0 {
			list.Children(ui.Div().Class("qr-empty").Text("No messages yet."))
		}
		for i, m := range t.messages {
			list.Children(ui.Div().ID(fmt.Sprintf("msg-%d", i)).Class("qr-msg qr-mono").Text(m))
		}
		content = a.pane("messages", "Messages scroll", "", list)
	default:
		content = ui.Div().Class("qr-split-h").Children(a.renderGrid(t), a.inspectorPane(t))
	}
	return pane.Children(content)
}

func (a *App) errorBanner(t *workTab) *ui.Element {
	e := t.err
	title := "Error"
	if e.Code != "" {
		title = "Error " + e.Code
	}
	b := ui.Div().ID("query-error").Class("qr-errbanner").Attr("role", ariaRole("alert")).Children(
		ui.Span().Class("qr-badge qr-badge-err").Text(title),
		ui.Span().Class("qr-errmsg").Text(e.Message),
	)
	if e.Line > 0 {
		b.Children(ui.Span().Class("qr-muted").Text(fmt.Sprintf("line %d, col %d", e.Line, max(1, e.Col))))
		b.Children(a.tbButton("err-jump", "Go to error", "qr-btn", false, func(ctx context.Context) error { return a.jumpToError(ctx) }))
	}
	return b
}

// ---- table tab ----

func (a *App) tableTabView(t *workTab) *ui.Element {
	filterBar := ui.Div().ID("table-filter-bar").Class("qr-toolbar").Children(
		a.icon("filter", toneMuted, iconSmall),
		ui.Span().Class("qr-muted qr-nowrap").Text("WHERE"),
		ui.Input().ID("tbl-filter").Class("qr-input qr-filter-where").Attr("type", "text").
			Attr("placeholder", "status = 'paid' AND total > 100").Attr("aria-label", "Row filter").
			Attr("value", t.filter),
		a.tbButton("tbl-apply", "Apply", "qr-btn qr-btn-primary", false, func(ctx context.Context) error {
			t.filter = a.page.Form().FormValue("tbl-filter")
			t.page = 0
			return a.loadTablePage(ctx, t)
		}),
		a.tbButton("tbl-refresh", "Refresh", "qr-btn", false, func(ctx context.Context) error { return a.loadTablePage(ctx, t) }),
	)
	bar := ui.Div().ID("table-bar").Class("qr-toolbar").Children(ui.Span().Class("qr-muted qr-nowrap").Text("Rows"))
	sizes := ui.Div().Class("qr-seg").Attr("role", ariaRole("group")).Attr("aria-label", "Rows per page")
	for _, n := range []int{100, 500, 1000} {
		n := n
		cls := "qr-seg-item"
		if t.pageSize == n {
			cls += " qr-seg-on"
		}
		sizes.Children(ui.Div().ID(fmt.Sprintf("tbl-size-%d", n)).Class(cls).Text(fmt.Sprint(n)).Attr("role", "button").
			OnClick(func(ctx context.Context) error { t.pageSize, t.page = n, 0; return a.loadTablePage(ctx, t) }))
	}
	bar.Children(sizes)
	pages := max(1, int((t.totalRows+int64(t.pageSize)-1)/int64(max(1, t.pageSize))))
	bar.Children(
		a.tbButton("tbl-prev", "Prev", "qr-btn", t.page <= 0, func(ctx context.Context) error { t.page--; return a.loadTablePage(ctx, t) }),
		ui.Span().ID("tbl-page").Class("qr-muted qr-nowrap").Text(fmt.Sprintf("Page %d / %d  (%d rows)", t.page+1, pages, t.totalRows)),
		a.tbButton("tbl-next", "Next", "qr-btn", t.page+1 >= pages, func(ctx context.Context) error { t.page++; return a.loadTablePage(ctx, t) }),
		ui.Div().Class("qr-spacer"),
	)
	if t.writable {
		bar.Children(a.tbButton("tbl-add", "Add Row", "qr-btn", false, func(ctx context.Context) error { return a.stageInsert(ctx, t) }))
	}
	bar.Children(
		a.freezeIconButton(t),
		a.tbButton("tbl-import", "Import CSV", "qr-btn", a.cfg.ReadOnly, func(ctx context.Context) error { return a.openImport(ctx) }),
		a.tbButton("tbl-export", "Export", "qr-btn", false, func(ctx context.Context) error { return a.openExport(ctx, true) }),
		a.tbButton("tbl-inspector", "Inspector", "qr-btn", false, func(ctx context.Context) error {
			t.inspectorOpen = !t.inspectorOpen
			return a.redraw(ctx)
		}),
	)
	view := ui.Div().ID("table-tab").Class("qr-tab-body").Children(filterBar, bar)
	if t.err != nil {
		view.Children(a.errorBanner(t))
	}
	if !t.writable && t.kind == tabTable {
		view.Children(ui.Div().ID("tbl-ro-note").Class("qr-note qr-pad").Text(t.roReason()))
	}
	view.Children(ui.Div().Class("qr-split-h").Children(a.renderGrid(t), a.inspectorPane(t)))
	if t.cs != nil && t.cs.Pending() > 0 {
		view.Children(a.changesetBar(t))
	}
	return view
}

func (t *workTab) roReason() string {
	switch {
	case len(t.pk) == 0:
		return "Read-only: this table has no primary key, so rows cannot be edited safely."
	default:
		return "Read-only: the connection is read-only. Unlock writes from the title bar to edit rows."
	}
}

// sortedSelectionRows returns selected row indexes in order.
func (g *gridState) sortedSelectionRows() []int {
	if g.ar < 0 {
		return nil
	}
	var out []int
	for r := min(g.ar, g.fr); r <= max(g.ar, g.fr); r++ {
		out = append(out, r)
	}
	sort.Ints(out)
	return out
}

// ---- right-click menu ----

func (a *App) onHover(ctx context.Context, previous, next ownframe.Box) (bool, error) {
	return false, nil
}

func (a *App) onPress(ctx context.Context, box ownframe.Box, x, y float64) (bool, error) {
	if key, ok := a.trackKey(box.ID); ok {
		return true, a.pressTrack(ctx, key, x, y)
	}
	if a.modal != "" || box.ID == "" {
		return false, nil
	}
	if a.ac.open && !strings.HasPrefix(box.ID, "ac-") && !strings.HasPrefix(box.ID, "ed-line-") && box.ID != "editor-body" && box.ID != "editor-wrap" {
		a.ac.open = false // a press anywhere else is a blur
		if err := a.redraw(ctx); err != nil {
			return false, err
		}
	}
	if strings.HasPrefix(box.ID, "ed-line-") || box.ID == "editor-body" || box.ID == "editor-wrap" {
		a.focus = "editor"
		shift, _ := a.readMods()
		a.pressEditor(box, x, y, shift)
		a.pressedEditor = true
		return false, a.redraw(ctx)
	}
	a.pressedEditor = false
	return false, nil
}

// onClick receives clicks that no builder callback claimed; they decide which
// surface owns the keyboard.
func (a *App) onClick(ctx context.Context, box ownframe.Box) error {
	switch box.Action {
	case "win-min":
		a.page.Window().Minimize()
		return nil
	case "win-max":
		a.page.Window().ToggleMaximize()
		return nil
	case "win-close":
		a.page.Window().Close()
		return nil
	}
	switch {
	case strings.HasPrefix(box.ID, "ed-line-") || box.ID == "editor-body" || box.ID == "editor-wrap":
		a.focus = "editor"
	case strings.HasPrefix(box.ID, "grid") || strings.HasPrefix(box.ID, "cell-"):
		a.focus = "grid"
	}
	return nil
}

// cellAlign keeps only the layout part of a cell's type class. The tint of
// NULL, boolean, JSON and BLOB values belongs to the inner pill, never the cell.
func cellAlign(tcls string) string {
	if tcls == "qr-num" || tcls == "qr-time" {
		return tcls
	}
	return ""
}

// rowLabel is the accessible name of a grid row: its number and every value.
func rowLabel(g *gridState, cells []driver.CellValue, r int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Row %d", r+1)
	for i, c := range g.cols {
		if i >= len(cells) {
			break
		}
		txt, _ := cellDisplay(cells[i])
		fmt.Fprintf(&b, ", %s %s", c.Name, txt)
	}
	return b.String()
}

// emptyState is a centred icon, title and hint, with an optional call to action.
func (a *App) emptyState(id, icon, title, hint, btnID, btnLabel string, fn func(context.Context) error) *ui.Element {
	e := ui.Div().ID(id).Class("qr-empty").Children(
		ui.Div().Class("qr-empty-icon").Children(a.icon(icon, toneBrand, emptyIconSize)),
		ui.Div().Class("qr-empty-title").Text(title),
		ui.Div().Class("qr-empty-hint").Text(hint),
	)
	if btnID != "" {
		e.Children(a.tbButton(btnID, btnLabel, "qr-btn qr-btn-primary", false, fn))
	}
	return e
}

// gridLayout decides which columns are painted. Columns scroll one whole
// column at a time (so no cell is ever cut at an edge), the row-number gutter
// and an optionally frozen first column stay put, and only columns that fully
// fit are rendered. It returns the painted columns (frozen first), the fixed
// width, the pixel offset and total width of the scrolled columns, and the
// scrolled viewport width.
func (a *App) gridLayout(t *workTab, g *gridState) (cols []int, fixedW, off, content, viewport int) {
	fixedW = gutterWidth
	start := 0
	if t.freeze && len(g.cols) > 1 {
		cols = append(cols, 0)
		fixedW += g.widths[0]
		start = 1
	}
	viewport = max(gridMinColWidth, a.gridViewW(t)-fixedW)
	for i := start; i < len(g.widths); i++ {
		content += g.widths[i]
	}
	first := min(max(g.hcol, start), a.maxHcol(t, g))
	g.hcol = first
	for i := start; i < first; i++ {
		off += g.widths[i]
	}
	used := 0
	shown := 0
	for i := first; i < len(g.cols); i++ {
		if used+g.widths[i] > viewport && shown > 0 {
			break
		}
		cols = append(cols, i)
		used += g.widths[i]
		shown++
	}
	return cols, fixedW, off, content, viewport
}

// maxHcol is the first scrolled column at which every remaining column fits,
// so scrolling can always reach the last column.
func (a *App) maxHcol(t *workTab, g *gridState) int {
	start, fixed := 0, gutterWidth
	if t.freeze && len(g.cols) > 1 {
		start, fixed = 1, fixed+g.widths[0]
	}
	viewport := max(gridMinColWidth, a.gridViewW(t)-fixed)
	sum, i := 0, len(g.widths)
	for i > start && sum+g.widths[i-1] <= viewport {
		i--
		sum += g.widths[i]
	}
	return max(start, min(i, len(g.cols)-1))
}

// colAtOffset maps a horizontal pixel offset to the nearest column start.
func (a *App) colAtOffset(t *workTab, g *gridState, o int) int {
	start := 0
	if t.freeze && len(g.cols) > 1 {
		start = 1
	}
	if o >= g.contentWidth(start)-max(gridMinColWidth, a.gridViewW(t)-g.fixedWidth(t)) {
		return a.maxHcol(t, g)
	}
	pos, best := 0, start
	for i := start; i < len(g.widths); i++ {
		if o >= pos+g.widths[i]/2 {
			best = i
		}
		pos += g.widths[i]
	}
	return best
}

// revealCol scrolls horizontally just far enough to show column c fully.
func (a *App) revealCol(t *workTab, g *gridState, c int) {
	if t.freeze && c == 0 {
		return
	}
	if c < g.hcol {
		g.hcol = c
		return
	}
	for g.hcol < c {
		cols, _, _, _, _ := a.gridLayout(t, g)
		seen := false
		for _, x := range cols {
			if x == c {
				seen = true
			}
		}
		if seen {
			return
		}
		g.hcol++
	}
}

func freezeLabel(t *workTab) string {
	if t.freeze {
		return "Unfreeze"
	}
	return "Freeze column"
}

// contentWidth is the summed width of columns from start on.
func (g *gridState) contentWidth(start int) int {
	w := 0
	for i := start; i < len(g.widths); i++ {
		w += g.widths[i]
	}
	return w
}

// fixedWidth is the gutter plus a frozen first column.
func (g *gridState) fixedWidth(t *workTab) int {
	if t.freeze && len(g.cols) > 1 {
		return gutterWidth + g.widths[0]
	}
	return gutterWidth
}

// freezeIconButton is the compact freeze toggle for the crowded table toolbar.
func (a *App) freezeIconButton(t *workTab) *ui.Element {
	cls := "qr-iconbtn"
	tone := toneMuted
	if t.freeze {
		cls += " qr-iconbtn-on"
		tone = toneBrand
	}
	return ui.Button("").ID("tbl-freeze").Class(cls).Attr("type", "button").
		Attr("aria-label", freezeLabel(t)).Attr("title", freezeLabel(t)).Attr("aria-pressed", fmt.Sprint(t.freeze)).
		Children(a.icon("lock", tone, iconMedium)).
		OnClick(func(ctx context.Context) error { t.freeze = !t.freeze; return a.redraw(ctx) })
}
