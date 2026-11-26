package ui

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/chinmay-sawant/ownframe/ui"

	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/driver"
	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/sqlkit"
)

// cellEditing is an open inline editor on one grid cell.
type cellEditing struct{ row, col int }

// selectedCell returns the focus cell of t's grid.
func (t *workTab) selectedCell() (driver.CellValue, driver.ColumnMeta, bool) {
	g := t.grid
	if g == nil || g.fr < 0 || g.fc < 0 || g.fr >= len(g.rows) || g.fc >= len(g.cols) {
		return driver.CellValue{}, driver.ColumnMeta{}, false
	}
	return g.rows[g.fr][g.fc], g.cols[g.fc], true
}

// inspectorPane is the right-hand drawer with the full value of the selected cell.
func (a *App) inspectorPane(t *workTab) *ui.Element {
	if !t.inspectorOpen {
		return nil
	}
	cell, col, ok := t.selectedCell()
	pane := ui.Div().ID("inspector").Class("qr-inspector").Width(ui.Px(float64(a.inspectorW))).
		Attr("role", ariaRole("complementary")).Attr("aria-label", "Cell inspector")
	head := ui.Div().Class("qr-inspector-head").Children(
		ui.Span().Class("qr-sidebar-title").Text("Cell Inspector"),
		ui.Button("Close").ID("insp-close").Class("qr-btn qr-btn-ghost").Attr("type", "button").
			OnClick(func(ctx context.Context) error { t.inspectorOpen = false; return a.redraw(ctx) }),
	)
	pane.Children(ui.Div().ID("split-inspector").Class("qr-drag-handle").Attr("aria-label", "Resize inspector"), head)
	if !ok {
		return pane.Children(ui.Div().ID("insp-empty").Class("qr-empty").Text("Select a cell to inspect it."))
	}
	text := cell.String()
	isJSON := false
	if cell.Kind == driver.KindJSON || cell.Kind == driver.KindText {
		if p, valid := prettyJSON(cell.S); valid && (strings.HasPrefix(strings.TrimSpace(cell.S), "{") || strings.HasPrefix(strings.TrimSpace(cell.S), "[")) {
			text, isJSON = p, true
		}
	}
	size := len(text)
	if cell.Kind == driver.KindBlob {
		size = len(cell.Blob)
	}
	pane.Children(ui.Div().Class("qr-insp-meta qr-muted").Children(
		ui.Span().Text(col.Name),
		ui.Span().Text(col.Type),
		ui.Span().ID("insp-size").Text(humanBytes(size)),
	))
	body := ui.Div().ID("insp-body").Class("qr-insp-body qr-mono")
	if isJSON {
		for _, ln := range highlightedBlock(text, sqlkit.SQLite) {
			body.Children(ln)
		}
		body.Attr("data-json", "true")
	} else {
		for _, ln := range strings.Split(text, "\n") {
			body.Children(ui.Div().Class("qr-ddl-line").Children(ui.Span().Class("qr-run").Text(ln)))
		}
	}
	editable := t.writable && t.kind == tabTable && (cell.Kind == driver.KindText || cell.Kind == driver.KindJSON)
	if !editable || isJSON {
		pane.Children(a.pane("insp", "Inspector scroll", "qr-insp-pane", body)) // an editable plain value shows only the editor, not a second copy
	}
	if t.writable && t.kind == tabTable && (cell.Kind == driver.KindText || cell.Kind == driver.KindJSON) {
		pane.Children(
			ui.Tag("textarea").ID("insp-edit").Class("qr-input qr-insp-edit").Attr("aria-label", "Edit cell value").Text(cell.S),
			a.tbButton("insp-apply", "Apply to cell", "qr-btn qr-btn-primary", false, func(ctx context.Context) error {
				return a.stageCell(ctx, t, t.grid.fr, t.grid.fc, editedCell(cell.Kind, a.page.Form().FormValue("insp-edit")))
			}),
		)
	}
	pane.Children(a.tbButton("insp-copy", "Copy", "qr-btn", false, func(ctx context.Context) error {
		a.copyClipboard(text)
		a.notify("Value copied")
		return a.redraw(ctx)
	}))
	return pane
}

// ---- inline editing and staged changes ----

func (a *App) beginInlineEdit(ctx context.Context, t *workTab, r, c int) error {
	if !t.writable || a.cfg.ReadOnly {
		a.notify("Table is read-only")
		return a.redraw(ctx)
	}
	t.editing = &cellEditing{row: r, col: c}
	a.page.Form().RequestFocus("cell-editor")
	return a.redraw(ctx)
}

func (a *App) inlineEditor(t *workTab) *ui.Element {
	cur := t.grid.rows[t.editing.row][t.editing.col]
	if v, ok := a.stagedValue(t, t.editing.row, t.editing.col); ok {
		cur = v
	}
	val := cur.String()
	if cur.IsNull() {
		val = ""
	}
	return ui.Div().Class("qr-inline-edit").Children(
		ui.Input().ID("cell-editor").Class("qr-input qr-cell-input").Attr("type", "text").Attr("value", val).
			Attr("aria-label", "Edit cell value"),
		ui.Button("NULL").ID("cell-set-null").Class("qr-btn").Attr("type", "button").Attr("aria-label", "Set NULL").
			OnClick(func(ctx context.Context) error {
				return a.stageCell(ctx, t, t.editing.row, t.editing.col, driver.Null())
			}),
	)
}

func (a *App) stagedValue(t *workTab, r, c int) (driver.CellValue, bool) {
	if t.cs == nil {
		return driver.CellValue{}, false
	}
	id, ok := a.rowID(t, t.grid, r)
	if !ok {
		return driver.CellValue{}, false
	}
	v, ok := t.cs.EditedValue(id, t.grid.cols[c].Name)
	if !ok {
		return driver.CellValue{}, false
	}
	return anyCell(v), true
}

func (a *App) changeset(t *workTab) *sqlkit.Changeset {
	if t.cs == nil {
		t.cs = sqlkit.NewChangeset(sqlkit.Config{
			Dialect: a.dialect(), Schema: qualSchema(a.cfg, t.schema), Table: t.table,
			ReadOnly: a.cfg.ReadOnly, HasRowID: len(t.pk) > 0,
		})
	}
	return t.cs
}

func (a *App) commitInlineEdit(ctx context.Context, t *workTab) error {
	e := t.editing
	if e == nil {
		return nil
	}
	old := t.grid.rows[e.row][e.col]
	nv := parseCell(a.page.Form().FormValue("cell-editor"), old)
	return a.stageCell(ctx, t, e.row, e.col, nv)
}

// stageCell records one cell edit in the changeset.
func (a *App) stageCell(ctx context.Context, t *workTab, r, c int, nv driver.CellValue) error {
	t.editing = nil
	id, ok := a.rowID(t, t.grid, r)
	if !ok {
		a.notify("Row has no primary key")
		return a.redraw(ctx)
	}
	old := t.grid.rows[r][c]
	if err := a.changeset(t).UpdateCell(id, t.grid.cols[c].Name, cellAny(old), cellAny(nv)); err != nil {
		a.notify(err.Error())
	}
	if t.cs.Pending() == 0 {
		t.cs = nil
	}
	return a.redraw(ctx)
}

func (a *App) stageDeleteSelected(ctx context.Context, t *workTab) error {
	g := t.grid
	for _, r := range g.sortedSelectionRows() {
		if id, ok := a.rowID(t, g, r); ok {
			if t.cs != nil && t.cs.IsDeleted(id) {
				t.cs.UnstageDelete(id)
				continue
			}
			if err := a.changeset(t).DeleteRow(id); err != nil {
				a.notify(err.Error())
			}
		}
	}
	return a.redraw(ctx)
}

func (a *App) stageInsert(ctx context.Context, t *workTab) error {
	var cols []string
	var vals []any
	for _, c := range t.grid.cols {
		if c.PrimaryKey && strings.Contains(strings.ToLower(c.Type), "int") {
			continue // let the database assign the key
		}
		cols = append(cols, c.Name)
		vals = append(vals, nil)
	}
	if _, err := a.changeset(t).InsertRow(cols, vals); err != nil {
		a.notify(err.Error())
	}
	return a.redraw(ctx)
}

func (a *App) duplicateRow(ctx context.Context, t *workTab, r int) error {
	g := t.grid
	var cols []string
	var vals []any
	for i, c := range g.cols {
		if c.PrimaryKey {
			continue
		}
		cols = append(cols, c.Name)
		vals = append(vals, cellAny(g.rows[r][i]))
	}
	if _, err := a.changeset(t).InsertRow(cols, vals); err != nil {
		a.notify(err.Error())
	}
	return a.redraw(ctx)
}

// changesetBar is the floating bar listing pending changes with actions.
func (a *App) changesetBar(t *workTab) *ui.Element {
	n := t.cs.Pending()
	bar := ui.Div().ID("changeset-bar").Class("qr-changebar").Attr("role", ariaRole("status")).Children(
		ui.Span().ID("cs-count").Class("qr-changebar-count").Text(fmt.Sprintf("%d pending changes", n)),
	)
	if t.csErr != "" {
		bar.Children(ui.Span().ID("cs-error").Class("qr-note-err").Text(t.csErr))
	}
	bar.Children(ui.Div().Class("qr-spacer"),
		a.tbButton("cs-preview", "Preview SQL", "qr-btn", false, func(ctx context.Context) error {
			p, err := t.cs.Preview()
			if err != nil {
				a.notify(err.Error())
			}
			a.modalNote, a.modal = p, "sqlpreview"
			return a.redraw(ctx)
		}),
		a.tbButton("cs-discard", "Discard", "qr-btn", false, func(ctx context.Context) error {
			t.cs.Discard()
			t.cs, t.csErr = nil, ""
			return a.redraw(ctx)
		}),
		a.tbButton("cs-commit", "Commit", "qr-btn qr-btn-primary", false, func(ctx context.Context) error { return a.commitChangeset(ctx, t) }),
	)
	return bar
}

// commitChangeset writes every staged change in one transaction. On failure the
// transaction is rolled back, the changeset is kept, and the failing statement
// is reported.
func (a *App) commitChangeset(ctx context.Context, t *workTab) error {
	sts, err := t.cs.Statements()
	if err != nil {
		t.csErr = err.Error()
		return a.redraw(ctx)
	}
	failed, err := a.execAtomic(ctx, sts)
	if err != nil {
		stmt := ""
		if failed >= 0 && failed < len(sts) {
			stmt = sqlkit.Inline(sts[failed].SQL, sts[failed].Args, a.dialect())
		}
		t.csErr = fmt.Sprintf("Commit rolled back: %v (statement: %s)", err, stmt)
		return a.redraw(ctx)
	}
	t.csErr = ""
	t.cs, t.editing = nil, nil
	a.notify(fmt.Sprintf("Committed %d statements", len(sts)))
	return a.loadTablePage(ctx, t)
}

func (a *App) execAtomic(ctx context.Context, sts []sqlkit.Statement) (int, error) {
	if tx, ok := a.drv.(driver.TxExecer); ok {
		sqls := make([]string, len(sts))
		args := make([][]any, len(sts))
		for i, s := range sts {
			sqls[i], args[i] = s.SQL, s.Args
		}
		return tx.ExecTx(ctx, sqls, args)
	}
	// Fallback: one script with literal values inside BEGIN/COMMIT.
	var b strings.Builder
	b.WriteString("BEGIN;\n")
	for _, s := range sts {
		b.WriteString(sqlkit.Inline(s.SQL, s.Args, a.dialect()))
		b.WriteString(";\n")
	}
	b.WriteString("COMMIT;")
	if _, err := a.drv.Exec(ctx, b.String()); err != nil {
		_, _ = a.drv.Exec(ctx, "ROLLBACK")
		return a.guessFailed(ctx, sts), err
	}
	return -1, nil
}

func (a *App) guessFailed(ctx context.Context, sts []sqlkit.Statement) int { return len(sts) - 1 }

// ---- grid context menu ----

func (a *App) gridMenuAction(ctx context.Context, id string, r, c int) error {
	t := a.activeTabPtr()
	if t == nil || t.grid == nil || r >= len(t.grid.rows) {
		return nil
	}
	g := t.grid
	switch id {
	case "g-copy":
		a.copyClipboard(cellPlain(g.rows[r][c]))
	case "g-json":
		a.copyClipboard(rowJSON(g, r))
	case "g-csv":
		a.copyClipboard(rowCSV(g, r))
	case "g-filter":
		return a.filterByValue(ctx, t, r, c)
	case "g-inspect":
		g.ar, g.ac, g.fr, g.fc, g.rowSel = r, c, r, c, false
		t.inspectorOpen = true
		return a.redraw(ctx)
	case "g-insert":
		a.copyClipboard(rowInsert(a, t, r))
	case "g-md":
		a.copyClipboard(selectionMarkdown(g, r, c))
	case "g-null":
		return a.stageCell(ctx, t, r, c, driver.Null())
	case "g-dup":
		return a.duplicateRow(ctx, t, r)
	case "g-del":
		g.ar, g.ac, g.fr, g.fc, g.rowSel = r, 0, r, len(g.cols)-1, true
		return a.stageDeleteSelected(ctx, t)
	}
	return a.redraw(ctx)
}

func rowJSON(g *gridState, r int) string {
	m := map[string]any{}
	for i, c := range g.cols {
		m[c.Name] = cellJSONValue(g.rows[r][i])
	}
	b, _ := json.MarshalIndent(m, "", "  ")
	return string(b)
}

func cellJSONValue(c driver.CellValue) any {
	switch c.Kind {
	case driver.KindNull:
		return nil
	case driver.KindJSON:
		var v any
		if json.Unmarshal([]byte(c.S), &v) == nil {
			return v
		}
		return c.S
	case driver.KindTime:
		return c.String()
	case driver.KindBlob:
		return c.String()
	}
	return cellAny(c)
}

func rowInsert(a *App, t *workTab, r int) string {
	g := t.grid
	names := make([]string, len(g.cols))
	vals := make([]string, len(g.cols))
	for i, c := range g.cols {
		names[i] = driver.QuoteIdent(c.Name)
		vals[i] = g.rows[r][i].SQLLiteral()
	}
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);", driver.QualifiedTable(qualSchema(a.cfg, t.schema), t.table),
		strings.Join(names, ", "), strings.Join(vals, ", "))
}

func selectionMarkdown(g *gridState, r, c int) string {
	r0, r1, c0, c1 := r, r, 0, len(g.cols)-1
	if g.ar >= 0 && g.selected(r, c) {
		r0, r1 = min(g.ar, g.fr), max(g.ar, g.fr)
		c0, c1 = min(g.ac, g.fc), max(g.ac, g.fc)
		if g.rowSel {
			c0, c1 = 0, len(g.cols)-1
		}
	}
	var b strings.Builder
	esc := func(s string) string { return strings.ReplaceAll(strings.ReplaceAll(s, "|", `\|`), "\n", " ") }
	b.WriteString("|")
	for i := c0; i <= c1; i++ {
		b.WriteString(" " + esc(g.cols[i].Name) + " |")
	}
	b.WriteString("\n|")
	for i := c0; i <= c1; i++ {
		_ = i
		b.WriteString(" --- |")
	}
	for i := r0; i <= r1 && i < len(g.rows); i++ {
		b.WriteString("\n|")
		for j := c0; j <= c1; j++ {
			b.WriteString(" " + esc(g.rows[i][j].String()) + " |")
		}
	}
	return b.String()
}

// editedCell wraps edited inspector text in a cell of the original text kind.
func editedCell(kind driver.CellKind, text string) driver.CellValue {
	if kind == driver.KindJSON {
		return driver.JSON(text)
	}
	return driver.Text(text)
}

// rowCSV renders row r as one CSV record (NULL as an empty field).
func rowCSV(g *gridState, r int) string {
	var b strings.Builder
	w := csv.NewWriter(&b)
	rec := make([]string, len(g.cols))
	for i := range g.cols {
		rec[i] = cellPlain(g.rows[r][i])
	}
	_ = w.Write(rec)
	w.Flush()
	return strings.TrimRight(b.String(), "\n")
}

// filterByValue narrows the grid to rows equal to the cell at (r, c): a WHERE
// condition for table tabs, the quick filter for query results.
func (a *App) filterByValue(ctx context.Context, t *workTab, r, c int) error {
	g := t.grid
	cell := g.rows[r][c]
	if t.kind == tabQuery {
		t.quick = cellPlain(cell)
		t.applyFilters()
		return a.redraw(ctx)
	}
	col := driver.QuoteIdent(g.cols[c].Name)
	if cell.IsNull() {
		t.filter = col + " IS NULL"
	} else {
		t.filter = col + " = " + cell.SQLLiteral()
	}
	a.page.Form().SetFormValue("tbl-filter", t.filter)
	t.page = 0
	return a.loadTablePage(ctx, t)
}
