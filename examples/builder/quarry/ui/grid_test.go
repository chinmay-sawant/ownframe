package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/driver"
)

func syntheticResult(n int) driver.QueryResult {
	cols := []driver.ColumnMeta{
		{Name: "id", Type: "INTEGER", PrimaryKey: true},
		{Name: "name", Type: "TEXT"},
		{Name: "score", Type: "REAL"},
		{Name: "active", Type: "BOOLEAN"},
		{Name: "at", Type: "TEXT"},
		{Name: "note", Type: "TEXT", Nullable: true},
	}
	rows := make([][]driver.CellValue, n)
	for i := range rows {
		note := driver.Null()
		if i%3 == 0 {
			note = driver.Text("n" + fmt.Sprint(i))
		}
		rows[i] = []driver.CellValue{driver.Int(int64(i + 1)), driver.Text(fmt.Sprintf("row %d", i+1)), driver.Float(float64(i) / 4),
			driver.Bool(i%2 == 0), driver.Time(time.Unix(1_700_000_000+int64(i), 0).UTC()), note}
	}
	return driver.QueryResult{Columns: cols, Rows: rows}
}

func (r *rig) countBoxes(prefix string) int {
	n := 0
	for _, b := range r.a.Page().Boxes() {
		if strings.HasPrefix(b.ID, prefix) {
			n++
		}
	}
	return n
}

func TestVirtualizedGridWindowing(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	tab := r.a.activeTabPtr()
	tab.grid = newGridState(syntheticResult(100_000))
	r.must(r.a.redraw(r.ctx))
	viewH := r.a.gridViewH(tab)
	maxRows := viewH/gridRowHeight + 2 + 2*gridOverscan
	rows := r.countBoxes("grid-row-")
	if rows == 0 || rows > maxRows {
		t.Fatalf("rendered %d rows, want 1..%d", rows, maxRows)
	}
	cells := r.countBoxes("cell-")
	if cells != rows*len(tab.grid.cols) {
		t.Fatalf("cells = %d, rows = %d", cells, rows)
	}
	// Scroll deep: row count stays bounded, window follows.
	body := r.box("grid-body")
	for i := 0; i < 400; i++ {
		_, err := r.a.onWheel(r.ctx, body.X+50, body.Y+50, 0, -10)
		r.must(err)
	}
	if got := r.countBoxes("grid-row-"); got > maxRows {
		t.Fatalf("after scrolling %d rows", got)
	}
	first := -1
	for _, b := range r.a.Page().Boxes() {
		if strings.HasPrefix(b.ID, "grid-row-") {
			var n int
			_, _ = fmt.Sscanf(b.ID, "grid-row-%d", &n)
			if first < 0 || n < first {
				first = n
			}
		}
	}
	if first < 1000 {
		t.Fatalf("window did not move: first row %d", first)
	}
	// Scroll to the very end.
	tab.grid.scrollY = 1e12
	tab.grid.clampScroll(r.a.gridViewW(tab), viewH, 0)
	r.must(r.a.redraw(r.ctx))
	if !r.has("grid-row-99999") {
		t.Fatal("last row not reachable")
	}
	if r.has("grid-row-0") {
		t.Fatal("first row should be out of the window")
	}
	// Horizontal scroll shifts the header.
	_, err := r.a.onWheel(r.ctx, body.X+50, body.Y+50, -3, 0)
	r.must(err)
	r.png("grid-100k")
}

func TestTypedCellRendering(t *testing.T) {
	jsonCell := driver.JSON(`{"a":1,"b":[1,2]}`)
	cases := []struct {
		cell        driver.CellValue
		text, class string
	}{
		{driver.Null(), "NULL", "qr-null"},
		{driver.Int(-42), "-42", "qr-num"},
		{driver.Float(3.5), "3.5", "qr-num"},
		{driver.Bool(true), "true", "qr-bool qr-bool-true"},
		{driver.Bool(false), "false", "qr-bool qr-bool-false"},
		{driver.Time(time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)), "2025-01-02 03:04:05", "qr-time"},
		{jsonCell, "{...} 2 keys", "qr-json"},
		{driver.JSON(`[1,2,3]`), "[...] 3 items", "qr-json"},
		{driver.Blob(make([]byte, 4300)), "<BLOB 4.2 KB>", "qr-blob"},
		{driver.Text("multi\nline"), "multi line", ""},
	}
	for _, c := range cases {
		txt, cls := cellDisplay(c.cell)
		if txt != c.text || cls != c.class {
			t.Errorf("%v -> %q/%q, want %q/%q", c.cell.Kind, txt, cls, c.text, c.class)
		}
	}
	// Every kind reaches the painted grid.
	r := newRig(t, 1440, 960, "")
	cols := []driver.ColumnMeta{{Name: "n", Type: "INT"}, {Name: "f", Type: "REAL"}, {Name: "t", Type: "TEXT"}, {Name: "b", Type: "BOOL"},
		{Name: "ts", Type: "TS"}, {Name: "j", Type: "JSON"}, {Name: "x", Type: "BLOB"}, {Name: "z", Type: "TEXT"}}
	row := []driver.CellValue{driver.Int(7), driver.Float(2.5), driver.Text("hi"), driver.Bool(true),
		driver.Time(time.Date(2025, 5, 6, 7, 8, 9, 0, time.UTC)), jsonCell, driver.Blob(make([]byte, 2048)), driver.Null()}
	r.a.activeTabPtr().grid = newGridState(driver.QueryResult{Columns: cols, Rows: [][]driver.CellValue{row}})
	r.must(r.a.redraw(r.ctx))
	txt := r.texts()
	for _, want := range []string{"NULL", "true", "2025-05-06 07:08:09", "{...} 2 keys", "<BLOB 2.0 KB>", "hi", "2.5"} {
		if !strings.Contains(txt, want) {
			t.Errorf("grid did not paint %q", want)
		}
	}
	// Numbers are right-aligned: the cell uses the qr-num class and flex-end.
	num := r.box("cell-0-0")
	var numX float64
	for _, op := range r.a.Page().Probe().Display().Ops {
		if op.Kind == ownframe.DisplayOpText && op.Text == "7" {
			numX = op.X * 4 / 3
		}
	}
	if numX < num.X+num.W/2 {
		t.Errorf("number not right aligned: text x=%.1f in cell %.1f..%.1f", numX, num.X, num.X+num.W)
	}
	r.png("typed-cells")
}

func TestGridNavigationAndCopy(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	tab := r.a.activeTabPtr()
	tab.ed = newEditor("SELECT id, full_name, country FROM users ORDER BY id")
	r.chord("enter", false)
	r.settle()
	r.tap("cell-2-1")
	g := tab.grid
	if g.fr != 2 || g.fc != 1 || r.a.focus != "grid" {
		t.Fatalf("selection = %d,%d focus=%q", g.fr, g.fc, r.a.focus)
	}
	r.tap("cell-2-1") // blur any control focus, keep grid focus
	r.key("arrowdown")
	r.key("arrowright")
	if g.fr != 3 || g.fc != 2 {
		t.Fatalf("arrows -> %d,%d", g.fr, g.fc)
	}
	r.key("home")
	if g.fc != 0 {
		t.Fatal("Home")
	}
	r.key("end")
	if g.fc != 2 {
		t.Fatal("End")
	}
	r.key("pagedown")
	if g.fr <= 10 {
		t.Fatalf("PageDown -> %d", g.fr)
	}
	r.key("pageup")
	r.key("arrowup")
	// Single cell copies as plain text.
	g.ar, g.ac, g.fr, g.fc = 0, 1, 0, 1
	txt, ok, err := r.a.onCopy(r.ctx)
	r.must(err)
	if !ok || txt != g.rows[0][1].S {
		t.Fatalf("single copy = %q", txt)
	}
	// A block copies as TSV.
	g.ar, g.ac, g.fr, g.fc = 0, 0, 1, 2
	txt, _, _ = r.a.onCopy(r.ctx)
	lines := strings.Split(txt, "\n")
	if len(lines) != 2 || strings.Count(lines[0], "\t") != 2 || !strings.HasPrefix(lines[0], "1\t") {
		t.Fatalf("TSV = %q", txt)
	}
	// Row selection from the gutter.
	r.tap("grid-gutter-4")
	txt, _, _ = r.a.onCopy(r.ctx)
	if strings.Count(txt, "\t") != 2 || !strings.HasPrefix(txt, "5\t") {
		t.Fatalf("row copy = %q", txt)
	}
}

func TestJSONInspectorDrawer(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.openTableByName("audit_events")
	tab := r.a.activeTabPtr()
	pi := -1
	for i, c := range tab.grid.cols {
		if c.Name == "payload" {
			pi = i
		}
	}
	if pi < 0 {
		t.Fatal("payload column missing")
	}
	r.tap(fmt.Sprintf("cell-0-%d", pi))
	r.tap("tbl-inspector")
	if !r.has("inspector") {
		t.Fatal("drawer did not open")
	}
	body := r.box("insp-body")
	if body.H < 40 {
		t.Fatalf("pretty JSON should span several lines, body height %v", body.H)
	}
	if !strings.Contains(r.texts(), "\"") || r.countBoxes("insp-size") != 1 {
		t.Fatal("JSON text or byte size missing")
	}
	orig := tab.grid.rows[0][pi]
	r.a.page.Form().SetFormValue("insp-edit", `{"edited":true}`)
	r.tap("insp-apply")
	if tab.cs == nil || tab.cs.Pending() != 1 {
		t.Fatalf("edit not staged: %v", tab.cs)
	}
	sql, _ := tab.cs.Preview()
	if !strings.Contains(sql, `{"edited":true}`) || !strings.Contains(sql, "UPDATE") {
		t.Fatalf("preview = %q", sql)
	}
	_ = orig
	r.tap("insp-close")
	if r.has("inspector") {
		t.Fatal("close")
	}
	// Double-click opens the drawer too.
	r.a.lastClickID = ""
	r.tap("cell-1-0")
	r.now = r.now.Add(50 * time.Millisecond)
	r.tap("cell-1-0")
	r.png("inspector")
}

func TestTableFilterSortPagination(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.openTableByName("orders")
	tab := r.a.activeTabPtr()
	if tab.totalRows != 2500 || len(tab.grid.rows) != 100 {
		t.Fatalf("total=%d rows=%d", tab.totalRows, len(tab.grid.rows))
	}
	want, err := r.a.drv.Query(r.ctx, "SELECT COUNT(*) FROM orders WHERE total_cents > 100000", 1)
	r.must(err)
	r.a.page.Form().SetFormValue("tbl-filter", "total_cents > 100000")
	r.tap("tbl-apply")
	r.settle()
	if tab.totalRows != want.Rows[0][0].I {
		t.Fatalf("filtered total %d, want %d", tab.totalRows, want.Rows[0][0].I)
	}
	tc := -1
	for i, c := range tab.grid.cols {
		if c.Name == "total_cents" {
			tc = i
		}
	}
	for _, row := range tab.grid.rows {
		if row[tc].I <= 100000 {
			t.Fatalf("row violates filter: %v", row[tc])
		}
	}
	r.tap(fmt.Sprintf("grid-h-%d", tc)) // ASC
	r.settle()
	for i := 1; i < len(tab.grid.rows); i++ {
		if tab.grid.rows[i-1][tc].I > tab.grid.rows[i][tc].I {
			t.Fatal("ASC order broken")
		}
	}
	r.tap(fmt.Sprintf("grid-h-%d", tc)) // DESC
	r.settle()
	for i := 1; i < len(tab.grid.rows); i++ {
		if tab.grid.rows[i-1][tc].I < tab.grid.rows[i][tc].I {
			t.Fatal("DESC order broken")
		}
	}
	r.tap(fmt.Sprintf("grid-h-%d", tc)) // none
	r.settle()
	if tab.sortDir != 0 || tab.sortCol != 0 {
		t.Fatal("third click should clear the sort")
	}
	r.tap("tbl-size-500")
	r.settle()
	if len(tab.grid.rows) != min(500, int(tab.totalRows)) {
		t.Fatalf("page size 500 -> %d rows", len(tab.grid.rows))
	}
	r.tap("tbl-size-100")
	r.settle()
	first := tab.grid.rows[0][0].I
	r.tap("tbl-next")
	r.settle()
	if tab.page != 1 || tab.grid.rows[0][0].I == first {
		t.Fatalf("Next did not page: page=%d", tab.page)
	}
	r.tap("tbl-prev")
	r.settle()
	if tab.page != 0 {
		t.Fatal("Prev")
	}
	// Unsafe filters are refused before they reach the database.
	r.a.page.Form().SetFormValue("tbl-filter", "1=1; DROP TABLE orders")
	r.tap("tbl-apply")
	if tab.err == nil {
		t.Fatal("a filter with ';' must be rejected")
	}
	if got, _ := r.a.drv.Query(r.ctx, "SELECT COUNT(*) FROM orders", 1); got.Rows[0][0].I != 2500 {
		t.Fatal("orders was modified")
	}
}

func (r *rig) editCell(row, col int, value string) {
	r.t.Helper()
	id := fmt.Sprintf("cell-%d-%d", row, col)
	r.a.lastClickID = ""
	r.tap(id)
	r.now = r.now.Add(60 * time.Millisecond)
	r.tap(id)
	if r.a.activeTabPtr().editing == nil {
		r.t.Fatal("double-click did not open the inline editor")
	}
	r.a.page.Form().SetFormValue("cell-editor", value)
	r.must(r.a.onSubmit(r.ctx))
}

func (r *rig) colIndex(name string) int {
	for i, c := range r.a.activeTabPtr().grid.cols {
		if c.Name == name {
			return i
		}
	}
	r.t.Fatalf("no column %q", name)
	return -1
}

func TestStagedEditCommitAndDiscard(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.openTableByName("users")
	tab := r.a.activeTabPtr()
	nameCol := r.colIndex("full_name")
	orig := tab.grid.rows[0][nameCol].S
	r.editCell(0, nameCol, "Renamed Person")
	if tab.cs == nil || tab.cs.Pending() != 1 || !r.has("changeset-bar") {
		t.Fatal("changeset bar expected after one edit")
	}
	if !strings.Contains(r.texts(), "1 pending changes") {
		t.Fatalf("count label: %q", r.texts())
	}
	cell := r.box(fmt.Sprintf("cell-0-%d", nameCol))
	if cell.W <= 0 {
		t.Fatal("cell missing")
	}
	r.png("staged-edit")
	r.tap("cs-preview")
	if r.a.modal != "sqlpreview" || !strings.Contains(r.texts(), "UPDATE") || !strings.Contains(r.texts(), "Renamed Person") {
		t.Fatalf("preview modal: %q", r.a.modal)
	}
	r.tap("pv-close")
	// Discard restores the original.
	r.tap("cs-discard")
	if tab.cs != nil || r.has("changeset-bar") {
		t.Fatal("discard should drop the bar")
	}
	got, _ := r.a.drv.Query(r.ctx, "SELECT full_name FROM users WHERE id = 1", 1)
	if got.Rows[0][0].S != orig {
		t.Fatal("discard must not write")
	}
	// Commit persists inside a transaction.
	r.editCell(0, nameCol, "Renamed Person")
	r.editCell(1, nameCol, "Second One")
	r.tap("cs-commit")
	r.settle()
	got, _ = r.a.drv.Query(r.ctx, "SELECT full_name FROM users WHERE id IN (1, 2) ORDER BY id", 2)
	if got.Rows[0][0].S != "Renamed Person" || got.Rows[1][0].S != "Second One" {
		t.Fatalf("not persisted: %v", got.Rows)
	}
	if tab.cs != nil || tab.grid.rows[0][nameCol].S != "Renamed Person" {
		t.Fatal("grid should reload with the committed value")
	}
	// Insert and delete rows.
	r.tap("tbl-add")
	if tab.cs.Pending() != 1 || !strings.Contains(strings.ToUpper(func() string { s, _ := tab.cs.Preview(); return s }()), "INSERT") {
		t.Fatal("Add Row should stage an INSERT")
	}
	r.tap("cs-discard")
	r.tap("grid-gutter-3")
	r.key("delete")
	id3, _ := r.a.rowID(tab, tab.grid, 3)
	if tab.cs == nil || !tab.cs.IsDeleted(id3) {
		t.Fatal("Delete should stage a row deletion")
	}
	rowEl := r.box("grid-row-3")
	if rowEl.H != gridRowHeight {
		t.Fatal("row geometry")
	}
}

func TestChangesetCommitRollbackOnError(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.openTableByName("users")
	tab := r.a.activeTabPtr()
	nameCol, emailCol := r.colIndex("full_name"), r.colIndex("email")
	before, _ := r.a.drv.Query(r.ctx, "SELECT full_name, email FROM users WHERE id IN (1, 2) ORDER BY id", 2)
	dup := tab.grid.rows[0][emailCol].S
	r.editCell(0, nameCol, "Should Roll Back")
	r.editCell(1, emailCol, dup) // UNIQUE violation
	r.tap("cs-commit")
	r.settle()
	if tab.cs == nil || tab.cs.Pending() != 2 {
		t.Fatal("staged changes must survive a failed commit")
	}
	if !r.has("cs-error") || !strings.Contains(strings.ToLower(r.texts()), "unique") {
		t.Fatalf("failure not shown: %q", r.texts())
	}
	if !strings.Contains(r.texts(), "UPDATE") {
		t.Fatal("the failing statement should be quoted")
	}
	after, _ := r.a.drv.Query(r.ctx, "SELECT full_name, email FROM users WHERE id IN (1, 2) ORDER BY id", 2)
	for i := range before.Rows {
		for j := range before.Rows[i] {
			if !before.Rows[i][j].Equal(after.Rows[i][j]) {
				t.Fatalf("DB changed after rollback: %v vs %v", before.Rows, after.Rows)
			}
		}
	}
	// The connection must still be usable for writes afterwards.
	if _, err := r.a.drv.Exec(r.ctx, "UPDATE users SET plan = plan WHERE id = 3"); err != nil {
		t.Fatalf("connection left in a broken transaction: %v", err)
	}
}

func TestGridContextMenuCopyFormats(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.openTableByName("products")
	tab := r.a.activeTabPtr()
	items := r.a.contextMenuFor("cell-1-1")
	if len(items) != 10 {
		t.Fatalf("menu = %+v", items)
	}
	do := func(id string) string {
		ownframe.Clear()
		r.a.contextMenuFor("cell-1-1")
		r.must(r.a.menuAction(r.ctx, id))
		s, _ := ownframe.Read()
		return s
	}
	if got := do("g-copy"); got != tab.grid.rows[1][1].S {
		t.Fatalf("copy value = %q", got)
	}
	js := do("g-json")
	if !strings.Contains(js, `"sku"`) || !strings.Contains(js, `"price_cents"`) || !strings.HasPrefix(js, "{") {
		t.Fatalf("json = %q", js)
	}
	ins := do("g-insert")
	if !strings.HasPrefix(ins, `INSERT INTO "products" (`) || !strings.Contains(ins, "VALUES (") || !strings.HasSuffix(ins, ");") {
		t.Fatalf("insert = %q", ins)
	}
	if _, err := r.a.drv.Exec(r.ctx, "CREATE TABLE p2 AS SELECT * FROM products WHERE 0"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.a.drv.Exec(r.ctx, strings.Replace(ins, `"products"`, `"p2"`, 1)); err != nil {
		t.Fatalf("copied INSERT must re-execute: %v", err)
	}
	tab.grid.ar, tab.grid.ac, tab.grid.fr, tab.grid.fc = 0, 0, 1, 1
	md := do("g-md")
	lines := strings.Split(md, "\n")
	if len(lines) != 4 || !strings.HasPrefix(lines[0], "| id | sku |") || !strings.HasPrefix(lines[1], "| --- | --- |") {
		t.Fatalf("markdown = %q", md)
	}
	// Writable actions stage changes.
	do("g-null")
	do("g-dup")
	do("g-del")
	if tab.cs == nil {
		t.Fatal("writable menu actions should stage changes")
	}
	pv, _ := tab.cs.Preview()
	if !strings.Contains(pv, "INSERT INTO") || !strings.Contains(pv, "DELETE FROM") {
		t.Fatalf("staged preview = %q", pv)
	}
}

func TestExportImportDialogs(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.openTableByName("products")
	r.tap("tbl-export")
	if r.a.modal != "export" {
		t.Fatal("export dialog")
	}
	dir := t.TempDir()
	out := filepath.Join(dir, "p.csv")
	r.a.page.Form().SetFormValue("ex-path", out)
	r.a.page.Form().SetFormValue("ex-null", "NULL")
	r.tap("ex-scope-fulltable")
	r.tap("ex-run")
	data, err := os.ReadFile(filepath.Clean(out))
	if err != nil {
		t.Fatalf("export file: %v (%s)", err, r.a.exportOpts.Result)
	}
	if lines := strings.Count(string(data), "\n"); lines != 121 {
		t.Fatalf("full-table CSV has %d lines, want 121", lines)
	}
	r.tap("ex-fmt-ndjson")
	r.tap("ex-copy")
	if got, _ := ownframe.Read(); !strings.Contains(got, `"sku"`) || strings.Count(got, "\n") < 120 {
		t.Fatalf("clipboard ndjson = %.80q", got)
	}
	r.tap("ex-close")

	// CSV import creates a table in one transaction with inferred types.
	csv := filepath.Join(dir, "imp.csv")
	_ = os.WriteFile(csv, []byte("id,price,active,name\n1,1.5,true,a\n2,2.5,false,b\n3,,true,c\n"), 0o600)
	r.tap("tbl-import")
	r.a.page.Form().SetFormValue("im-path", csv)
	r.a.page.Form().SetFormValue("im-target", "imported_things")
	r.tap("im-load")
	if r.a.importSt.Data == nil || !r.has("im-preview") || !r.has("im-row-2") || !strings.Contains(r.texts(), "REAL") {
		t.Fatalf("preview missing: %q err=%s", r.texts(), r.a.importSt.Err)
	}
	r.tap("im-run")
	if r.a.importSt.Err != "" {
		t.Fatal(r.a.importSt.Err)
	}
	res, err := r.a.drv.Query(r.ctx, "SELECT id, price, active, name FROM imported_things ORDER BY id", 10)
	r.must(err)
	if len(res.Rows) != 3 || res.Rows[0][1].F != 1.5 || !res.Rows[2][1].IsNull() {
		t.Fatalf("imported = %v", res.Rows)
	}
	types, _ := r.a.drv.TableColumns(r.ctx, "main", "imported_things")
	if types[0].Type != "INTEGER" || types[1].Type != "REAL" || types[2].Type != "BOOLEAN" || types[3].Type != "TEXT" {
		t.Fatalf("types = %+v", types)
	}
	// Append mode adds rows; the sidebar learned the new table.
	r.tap("im-mode-append")
	r.tap("im-run")
	res, _ = r.a.drv.Query(r.ctx, "SELECT COUNT(*) FROM imported_things", 1)
	if res.Rows[0][0].I != 6 {
		t.Fatalf("append count = %d", res.Rows[0][0].I)
	}
	// A failing batch (CREATE of an existing table) rolls back entirely.
	r.tap("im-mode-createnew")
	r.tap("im-run")
	if r.a.importSt.Err == "" {
		t.Fatal("creating an existing table must fail")
	}
	res, _ = r.a.drv.Query(r.ctx, "SELECT COUNT(*) FROM imported_things", 1)
	if res.Rows[0][0].I != 6 {
		t.Fatalf("failed import changed rows: %d", res.Rows[0][0].I)
	}
	// Dropping a CSV opens the importer.
	r.a.modal = ""
	r.must(r.a.onDrop(r.ctx, []ownframe.Drop{{Name: "d.csv", Read: func() ([]byte, error) { return []byte("a,b\n1,2\n"), nil }}}))
	if r.a.modal != "import" || r.a.importSt.Data == nil {
		t.Fatal("CSV drop should open the import dialog")
	}
	var _ context.Context
}
