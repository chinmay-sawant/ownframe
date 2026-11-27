package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/driver"
)

// wideQuery returns a SELECT of n integer columns over three rows.
func wideQuery(n int) string {
	var cols []string
	for i := 1; i <= n; i++ {
		cols = append(cols, fmt.Sprintf("%d AS col_%02d", i, i))
	}
	sel := "SELECT " + strings.Join(cols, ", ")
	return sel + " UNION ALL " + sel + " UNION ALL " + sel
}

func (r *rig) gridCols(prefix string) []int {
	var out []int
	for _, b := range r.a.Page().Boxes() {
		if strings.HasPrefix(b.ID, prefix) {
			var n int
			if _, err := fmt.Sscanf(b.ID[len(prefix):], "%d", &n); err == nil && !strings.Contains(b.ID[len(prefix):], "-") {
				out = append(out, n)
			}
		}
	}
	return out
}

func TestGridHorizontalScrollKeepsGutterAndHeader(t *testing.T) {
	r := newRig(t, 1280, 800, "")
	r.must(r.a.RunSQL(r.ctx, wideQuery(30)))
	tab := r.a.activeTabPtr()
	body := r.box("grid-body")
	if !r.has("grid-hthumb") {
		t.Fatal("a wide result needs a horizontal scrollbar")
	}
	first := func() int {
		best := 1 << 20
		for _, c := range r.gridCols("grid-h-") {
			best = min(best, c)
		}
		return best
	}
	gutter0 := r.box("grid-gutter-0")
	if first() != 0 {
		t.Fatal("scrolling starts at the first column")
	}
	_, err := r.a.onWheel(r.ctx, body.X+120, body.Y+60, 3, 0)
	r.must(err)
	if tab.grid.hcol != 0 {
		t.Fatalf("scrolling back past the first column moves nothing, hcol=%d", tab.grid.hcol)
	}
	for i := 0; i < 5; i++ {
		_, err = r.a.onWheel(r.ctx, body.X+120, body.Y+60, -3, 0)
		r.must(err)
	}
	if tab.grid.hcol != 5 || first() != 5 {
		t.Fatalf("five notches should hide five columns, hcol=%d first=%d", tab.grid.hcol, first())
	}
	// Header and body paint the same columns, and the gutter has not moved.
	hc, cc := r.gridCols("grid-h-"), map[int]bool{}
	for _, b := range r.a.Page().Boxes() {
		var row, col int
		if n, _ := fmt.Sscanf(b.ID, "cell-0-%d", &col); n == 1 && !strings.Contains(b.ID[7:], "-") {
			cc[col] = true
		}
		_ = row
	}
	for _, c := range hc {
		if !cc[c] {
			t.Fatalf("header shows column %d but row 0 has no such cell", c)
		}
	}
	if len(hc) != len(cc) {
		t.Fatalf("header columns %v vs cells %v", hc, cc)
	}
	if g := r.box("grid-gutter-0"); g.X != gutter0.X || g.W != gutter0.W {
		t.Fatalf("gutter moved: %+v -> %+v", gutter0, g)
	}
	// No painted cell is cut by the right edge of the grid.
	for _, b := range r.a.Page().Boxes() {
		if strings.HasPrefix(b.ID, "cell-0-") && b.X+b.W > body.X+body.W+0.5 {
			t.Fatalf("cell %s straddles the right edge: %.0f > %.0f", b.ID, b.X+b.W, body.X+body.W)
		}
	}
	// Dragging the thumb to the end shows the last column; to the start, the first.
	thumb := r.box("grid-hthumb")
	claimed, err := r.a.onDragStart(r.ctx, thumb)
	r.must(err)
	if !claimed {
		t.Fatal("thumb claims the drag")
	}
	r.must(r.a.onDragMove(r.ctx, 1e5, 0))
	r.must(r.a.onDragEnd(r.ctx))
	if !r.has("cell-0-29") {
		t.Fatal("the last column should be visible after dragging to the end")
	}
	_, _ = r.a.onDragStart(r.ctx, r.box("grid-hthumb"))
	r.must(r.a.onDragMove(r.ctx, -1e5, 0))
	r.must(r.a.onDragEnd(r.ctx))
	if tab.grid.hcol != 0 || !r.has("cell-0-0") {
		t.Fatalf("dragging back to the start: hcol=%d", tab.grid.hcol)
	}
	// A track press jumps.
	after := r.box("grid-hbar-after")
	claim, err := r.a.onPress(r.ctx, after, after.X+after.W-3, after.Y+2)
	r.must(err)
	if !claim || tab.grid.hcol == 0 {
		t.Fatalf("track press: claim=%v hcol=%d", claim, tab.grid.hcol)
	}
	// Narrow results have no bar.
	r.must(r.a.RunSQL(r.ctx, "SELECT 1 AS a, 2 AS b"))
	if r.has("grid-hthumb") {
		t.Fatal("no horizontal bar when the columns fit")
	}
}

func TestGridFreezeFirstColumnAndKeyboardReveal(t *testing.T) {
	r := newRig(t, 1280, 800, "")
	r.must(r.a.RunSQL(r.ctx, wideQuery(30)))
	tab := r.a.activeTabPtr()
	r.tap("rs-freeze")
	if !tab.freeze {
		t.Fatal("freeze toggle")
	}
	r.tap("cell-0-0")
	x0 := r.box("cell-0-0").X
	for i := 0; i < 12; i++ {
		r.key("arrowright")
	}
	if tab.grid.fc != 12 {
		t.Fatalf("focus column %d", tab.grid.fc)
	}
	if !r.has("cell-0-12") {
		t.Fatal("arrowing right must scroll the focused column into view")
	}
	if !r.has("cell-0-0") || r.box("cell-0-0").X != x0 {
		t.Fatal("the frozen column must stay where it is")
	}
	if !r.has("grid-h-0") {
		t.Fatal("the frozen header must stay")
	}
	if r.has("cell-0-1") {
		t.Fatal("column 1 scrolled away behind the frozen column")
	}
	for i := 0; i < 30; i++ {
		r.key("arrowleft")
	}
	if tab.grid.fc != 0 || tab.grid.hcol > 1 || !r.has("cell-0-1") {
		t.Fatalf("back at the start: fc=%d hcol=%d", tab.grid.fc, tab.grid.hcol)
	}
	r.tap("rs-freeze")
	if tab.freeze {
		t.Fatal("unfreeze")
	}
}

func TestGridColumnResizeByDraggingHeaderEdge(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.must(r.a.RunSQL(r.ctx, "SELECT id, full_name, email FROM users"))
	before := r.box("grid-h-1").W
	handle := r.box("colresize-1")
	claimed, err := r.a.onDragStart(r.ctx, handle)
	r.must(err)
	if !claimed {
		t.Fatal("the header edge claims the drag")
	}
	r.must(r.a.onDragMove(r.ctx, 60, 0))
	r.must(r.a.onDragEnd(r.ctx))
	if got := r.box("grid-h-1").W; got != before+60 {
		t.Fatalf("header width %v -> %v, want +60", before, got)
	}
	if got := r.box("cell-0-1").W; got != before+60 {
		t.Fatalf("cells follow the header: %v", got)
	}
	_, _ = r.a.onDragStart(r.ctx, r.box("colresize-1"))
	r.must(r.a.onDragMove(r.ctx, -1000, 0))
	r.must(r.a.onDragEnd(r.ctx))
	if got := r.box("grid-h-1").W; got != float64(minColW) {
		t.Fatalf("minimum column width %v, want %d", got, minColW)
	}
}

func TestGridClientSideSort(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.must(r.a.RunSQL(r.ctx, "SELECT id, full_name, CASE WHEN id % 7 = 0 THEN NULL ELSE country END AS country FROM users WHERE id <= 20"))
	g := r.a.activeTabPtr().grid
	orig := make([]int64, len(g.rows))
	for i, row := range g.rows {
		orig[i] = row[0].I
	}
	r.tap("grid-h-0") // ascending
	for i := 1; i < len(g.rows); i++ {
		if g.rows[i-1][0].I > g.rows[i][0].I {
			t.Fatal("ascending order broken")
		}
	}
	r.tap("grid-h-0") // descending
	if g.rows[0][0].I != 20 || g.rows[19][0].I != 1 {
		t.Fatalf("descending: first %d last %d", g.rows[0][0].I, g.rows[19][0].I)
	}
	r.tap("grid-h-0") // none
	for i, row := range g.rows {
		if row[0].I != orig[i] {
			t.Fatal("the third click restores the original order")
		}
	}
	// Text sorts case-insensitively and NULLs always sort last.
	r.tap("grid-h-2")
	seenNull := false
	for _, row := range g.rows {
		if row[2].IsNull() {
			seenNull = true
		} else if seenNull {
			t.Fatal("a non-NULL row follows a NULL row")
		}
	}
	r.tap("grid-h-2")
	if !g.rows[len(g.rows)-1][2].IsNull() {
		t.Fatal("NULLs stay last when descending")
	}
	if !strings.Contains(r.texts(), "country") {
		t.Fatal("header text")
	}
	// The sort marker follows the column.
	if r.a.activeTabPtr().sortCol != 3 || r.a.activeTabPtr().sortDir != 2 {
		t.Fatalf("sort state %d/%d", r.a.activeTabPtr().sortCol, r.a.activeTabPtr().sortDir)
	}
}

func TestGridQuickFilterOnQueryResults(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.must(r.a.RunSQL(r.ctx, "SELECT id, full_name, country FROM users WHERE id <= 100"))
	tab := r.a.activeTabPtr()
	total := len(tab.grid.rows)
	r.a.page.Form().SetFormValue("grid-filter", "DE")
	r.must(r.a.onChange(r.ctx, r.box("grid-filter")))
	n := len(tab.grid.rows)
	if n == 0 || n >= total {
		t.Fatalf("filter kept %d of %d", n, total)
	}
	for _, row := range tab.grid.rows {
		if !rowContains(row, "de") {
			t.Fatalf("row %v does not match", row)
		}
	}
	if !strings.Contains(r.texts(), fmt.Sprintf("%d of %d rows", n, total)) {
		t.Fatalf("info label: %q", r.texts())
	}
	// Sorting applies to the filtered rows, and clearing restores everything.
	r.tap("grid-h-0")
	r.a.page.Form().SetFormValue("grid-filter", "")
	r.must(r.a.onChange(r.ctx, r.box("grid-filter")))
	if len(tab.grid.rows) != total {
		t.Fatalf("cleared filter shows %d rows", len(tab.grid.rows))
	}
	if tab.grid.rows[0][0].I != 1 {
		t.Fatal("ascending sort survives the filter change")
	}
	// A new run resets the filter and the sort.
	r.must(r.a.RunSQL(r.ctx, "SELECT 1 AS one"))
	if tab.quick != "" || tab.sortCol != 0 {
		t.Fatal("a new result starts unfiltered and unsorted")
	}
}

func TestGridEnterAndF2EditAndInspect(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.openTableByName("users")
	tab := r.a.activeTabPtr()
	nameCol := r.colIndex("full_name")
	r.tap(fmt.Sprintf("cell-1-%d", nameCol))
	r.key("f2")
	if tab.editing == nil || tab.editing.row != 1 || tab.editing.col != nameCol {
		t.Fatalf("F2 should open the inline editor: %+v", tab.editing)
	}
	r.a.page.Form().SetFormValue("cell-editor", "Edited By Keyboard")
	r.must(r.a.onSubmit(r.ctx))
	if tab.cs == nil || tab.cs.Pending() != 1 {
		t.Fatal("Enter in the editor stages the change")
	}
	r.key("escape")
	r.a.focus = "grid"
	r.tap(fmt.Sprintf("cell-2-%d", nameCol))
	r.key("enter")
	if tab.editing == nil || tab.editing.row != 2 {
		t.Fatal("Enter edits the selected cell of a writable table")
	}
	r.key("escape")
	if tab.editing != nil {
		t.Fatal("Escape cancels the edit")
	}
	// On a read-only result, Enter opens the inspector instead.
	r.a.activeTab = 0
	r.must(r.a.RunSQL(r.ctx, "SELECT id, full_name FROM users"))
	r.tap("cell-0-1")
	r.key("enter")
	if !r.a.activeTabPtr().inspectorOpen || !r.has("inspector") {
		t.Fatal("Enter on a read-only cell opens the inspector")
	}
}

func TestGridContextMenuCsvFilterAndInspector(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.must(r.a.RunSQL(r.ctx, "SELECT id, full_name, country FROM users WHERE id <= 50"))
	tab := r.a.activeTabPtr()
	pick := func(id string, row, col int) {
		r.a.contextMenuFor(fmt.Sprintf("cell-%d-%d", row, col))
		r.must(r.a.menuAction(r.ctx, id))
	}
	pick("g-csv", 0, 1)
	got, _ := ownframe.Read()
	if want := fmt.Sprintf("%d,%s,%s", tab.grid.rows[0][0].I, tab.grid.rows[0][1].S, tab.grid.rows[0][2].S); got != want {
		t.Fatalf("row CSV = %q, want %q", got, want)
	}
	country := tab.grid.rows[3][2].S
	total := len(tab.grid.rows)
	pick("g-filter", 3, 2)
	if tab.quick != country || len(tab.grid.rows) >= total {
		t.Fatalf("filter by value: quick=%q rows=%d/%d", tab.quick, len(tab.grid.rows), total)
	}
	pick("g-inspect", 0, 1)
	if !tab.inspectorOpen || tab.grid.fc != 1 || !r.has("inspector") {
		t.Fatal("Open in Inspector selects the cell and shows the drawer")
	}
	// Table tab: Filter by Value becomes a WHERE condition and reloads.
	r.openTableByName("users")
	tt := r.a.activeTabPtr()
	cc := r.colIndex("country")
	want := tt.grid.rows[0][cc].S
	pick("g-filter", 0, cc)
	r.settle()
	if !strings.Contains(tt.filter, `"country" = '`+want+`'`) {
		t.Fatalf("WHERE = %q", tt.filter)
	}
	for _, row := range tt.grid.rows {
		if row[cc].S != want {
			t.Fatalf("server filter returned %q", row[cc].S)
		}
	}
	if r.a.page.Form().FormValue("tbl-filter") != tt.filter {
		t.Fatal("the WHERE box should show the generated condition")
	}
}

func TestGridPerColumnFilters(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.must(r.a.RunSQL(r.ctx, "SELECT id, full_name, CASE WHEN id % 5 = 0 THEN NULL ELSE country END AS country, plan FROM users WHERE id <= 100"))
	tab := r.a.activeTabPtr()
	total := len(tab.grid.rows)
	if r.has("grid-filters") {
		t.Fatal("the filter row starts closed")
	}
	r.tap("rs-filters")
	if !r.has("grid-filters") || !r.has("colf-0") || !r.has("colf-3") {
		t.Fatal("the filter row shows one box per column")
	}
	set := func(i int, v string) {
		id := fmt.Sprintf("colf-%d", i)
		r.a.page.Form().SetFormValue(id, v)
		r.must(r.a.onChange(r.ctx, r.box(id)))
	}
	// contains
	set(3, "PRO")
	n := len(tab.grid.rows)
	if n == 0 || n >= total {
		t.Fatalf("contains kept %d of %d", n, total)
	}
	for _, row := range tab.grid.rows {
		if !strings.Contains(strings.ToLower(row[3].S), "pro") {
			t.Fatalf("row %v does not contain pro", row)
		}
	}
	// equals combines with contains (AND)
	set(3, "=pro")
	for _, row := range tab.grid.rows {
		if row[3].S != "pro" {
			t.Fatalf("equals matched %q", row[3].S)
		}
	}
	set(3, "=pr") // partial text does not equal
	if len(tab.grid.rows) != 0 {
		t.Fatal("= needs the whole value")
	}
	set(3, "")
	// null / not null
	set(2, "null")
	if len(tab.grid.rows) != total/5 {
		t.Fatalf("null kept %d, want %d", len(tab.grid.rows), total/5)
	}
	set(2, "not null")
	if len(tab.grid.rows) != total-total/5 {
		t.Fatalf("not null kept %d", len(tab.grid.rows))
	}
	// combines with the quick filter, and the label shows N of M
	r.a.page.Form().SetFormValue("grid-filter", "pro")
	r.must(r.a.onChange(r.ctx, r.box("grid-filter")))
	m := len(tab.grid.rows)
	if m >= total-total/5 || m == 0 {
		t.Fatalf("quick + column filter kept %d", m)
	}
	if !strings.Contains(r.texts(), fmt.Sprintf("%d of %d rows", m, total)) {
		t.Fatalf("label should read %d of %d rows", m, total)
	}
	if !r.has("rs-clear-filters") {
		t.Fatal("Clear filters shows while a filter is set")
	}
	// sorting keeps working on the filtered rows
	r.tap("grid-h-0")
	for i := 1; i < len(tab.grid.rows); i++ {
		if tab.grid.rows[i-1][0].I > tab.grid.rows[i][0].I {
			t.Fatal("sort order lost under filters")
		}
	}
	r.tap("rs-clear-filters")
	if len(tab.grid.rows) != total || tab.filtersActive() || r.has("rs-clear-filters") {
		t.Fatalf("clear-all restores %d rows, got %d", total, len(tab.grid.rows))
	}
	if r.a.page.Form().FormValue("colf-2") != "" || r.a.page.Form().FormValue("grid-filter") != "" {
		t.Fatal("clear-all empties the input boxes")
	}
	// A new run resets everything; closing the row hides it.
	set(1, "a")
	r.must(r.a.RunSQL(r.ctx, "SELECT 1 AS one"))
	if tab.filtersActive() {
		t.Fatal("a new result starts unfiltered")
	}
	r.tap("rs-filters")
	if r.has("grid-filters") {
		t.Fatal("the toggle closes the row")
	}
}

func TestColMatchesRules(t *testing.T) {
	null, abc := driver.Null(), driver.Text("Abc")
	for _, c := range []struct {
		cell driver.CellValue
		spec string
		want bool
	}{
		{abc, "b", true}, {abc, "B", true}, {abc, "x", false}, {abc, "=abc", true}, {abc, "= abc", true}, {abc, "=ab", false},
		{null, "null", true}, {null, "NOT NULL", false}, {abc, "not null", true}, {abc, "null", false}, {null, "x", false},
	} {
		if got := colMatches(c.cell, c.spec); got != c.want {
			t.Errorf("colMatches(%v, %q) = %v, want %v", c.cell, c.spec, got, c.want)
		}
	}
}
