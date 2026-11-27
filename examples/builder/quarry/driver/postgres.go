package driver

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// pgDriver implements Driver over a pgx/v5 connection pool.
type pgDriver struct {
	cfg  ConnectionConfig
	pool *pgxpool.Pool

	mu       sync.Mutex
	prevAt   time.Time
	prevXact int64
}

// PostgresDSN builds the connection URL for cfg. The password never appears in
// error messages produced by this package.
func PostgresDSN(cfg ConnectionConfig) string {
	host := cfg.Host
	if host == "" {
		host = "localhost"
	}
	port := cfg.Port
	if port == 0 {
		port = 5432
	}
	ssl := cfg.SSLMode
	if ssl == "" {
		ssl = "prefer"
	}
	u := url.URL{
		Scheme: "postgres",
		Host:   net.JoinHostPort(host, strconv.Itoa(port)),
		Path:   "/" + cfg.Database,
	}
	if cfg.User != "" {
		if cfg.Password != "" {
			u.User = url.UserPassword(cfg.User, cfg.Password)
		} else {
			u.User = url.User(cfg.User)
		}
	}
	q := url.Values{}
	q.Set("sslmode", ssl)
	q.Set("connect_timeout", strconv.Itoa(int(connectTimeout(cfg).Seconds())))
	q.Set("application_name", "quarry")
	u.RawQuery = q.Encode()
	return u.String()
}

// OpenPostgres connects to cfg and verifies the server with Ping.
func OpenPostgres(ctx context.Context, cfg ConnectionConfig) (Driver, error) {
	pc, err := pgxpool.ParseConfig(PostgresDSN(cfg))
	if err != nil {
		return nil, fmt.Errorf("driver: invalid postgres settings: %w", err)
	}
	pc.MaxConns = 4
	cctx, cancel := context.WithTimeout(ctx, connectTimeout(cfg))
	defer cancel()
	pool, err := pgxpool.NewWithConfig(cctx, pc)
	if err != nil {
		return nil, fmt.Errorf("driver: connect %s:%d: %w", cfg.Host, cfg.Port, err)
	}
	d := &pgDriver{cfg: cfg, pool: pool}
	if err := d.Ping(cctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("driver: ping %s:%d: %w", cfg.Host, cfg.Port, err)
	}
	return d, nil
}

func (d *pgDriver) Engine() EngineKind       { return EnginePostgres }
func (d *pgDriver) Config() ConnectionConfig { return d.cfg }
func (d *pgDriver) Close() error             { d.pool.Close(); return nil }
func (d *pgDriver) Ping(ctx context.Context) error {
	return d.pool.Ping(ctx)
}

func (d *pgDriver) Version(ctx context.Context) (string, error) {
	var v string
	if err := d.pool.QueryRow(ctx, "SHOW server_version").Scan(&v); err != nil {
		return "", err
	}
	return "PostgreSQL " + v, nil
}

func (d *pgDriver) Query(ctx context.Context, sqlText string, maxRows int) (QueryResult, error) {
	if d.cfg.ReadOnly {
		if err := CheckReadOnly(sqlText); err != nil {
			return QueryResult{}, err
		}
	}
	if maxRows <= 0 {
		maxRows = DefaultMaxRows
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	start := time.Now()
	var q interface {
		Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	} = d.pool
	if d.cfg.ReadOnly { // server-side guard on top of CheckReadOnly
		tx, err := d.pool.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
		if err != nil {
			return QueryResult{}, err
		}
		defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
		q = tx
	}
	rows, err := q.Query(ctx, sqlText)
	if err != nil {
		return QueryResult{}, err
	}
	var res QueryResult
	tm := rows.Conn().TypeMap()
	fds := rows.FieldDescriptions()
	res.Columns = make([]ColumnMeta, len(fds))
	for i, fd := range fds {
		name := strconv.FormatUint(uint64(fd.DataTypeOID), 10)
		if t, ok := tm.TypeForOID(fd.DataTypeOID); ok {
			name = t.Name
		}
		res.Columns[i] = ColumnMeta{Name: fd.Name, Type: name, Nullable: true}
	}
	for rows.Next() {
		if len(res.Rows) >= maxRows {
			res.Truncated = true
			cancel() // abort the remaining stream instead of draining it
			break
		}
		vals, err := rows.Values()
		if err != nil {
			rows.Close()
			return res, err
		}
		row := make([]CellValue, len(vals))
		for i, v := range vals {
			row[i] = pgCell(v, res.Columns[i].Type)
		}
		res.Rows = append(res.Rows, row)
	}
	if !res.Truncated {
		if err := rows.Err(); err != nil {
			rows.Close()
			return res, err
		}
	}
	rows.Close()
	res.RowsAffected = rows.CommandTag().RowsAffected()
	res.Duration = time.Since(start)
	return res, nil
}

// pgCell converts a pgx decoded value into a CellValue.
func pgCell(v any, typ string) CellValue {
	switch x := v.(type) {
	case nil:
		return Null()
	case [16]byte:
		return Text(fmt.Sprintf("%x-%x-%x-%x-%x", x[0:4], x[4:6], x[6:8], x[8:10], x[10:16]))
	case map[string]any, []any:
		if typ == "json" || typ == "jsonb" {
			b, _ := json.Marshal(x)
			return JSON(string(b))
		}
		b, _ := json.Marshal(x)
		return Text(string(b))
	case driver.Valuer:
		dv, err := x.Value()
		if err != nil || dv == nil {
			return Null()
		}
		return FromAny(dv, typ)
	case fmt.Stringer:
		return Text(x.String())
	default:
		return FromAny(v, typ)
	}
}

func (d *pgDriver) Exec(ctx context.Context, sqlText string) (int64, error) {
	if d.cfg.ReadOnly {
		return 0, ErrReadOnly
	}
	tag, err := d.pool.Exec(ctx, sqlText)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (d *pgDriver) CancelQuery(ctx context.Context, pid int) error {
	var ok bool
	if err := d.pool.QueryRow(ctx, "SELECT pg_cancel_backend($1)", pid).Scan(&ok); err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("driver: backend %d was not cancelled", pid)
	}
	return nil
}

// TerminateQuery ends the whole backend session (pg_terminate_backend), the
// forceful counterpart of CancelQuery.
func (d *pgDriver) TerminateQuery(ctx context.Context, pid int) error {
	var ok bool
	if err := d.pool.QueryRow(ctx, "SELECT pg_terminate_backend($1)", pid).Scan(&ok); err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("driver: backend %d was not terminated", pid)
	}
	return nil
}

const pgSchemaFilter = `nspname NOT LIKE 'pg\_%' AND nspname <> 'information_schema'`

func (d *pgDriver) ListSchemas(ctx context.Context) ([]SchemaNode, error) {
	nss, err := d.pool.Query(ctx, `SELECT nspname FROM pg_namespace WHERE `+pgSchemaFilter+
		` ORDER BY (nspname = 'public') DESC, nspname`)
	if err != nil {
		return nil, err
	}
	var order []string
	for nss.Next() {
		var s string
		if err := nss.Scan(&s); err != nil {
			return nil, err
		}
		order = append(order, s)
	}
	nss.Close()
	if err := nss.Err(); err != nil {
		return nil, err
	}
	type bucket struct{ tables, views, matviews, indexes, triggers, funcs []SchemaNode }
	by := map[string]*bucket{}
	for _, s := range order {
		by[s] = &bucket{}
	}
	rel, err := d.pool.Query(ctx, `
SELECT n.nspname, c.relname, c.relkind::text, c.reltuples::bigint,
       COALESCE(t.relname, ''), COALESCE(x.indisunique, false)
FROM pg_class c
JOIN pg_namespace n ON n.oid = c.relnamespace
LEFT JOIN pg_index x ON x.indexrelid = c.oid
LEFT JOIN pg_class t ON t.oid = x.indrelid
WHERE c.relkind IN ('r','p','f','v','m','i') AND `+strings.ReplaceAll(pgSchemaFilter, "nspname", "n.nspname")+`
ORDER BY n.nspname, c.relname`)
	if err != nil {
		return nil, err
	}
	for rel.Next() {
		var schema, name, kind, tbl string
		var tuples int64
		var uniq bool
		if err := rel.Scan(&schema, &name, &kind, &tuples, &tbl, &uniq); err != nil {
			return nil, err
		}
		b := by[schema]
		if b == nil {
			continue
		}
		n := SchemaNode{Name: name, Schema: schema, Table: tbl, RowCount: tuples}
		switch kind {
		case "r", "p", "f":
			n.Kind = NodeTable
			b.tables = append(b.tables, n)
		case "v":
			n.Kind, n.RowCount = NodeView, -1
			b.views = append(b.views, n)
		case "m":
			n.Kind = NodeMatView
			b.matviews = append(b.matviews, n)
		case "i":
			n.Kind, n.RowCount = NodeIndex, -1
			if uniq {
				n.Detail = "UNIQUE"
			}
			b.indexes = append(b.indexes, n)
		}
	}
	rel.Close()
	if err := rel.Err(); err != nil {
		return nil, err
	}
	fn, err := d.pool.Query(ctx, `
SELECT n.nspname, p.proname, pg_get_function_identity_arguments(p.oid)
FROM pg_proc p JOIN pg_namespace n ON n.oid = p.pronamespace
WHERE p.prokind = 'f' AND `+strings.ReplaceAll(pgSchemaFilter, "nspname", "n.nspname")+`
ORDER BY n.nspname, p.proname`)
	if err == nil {
		for fn.Next() {
			var schema, name, args string
			if fn.Scan(&schema, &name, &args) == nil && by[schema] != nil {
				by[schema].funcs = append(by[schema].funcs, SchemaNode{
					Name: name, Kind: NodeFunction, Schema: schema, RowCount: -1, Detail: "(" + args + ")",
				})
			}
		}
		fn.Close()
	}
	tg, err := d.pool.Query(ctx, `
SELECT n.nspname, g.tgname, c.relname
FROM pg_trigger g JOIN pg_class c ON c.oid = g.tgrelid
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE NOT g.tgisinternal AND `+strings.ReplaceAll(pgSchemaFilter, "nspname", "n.nspname")+`
ORDER BY n.nspname, g.tgname`)
	if err == nil {
		for tg.Next() {
			var schema, name, tbl string
			if tg.Scan(&schema, &name, &tbl) == nil && by[schema] != nil {
				by[schema].triggers = append(by[schema].triggers, SchemaNode{
					Name: name, Kind: NodeTrigger, Schema: schema, Table: tbl, RowCount: -1,
				})
			}
		}
		tg.Close()
	}
	out := make([]SchemaNode, 0, len(order))
	for _, s := range order {
		b := by[s]
		node := SchemaNode{Name: s, Kind: NodeSchema, Schema: s, RowCount: -1}
		node.Children = appendGroups(nil,
			"Tables", b.tables, "Views", b.views, "Materialized Views", b.matviews,
			"Indexes", b.indexes, "Triggers", b.triggers, "Functions", b.funcs)
		out = append(out, node)
	}
	return out, nil
}

func pgSchema(s string) string {
	if s == "" {
		return "public"
	}
	return s
}

func (d *pgDriver) TableColumns(ctx context.Context, schema, table string) ([]ColumnMeta, error) {
	rows, err := d.pool.Query(ctx, `
SELECT a.attname, format_type(a.atttypid, a.atttypmod), NOT a.attnotnull,
       COALESCE(pg_get_expr(d.adbin, d.adrelid), ''),
       EXISTS (SELECT 1 FROM pg_constraint k WHERE k.conrelid = c.oid AND k.contype = 'p' AND a.attnum = ANY (k.conkey)),
       EXISTS (SELECT 1 FROM pg_constraint k WHERE k.conrelid = c.oid AND k.contype = 'f' AND a.attnum = ANY (k.conkey))
FROM pg_attribute a
JOIN pg_class c ON c.oid = a.attrelid
JOIN pg_namespace n ON n.oid = c.relnamespace
LEFT JOIN pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum
WHERE n.nspname = $1 AND c.relname = $2 AND a.attnum > 0 AND NOT a.attisdropped
ORDER BY a.attnum`, pgSchema(schema), table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cols []ColumnMeta
	for rows.Next() {
		var c ColumnMeta
		if err := rows.Scan(&c.Name, &c.Type, &c.Nullable, &c.DefaultVal, &c.PrimaryKey, &c.ForeignKey); err != nil {
			return nil, err
		}
		cols = append(cols, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(cols) == 0 {
		return nil, fmt.Errorf("driver: table %s.%s not found", pgSchema(schema), table)
	}
	return cols, nil
}

func (d *pgDriver) TableIndexes(ctx context.Context, schema, table string) ([]IndexMeta, error) {
	rows, err := d.pool.Query(ctx, `
SELECT ic.relname, i.indisunique, i.indisprimary, am.amname, pg_get_indexdef(i.indexrelid),
       ARRAY(SELECT a.attname::text FROM unnest(i.indkey::int2[]) WITH ORDINALITY k(attnum, ord)
             JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = k.attnum ORDER BY k.ord)
FROM pg_index i
JOIN pg_class ic ON ic.oid = i.indexrelid
JOIN pg_class t ON t.oid = i.indrelid
JOIN pg_namespace n ON n.oid = t.relnamespace
JOIN pg_am am ON am.oid = ic.relam
WHERE n.nspname = $1 AND t.relname = $2
ORDER BY i.indisprimary DESC, ic.relname`, pgSchema(schema), table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []IndexMeta
	for rows.Next() {
		var ix IndexMeta
		if err := rows.Scan(&ix.Name, &ix.Unique, &ix.Primary, &ix.Method, &ix.Definition, &ix.Columns); err != nil {
			return nil, err
		}
		out = append(out, ix)
	}
	return out, rows.Err()
}

func (d *pgDriver) TableForeignKeys(ctx context.Context, schema, table string) ([]ForeignKeyMeta, error) {
	rows, err := d.pool.Query(ctx, `
SELECT con.conname,
       ARRAY(SELECT a.attname::text FROM unnest(con.conkey) WITH ORDINALITY k(attnum, ord)
             JOIN pg_attribute a ON a.attrelid = con.conrelid AND a.attnum = k.attnum ORDER BY k.ord),
       rn.nspname, rc.relname,
       ARRAY(SELECT a.attname::text FROM unnest(con.confkey) WITH ORDINALITY k(attnum, ord)
             JOIN pg_attribute a ON a.attrelid = con.confrelid AND a.attnum = k.attnum ORDER BY k.ord),
       con.confupdtype::text, con.confdeltype::text
FROM pg_constraint con
JOIN pg_class c ON c.oid = con.conrelid
JOIN pg_namespace n ON n.oid = c.relnamespace
JOIN pg_class rc ON rc.oid = con.confrelid
JOIN pg_namespace rn ON rn.oid = rc.relnamespace
WHERE con.contype = 'f' AND n.nspname = $1 AND c.relname = $2
ORDER BY con.conname`, pgSchema(schema), table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ForeignKeyMeta
	for rows.Next() {
		var fk ForeignKeyMeta
		var upd, del string
		if err := rows.Scan(&fk.Name, &fk.Columns, &fk.RefSchema, &fk.RefTable, &fk.RefColumns, &upd, &del); err != nil {
			return nil, err
		}
		fk.OnUpdate, fk.OnDelete = pgFKAction(upd), pgFKAction(del)
		out = append(out, fk)
	}
	return out, rows.Err()
}

func pgFKAction(c string) string {
	switch c {
	case "r":
		return "RESTRICT"
	case "c":
		return "CASCADE"
	case "n":
		return "SET NULL"
	case "d":
		return "SET DEFAULT"
	}
	return "NO ACTION"
}

func (d *pgDriver) TableDDL(ctx context.Context, schema, table string) (string, error) {
	schema = pgSchema(schema)
	var kind string
	var oid uint32
	err := d.pool.QueryRow(ctx, `
SELECT c.relkind::text, c.oid FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = $1 AND c.relname = $2`, schema, table).Scan(&kind, &oid)
	if err != nil {
		return "", fmt.Errorf("driver: %s.%s not found: %w", schema, table, err)
	}
	if kind == "v" || kind == "m" {
		var def string
		if err := d.pool.QueryRow(ctx, "SELECT pg_get_viewdef($1::oid, true)", oid).Scan(&def); err != nil {
			return "", err
		}
		word := "VIEW"
		if kind == "m" {
			word = "MATERIALIZED VIEW"
		}
		return fmt.Sprintf("CREATE %s %s AS\n%s\n", word, QualifiedTable(schema, table), strings.TrimSpace(def)), nil
	}
	cols, err := d.TableColumns(ctx, schema, table)
	if err != nil {
		return "", err
	}
	var cons []string
	crow, err := d.pool.Query(ctx, `SELECT conname, pg_get_constraintdef(oid, true) FROM pg_constraint
WHERE conrelid = $1::oid ORDER BY contype = 'p' DESC, contype, conname`, oid)
	if err != nil {
		return "", err
	}
	for crow.Next() {
		var name, def string
		if err := crow.Scan(&name, &def); err != nil {
			return "", err
		}
		cons = append(cons, fmt.Sprintf("CONSTRAINT %s %s", QuoteIdent(name), def))
	}
	crow.Close()
	var idx []string
	irow, err := d.pool.Query(ctx, `SELECT pg_get_indexdef(i.indexrelid) FROM pg_index i
WHERE i.indrelid = $1::oid AND NOT EXISTS (SELECT 1 FROM pg_constraint k WHERE k.conindid = i.indexrelid)
ORDER BY 1`, oid)
	if err == nil {
		for irow.Next() {
			var def string
			if irow.Scan(&def) == nil {
				idx = append(idx, def)
			}
		}
		irow.Close()
	}
	return BuildCreateTable(schema, table, cols, cons, idx), nil
}

// BuildCreateTable synthesizes a CREATE TABLE statement. constraints are full
// clauses ("CONSTRAINT x PRIMARY KEY (id)"); indexDefs are CREATE INDEX text.
func BuildCreateTable(schema, table string, cols []ColumnMeta, constraints, indexDefs []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "CREATE TABLE %s (\n", QualifiedTable(schema, table))
	n := len(cols) + len(constraints)
	i := 0
	line := func(s string) {
		i++
		b.WriteString("    " + s)
		if i < n {
			b.WriteByte(',')
		}
		b.WriteByte('\n')
	}
	for _, c := range cols {
		s := QuoteIdent(c.Name) + " " + c.Type
		if c.DefaultVal != "" {
			s += " DEFAULT " + c.DefaultVal
		}
		if !c.Nullable && !c.PrimaryKey {
			s += " NOT NULL"
		}
		line(s)
	}
	for _, c := range constraints {
		line(c)
	}
	b.WriteString(");\n")
	for _, d := range indexDefs {
		b.WriteString("\n" + d + ";\n")
	}
	return b.String()
}

func (d *pgDriver) CollectMetrics(ctx context.Context) (MetricSnapshot, error) {
	snap := MetricSnapshot{CapturedAt: time.Now(), CacheHitRatio: -1, Extra: map[string]float64{}}
	var commit, rollback, hit, read int64
	err := d.pool.QueryRow(ctx, `
SELECT COALESCE(xact_commit, 0), COALESCE(xact_rollback, 0), COALESCE(blks_hit, 0), COALESCE(blks_read, 0),
       pg_database_size(current_database())
FROM pg_stat_database WHERE datname = current_database()`).
		Scan(&commit, &rollback, &hit, &read, &snap.DBSizeBytes)
	if err != nil {
		return snap, err
	}
	snap.CommitTotal, snap.RollbackTotal, snap.BlocksHit, snap.BlocksRead = commit, rollback, hit, read
	if hit+read > 0 {
		snap.CacheHitRatio = float64(hit) / float64(hit+read)
	}
	d.mu.Lock()
	total := commit + rollback
	if !d.prevAt.IsZero() {
		if dt := snap.CapturedAt.Sub(d.prevAt).Seconds(); dt > 0 {
			snap.TPS = float64(total-d.prevXact) / dt
		}
	}
	d.prevAt, d.prevXact = snap.CapturedAt, total
	d.mu.Unlock()

	_ = d.pool.QueryRow(ctx, `
SELECT count(*) FILTER (WHERE state = 'active'),
       count(*) FILTER (WHERE state = 'idle')
FROM pg_stat_activity WHERE datname = current_database() AND backend_type = 'client backend'`).
		Scan(&snap.ActiveConns, &snap.IdleConns)
	_ = d.pool.QueryRow(ctx, "SELECT count(*) FROM pg_locks WHERE NOT granted").Scan(&snap.WaitingLocks)

	if rows, err := d.pool.Query(ctx, `
SELECT pid, COALESCE(usename, ''), COALESCE(datname, ''), COALESCE(state, ''), COALESCE(query, ''),
       COALESCE(query_start, now()), COALESCE(wait_event_type, ''), COALESCE(wait_event, '')
FROM pg_stat_activity
WHERE datname = current_database() AND pid <> pg_backend_pid() AND state IS DISTINCT FROM 'idle'
  AND backend_type = 'client backend'
ORDER BY query_start NULLS LAST LIMIT 50`); err == nil {
		for rows.Next() {
			var q ActiveQuery
			var wt, we string
			if rows.Scan(&q.PID, &q.User, &q.Database, &q.State, &q.Query, &q.StartedAt, &wt, &we) == nil {
				q.Duration = snap.CapturedAt.Sub(q.StartedAt)
				if wt != "" {
					q.WaitEvent = wt + ":" + we
				}
				q.Blocked = wt == "Lock"
				snap.ActiveQueries = append(snap.ActiveQueries, q)
			}
		}
		rows.Close()
	}
	if rows, err := d.pool.Query(ctx, `
SELECT n.nspname || '.' || c.relname, pg_total_relation_size(c.oid)
FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE c.relkind IN ('r','m') AND n.nspname NOT IN ('pg_catalog','information_schema') AND n.nspname NOT LIKE 'pg\_toast%'
ORDER BY 2 DESC LIMIT 10`); err == nil {
		for rows.Next() {
			var it StorageItem
			if rows.Scan(&it.Name, &it.Bytes) == nil {
				snap.Storage = append(snap.Storage, it)
			}
		}
		rows.Close()
	}
	return snap, nil
}

// Explain runs EXPLAIN (FORMAT JSON), or EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON)
// inside a transaction that is always rolled back so analyzing a write
// statement leaves no trace.
func (d *pgDriver) Explain(ctx context.Context, sqlText string, analyze bool) (ExplainPlan, error) {
	if d.cfg.ReadOnly {
		if err := CheckReadOnly(sqlText); err != nil {
			return ExplainPlan{}, err
		}
	}
	prefix := "EXPLAIN (FORMAT JSON) "
	if analyze {
		prefix = "EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) "
	}
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return ExplainPlan{}, err
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	var raw []byte
	if err := tx.QueryRow(ctx, prefix+sqlText).Scan(&raw); err != nil {
		return ExplainPlan{}, err
	}
	return ParsePGExplainJSON(raw)
}

type pgPlanJSON struct {
	Plan          pgNodeJSON `json:"Plan"`
	PlanningTime  float64    `json:"Planning Time"`
	ExecutionTime *float64   `json:"Execution Time"`
}

type pgNodeJSON struct {
	NodeType    string       `json:"Node Type"`
	Relation    string       `json:"Relation Name"`
	Alias       string       `json:"Alias"`
	IndexName   string       `json:"Index Name"`
	JoinType    string       `json:"Join Type"`
	Filter      string       `json:"Filter"`
	IndexCond   string       `json:"Index Cond"`
	HashCond    string       `json:"Hash Cond"`
	SortKey     []string     `json:"Sort Key"`
	StartupCost float64      `json:"Startup Cost"`
	TotalCost   float64      `json:"Total Cost"`
	PlanRows    float64      `json:"Plan Rows"`
	ActualRows  *float64     `json:"Actual Rows"`
	ActualTotal *float64     `json:"Actual Total Time"`
	ActualLoops *float64     `json:"Actual Loops"`
	SharedHit   int64        `json:"Shared Hit Blocks"`
	SharedRead  int64        `json:"Shared Read Blocks"`
	Plans       []pgNodeJSON `json:"Plans"`
}

// ParsePGExplainJSON decodes the output of EXPLAIN (FORMAT JSON) into the
// shared ExplainPlan tree. Node IDs are assigned in pre-order starting at 1.
func ParsePGExplainJSON(raw []byte) (ExplainPlan, error) {
	var top []pgPlanJSON
	if err := json.Unmarshal(raw, &top); err != nil {
		return ExplainPlan{}, fmt.Errorf("driver: decode explain json: %w", err)
	}
	if len(top) == 0 {
		return ExplainPlan{}, fmt.Errorf("driver: empty explain output")
	}
	plan := ExplainPlan{Engine: EnginePostgres, Raw: string(raw), PlanningMs: top[0].PlanningTime}
	if top[0].ExecutionTime != nil {
		plan.Analyzed = true
		plan.ExecutionMs = *top[0].ExecutionTime
	}
	id := 0
	var conv func(n pgNodeJSON, parent int) *PlanNode
	conv = func(n pgNodeJSON, parent int) *PlanNode {
		id++
		pn := &PlanNode{
			ID: id, ParentID: parent, NodeType: n.NodeType, Relation: n.Relation, Index: n.IndexName,
			StartupCost: n.StartupCost, TotalCost: n.TotalCost, PlanRows: n.PlanRows, ActualMs: -1,
			SharedHit: n.SharedHit, SharedRead: n.SharedRead,
		}
		if n.ActualRows != nil {
			pn.ActualRows = *n.ActualRows
		}
		if n.ActualTotal != nil {
			pn.ActualMs = *n.ActualTotal
		}
		if n.ActualLoops != nil {
			pn.Loops = *n.ActualLoops
		}
		var det []string
		if n.JoinType != "" {
			det = append(det, n.JoinType+" join")
		}
		for _, c := range []string{n.IndexCond, n.HashCond, n.Filter} {
			if c != "" {
				det = append(det, c)
			}
		}
		if len(n.SortKey) > 0 {
			det = append(det, "by "+strings.Join(n.SortKey, ", "))
		}
		pn.Detail = strings.Join(det, "; ")
		switch {
		case n.NodeType == "Seq Scan":
			pn.Scan = ScanFull
			plan.FullScans++
		case strings.Contains(n.NodeType, "Index"):
			pn.Scan = ScanSearch
			plan.IndexSearches++
		}
		for _, c := range n.Plans {
			pn.Children = append(pn.Children, conv(c, pn.ID))
		}
		return pn
	}
	plan.Root = conv(top[0].Plan, -1)
	return plan, nil
}

var pgVersionRe = regexp.MustCompile(`PostgreSQL\s+([0-9.]+)`)

// ShortPGVersion extracts "PostgreSQL 16.2" from a version() banner.
func ShortPGVersion(banner string) string {
	if m := pgVersionRe.FindStringSubmatch(banner); m != nil {
		return "PostgreSQL " + m[1]
	}
	return banner
}

// ConfigFromDSN parses a postgres:// or postgresql:// URL into a
// ConnectionConfig (name "host/database", EnvTag by host).
func ConfigFromDSN(dsn string) (ConnectionConfig, error) {
	u, err := url.Parse(dsn)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") {
		return ConnectionConfig{}, fmt.Errorf("driver: not a postgres:// URL")
	}
	cfg := ConnectionConfig{
		Engine: EnginePostgres, Host: u.Hostname(), Database: strings.TrimPrefix(u.Path, "/"),
		SSLMode: u.Query().Get("sslmode"), EnvTag: EnvLocal,
	}
	if p := u.Port(); p != "" {
		cfg.Port, _ = strconv.Atoi(p)
	}
	if u.User != nil {
		cfg.User = u.User.Username()
		cfg.Password, _ = u.User.Password()
	}
	if h := cfg.Host; h != "" && h != "localhost" && h != "127.0.0.1" && h != "::1" {
		cfg.EnvTag = EnvStaging
	}
	cfg.Name = cfg.Host + "/" + cfg.Database
	return cfg, nil
}

// ConfigFromArg interprets the -db flag: a postgres:// URL becomes a
// PostgreSQL profile, anything else a SQLite file path.
func ConfigFromArg(arg string) (ConnectionConfig, error) {
	if strings.HasPrefix(arg, "postgres://") || strings.HasPrefix(arg, "postgresql://") {
		return ConfigFromDSN(arg)
	}
	name := arg
	if i := strings.LastIndexAny(arg, `/\`); i >= 0 {
		name = arg[i+1:]
	}
	return ConnectionConfig{Engine: EngineSQLite, FilePath: arg, Name: name, EnvTag: EnvLocal}, nil
}

// ExecTx runs every statement with its bound arguments inside one transaction.
// On failure it rolls everything back and returns the index of the failing
// statement; on success it returns -1.
func (d *pgDriver) ExecTx(ctx context.Context, sqls []string, args [][]any) (int, error) {
	if d.cfg.ReadOnly {
		return -1, ErrReadOnly
	}
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return -1, err
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	for i, s := range sqls {
		var a []any
		if i < len(args) {
			a = args[i]
		}
		if _, err := tx.Exec(ctx, s, a...); err != nil {
			return i, err
		}
	}
	return -1, tx.Commit(ctx)
}
