package driver

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	_ "modernc.org/sqlite" // pure-Go SQLite driver ("sqlite")
)

// sqliteDriver implements Driver over modernc.org/sqlite.
type sqliteDriver struct {
	cfg ConnectionConfig
	db  *sql.DB
	trk *tracker
}

// OpenSQLite opens cfg.FilePath (or ":memory:") and verifies it with Ping.
// ReadOnly connections set PRAGMA query_only and also reject writes in Go.
func OpenSQLite(ctx context.Context, cfg ConnectionConfig) (Driver, error) {
	if cfg.FilePath == "" {
		return nil, fmt.Errorf("driver: sqlite connection %q has no file path", cfg.Name)
	}
	mem := cfg.FilePath == ":memory:"
	if !mem && cfg.ReadOnly {
		if _, err := os.Stat(cfg.FilePath); err != nil {
			return nil, fmt.Errorf("driver: open %s: %w", cfg.FilePath, err)
		}
	}
	db, err := sql.Open("sqlite", sqliteDSN(cfg))
	if err != nil {
		return nil, err
	}
	if mem {
		db.SetMaxOpenConns(1) // each :memory: connection is its own database
	} else {
		db.SetMaxOpenConns(sqliteMaxConns)
	}
	db.SetMaxIdleConns(sqliteIdleConns)
	d := &sqliteDriver{cfg: cfg, db: db, trk: newTracker("", cfg.FilePath)}
	pctx, cancel := context.WithTimeout(ctx, connectTimeout(cfg))
	defer cancel()
	if err := d.Ping(pctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return d, nil
}

// Pool and timeout defaults for file-backed SQLite connections.
const (
	sqliteMaxConns        = 4
	sqliteIdleConns       = 2
	defaultConnectTimeout = 5 * time.Second
)

func connectTimeout(cfg ConnectionConfig) time.Duration {
	if cfg.ConnectTimeout > 0 {
		return cfg.ConnectTimeout
	}
	return defaultConnectTimeout
}

func sqliteDSN(cfg ConnectionConfig) string {
	path := strings.NewReplacer("%", "%25", "?", "%3f", "#", "%23").Replace(cfg.FilePath)
	q := "?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)"
	if cfg.ReadOnly {
		q += "&_pragma=query_only(1)"
	}
	return "file:" + path + q
}

func (d *sqliteDriver) Engine() EngineKind       { return EngineSQLite }
func (d *sqliteDriver) Config() ConnectionConfig { return d.cfg }
func (d *sqliteDriver) Close() error             { return d.db.Close() }

func (d *sqliteDriver) Ping(ctx context.Context) error { return d.db.PingContext(ctx) }

func (d *sqliteDriver) Version(ctx context.Context) (string, error) {
	var v string
	if err := d.db.QueryRowContext(ctx, "SELECT sqlite_version()").Scan(&v); err != nil {
		return "", err
	}
	return "SQLite " + v, nil
}

func (d *sqliteDriver) Query(ctx context.Context, sqlText string, maxRows int) (QueryResult, error) {
	if d.cfg.ReadOnly {
		if err := CheckReadOnly(sqlText); err != nil {
			return QueryResult{}, err
		}
	}
	if maxRows <= 0 {
		maxRows = DefaultMaxRows
	}
	ctx, done := d.trk.begin(ctx, sqlText)
	defer done()
	start := time.Now()
	rows, err := d.db.QueryContext(ctx, sqlText)
	if err != nil {
		return QueryResult{}, err
	}
	defer func() { _ = rows.Close() }()
	res, err := scanRows(rows, maxRows)
	res.Duration = time.Since(start)
	return res, err
}

// scanRows decodes rows into a QueryResult, stopping at maxRows and flagging
// Truncated when at least one more row exists.
func scanRows(rows *sql.Rows, maxRows int) (QueryResult, error) {
	var res QueryResult
	cts, err := rows.ColumnTypes()
	if err != nil {
		return res, err
	}
	decl := make([]string, len(cts))
	res.Columns = make([]ColumnMeta, len(cts))
	for i, ct := range cts {
		decl[i] = ct.DatabaseTypeName()
		nullable, _ := ct.Nullable()
		res.Columns[i] = ColumnMeta{Name: ct.Name(), Type: decl[i], Nullable: nullable}
	}
	vals := make([]any, len(cts))
	ptrs := make([]any, len(cts))
	for i := range vals {
		ptrs[i] = &vals[i]
	}
	for rows.Next() {
		if len(res.Rows) >= maxRows {
			res.Truncated = true
			break
		}
		if err := rows.Scan(ptrs...); err != nil {
			return res, err
		}
		row := make([]CellValue, len(cts))
		for i, v := range vals {
			row[i] = FromAny(v, decl[i])
		}
		res.Rows = append(res.Rows, row)
	}
	if !res.Truncated {
		if err := rows.Err(); err != nil {
			return res, err
		}
	}
	return res, nil
}

func (d *sqliteDriver) Exec(ctx context.Context, sqlText string) (int64, error) {
	if d.cfg.ReadOnly {
		return 0, ErrReadOnly
	}
	ctx, done := d.trk.begin(ctx, sqlText)
	defer done()
	r, err := d.db.ExecContext(ctx, sqlText)
	if err != nil {
		return 0, err
	}
	n, _ := r.RowsAffected()
	return n, nil
}

func (d *sqliteDriver) ListSchemas(ctx context.Context) ([]SchemaNode, error) {
	rows, err := d.db.QueryContext(ctx,
		`SELECT type, name, tbl_name, COALESCE(sql, '') FROM sqlite_master
		 WHERE name NOT LIKE 'sqlite\_%' ESCAPE '\' ORDER BY type, name`)
	if err != nil {
		return nil, err
	}
	var tables, views, indexes, triggers []SchemaNode
	for rows.Next() {
		var typ, name, tbl, ddl string
		if err := rows.Scan(&typ, &name, &tbl, &ddl); err != nil {
			_ = rows.Close()
			return nil, err
		}
		n := SchemaNode{Name: name, Schema: "main", Table: tbl, RowCount: -1}
		switch typ {
		case "table":
			n.Kind = NodeTable
			tables = append(tables, n)
		case "view":
			n.Kind = NodeView
			views = append(views, n)
		case "index":
			n.Kind = NodeIndex
			if strings.Contains(strings.ToUpper(ddl), "CREATE UNIQUE") {
				n.Detail = "UNIQUE"
			}
			indexes = append(indexes, n)
		case "trigger":
			n.Kind = NodeTrigger
			triggers = append(triggers, n)
		}
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, err
	}
	_ = rows.Close()
	for i := range tables {
		var c int64
		err := d.db.QueryRowContext(ctx, "SELECT count(*) FROM "+QuoteIdent(tables[i].Name)).Scan(&c)
		if err == nil {
			tables[i].RowCount = c
		} else if ctx.Err() != nil {
			return nil, ctx.Err()
		}
	}
	main := SchemaNode{Name: "main", Kind: NodeSchema, Schema: "main", RowCount: -1}
	main.Children = appendGroups(main.Children, "Tables", tables, "Views", views, "Indexes", indexes, "Triggers", triggers)
	return []SchemaNode{main}, nil
}

// appendGroups appends non-empty NodeGroup folders; args alternate title, nodes.
func appendGroups(dst []SchemaNode, args ...any) []SchemaNode {
	for i := 0; i+1 < len(args); i += 2 {
		title, ok := args[i].(string)
		nodes, ok2 := args[i+1].([]SchemaNode)
		if !ok || !ok2 || len(nodes) == 0 {
			continue
		}
		schema := nodes[0].Schema
		dst = append(dst, SchemaNode{Name: title, Kind: NodeGroup, Schema: schema, RowCount: int64(len(nodes)), Children: nodes})
	}
	return dst
}

func sqliteSchema(schema string) string {
	if schema == "" {
		return "main"
	}
	return schema
}

func (d *sqliteDriver) TableColumns(ctx context.Context, schema, table string) ([]ColumnMeta, error) {
	schema = sqliteSchema(schema)
	rows, err := d.db.QueryContext(ctx,
		`SELECT name, type, "notnull", COALESCE(dflt_value, ''), pk FROM pragma_table_info(?, ?)`, table, schema)
	if err != nil {
		return nil, err
	}
	var cols []ColumnMeta
	for rows.Next() {
		var c ColumnMeta
		var notnull, pk int
		if err := rows.Scan(&c.Name, &c.Type, &notnull, &c.DefaultVal, &pk); err != nil {
			_ = rows.Close()
			return nil, err
		}
		c.PrimaryKey = pk > 0
		c.Nullable = notnull == 0 && pk == 0
		cols = append(cols, c)
	}
	err = rows.Err()
	_ = rows.Close()
	if err != nil {
		return nil, err
	}
	if len(cols) == 0 {
		return nil, fmt.Errorf("driver: table %s.%s not found", schema, table)
	}
	fks, err := d.TableForeignKeys(ctx, schema, table)
	if err == nil {
		for _, fk := range fks {
			for _, name := range fk.Columns {
				for i := range cols {
					if cols[i].Name == name {
						cols[i].ForeignKey = true
					}
				}
			}
		}
	}
	return cols, nil
}

func (d *sqliteDriver) TableIndexes(ctx context.Context, schema, table string) ([]IndexMeta, error) {
	schema = sqliteSchema(schema)
	rows, err := d.db.QueryContext(ctx,
		`SELECT name, "unique", origin FROM pragma_index_list(?, ?) ORDER BY seq`, table, schema)
	if err != nil {
		return nil, err
	}
	var out []IndexMeta
	for rows.Next() {
		var ix IndexMeta
		var uniq int
		var origin string
		if err := rows.Scan(&ix.Name, &uniq, &origin); err != nil {
			_ = rows.Close()
			return nil, err
		}
		ix.Unique = uniq != 0
		ix.Primary = origin == "pk"
		ix.Method = "btree"
		out = append(out, ix)
	}
	err = rows.Err()
	_ = rows.Close()
	if err != nil {
		return nil, err
	}
	for i := range out {
		cr, err := d.db.QueryContext(ctx, `SELECT COALESCE(name, '') FROM pragma_index_info(?, ?) ORDER BY seqno`, out[i].Name, schema)
		if err != nil {
			return nil, err
		}
		for cr.Next() {
			var c string
			if err := cr.Scan(&c); err != nil {
				_ = cr.Close()
				return nil, err
			}
			out[i].Columns = append(out[i].Columns, c)
		}
		_ = cr.Close()
		var def sql.NullString
		_ = d.db.QueryRowContext(ctx, `SELECT sql FROM sqlite_master WHERE type='index' AND name=?`, out[i].Name).Scan(&def)
		out[i].Definition = def.String
	}
	hasPK := false
	for _, ix := range out {
		if ix.Primary {
			hasPK = true
		}
	}
	if !hasPK { // INTEGER PRIMARY KEY aliases rowid and has no index entry
		cols, err := d.pkColumns(ctx, schema, table)
		if err == nil && len(cols) > 0 {
			out = append([]IndexMeta{{Name: "PRIMARY KEY", Columns: cols, Unique: true, Primary: true, Method: "rowid"}}, out...)
		}
	}
	return out, nil
}

func (d *sqliteDriver) pkColumns(ctx context.Context, schema, table string) ([]string, error) {
	rows, err := d.db.QueryContext(ctx, `SELECT name FROM pragma_table_info(?, ?) WHERE pk > 0 ORDER BY pk`, table, schema)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var cols []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		cols = append(cols, c)
	}
	return cols, rows.Err()
}

func (d *sqliteDriver) TableForeignKeys(ctx context.Context, schema, table string) ([]ForeignKeyMeta, error) {
	schema = sqliteSchema(schema)
	rows, err := d.db.QueryContext(ctx,
		`SELECT id, "table", "from", COALESCE("to", ''), on_update, on_delete
		 FROM pragma_foreign_key_list(?, ?) ORDER BY id, seq`, table, schema)
	if err != nil {
		return nil, err
	}
	var out []ForeignKeyMeta
	byID := map[int]int{}
	for rows.Next() {
		var id int
		var ref, from, to, upd, del string
		if err := rows.Scan(&id, &ref, &from, &to, &upd, &del); err != nil {
			_ = rows.Close()
			return nil, err
		}
		idx, ok := byID[id]
		if !ok {
			idx = len(out)
			byID[id] = idx
			out = append(out, ForeignKeyMeta{
				Name: fmt.Sprintf("fk_%s_%d", table, id), RefSchema: schema, RefTable: ref,
				OnUpdate: upd, OnDelete: del,
			})
		}
		out[idx].Columns = append(out[idx].Columns, from)
		out[idx].RefColumns = append(out[idx].RefColumns, to)
	}
	err = rows.Err()
	_ = rows.Close()
	if err != nil {
		return nil, err
	}
	for i := range out { // "to" is NULL when the FK targets the parent's primary key
		if len(out[i].RefColumns) > 0 && out[i].RefColumns[0] == "" {
			if pk, err := d.pkColumns(ctx, schema, out[i].RefTable); err == nil {
				out[i].RefColumns = pk
			}
		}
	}
	return out, nil
}

func (d *sqliteDriver) TableDDL(ctx context.Context, schema, table string) (string, error) {
	var ddl sql.NullString
	err := d.db.QueryRowContext(ctx,
		`SELECT sql FROM sqlite_master WHERE name = ? AND type IN ('table','view')`, table).Scan(&ddl)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("driver: %s not found", table)
	}
	if err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString(strings.TrimSpace(ddl.String))
	b.WriteString(";\n")
	rows, err := d.db.QueryContext(ctx,
		`SELECT sql FROM sqlite_master WHERE type IN ('index','trigger') AND tbl_name = ? AND sql IS NOT NULL ORDER BY type DESC, name`, table)
	if err != nil {
		return b.String(), nil
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var s string
		if rows.Scan(&s) == nil {
			b.WriteString("\n" + strings.TrimSpace(s) + ";\n")
		}
	}
	return b.String(), nil
}

func (d *sqliteDriver) CancelQuery(_ context.Context, pid int) error { return d.trk.cancel(pid) }

func (d *sqliteDriver) CollectMetrics(ctx context.Context) (MetricSnapshot, error) {
	snap := MetricSnapshot{CapturedAt: time.Now(), CacheHitRatio: -1, Extra: map[string]float64{}}
	st := d.db.Stats()
	snap.ActiveConns, snap.IdleConns = st.InUse, st.Idle
	var pages, pageSize, free int64
	_ = d.db.QueryRowContext(ctx, "PRAGMA page_count").Scan(&pages)
	_ = d.db.QueryRowContext(ctx, "PRAGMA page_size").Scan(&pageSize)
	_ = d.db.QueryRowContext(ctx, "PRAGMA freelist_count").Scan(&free)
	snap.DBSizeBytes = pages * pageSize
	snap.Extra["page_count"] = float64(pages)
	snap.Extra["page_size"] = float64(pageSize)
	snap.Extra["freelist_count"] = float64(free)
	var jm string
	_ = d.db.QueryRowContext(ctx, "PRAGMA journal_mode").Scan(&jm)
	if strings.EqualFold(jm, "wal") {
		var busy, logFrames, ckpt int64
		if d.db.QueryRowContext(ctx, "PRAGMA wal_checkpoint(PASSIVE)").Scan(&busy, &logFrames, &ckpt) == nil {
			snap.Extra["wal_pages"] = float64(logFrames)
			snap.Storage = append(snap.Storage, StorageItem{Name: "WAL", Bytes: logFrames * pageSize})
		}
	}
	if rows, err := d.db.QueryContext(ctx, `SELECT name, SUM(pgsize) FROM dbstat GROUP BY name ORDER BY 2 DESC LIMIT 12`); err == nil {
		for rows.Next() {
			var it StorageItem
			if rows.Scan(&it.Name, &it.Bytes) == nil {
				snap.Storage = append(snap.Storage, it)
			}
		}
		_ = rows.Close()
	}
	if free > 0 {
		snap.Storage = append(snap.Storage, StorageItem{Name: "free pages", Bytes: free * pageSize})
	}
	stmts, last, avg := d.trk.totals()
	snap.CommitTotal = stmts
	snap.Extra["last_query_ms"] = last
	snap.Extra["avg_query_ms"] = avg
	snap.ActiveQueries = d.trk.active()
	return snap, nil
}

// Explain runs EXPLAIN QUERY PLAN and maps the id/parent rows into a tree whose
// synthetic root (ID 0, "QUERY PLAN") parents the top-level steps. With
// analyze, the statement is also executed once (read statements only) to time
// it; SQLite has no per-node timing.
func (d *sqliteDriver) Explain(ctx context.Context, sqlText string, analyze bool) (ExplainPlan, error) {
	if d.cfg.ReadOnly {
		if err := CheckReadOnly(sqlText); err != nil {
			return ExplainPlan{}, err
		}
	}
	rows, err := d.db.QueryContext(ctx, "EXPLAIN QUERY PLAN "+sqlText)
	if err != nil {
		return ExplainPlan{}, err
	}
	plan := ExplainPlan{Engine: EngineSQLite}
	root := &PlanNode{ID: 0, ParentID: -1, NodeType: "QUERY PLAN", ActualMs: -1}
	nodes := map[int]*PlanNode{0: root}
	var raw strings.Builder
	type rec struct {
		id, parent int
		detail     string
	}
	var recs []rec
	for rows.Next() {
		var id, parent, unused int
		var detail string
		if err := rows.Scan(&id, &parent, &unused, &detail); err != nil {
			_ = rows.Close()
			return plan, err
		}
		recs = append(recs, rec{id, parent, detail})
		fmt.Fprintf(&raw, "%d|%d|%s\n", id, parent, detail)
	}
	err = rows.Err()
	_ = rows.Close()
	if err != nil {
		return plan, err
	}
	for _, r := range recs {
		n := classifySQLitePlan(r.detail)
		n.ID, n.ParentID = r.id, r.parent
		nodes[r.id] = n
	}
	for _, r := range recs {
		p, ok := nodes[r.parent]
		if !ok {
			p = root
			nodes[r.id].ParentID = 0
		}
		p.Children = append(p.Children, nodes[r.id])
		switch nodes[r.id].Scan {
		case ScanFull:
			plan.FullScans++
		case ScanSearch:
			plan.IndexSearches++
		}
	}
	plan.Root, plan.Raw = root, raw.String()
	if analyze {
		start := time.Now()
		if _, err := d.Query(ctx, sqlText, DefaultMaxRows); err != nil {
			return plan, err
		}
		plan.Analyzed = true
		plan.ExecutionMs = float64(time.Since(start)) / float64(time.Millisecond)
		root.ActualMs = plan.ExecutionMs
	}
	return plan, nil
}

// classifySQLitePlan parses details such as
//
//	SCAN orders
//	SEARCH users USING INDEX idx_users_email (email=?)
//	SEARCH o USING INTEGER PRIMARY KEY (rowid=?)
//	USE TEMP B-TREE FOR ORDER BY
func classifySQLitePlan(detail string) *PlanNode {
	n := &PlanNode{Detail: detail, NodeType: detail, ActualMs: -1}
	fields := strings.Fields(detail)
	if len(fields) == 0 {
		return n
	}
	switch fields[0] {
	case "SCAN":
		n.Scan = ScanFull
	case "SEARCH":
		n.Scan = ScanSearch
	default:
		return n
	}
	n.NodeType = fields[0]
	rest := fields[1:]
	if len(rest) > 0 && rest[0] == "TABLE" {
		rest = rest[1:]
	}
	if len(rest) > 0 {
		n.Relation = rest[0]
	}
	if i := strings.Index(detail, " USING "); i >= 0 {
		using := detail[i+len(" USING "):]
		switch {
		case strings.HasPrefix(using, "COVERING INDEX "):
			n.Index = firstWord(using[len("COVERING INDEX "):])
		case strings.HasPrefix(using, "INDEX "):
			n.Index = firstWord(using[len("INDEX "):])
		case strings.HasPrefix(using, "INTEGER PRIMARY KEY"):
			n.Index = "PRIMARY KEY"
		case strings.HasPrefix(using, "PRIMARY KEY"):
			n.Index = "PRIMARY KEY"
		}
	}
	return n
}

func firstWord(s string) string {
	if f := strings.Fields(s); len(f) > 0 {
		return f[0]
	}
	return ""
}

// ExecTx runs every statement with its bound arguments inside one transaction.
// On failure it rolls everything back and returns the index of the failing
// statement; on success it returns -1.
func (d *sqliteDriver) ExecTx(ctx context.Context, sqls []string, args [][]any) (int, error) {
	if d.cfg.ReadOnly {
		return -1, ErrReadOnly
	}
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return -1, err
	}
	defer func() { _ = tx.Rollback() }()
	for i, s := range sqls {
		var a []any
		if i < len(args) {
			a = args[i]
		}
		if _, err := tx.ExecContext(ctx, s, a...); err != nil {
			return i, err
		}
	}
	return -1, tx.Commit()
}
