package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/ui"

	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/driver"
	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/sqlkit"
)

func (a *App) copyClipboard(text string) { ownframe.Write(text) }

// openTable opens (or focuses) the data tab for a table or view and loads page 0.
func (a *App) openTable(ctx context.Context, n driver.SchemaNode) error {
	if i := a.findTab(tabTable, n.Schema, n.Name); i >= 0 {
		a.activeTab = i
		return a.redraw(ctx)
	}
	t := a.addTab(&workTab{kind: tabTable, title: n.Name, schema: n.Schema, table: n.Name, pageSize: defaultPageSize})
	a.focus = "grid"
	if a.drv != nil {
		cols, _ := a.drv.TableColumns(ctx, n.Schema, n.Name)
		for _, c := range cols {
			if c.PrimaryKey {
				t.pk = append(t.pk, c.Name)
			}
		}
		t.writable = n.Kind == driver.NodeTable && len(t.pk) > 0 && !a.cfg.ReadOnly
		t.cols = cols
	}
	return a.loadTablePage(ctx, t)
}

// tableSQL builds the data query: filter, sort, and one page of rows.
func (a *App) tableSQL(t *workTab) (page, count string, err error) {
	base := driver.QualifiedTable(qualSchema(a.cfg, t.schema), t.table)
	where := ""
	if f := strings.TrimSpace(t.filter); f != "" {
		if strings.Contains(f, ";") {
			return "", "", fmt.Errorf("filter cannot contain ';'")
		}
		if cerr := driver.CheckReadOnly("SELECT 1 WHERE " + f); cerr != nil {
			return "", "", fmt.Errorf("filter must be a plain condition: %w", cerr)
		}
		where = " WHERE " + f
	}
	order := ""
	if t.sortCol > 0 && t.sortCol-1 < len(t.cols) && t.sortDir != 0 {
		dir := map[int]string{1: "ASC", 2: "DESC"}[t.sortDir]
		order = " ORDER BY " + driver.QuoteIdent(t.cols[t.sortCol-1].Name) + " " + dir
	}
	page = fmt.Sprintf("SELECT * FROM %s%s%s LIMIT %d OFFSET %d", base, where, order, t.pageSize, t.page*t.pageSize)
	count = "SELECT COUNT(*) FROM " + base + where
	return page, count, nil
}

func (a *App) loadTablePage(ctx context.Context, t *workTab) error {
	if a.drv == nil {
		return a.redraw(ctx)
	}
	if t.pageSize == 0 {
		t.pageSize = 100
	}
	t.page = max(0, t.page)
	pageSQL, countSQL, err := a.tableSQL(t)
	if err != nil {
		t.err = &queryErr{Message: err.Error()}
		return a.redraw(ctx)
	}
	rctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), queryTimeout)
	a.running = &runState{cancel: cancel, started: a.clock(), shown: a.clock(), tabID: t.id}
	drv, tabID, limit := a.drv, t.id, t.pageSize
	go func() {
		out := queryDone{tabID: tabID, kind: "table", sql: pageSQL}
		cnt := execOne(rctx, drv, countSQL, 1)
		if cnt.Err == nil && len(cnt.Res.Rows) == 1 && len(cnt.Res.Rows[0]) == 1 {
			out.total = cnt.Res.Rows[0][0].I
		}
		r := execOne(rctx, drv, pageSQL, limit)
		out.results = []stmtResult{r}
		out.err = r.Err
		cancel()
		a.doneCh <- out
	}()
	return a.redraw(ctx)
}

func (a *App) applyTablePage(t *workTab, r queryDone) {
	if t == nil || len(r.results) == 0 {
		return
	}
	s := r.results[0]
	a.lastQuery = lastQueryInfo{Dur: s.Dur, Rows: len(s.Res.Rows), Set: true, Err: s.Err != nil}
	if s.Err != nil {
		t.err = a.describeErr(s.Err, r.sql, nil)
		return
	}
	t.err = nil
	prev := t.grid
	t.grid = newGridState(s.Res)
	if prev != nil && len(prev.widths) == len(t.grid.widths) {
		copy(t.grid.widths, prev.widths)
	}
	t.totalRows = r.total
	t.cs, t.editing = nil, nil
	if len(s.Res.Columns) > 0 && len(t.cols) == 0 {
		t.cols = s.Res.Columns
	}
	a.captureSlow(s)
}

// openStructure opens the Structure / DDL inspector for a table.
func (a *App) openStructure(ctx context.Context, n driver.SchemaNode) error {
	if i := a.findTab(tabStructure, n.Schema, n.Name); i >= 0 {
		a.activeTab = i
		return a.redraw(ctx)
	}
	t := a.addTab(&workTab{kind: tabStructure, title: n.Name + " (structure)", schema: n.Schema, table: n.Name, structSub: "columns"})
	if a.drv != nil {
		t.cols, _ = a.drv.TableColumns(ctx, n.Schema, n.Name)
		t.indexes, _ = a.drv.TableIndexes(ctx, n.Schema, n.Name)
		t.fks, _ = a.drv.TableForeignKeys(ctx, n.Schema, n.Name)
		t.ddl, _ = a.drv.TableDDL(ctx, n.Schema, n.Name)
	}
	return a.redraw(ctx)
}

func (a *App) structureView(t *workTab) *ui.Element {
	sub := func(id, label string) *ui.Element {
		cls := "qr-seg-item"
		if t.structSub == id {
			cls += " qr-seg-on"
		}
		return ui.Div().ID("st-sub-"+id).Class(cls).Text(label).Attr("role", ariaRole("tab")).
			OnClick(func(ctx context.Context) error { t.structSub = id; return a.redraw(ctx) })
	}
	head := ui.Div().Class("qr-toolbar").Children(
		ui.Span().Class("qr-sidebar-title").Text(structureTitle(a.cfg, t)),
		ui.Div().Class("qr-seg").Children(sub("columns", "Columns"), sub("indexes", "Indexes & Keys"), sub("ddl", "DDL")),
	)
	view := ui.Div().ID("structure-tab").Class("qr-tab-body").Children(head)
	body := ui.Div().ID("structure-body").Class("qr-structure")
	switch t.structSub {
	case "indexes":
		body.Children(a.indexesTable(t))
	case "ddl":
		body.Children(
			a.tbButton("ddl-copy", "Copy DDL", "qr-btn", false, func(ctx context.Context) error {
				a.copyClipboard(t.ddl)
				a.notify("DDL copied")
				return a.redraw(ctx)
			}),
			ui.Div().ID("ddl-code").Class("qr-ddl").Children(highlightedBlock(t.ddl, a.dialect())...),
		)
	default:
		body.Children(a.columnsTable(t))
	}
	return view.Children(a.pane("structure", "Structure scroll", "", body))
}

func tableRow(cls string, cells ...*ui.Element) *ui.Element {
	return ui.Div().Class("qr-trow "+cls).Attr("role", ariaRole("row")).Children(cells...)
}

func tcell(text string, w int, cls string) *ui.Element {
	return ui.Div().Class("qr-tcell "+cls).Width(ui.Px(float64(w))).Attr("role", ariaRole("cell")).Text(text)
}

func (a *App) columnsTable(t *workTab) *ui.Element {
	tbl := ui.Div().ID("columns-table").Class("qr-table").Attr("role", ariaRole("table")).Attr("aria-label", "Columns")
	tbl.Children(tableRow("qr-thead", tcell("Name", colName, ""), tcell("Type", colType, ""), tcell("Nullable", colNullable, ""), tcell("Default", colDefault, ""), tcell("Keys", colKeys, "")))
	fk := map[string]bool{}
	for _, f := range t.fks {
		for _, c := range f.Columns {
			fk[c] = true
		}
	}
	for i, c := range t.cols {
		keys := ""
		if c.PrimaryKey {
			keys = "PK"
		}
		if c.ForeignKey || fk[c.Name] {
			keys = strings.TrimSpace(keys + " FK")
		}
		null := "NO"
		if c.Nullable {
			null = "YES"
		}
		row := tableRow("", tcell(c.Name, colName, "qr-mono"), tcell(c.Type, colType, "qr-mono qr-muted"), tcell(null, colNullable, ""), tcell(c.DefaultVal, colDefault, "qr-mono qr-muted"), tcell(keys, colKeys, "qr-keytext"))
		row.ID(fmt.Sprintf("col-row-%d", i))
		tbl.Children(row)
	}
	return tbl
}

func (a *App) indexesTable(t *workTab) *ui.Element {
	tbl := ui.Div().ID("indexes-table").Class("qr-table").Attr("role", ariaRole("table")).Attr("aria-label", "Indexes and keys")
	tbl.Children(tableRow("qr-thead", tcell("Name", colIndexName, ""), tcell("Columns", colIndexCols, ""), tcell("Unique", colUnique, ""), tcell("References", colRefs, "")))
	for i, ix := range t.indexes {
		u := ""
		if ix.Unique || ix.Primary {
			u = "YES"
		}
		row := tableRow("", tcell(ix.Name, colIndexName, "qr-mono"), tcell(strings.Join(ix.Columns, ", "), colIndexCols, "qr-mono"), tcell(u, colUnique, ""), tcell("", colRefs, ""))
		row.ID(fmt.Sprintf("idx-row-%d", i))
		tbl.Children(row)
	}
	for i, f := range t.fks {
		ref := f.RefTable + "(" + strings.Join(f.RefColumns, ", ") + ")"
		row := tableRow("", tcell(f.Name, colIndexName, "qr-mono"), tcell(strings.Join(f.Columns, ", "), colIndexCols, "qr-mono"), tcell("FK", colUnique, "qr-keytext"), tcell(ref, colRefs, "qr-mono"))
		row.ID(fmt.Sprintf("fk-row-%d", i))
		tbl.Children(row)
	}
	return tbl
}

// highlightedBlock renders SQL as colored spans, one div per line.
func highlightedBlock(src string, d sqlkit.Dialect) []*ui.Element {
	toks := sqlkit.Tokenize(src, d)
	var lines []*ui.Element
	cur := ui.Div().Class("qr-ddl-line")
	flush := func() { lines = append(lines, cur); cur = ui.Div().Class("qr-ddl-line") }
	for _, tk := range toks {
		txt := tk.Text(src)
		parts := strings.Split(txt, "\n")
		for i, p := range parts {
			if i > 0 {
				flush()
			}
			if p != "" {
				cur.Children(ui.Span().Class("qr-run " + synClass(tk.Kind)).Text(p))
			}
		}
	}
	flush()
	return lines
}

var _ = time.Second

// structureTitle is the plain, unquoted table name shown above the inspector.
func structureTitle(cfg driver.ConnectionConfig, t *workTab) string {
	if sch := qualSchema(cfg, t.schema); sch != "" {
		return sch + "." + t.table
	}
	return t.table
}
