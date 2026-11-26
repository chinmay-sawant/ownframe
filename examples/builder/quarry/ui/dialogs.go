package ui

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/ui"

	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/driver"
	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/sqlkit"
	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/store"
)

// connDraft is the connection manager form.
type connDraft struct {
	ID       string
	Engine   driver.EngineKind
	Name     string
	Env      string
	Path     string
	Host     string
	Port     string
	Database string
	User     string
	Password string
	SSL      string
	ReadOnly bool
	Mock     bool
	Save     bool
	Result   string
	ResultOK bool

	// Form state: edited fields, whether the user tried to submit, the
	// password reveal toggle, the pasted URL, and a running Test Connection.
	Touched   map[string]bool
	Submitted bool
	ShowPass  bool
	DSN       string
	DSNErr    string
	Testing   bool
}

type exportOpts struct {
	Format    string // csv | json | ndjson | sql
	FullTable bool
	Header    bool
	Delim     string
	Null      string
	Path      string
	Result    string
}

type importState struct {
	Path   string
	Data   *sqlkit.ImportData
	Target string
	Append bool
	Err    string
	Done   string
}

type confirmState struct {
	Title, Body string
	Typed       string // text the user must type; empty for a plain confirm
	Label       string
	OK          func(ctx context.Context) error
}

// ---- modal host ----

func (a *App) modalView() *ui.Element {
	var body *ui.Element
	wide := false
	switch a.modal {
	case "":
		return nil
	case "conn":
		body, wide = a.connDialog(), true
	case "export":
		body = a.exportDialog()
	case "import":
		body, wide = a.importDialog(), true
	case "history":
		body, wide = a.historyDialog(), true
	case "settings":
		body = a.settingsDialog()
	case "sqlpreview":
		body, wide = a.sqlPreviewDialog(), true
	case "confirm":
		body = a.confirmDialog()
	case "palette":
		return a.paletteView()
	default:
		return nil
	}
	cls := "qr-modal"
	if wide {
		cls += " qr-modal-wide"
	}
	return ui.Div().ID("modal-scrim").Class("qr-scrim").Children(
		ui.Div().ID("modal").Class(cls).Attr("role", ariaRole("dialog")).Attr("aria-modal", "true").Attr("aria-label", a.modalTitle()).Children(body),
	).OnClick(a.closeModal)
}

func (a *App) modalTitle() string {
	return map[string]string{"conn": "Connection manager", "export": "Export data", "import": "Import CSV",
		"history": "Query history and snippets", "settings": "Settings", "sqlpreview": "SQL preview", "confirm": "Confirm"}[a.modal]
}

func (a *App) closeModal(ctx context.Context) error {
	a.modal, a.modalErr = "", ""
	_ = a.page.Form().Focus(ctx, "")
	return a.redraw(ctx)
}

// modalFoot is the button row every dialog ends with: secondary actions first,
// the primary action last (rightmost).
func modalFoot(buttons ...*ui.Element) *ui.Element {
	return ui.Div().Class("qr-modal-foot").Children(buttons...)
}

// segItem is one option of a segmented control. It is a button so Tab reaches
// it and Space or Enter selects it.
func segItem(prefix, label, cls string) *ui.Element {
	return ui.Button(label).ID(prefix+"-"+strings.ToLower(strings.ReplaceAll(label, " ", ""))).Class(cls).Attr("type", "button").
		Attr("aria-pressed", fmt.Sprint(strings.Contains(cls, "qr-seg-on")))
}

func modalButton(id, label, class string, fn func(context.Context) error) *ui.Element {
	return ui.Button(label).ID(id).Class(class).Attr("type", "button").Attr("aria-label", label).OnClick(fn)
}

func row(label string, control *ui.Element) *ui.Element {
	return ui.Div().Class("qr-form-row").Children(ui.Span().Class("qr-form-label").Text(label), control)
}

func input(id, value, placeholder string) *ui.Element {
	return ui.Input().ID(id).Class("qr-input").Attr("type", "text").Attr("value", value).
		Attr("placeholder", placeholder).Attr("aria-label", placeholder)
}

func (a *App) segment(prefix string, cur string, opts []string, set func(string)) *ui.Element {
	seg := ui.Div().Class("qr-seg").Attr("role", ariaRole("group"))
	for _, o := range opts {
		o := o
		cls := "qr-seg-item"
		if cur == o {
			cls += " qr-seg-on"
		}
		seg.Children(segItem(prefix, o, cls).
			OnClick(func(ctx context.Context) error { a.syncConnDraft(); set(o); return a.redraw(ctx) }))
	}
	return seg
}

// ---- connection manager ----

func (a *App) openConnManager(ctx context.Context) error {
	a.menuOpen = ""
	a.modal = "conn"
	if a.connDraft.Env == "" {
		a.connDraft = connDraft{Engine: driver.EngineSQLite, Env: driver.EnvLocal, Port: "5432", SSL: "prefer", Name: "New connection"}
	}
	return a.redraw(ctx)
}

func (a *App) syncConnDraft() {
	if a.modal != "conn" {
		return
	}
	f := a.page.Form()
	get := func(id string, dst *string) {
		if v := f.FormValue(id); v != "" || f.FocusID() == id {
			*dst = v
		}
	}
	d := &a.connDraft
	get("cd-name", &d.Name)
	get("cd-path", &d.Path)
	get("cd-host", &d.Host)
	get("cd-port", &d.Port)
	get("cd-db", &d.Database)
	get("cd-user", &d.User)
	get("cd-pass", &d.Password)
	get("cd-dsn", &d.DSN)
}

func (d connDraft) config() driver.ConnectionConfig {
	port, _ := strconv.Atoi(d.Port)
	cfg := driver.ConnectionConfig{
		ID: d.ID, Name: strings.TrimSpace(d.Name), Engine: d.Engine, EnvTag: d.Env, ReadOnly: d.ReadOnly || d.Env == driver.EnvProd,
		FilePath: strings.TrimSpace(d.Path), Host: d.Host, Port: port, Database: d.Database, User: d.User,
		Password: d.Password, SSLMode: d.SSL, Mock: d.Mock && d.Engine == driver.EnginePostgres,
	}
	if cfg.Name == "" {
		cfg.Name = "Untitled"
	}
	return cfg
}

func (a *App) connDialog() *ui.Element {
	d := &a.connDraft
	errs := validateConn(*d)
	left := ui.Div().ID("conn-list").Class("qr-conn-list").Attr("aria-label", "Saved connections")
	for i, sc := range a.savedConnections() {
		sc := sc
		cls := "qr-list-row"
		if sc.ID == d.ID {
			cls += " qr-list-row-active"
		}
		left.Children(ui.Div().ID(fmt.Sprintf("conn-saved-%d", i)).Class(cls).Children(
			ui.Span().Class("qr-badge").Text(engineIcon(sc.Engine)),
			ui.Span().Class("qr-tree-label").Text(sc.Name),
			ui.Span().Class("qr-muted").Text(sc.EnvTag),
			ui.Span().ID(fmt.Sprintf("conn-del-%d", i)).Class("qr-tab-close").Attr("role", "button").Attr("aria-label", "Delete "+sc.Name).Children(a.icon("x", toneMuted, iconSmall)).
				OnClick(func(ctx context.Context) error {
					_ = a.prefs.DeleteConnection(ctx, sc.ID)
					return a.redraw(ctx)
				}),
		).OnClick(func(ctx context.Context) error {
			a.connDraft = draftFrom(sc.ConnectionConfig)
			return a.redraw(ctx)
		}))
	}
	if len(a.savedConnections()) == 0 {
		left.Children(ui.Div().Class("qr-note qr-pad").Text("No saved connections yet."))
	}
	shown := func(id string) string { return d.shown(errs, id) }
	form := ui.Div().ID("conn-form").Class("qr-conn-form")
	form.Children(
		row("Engine", a.segment("cd-engine", engineName(d.Engine), []string{"SQLite", "PostgreSQL"}, func(s string) {
			d.Engine = driver.EngineSQLite
			if s == "PostgreSQL" {
				d.Engine = driver.EnginePostgres
			}
		})),
		field("cd-name", "Name", marked(input("cd-name", d.Name, "Connection name"), shown("cd-name")), shown("cd-name"), ""),
		row("Environment", a.segment("cd-env", d.Env, []string{driver.EnvLocal, driver.EnvStaging, driver.EnvProd}, func(s string) {
			d.Env = s
			d.ReadOnly = s == driver.EnvProd
		})),
	)
	if d.Engine == driver.EngineSQLite {
		form.Children(
			field("cd-path", "Database file", marked(input("cd-path", d.Path, "/path/to/database.db"), shown("cd-path")), shown("cd-path"),
				"Tip: drop a .db, .sqlite, or .sqlite3 file onto the window to open it."),
		)
	} else {
		dsnErr := ""
		if d.Touched["cd-dsn"] {
			dsnErr = d.DSNErr
		}
		dsn := ui.Div().Class("qr-input-pass").Children(
			marked(input("cd-dsn", d.DSN, dsnPlaceholder), dsnErr),
			modalButton("cd-dsn-fill", "Fill", "qr-btn", a.fillFromDSN),
		)
		pass := "password"
		toggle := "Show"
		if d.ShowPass {
			pass, toggle = "text", "Hide"
		}
		passRow := ui.Div().Class("qr-input-pass").Children(
			input("cd-pass", d.Password, "Password").Attr("type", pass),
			modalButton("cd-pass-toggle", toggle, "qr-btn", a.togglePassword).Attr("aria-pressed", fmt.Sprint(d.ShowPass)),
		)
		form.Children(
			field("cd-dsn", "Connection URL", dsn, dsnErr, "Optional. Paste a URL and press Fill to populate the fields below."),
			field("cd-host", "Host", marked(input("cd-host", d.Host, "localhost"), shown("cd-host")), shown("cd-host"), ""),
			field("cd-port", "Port", marked(input("cd-port", d.Port, "5432"), shown("cd-port")), shown("cd-port"), ""),
			field("cd-db", "Database", marked(input("cd-db", d.Database, "postgres"), shown("cd-db")), shown("cd-db"), ""),
			field("cd-user", "User", input("cd-user", d.User, "postgres"), "", ""),
			field("cd-pass", "Password", passRow, "", "Stored only in memory; saved connections never keep it."),
			row("SSL mode", a.segment("cd-ssl", d.SSL, []string{"disable", "prefer", "require", "verify-full"}, func(s string) { d.SSL = s })),
			row("Offline mock", a.segment("cd-mock", map[bool]string{true: "On", false: "Off"}[d.Mock], []string{"Off", "On"}, func(s string) { d.Mock = s == "On" })),
		)
	}
	form.Children(row("Read-only", a.segment("cd-ro", map[bool]string{true: "Yes", false: "No"}[d.ReadOnly || d.Env == driver.EnvProd], []string{"No", "Yes"}, func(s string) {
		d.ReadOnly = s == "Yes"
	})))
	if st := a.connStatus(); st != nil {
		form.Children(st)
	}
	test := modalButton("conn-test", "Test Connection", "qr-btn", func(ctx context.Context) error { return a.testConn(ctx) })
	if d.Testing {
		test = ui.Button("Test Connection").ID("conn-test").Class("qr-btn qr-btn-off").Attr("type", "button").Attr("aria-disabled", "true")
	}
	return ui.Div().Class("qr-conn-body").Children(
		ui.Div().Class("qr-modal-title").Text("Connection manager"),
		ui.Div().Class("qr-split-h").Children(left, form),
		modalFoot(test,
			modalButton("conn-save", "Save", "qr-btn", func(ctx context.Context) error { return a.saveConn(ctx) }),
			modalButton("conn-close", "Close", "qr-btn", a.closeModal),
			modalButton("conn-connect", "Connect", "qr-btn qr-btn-primary", func(ctx context.Context) error { return a.connectDraft(ctx) })),
	)
}

func engineName(e driver.EngineKind) string {
	if e == driver.EnginePostgres {
		return "PostgreSQL"
	}
	return "SQLite"
}

func draftFrom(c driver.ConnectionConfig) connDraft {
	return connDraft{ID: c.ID, Engine: c.Engine, Name: c.Name, Env: c.EnvTag, Path: c.FilePath, Host: c.Host,
		Port: strconv.Itoa(max(c.Port, 0)), Database: c.Database, User: c.User, Password: c.Password, SSL: c.SSLMode,
		ReadOnly: c.ReadOnly, Mock: c.Mock}
}

// submitDraft syncs the form and reports whether it is valid; an invalid draft
// shows every message under its field.
func (a *App) submitDraft(ctx context.Context) (bool, error) {
	a.syncConnDraft()
	d := &a.connDraft
	d.Submitted = true
	if len(validateConn(*d)) > 0 {
		d.Result, d.ResultOK = "Fix the highlighted fields first.", false
		return false, a.redraw(ctx)
	}
	return true, nil
}

func (a *App) saveConn(ctx context.Context) error {
	if ok, err := a.submitDraft(ctx); !ok {
		return err
	}
	sc, err := a.prefs.SaveConnection(ctx, store.SavedConnection{ConnectionConfig: a.connDraft.config()})
	if err != nil {
		a.connDraft.Result, a.connDraft.ResultOK = err.Error(), false
	} else {
		a.connDraft.ID = sc.ID
		a.connDraft.Result, a.connDraft.ResultOK = "Saved "+sc.Name, true
	}
	return a.redraw(ctx)
}

func (a *App) connectDraft(ctx context.Context) error {
	if ok, err := a.submitDraft(ctx); !ok {
		return err
	}
	cfg := a.connDraft.config()
	if err := a.connect(ctx, cfg); err != nil {
		a.connDraft.Result, a.connDraft.ResultOK = "Connection failed: "+err.Error(), false
		return a.redraw(ctx)
	}
	return a.closeModal(ctx)
}

// ---- drag and drop ----

func isSQLiteFile(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".db", ".sqlite", ".sqlite3":
		return true
	}
	return false
}

// onDrop opens a dropped SQLite file as a connection, or a CSV for import.
func (a *App) onDrop(ctx context.Context, files []ownframe.Drop) error {
	for _, f := range files {
		switch {
		case f.IsDir:
			continue
		case isSQLiteFile(f.Name) && f.Path != "":
			return a.openSQLiteFile(ctx, f.Path)
		case strings.EqualFold(filepath.Ext(f.Name), ".csv") && f.Read != nil:
			data, err := f.Read()
			if err != nil {
				a.notify(err.Error())
				return a.redraw(ctx)
			}
			return a.loadImport(ctx, f.Name, data)
		}
	}
	return nil
}

func (a *App) openSQLiteFile(ctx context.Context, path string) error {
	cfg := driver.ConnectionConfig{Name: filepath.Base(path), Engine: driver.EngineSQLite, EnvTag: driver.EnvLocal, FilePath: path}
	if a.prefs != nil {
		if sc, err := a.prefs.RecentSQLiteFile(ctx, path); err == nil {
			cfg = sc.ConnectionConfig
		}
	}
	if err := a.connect(ctx, cfg); err != nil {
		a.notify("Cannot open " + filepath.Base(path) + ": " + err.Error())
	}
	return a.redraw(ctx)
}

// ---- export ----

func (a *App) openExport(ctx context.Context, full bool) error {
	t := a.activeTabPtr()
	if t == nil || t.grid == nil {
		return nil
	}
	a.exportOpts = exportOpts{Format: "csv", FullTable: full && t.kind == tabTable, Header: true, Delim: ",",
		Path: filepath.Join(os.TempDir(), "quarry-export.csv")}
	a.modal = "export"
	return a.redraw(ctx)
}

func (a *App) exportDialog() *ui.Element {
	o := &a.exportOpts
	t := a.activeTabPtr()
	scope := []string{"Current result"}
	if t != nil && t.kind == tabTable {
		scope = append(scope, "Full table")
	}
	return ui.Div().Class("qr-conn-body").Children(
		ui.Div().Class("qr-modal-title").Text("Export data"),
		row("Format", a.expSeg("ex-fmt", map[string]string{"csv": "CSV", "json": "JSON", "ndjson": "NDJSON", "sql": "SQL"}[o.Format], []string{"CSV", "JSON", "NDJSON", "SQL"},
			func(s string) { o.Format = strings.ToLower(s); o.Path = swapExt(o.Path, o.Format) })),
		row("Rows", a.expSeg("ex-scope", map[bool]string{true: "Full table", false: "Current result"}[o.FullTable], scope, func(s string) { o.FullTable = s == "Full table" })),
		row("Header row", a.expSeg("ex-hdr", map[bool]string{true: "Yes", false: "No"}[o.Header], []string{"Yes", "No"}, func(s string) { o.Header = s == "Yes" })),
		row("Delimiter", input("ex-delim", o.Delim, ",")),
		row("NULL as", input("ex-null", o.Null, "(empty)")),
		row("File", input("ex-path", o.Path, "/path/to/export")),
		ui.Div().ID("ex-result").Class("qr-note").Text(o.Result),
		modalFoot(
			modalButton("ex-copy", "Copy", "qr-btn", func(ctx context.Context) error { return a.runExport(ctx, false) }),
			modalButton("ex-close", "Close", "qr-btn", a.closeModal),
			modalButton("ex-run", "Export", "qr-btn qr-btn-primary", func(ctx context.Context) error { return a.runExport(ctx, true) }),
		),
	)
}

func (a *App) expSeg(prefix, cur string, opts []string, set func(string)) *ui.Element {
	seg := ui.Div().Class("qr-seg").Attr("role", ariaRole("group"))
	for _, o := range opts {
		o := o
		cls := "qr-seg-item"
		if cur == o {
			cls += " qr-seg-on"
		}
		seg.Children(segItem(prefix, o, cls).
			OnClick(func(ctx context.Context) error { a.syncExport(); set(o); return a.redraw(ctx) }))
	}
	return seg
}

func swapExt(p, format string) string {
	ext := map[string]string{"csv": ".csv", "json": ".json", "ndjson": ".ndjson", "sql": ".sql"}[format]
	return strings.TrimSuffix(p, filepath.Ext(p)) + ext
}

func (a *App) syncExport() {
	f := a.page.Form()
	if v := f.FormValue("ex-delim"); v != "" {
		a.exportOpts.Delim = v
	}
	a.exportOpts.Null = f.FormValue("ex-null")
	if v := f.FormValue("ex-path"); v != "" {
		a.exportOpts.Path = v
	}
}

// exportBytes renders the current result or the whole table in the chosen format.
func (a *App) exportBytes(ctx context.Context, t *workTab, o exportOpts) ([]byte, error) {
	g := t.grid
	cols := make([]string, len(g.cols))
	for i, c := range g.cols {
		cols[i] = c.Name
	}
	rows := g.rows
	if o.FullTable && t.kind == tabTable && a.drv != nil {
		res, err := a.drv.Query(ctx, "SELECT * FROM "+driver.QualifiedTable(qualSchema(a.cfg, t.schema), t.table), driver.DefaultMaxRows)
		if err != nil {
			return nil, err
		}
		rows = res.Rows
	}
	anyRows := make([][]any, len(rows))
	for i, r := range rows {
		anyRows[i] = make([]any, len(r))
		for j, c := range r {
			anyRows[i][j] = cellAny(c)
		}
	}
	var buf bytes.Buffer
	eo := sqlkit.ExportOptions{NoHeader: !o.Header, Null: o.Null, Table: t.table, Schema: qualSchema(a.cfg, t.schema), Dialect: a.dialect(), BatchSize: exportBatchSize}
	if r := []rune(o.Delim); len(r) > 0 {
		eo.Delimiter = r[0]
	}
	if eo.Table == "" {
		eo.Table = "result"
	}
	var err error
	switch o.Format {
	case "json":
		err = sqlkit.WriteJSON(&buf, cols, sqlkit.SliceRows(anyRows))
	case "ndjson":
		err = sqlkit.WriteNDJSON(&buf, cols, sqlkit.SliceRows(anyRows))
	case "sql":
		err = sqlkit.WriteSQL(&buf, cols, sqlkit.SliceRows(anyRows), eo)
	default:
		err = sqlkit.WriteCSV(&buf, cols, sqlkit.SliceRows(anyRows), eo)
	}
	return buf.Bytes(), err
}

func (a *App) runExport(ctx context.Context, toFile bool) error {
	a.syncExport()
	t := a.activeTabPtr()
	if t == nil || t.grid == nil {
		return nil
	}
	data, err := a.exportBytes(ctx, t, a.exportOpts)
	switch {
	case err != nil:
		a.exportOpts.Result = "Export failed: " + err.Error()
	case !toFile:
		a.copyClipboard(string(data))
		a.exportOpts.Result = fmt.Sprintf("Copied %s to the clipboard", humanBytes(len(data)))
	default:
		if werr := os.WriteFile(a.exportOpts.Path, data, exportFileMode); werr != nil {
			a.exportOpts.Result = "Export failed: " + werr.Error()
		} else {
			a.exportOpts.Result = fmt.Sprintf("Wrote %s to %s", humanBytes(len(data)), a.exportOpts.Path)
		}
	}
	return a.redraw(ctx)
}

// ---- import ----

func (a *App) openImport(ctx context.Context) error {
	a.importSt = importState{Target: "imported", Append: false}
	a.modal = "import"
	return a.redraw(ctx)
}

func (a *App) loadImport(ctx context.Context, name string, data []byte) error {
	d, err := sqlkit.ParseCSV(bytes.NewReader(data), sqlkit.ImportOptions{})
	a.modal = "import"
	a.importSt = importState{Path: name, Data: d, Target: strings.TrimSuffix(filepath.Base(name), filepath.Ext(name))}
	if err != nil {
		a.importSt.Err = err.Error()
	}
	return a.redraw(ctx)
}

func (a *App) importDialog() *ui.Element {
	s := &a.importSt
	body := ui.Div().Class("qr-conn-body").Children(
		ui.Div().Class("qr-modal-title").Text("Import CSV"),
		row("CSV file", input("im-path", s.Path, "/path/to/data.csv")),
		row("Target table", input("im-target", s.Target, "table name")),
		row("Mode", a.impSeg("im-mode", map[bool]string{true: "Append", false: "Create new"}[s.Append], []string{"Create new", "Append"}, func(v string) { s.Append = v == "Append" })),
	)
	body.Children(ui.Div().Class("qr-modal-foot").Children(modalButton("im-load", "Load Preview", "qr-btn", func(ctx context.Context) error { return a.previewImport(ctx) })))
	if s.Err != "" {
		body.Children(ui.Div().ID("im-error").Class("qr-note qr-note-err").Text(s.Err))
	}
	if s.Data != nil {
		pv := ui.Div().ID("im-preview").Class("qr-table").Attr("role", ariaRole("table")).Attr("aria-label", "Import preview")
		var hdr []*ui.Element
		for _, c := range s.Data.Columns {
			hdr = append(hdr, tcell(c.Name+" "+c.Type.SQL(a.dialect()), importColWidth, ""))
		}
		pv.Children(tableRow("qr-thead", hdr...))
		for i, r := range s.Data.Preview(importPreviewRows) {
			var cells []*ui.Element
			for _, v := range r {
				cells = append(cells, tcell(anyCell(v).String(), importColWidth, "qr-mono"))
			}
			pr := tableRow("", cells...)
			pr.ID(fmt.Sprintf("im-row-%d", i))
			pv.Children(pr)
		}
		body.Children(pv, ui.Div().ID("im-count").Class("qr-note").Text(fmt.Sprintf("%d rows, %d columns", len(s.Data.Rows), len(s.Data.Columns))))
	}
	if s.Done != "" {
		body.Children(ui.Div().ID("im-done").Class("qr-note qr-note-ok").Text(s.Done))
	}
	body.Children(modalFoot(
		modalButton("im-close", "Close", "qr-btn", a.closeModal),
		modalButton("im-run", "Import", "qr-btn qr-btn-primary", func(ctx context.Context) error { return a.runImport(ctx) }),
	))
	return body
}

func (a *App) impSeg(prefix, cur string, opts []string, set func(string)) *ui.Element {
	seg := ui.Div().Class("qr-seg").Attr("role", ariaRole("group"))
	for _, o := range opts {
		o := o
		cls := "qr-seg-item"
		if cur == o {
			cls += " qr-seg-on"
		}
		seg.Children(segItem(prefix, o, cls).
			OnClick(func(ctx context.Context) error { set(o); return a.redraw(ctx) }))
	}
	return seg
}

func (a *App) previewImport(ctx context.Context) error {
	f := a.page.Form()
	a.importSt.Path = f.FormValue("im-path")
	if v := f.FormValue("im-target"); v != "" {
		a.importSt.Target = v
	}
	data, err := os.ReadFile(a.importSt.Path)
	if err != nil {
		a.importSt.Err, a.importSt.Data = err.Error(), nil
		return a.redraw(ctx)
	}
	d, perr := sqlkit.ParseCSV(bytes.NewReader(data), sqlkit.ImportOptions{})
	a.importSt.Data, a.importSt.Err, a.importSt.Done = d, "", ""
	if perr != nil {
		a.importSt.Err = perr.Error()
	}
	return a.redraw(ctx)
}

// runImport creates or appends to the target table inside one transaction.
func (a *App) runImport(ctx context.Context) error {
	s := &a.importSt
	if v := a.page.Form().FormValue("im-target"); v != "" {
		s.Target = v
	}
	if s.Data == nil || a.drv == nil {
		s.Err = "Load a CSV file first"
		return a.redraw(ctx)
	}
	if a.cfg.ReadOnly {
		s.Err = "Connection is read-only"
		return a.redraw(ctx)
	}
	d := a.dialect()
	var sts []sqlkit.Statement
	if !s.Append {
		sts = append(sts, sqlkit.Statement{Kind: "CREATE", SQL: s.Data.CreateTableSQL(s.Target, d)})
	}
	ins := s.Data.InsertSQL(s.Target, d)
	for _, r := range s.Data.Rows {
		sts = append(sts, sqlkit.Statement{Kind: "INSERT", SQL: ins, Args: r})
	}
	if _, err := a.execAtomic(ctx, sts); err != nil {
		s.Err = "Import rolled back: " + err.Error()
		return a.redraw(ctx)
	}
	s.Err, s.Done = "", fmt.Sprintf("Imported %d rows into %s", len(s.Data.Rows), s.Target)
	a.refreshSchema(ctx)
	return a.redraw(ctx)
}

// ---- history and snippets ----

func (a *App) openHistory(ctx context.Context) error {
	a.modal = "history"
	a.histQuery = ""
	return a.redraw(ctx)
}

func (a *App) saveSnippet(ctx context.Context) error {
	t := a.activeTabPtr()
	if t == nil || t.ed == nil || strings.TrimSpace(t.ed.text) == "" || a.prefs == nil {
		return nil
	}
	name := strings.TrimSpace(strings.SplitN(strings.TrimSpace(t.ed.text), "\n", snippetFirstLines)[0])
	if len(name) > snippetNameMax {
		name = name[:snippetNameMax]
	}
	if _, err := a.prefs.SaveSnippet(ctx, store.Snippet{Name: name, SQL: t.ed.text}); err != nil {
		a.notify(err.Error())
	} else {
		a.notify("Snippet saved")
	}
	return a.redraw(ctx)
}

func (a *App) historyDialog() *ui.Element {
	q := strings.ToLower(a.histQuery)
	body := ui.Div().Class("qr-conn-body").Children(
		ui.Div().Class("qr-modal-title").Text("Query history and snippets"),
		input("hist-search", a.histQuery, "Search history"),
	)
	list := ui.Div().ID("hist-list").Class("qr-listcol")
	hist, _ := a.prefs.ListHistory(a.ctx, a.cfg.ID, historyLimit)
	n := 0
	for _, h := range hist {
		if q != "" && !strings.Contains(strings.ToLower(h.SQL), q) {
			continue
		}
		h := h
		status := "ok"
		cls := "qr-badge qr-badge-ok"
		if h.Error != "" {
			status, cls = "error", "qr-badge qr-badge-err"
		}
		list.Children(ui.Div().ID(fmt.Sprintf("hist-%d", n)).Class("qr-list-row").Children(
			ui.Span().Class(cls).Text(status),
			ui.Span().Class("qr-mono qr-tree-label qr-oneline").Text(truncate(oneLine(h.SQL), historyChars)),
			ui.Span().Class("qr-muted").Text(fmt.Sprintf("%s, %d rows", fmtDur(h.Duration), h.Rows)),
			ui.Span().ID(fmt.Sprintf("hist-star-%d", n)).Class("qr-iconbtn").Attr("role", "button").Attr("aria-label", "Save as snippet").Attr("title", "Save as snippet").Children(a.icon("star", toneMuted, iconSmall)).
				OnClick(func(ctx context.Context) error {
					_, _ = a.prefs.SaveSnippet(ctx, store.Snippet{Name: oneLine(h.SQL)[:min(snippetNameMax, len(oneLine(h.SQL)))], SQL: h.SQL})
					return a.redraw(ctx)
				}),
		).OnClick(func(ctx context.Context) error { return a.loadSQL(ctx, h.SQL) }))
		n++
	}
	if n == 0 {
		list.Children(ui.Div().ID("hist-empty").Class("qr-note qr-pad").Text("No matching history."))
	}
	body.Children(ui.Div().Class("qr-menu-head qr-muted").Text("History"), a.rowPane("hist", "History scroll", list, histPaneH))
	snips, _ := a.prefs.ListSnippets(a.ctx)
	sl := ui.Div().ID("snippet-list").Class("qr-listcol")
	for i, sn := range snips {
		sn := sn
		sl.Children(ui.Div().ID(fmt.Sprintf("snippet-%d", i)).Class("qr-list-row").Children(
			ui.Span().Class("qr-badge qr-badge-brand").Text("SNIP"),
			ui.Span().Class("qr-tree-label").Text(sn.Name),
		).OnClick(func(ctx context.Context) error { return a.loadSQL(ctx, sn.SQL) }))
	}
	body.Children(ui.Div().Class("qr-menu-head qr-muted").Text("Snippets"), a.rowPane("snip", "Snippet scroll", sl, snipPaneH),
		modalFoot(modalButton("hist-close", "Close", "qr-btn qr-btn-primary", a.closeModal)))
	return body
}

func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// loadSQL puts text in the current query tab (or a new one).
func (a *App) loadSQL(ctx context.Context, sql string) error {
	t := a.activeTabPtr()
	if t == nil || t.kind != tabQuery || strings.TrimSpace(t.ed.text) != "" {
		t = a.newQueryTab(sql)
	} else {
		t.ed = newEditor(sql)
	}
	t.dirty = true
	a.modal = ""
	a.focus = "editor"
	return a.redraw(ctx)
}

// ---- settings, previews, confirmations ----

// settingRow is one setting: its label and a one-line description on the left,
// the control on the right.
func settingRow(id, label, desc string, control *ui.Element) *ui.Element {
	return ui.Div().ID(id+"-row").Class("qr-set-row").Children(
		ui.Div().Class("qr-set-text").Children(
			ui.Span().Class("qr-set-label").Text(label),
			ui.Span().Class("qr-set-desc").Text(desc)),
		control)
}

// settingSection groups related rows under a heading and a description.
func settingSection(id, title, desc string, rows ...*ui.Element) *ui.Element {
	return ui.Div().ID(id).Class("qr-set-section").Attr("role", ariaRole("group")).Attr("aria-label", title).Children(
		ui.Div().Class("qr-set-head").Children(ui.Span().Class("qr-set-title").Text(title), ui.Span().Class("qr-set-desc").Text(desc)),
		ui.Div().Class("qr-set-rows").Children(rows...))
}

// toggleSwitch is an on/off switch built from a focusable button, so it takes
// Tab focus and Space/Enter like any control.
func toggleSwitch(id, label string, on bool, fn func(context.Context) error) *ui.Element {
	cls := "qr-switch"
	if on {
		cls += " qr-switch-on"
	}
	return ui.Button("").ID(id).Class(cls).Attr("type", "button").Attr("role", "checkbox").
		Attr("aria-checked", fmt.Sprint(on)).Attr("aria-label", label).
		Children(ui.Span().Class("qr-switch-knob")).OnClick(fn)
}

func (a *App) applySlowThreshold() {
	if n, err := strconv.Atoi(strings.TrimSpace(a.page.Form().FormValue("set-slow"))); err == nil && n >= 0 {
		a.slow.SetThreshold(time.Duration(n) * time.Millisecond)
		a.savePref("slow.ms", strconv.Itoa(n))
	}
}

// settingsDialog edits the persisted preferences. Every choice applies and is
// saved immediately; only the slow-query threshold needs Apply (or Done).
func (a *App) settingsDialog() *ui.Element {
	limits := []string{"100", "1,000", "10,000", "Unlimited"}
	polls := []string{"1s", "2s", "5s", "Paused"}
	pollCur := map[time.Duration]string{pollFast: "1s", pollNormal: "2s", pollSlow: "5s", 0: "Paused"}[a.mon.pollEvery]
	slow := ui.Div().Class("qr-input-pass").Children(
		input("set-slow", strconv.Itoa(int(a.slow.Threshold()/time.Millisecond)), "50").Class("qr-input qr-input-narrow"),
		ui.Span().Class("qr-muted").Text("ms"),
		modalButton("set-apply", "Apply", "qr-btn", func(ctx context.Context) error {
			a.applySlowThreshold()
			return a.redraw(ctx)
		}))
	return ui.Div().ID("settings").Class("qr-conn-body").Children(
		ui.Div().Class("qr-modal-title").Text("Settings"),
		settingSection("set-sec-appearance", "Appearance", "How Quarry looks.",
			settingRow("set-theme", "Theme", "QuarryDark or QuarryLight. Ctrl+Shift+T toggles it anywhere.",
				a.impSeg("set-theme", map[bool]string{true: "Dark", false: "Light"}[a.dark], []string{"Dark", "Light"}, func(v string) {
					_ = a.setThemeDark(a.ctx, v == "Dark")
				}))),
		settingSection("set-sec-query", "Queries", "Defaults for new query and table tabs.",
			settingRow("set-limit", "Default row limit", "Rows fetched per query before paging.",
				a.impSeg("set-limit", limitLabel(a.defLimit), limits, func(v string) {
					a.defLimit = map[string]int{"100": smallRows, "1,000": thousandRows, "10,000": tenThousandRows, "Unlimited": 0}[v]
					a.savePref("default.limit", strconv.Itoa(a.defLimit))
				})),
			settingRow("set-slowq", "Slow query threshold", "Statements slower than this are added to the slow log.", slow)),
		settingSection("set-sec-monitor", "Monitoring", "Live metrics on the Monitor tab.",
			settingRow("set-pollrow", "Polling interval", "How often the driver is sampled.",
				a.impSeg("set-poll", pollCur, polls, func(v string) {
					d := map[string]time.Duration{"1s": pollFast, "2s": pollNormal, "5s": pollSlow, "Paused": 0}[v]
					a.mon.pollEvery = d
					if a.mon.coll != nil {
						a.mon.coll.SetInterval(d)
					}
					a.savePref("monitor.poll", strconv.FormatInt(d.Milliseconds(), 10))
				}))),
		settingSection("set-sec-safety", "Safety", "Guards against destructive statements.",
			settingRow("set-confirmrow", "Confirm DROP and TRUNCATE", "Production connections always confirm, whatever this says.",
				toggleSwitch("set-confirm", "Confirm DROP and TRUNCATE", a.confirmDestructive, func(ctx context.Context) error {
					a.confirmDestructive = !a.confirmDestructive
					a.savePref("confirm.destructive", map[bool]string{true: "1", false: "0"}[a.confirmDestructive])
					return a.redraw(ctx)
				}))),
		modalFoot(modalButton("set-close", "Done", "qr-btn qr-btn-primary", func(ctx context.Context) error {
			a.applySlowThreshold()
			return a.closeModal(ctx)
		})),
	)
}

func (a *App) sqlPreviewDialog() *ui.Element {
	return ui.Div().Class("qr-conn-body").Children(
		ui.Div().Class("qr-modal-title").Text("SQL preview"),
		ui.Div().ID("sql-preview").Class("qr-ddl").Children(highlightedBlock(a.modalNote, a.dialect())...),
		modalFoot(modalButton("pv-close", "Close", "qr-btn qr-btn-primary", a.closeModal)),
	)
}

func (a *App) askConfirm(ctx context.Context, c *confirmState) error {
	a.confirm = c
	a.modal = "confirm"
	return a.redraw(ctx)
}

func (a *App) confirmDialog() *ui.Element {
	c := a.confirm
	body := ui.Div().Class("qr-conn-body").Children(
		ui.Div().Class("qr-modal-title").Text(c.Title),
		ui.Div().Class("qr-note").Text(c.Body),
	)
	if c.Typed != "" {
		body.Children(row("Type "+c.Typed, input("confirm-typed", "", c.Typed)))
	}
	if a.modalErr != "" {
		body.Children(ui.Div().ID("confirm-error").Class("qr-note qr-note-err").Text(a.modalErr))
	}
	body.Children(modalFoot(
		modalButton("confirm-cancel", "Cancel", "qr-btn", a.closeModal),
		modalButton("confirm-ok", c.Label, "qr-btn qr-btn-danger", a.confirmOK),
	))
	return body
}

// confirmOK runs the confirmed action once the typed text, if required, matches.
func (a *App) confirmOK(ctx context.Context) error {
	c := a.confirm
	if c == nil {
		return a.closeModal(ctx)
	}
	if c.Typed != "" && a.page.Form().FormValue("confirm-typed") != c.Typed {
		a.modalErr = "Type " + c.Typed + " exactly to continue"
		return a.redraw(ctx)
	}
	a.modal, a.modalErr = "", ""
	_ = a.page.Form().Focus(ctx, "")
	if err := a.redraw(ctx); err != nil { // closes the dialog and returns focus to the opener
		return err
	}
	return c.OK(ctx)
}

// askUnlockProd confirms before lifting the Prod read-only lock.
func (a *App) askUnlockProd(ctx context.Context) error {
	return a.askConfirm(ctx, &confirmState{
		Title: "Unlock writes on production?", Label: "Unlock writes",
		Body: "Writes on a Prod connection are blocked by default. Unlock only if you are sure.",
		OK: func(ctx context.Context) error {
			if err := a.reopenWritable(ctx, true); err != nil {
				a.notify(err.Error())
			}
			return a.redraw(ctx)
		},
	})
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "..."
}
