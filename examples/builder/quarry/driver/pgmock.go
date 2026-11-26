package driver

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// mockPG is a deterministic in-memory stand-in for PostgreSQL so headless
// tests, CI, and offline demos exercise the PostgreSQL UI paths without a
// server. It has three schemas (public, analytics, billing), canned query
// results, EXPLAIN JSON fixtures, and synthetic telemetry that advances on
// every CollectMetrics call.
type mockPG struct {
	cfg ConnectionConfig
	trk *tracker

	tables []mockTable

	mu     sync.Mutex
	closed bool
	step   int
}

type mockTable struct {
	schema, name string
	kind         SchemaNodeKind
	rows         int64
	cols         []ColumnMeta
	fks          []ForeignKeyMeta
	idx          []IndexMeta
}

// mockEpoch anchors every generated timestamp.
func mockEpoch() time.Time { return time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC) }

func col(name, typ string, pk bool) ColumnMeta {
	return ColumnMeta{Name: name, Type: typ, Nullable: !pk, PrimaryKey: pk}
}

func notNull(c ColumnMeta) ColumnMeta { c.Nullable = false; return c }

func withDefault(c ColumnMeta, d string) ColumnMeta { c.DefaultVal = d; return c }

func fkCol(c ColumnMeta) ColumnMeta { c.ForeignKey = true; return c }

// Synthetic telemetry and value-generation parameters for the mock driver.
const (
	mockPriceCycle      = 10000
	mockCents           = 100
	mockPriceBase       = 9.99
	mockNullEvery       = 7
	mockIntStride       = 3
	mockIntBase         = 7
	mockInactiveEvery   = 5
	mockHashMul         = 2654435761
	mockHashMod         = 1 << 32
	mockHitBase         = 9_000_000
	mockHitStep         = 48_000
	mockReadBase        = 120_000
	mockReadStep        = 900
	mockPollGap         = 2 * time.Second
	mockConnBase        = 3
	mockConnSwing       = 2
	mockConnPeriod      = 3
	mockIdleBase        = 11
	mockIdlePeriod      = 5
	mockLockCycle       = 7
	mockLockMin         = 5
	mockTPSBase         = 420
	mockTPSSwing        = 90
	mockTPSPeriod       = 4
	mockTPSJitterMod    = 5
	mockTPSJitter       = 3
	mockSizeBase        = 2_400_000_000
	mockSizeStep        = 1_048_576
	mockCommitPerStep   = 840
	mockRollbackPerStep = 3
	mockBytesEvents     = 1_610_612_736
	mockBytesOrders     = 402_653_184
	mockBytesInvoices   = 104_857_600
	mockBytesUsers      = 52_428_800
	mockBytesPayments   = 41_943_040
	mockCheckpointEvery = 10
	mockPIDSlow         = 4101
	mockPIDFast         = 4102
	mockPIDBlocked      = 4103
	mockSlowDur         = 1840 * time.Millisecond
	mockFastDur         = 12 * time.Millisecond
	mockBlockedDur      = 4 * time.Second
)

// Row estimates the mock catalog reports for each relation.
const (
	mockRowsUsers    = 12840
	mockRowsProducts = 420
	mockRowsOrders   = 58210
	mockRowsEvents   = 1250000
	mockRowsRollups  = 730
	mockRowsInvoices = 9120
	mockRowsPayments = 8870
)

func mockTables() []mockTable {
	return []mockTable{
		{
			schema: "public", name: "users", kind: NodeTable, rows: mockRowsUsers,
			cols: []ColumnMeta{
				withDefault(col("id", "bigint", true), "nextval('users_id_seq'::regclass)"),
				notNull(col("email", "text", false)), col("full_name", "text", false),
				col("active", "boolean", false), withDefault(notNull(col("created_at", "timestamp with time zone", false)), "now()"),
			},
			idx: []IndexMeta{
				{Name: "users_pkey", Columns: []string{"id"}, Unique: true, Primary: true, Method: "btree", Definition: "CREATE UNIQUE INDEX users_pkey ON public.users USING btree (id)"},
				{Name: "users_email_key", Columns: []string{"email"}, Unique: true, Method: "btree", Definition: "CREATE UNIQUE INDEX users_email_key ON public.users USING btree (email)"},
			},
		},
		{
			schema: "public", name: "products", kind: NodeTable, rows: mockRowsProducts,
			cols: []ColumnMeta{
				withDefault(col("id", "bigint", true), "nextval('products_id_seq'::regclass)"),
				notNull(col("sku", "text", false)), col("title", "text", false), col("price", "numeric(10,2)", false),
				col("attrs", "jsonb", false),
			},
			idx: []IndexMeta{
				{Name: "products_pkey", Columns: []string{"id"}, Unique: true, Primary: true, Method: "btree", Definition: "CREATE UNIQUE INDEX products_pkey ON public.products USING btree (id)"},
				{Name: "products_attrs_gin", Columns: []string{"attrs"}, Method: "gin", Definition: "CREATE INDEX products_attrs_gin ON public.products USING gin (attrs)"},
			},
		},
		{
			schema: "public", name: "orders", kind: NodeTable, rows: mockRowsOrders,
			cols: []ColumnMeta{
				withDefault(col("id", "bigint", true), "nextval('orders_id_seq'::regclass)"),
				fkCol(notNull(col("user_id", "bigint", false))), notNull(col("status", "text", false)),
				col("total", "numeric(12,2)", false), withDefault(notNull(col("created_at", "timestamp with time zone", false)), "now()"),
			},
			fks: []ForeignKeyMeta{{Name: "orders_user_id_fkey", Columns: []string{"user_id"}, RefSchema: "public", RefTable: "users", RefColumns: []string{"id"}, OnUpdate: "NO ACTION", OnDelete: "CASCADE"}},
			idx: []IndexMeta{
				{Name: "orders_pkey", Columns: []string{"id"}, Unique: true, Primary: true, Method: "btree", Definition: "CREATE UNIQUE INDEX orders_pkey ON public.orders USING btree (id)"},
				{Name: "orders_user_idx", Columns: []string{"user_id", "created_at"}, Method: "btree", Definition: "CREATE INDEX orders_user_idx ON public.orders USING btree (user_id, created_at)"},
			},
		},
		{
			schema: "analytics", name: "events", kind: NodeTable, rows: mockRowsEvents,
			cols: []ColumnMeta{
				withDefault(col("id", "bigint", true), "nextval('analytics.events_id_seq'::regclass)"),
				fkCol(col("user_id", "bigint", false)), notNull(col("kind", "text", false)),
				col("payload", "jsonb", false), withDefault(notNull(col("occurred_at", "timestamp with time zone", false)), "now()"),
			},
			fks: []ForeignKeyMeta{{Name: "events_user_id_fkey", Columns: []string{"user_id"}, RefSchema: "public", RefTable: "users", RefColumns: []string{"id"}, OnUpdate: "NO ACTION", OnDelete: "SET NULL"}},
			idx: []IndexMeta{
				{Name: "events_pkey", Columns: []string{"id"}, Unique: true, Primary: true, Method: "btree", Definition: "CREATE UNIQUE INDEX events_pkey ON analytics.events USING btree (id)"},
				{Name: "events_kind_time_idx", Columns: []string{"kind", "occurred_at"}, Method: "btree", Definition: "CREATE INDEX events_kind_time_idx ON analytics.events USING btree (kind, occurred_at)"},
			},
		},
		{
			schema: "analytics", name: "daily_rollups", kind: NodeMatView, rows: mockRowsRollups,
			cols: []ColumnMeta{col("day", "date", false), col("kind", "text", false), col("events", "bigint", false)},
		},
		{
			schema: "analytics", name: "v_funnel", kind: NodeView, rows: -1,
			cols: []ColumnMeta{col("step", "text", false), col("users", "bigint", false)},
		},
		{
			schema: "billing", name: "invoices", kind: NodeTable, rows: mockRowsInvoices,
			cols: []ColumnMeta{
				withDefault(col("id", "bigint", true), "nextval('billing.invoices_id_seq'::regclass)"),
				fkCol(notNull(col("user_id", "bigint", false))), notNull(col("amount", "numeric(12,2)", false)),
				notNull(col("status", "text", false)), withDefault(notNull(col("issued_at", "timestamp with time zone", false)), "now()"),
			},
			fks: []ForeignKeyMeta{{Name: "invoices_user_id_fkey", Columns: []string{"user_id"}, RefSchema: "public", RefTable: "users", RefColumns: []string{"id"}, OnUpdate: "NO ACTION", OnDelete: "RESTRICT"}},
			idx: []IndexMeta{
				{Name: "invoices_pkey", Columns: []string{"id"}, Unique: true, Primary: true, Method: "btree", Definition: "CREATE UNIQUE INDEX invoices_pkey ON billing.invoices USING btree (id)"},
			},
		},
		{
			schema: "billing", name: "payments", kind: NodeTable, rows: mockRowsPayments,
			cols: []ColumnMeta{
				withDefault(col("id", "bigint", true), "nextval('billing.payments_id_seq'::regclass)"),
				fkCol(notNull(col("invoice_id", "bigint", false))), notNull(col("amount", "numeric(12,2)", false)),
				col("paid_at", "timestamp with time zone", false), col("receipt", "uuid", false),
			},
			fks: []ForeignKeyMeta{{Name: "payments_invoice_id_fkey", Columns: []string{"invoice_id"}, RefSchema: "billing", RefTable: "invoices", RefColumns: []string{"id"}, OnUpdate: "NO ACTION", OnDelete: "CASCADE"}},
			idx: []IndexMeta{
				{Name: "payments_pkey", Columns: []string{"id"}, Unique: true, Primary: true, Method: "btree", Definition: "CREATE UNIQUE INDEX payments_pkey ON billing.payments USING btree (id)"},
			},
		},
	}
}

func mockFuncs(schema string) []string {
	switch schema {
	case "public":
		return []string{"touch_updated_at"}
	case "analytics":
		return []string{"rollup_day"}
	case "billing":
		return []string{"invoice_total"}
	}
	return nil
}

// NewPGMock returns the offline PostgreSQL driver for cfg. Engine is forced to
// EnginePostgres.
func NewPGMock(cfg ConnectionConfig) Driver {
	cfg.Engine, cfg.Mock = EnginePostgres, true
	if cfg.Database == "" {
		cfg.Database = "quarry_mock"
	}
	if cfg.Name == "" {
		cfg.Name = "Mock PostgreSQL"
	}
	return &mockPG{cfg: cfg, tables: mockTables(), trk: newTracker("quarry", cfg.Database)}
}

func (m *mockPG) Engine() EngineKind       { return EnginePostgres }
func (m *mockPG) Config() ConnectionConfig { return m.cfg }

func (m *mockPG) live() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return fmt.Errorf("driver: mock connection is closed")
	}
	return nil
}

func (m *mockPG) Close() error {
	m.mu.Lock()
	m.closed = true
	m.mu.Unlock()
	return nil
}

func (m *mockPG) Ping(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return m.live()
}

func (m *mockPG) Version(context.Context) (string, error) { return "PostgreSQL 16.3 (mock)", nil }

func (m *mockPG) find(schema, table string) (*mockTable, error) {
	if schema == "" {
		schema = "public"
	}
	for i := range m.tables {
		if m.tables[i].schema == schema && m.tables[i].name == table {
			return &m.tables[i], nil
		}
	}
	return nil, fmt.Errorf("driver: relation %s.%s does not exist", schema, table)
}

func (m *mockPG) ListSchemas(ctx context.Context) ([]SchemaNode, error) {
	if err := m.Ping(ctx); err != nil {
		return nil, err
	}
	var out []SchemaNode
	for _, s := range []string{"public", "analytics", "billing"} {
		var tables, views, mviews, idxs, funcs []SchemaNode
		for _, t := range m.tables {
			if t.schema != s {
				continue
			}
			n := SchemaNode{Name: t.name, Kind: t.kind, Schema: s, RowCount: t.rows}
			switch t.kind {
			case NodeTable:
				tables = append(tables, n)
			case NodeView:
				views = append(views, n)
			case NodeMatView:
				mviews = append(mviews, n)
			}
			for _, ix := range t.idx {
				d := ""
				if ix.Unique {
					d = "UNIQUE"
				}
				idxs = append(idxs, SchemaNode{Name: ix.Name, Kind: NodeIndex, Schema: s, Table: t.name, RowCount: -1, Detail: d})
			}
		}
		for _, f := range mockFuncs(s) {
			funcs = append(funcs, SchemaNode{Name: f, Kind: NodeFunction, Schema: s, RowCount: -1, Detail: "()"})
		}
		node := SchemaNode{Name: s, Kind: NodeSchema, Schema: s, RowCount: -1}
		node.Children = appendGroups(nil, "Tables", tables, "Views", views, "Materialized Views", mviews, "Indexes", idxs, "Functions", funcs)
		out = append(out, node)
	}
	return out, nil
}

func (m *mockPG) TableColumns(ctx context.Context, schema, table string) ([]ColumnMeta, error) {
	if err := m.Ping(ctx); err != nil {
		return nil, err
	}
	t, err := m.find(schema, table)
	if err != nil {
		return nil, err
	}
	return append([]ColumnMeta(nil), t.cols...), nil
}

func (m *mockPG) TableIndexes(ctx context.Context, schema, table string) ([]IndexMeta, error) {
	if err := m.Ping(ctx); err != nil {
		return nil, err
	}
	t, err := m.find(schema, table)
	if err != nil {
		return nil, err
	}
	return append([]IndexMeta(nil), t.idx...), nil
}

func (m *mockPG) TableForeignKeys(ctx context.Context, schema, table string) ([]ForeignKeyMeta, error) {
	if err := m.Ping(ctx); err != nil {
		return nil, err
	}
	t, err := m.find(schema, table)
	if err != nil {
		return nil, err
	}
	return append([]ForeignKeyMeta(nil), t.fks...), nil
}

func (m *mockPG) TableDDL(ctx context.Context, schema, table string) (string, error) {
	if err := m.Ping(ctx); err != nil {
		return "", err
	}
	t, err := m.find(schema, table)
	if err != nil {
		return "", err
	}
	switch t.kind {
	case NodeView:
		return fmt.Sprintf("CREATE VIEW %s AS\n SELECT 'visit'::text AS step, count(DISTINCT user_id) AS users\n   FROM analytics.events;\n", QualifiedTable(t.schema, t.name)), nil
	case NodeMatView:
		return fmt.Sprintf("CREATE MATERIALIZED VIEW %s AS\n SELECT occurred_at::date AS day, kind, count(*) AS events\n   FROM analytics.events\n  GROUP BY 1, 2;\n", QualifiedTable(t.schema, t.name)), nil
	}
	var cons, idx []string
	for _, ix := range t.idx {
		switch {
		case ix.Primary:
			cons = append(cons, fmt.Sprintf("CONSTRAINT %s PRIMARY KEY (%s)", QuoteIdent(ix.Name), strings.Join(ix.Columns, ", ")))
		case ix.Unique && strings.HasSuffix(ix.Name, "_key"):
			cons = append(cons, fmt.Sprintf("CONSTRAINT %s UNIQUE (%s)", QuoteIdent(ix.Name), strings.Join(ix.Columns, ", ")))
		default:
			idx = append(idx, ix.Definition)
		}
	}
	for _, fk := range t.fks {
		cons = append(cons, fmt.Sprintf("CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s(%s) ON DELETE %s",
			QuoteIdent(fk.Name), strings.Join(fk.Columns, ", "), QualifiedTable(fk.RefSchema, fk.RefTable),
			strings.Join(fk.RefColumns, ", "), fk.OnDelete))
	}
	return BuildCreateTable(t.schema, t.name, t.cols, cons, idx), nil
}

var (
	mockSelectRe = regexp.MustCompile(`(?is)^\s*select\s+(.+?)\s+from\s+("?[\w]+"?(?:\."?[\w]+"?)?)\s*(.*)$`)
	mockLimitRe  = regexp.MustCompile(`(?i)\blimit\s+(\d+)`)
	mockSleepRe  = regexp.MustCompile(`(?i)pg_sleep\(\s*([0-9.]+)\s*\)`)
)

func (m *mockPG) Query(ctx context.Context, sqlText string, maxRows int) (QueryResult, error) {
	if err := m.Ping(ctx); err != nil {
		return QueryResult{}, err
	}
	if m.cfg.ReadOnly {
		if err := CheckReadOnly(sqlText); err != nil {
			return QueryResult{}, err
		}
	}
	if maxRows <= 0 {
		maxRows = DefaultMaxRows
	}
	ctx, done := m.trk.begin(ctx, sqlText)
	defer done()
	start := time.Now()
	if sm := mockSleepRe.FindStringSubmatch(sqlText); sm != nil {
		secs, _ := strconv.ParseFloat(sm[1], 64)
		select {
		case <-time.After(time.Duration(secs * float64(time.Second))):
		case <-ctx.Done():
			return QueryResult{}, ctx.Err()
		}
		return QueryResult{Columns: []ColumnMeta{{Name: "pg_sleep", Type: "void"}}, Rows: [][]CellValue{{Text("")}}, Duration: time.Since(start)}, nil
	}
	sm := mockSelectRe.FindStringSubmatch(sqlText)
	if sm == nil {
		return QueryResult{
			Columns:  []ColumnMeta{{Name: "result", Type: "text"}},
			Rows:     [][]CellValue{{Text("ok (mock)")}},
			Duration: time.Since(start),
		}, nil
	}
	ref := strings.ReplaceAll(sm[2], `"`, "")
	schema, name := "public", ref
	if i := strings.IndexByte(ref, '.'); i >= 0 {
		schema, name = ref[:i], ref[i+1:]
	}
	t, err := m.find(schema, name)
	if err != nil {
		return QueryResult{}, err
	}
	if strings.EqualFold(strings.TrimSpace(sm[1]), "count(*)") {
		n := t.rows
		if n < 0 {
			n = 5
		}
		return QueryResult{
			Columns:  []ColumnMeta{{Name: "count", Type: "int8"}},
			Rows:     [][]CellValue{{Int(n)}},
			Duration: time.Since(start),
		}, nil
	}
	total := t.rows
	if total < 0 {
		total = 5
	}
	if lm := mockLimitRe.FindStringSubmatch(sm[3]); lm != nil {
		if n, _ := strconv.ParseInt(lm[1], 10, 64); n < total {
			total = n
		}
	}
	res := QueryResult{Columns: append([]ColumnMeta(nil), t.cols...)}
	for i := int64(0); i < total; i++ {
		if len(res.Rows) >= maxRows {
			res.Truncated = true
			break
		}
		if i%1024 == 0 && ctx.Err() != nil {
			return QueryResult{}, ctx.Err()
		}
		row := make([]CellValue, len(t.cols))
		for c, cm := range t.cols {
			row[c] = mockCell(t, cm, i)
		}
		res.Rows = append(res.Rows, row)
	}
	res.Duration = time.Since(start)
	return res, nil
}

// mockCell derives a deterministic value from the column name and row index.
func mockCell(t *mockTable, c ColumnMeta, i int64) CellValue {
	statuses := []string{"pending", "paid", "shipped", "refunded"}
	kinds := []string{"page_view", "signup", "checkout", "click"}
	switch {
	case c.Name == "id":
		return Int(i + 1)
	case strings.HasSuffix(c.Name, "_id"):
		return Int(i%500 + 1)
	case c.Name == "email":
		return Text(fmt.Sprintf("user%05d@example.com", i+1))
	case c.Name == "full_name":
		if i%17 == 0 {
			return Null()
		}
		return Text(fmt.Sprintf("Customer %d", i+1))
	case c.Name == "sku":
		return Text(fmt.Sprintf("SKU-%04d", i+1))
	case c.Name == "title":
		return Text(fmt.Sprintf("Product %d", i+1))
	case c.Name == "status":
		return Text(statuses[i%int64(len(statuses))])
	case c.Name == "kind" || c.Name == "step":
		return Text(kinds[i%int64(len(kinds))])
	case c.Name == "active":
		return Bool(i%mockInactiveEvery != 0)
	case c.Type == "uuid":
		return Text(fmt.Sprintf("%08x-0000-4000-8000-%012x", (i*mockHashMul)%mockHashMod, i))
	case c.Type == "jsonb":
		return JSON(fmt.Sprintf(`{"seq":%d,"source":"mock","tags":["a","b"]}`, i+1))
	case strings.HasPrefix(c.Type, "numeric"):
		return Float(float64(i%mockPriceCycle)/mockCents + mockPriceBase)
	case c.Type == "date":
		return Time(mockEpoch().AddDate(0, 0, int(i)))
	case strings.HasPrefix(c.Type, "timestamp"):
		if c.Name == "paid_at" && i%mockNullEvery == 0 {
			return Null()
		}
		return Time(mockEpoch().Add(time.Duration(i) * time.Hour))
	case c.Type == "bigint" || c.Type == "integer":
		return Int(i*mockIntStride + mockIntBase)
	}
	return Text(fmt.Sprintf("%s-%d", c.Name, i+1))
}

func (m *mockPG) Exec(ctx context.Context, sqlText string) (int64, error) {
	if err := m.Ping(ctx); err != nil {
		return 0, err
	}
	if m.cfg.ReadOnly {
		return 0, ErrReadOnly
	}
	_, done := m.trk.begin(ctx, sqlText)
	defer done()
	return 1, nil
}

func (m *mockPG) CancelQuery(_ context.Context, pid int) error { return m.trk.cancel(pid) }

// TerminateQuery behaves like CancelQuery in the mock: it stops the statement.
func (m *mockPG) TerminateQuery(_ context.Context, pid int) error { return m.trk.cancel(pid) }

const mockSeqScanPlan = `[{"Plan": {"Node Type": "Seq Scan", "Relation Name": "%s", "Alias": "%s", "Startup Cost": 0.00, "Total Cost": 1834.00, "Plan Rows": 58210, "Plan Width": 40%s}%s}]`

const mockIndexPlan = `[{"Plan": {"Node Type": "Limit", "Startup Cost": 0.29, "Total Cost": 8.31, "Plan Rows": 1%s, "Plans": [
 {"Node Type": "Index Scan", "Parent Relationship": "Outer", "Relation Name": "%s", "Alias": "%s", "Index Name": "%s_pkey", "Startup Cost": 0.29, "Total Cost": 8.31, "Plan Rows": 1, "Index Cond": "(id = 42)"%s}]}%s}]`

// Explain returns an index-lookup plan when the statement filters on id and a
// sequential scan otherwise, mirroring real PostgreSQL JSON.
func (m *mockPG) Explain(ctx context.Context, sqlText string, analyze bool) (ExplainPlan, error) {
	if err := m.Ping(ctx); err != nil {
		return ExplainPlan{}, err
	}
	if m.cfg.ReadOnly {
		if err := CheckReadOnly(sqlText); err != nil {
			return ExplainPlan{}, err
		}
	}
	rel := "orders"
	if sm := mockSelectRe.FindStringSubmatch(sqlText); sm != nil {
		r := strings.ReplaceAll(sm[2], `"`, "")
		if i := strings.IndexByte(r, '.'); i >= 0 {
			r = r[i+1:]
		}
		rel = r
	}
	var nodeAct, limitAct, topAct string
	if analyze {
		nodeAct = `, "Actual Startup Time": 0.020, "Actual Total Time": 0.031, "Actual Rows": 1, "Actual Loops": 1, "Shared Hit Blocks": 4, "Shared Read Blocks": 0`
		limitAct = `, "Actual Startup Time": 0.021, "Actual Total Time": 0.033, "Actual Rows": 1, "Actual Loops": 1`
		topAct = `, "Planning Time": 0.120, "Execution Time": 0.060`
	} else {
		topAct = `, "Planning Time": 0.000`
	}
	var raw string
	if regexp.MustCompile(`(?i)\bwhere\b[^;]*\bid\s*=`).MatchString(sqlText) {
		raw = fmt.Sprintf(mockIndexPlan, limitAct, rel, rel, rel, nodeAct, topAct)
	} else {
		seq := nodeAct
		if analyze {
			seq = `, "Actual Startup Time": 0.015, "Actual Total Time": 14.200, "Actual Rows": 58210, "Actual Loops": 1, "Shared Hit Blocks": 834, "Shared Read Blocks": 0`
		}
		raw = fmt.Sprintf(mockSeqScanPlan, rel, rel, seq, topAct)
	}
	return ParsePGExplainJSON([]byte(raw))
}

// CollectMetrics returns synthetic telemetry that is a pure function of how
// many samples were taken, so tests are reproducible.
func (m *mockPG) CollectMetrics(ctx context.Context) (MetricSnapshot, error) {
	if err := m.Ping(ctx); err != nil {
		return MetricSnapshot{}, err
	}
	m.mu.Lock()
	m.step++
	n := m.step
	m.mu.Unlock()
	f := float64(n)
	hit := int64(mockHitBase + n*mockHitStep)
	read := int64(mockReadBase + n*mockReadStep)
	snap := MetricSnapshot{
		CapturedAt:    mockEpoch().Add(time.Duration(n) * mockPollGap),
		ActiveConns:   mockConnBase + int(mockConnSwing*math.Sin(f/mockConnPeriod)+mockConnSwing),
		IdleConns:     mockIdleBase - int(math.Cos(f/mockIdlePeriod)*mockConnSwing),
		WaitingLocks:  n % mockLockCycle / mockLockMin,
		TPS:           mockTPSBase + mockTPSSwing*math.Sin(f/mockTPSPeriod) + float64(n%mockTPSJitterMod)*mockTPSJitter,
		CacheHitRatio: float64(hit) / float64(hit+read),
		DBSizeBytes:   mockSizeBase + int64(n)*mockSizeStep,
		CommitTotal:   int64(n) * mockCommitPerStep,
		RollbackTotal: int64(n) * mockRollbackPerStep,
		BlocksHit:     hit,
		BlocksRead:    read,
		Storage: []StorageItem{
			{Name: "analytics.events", Bytes: mockBytesEvents},
			{Name: "public.orders", Bytes: mockBytesOrders},
			{Name: "billing.invoices", Bytes: mockBytesInvoices},
			{Name: "public.users", Bytes: mockBytesUsers},
			{Name: "billing.payments", Bytes: mockBytesPayments},
		},
		Extra: map[string]float64{"checkpoints_timed": float64(n / mockCheckpointEvery)},
	}
	snap.ActiveQueries = []ActiveQuery{
		{PID: mockPIDSlow, User: "app", Database: m.cfg.Database, State: "active", Query: "SELECT * FROM analytics.events WHERE kind = 'checkout' ORDER BY occurred_at DESC LIMIT 500", StartedAt: snap.CapturedAt.Add(-mockSlowDur), Duration: mockSlowDur, WaitEvent: "IO:DataFileRead"},
		{PID: mockPIDFast, User: "app", Database: m.cfg.Database, State: "active", Query: "UPDATE billing.invoices SET status = 'paid' WHERE id = 9120", StartedAt: snap.CapturedAt.Add(-mockFastDur), Duration: mockFastDur},
	}
	if snap.WaitingLocks > 0 {
		snap.ActiveQueries = append(snap.ActiveQueries, ActiveQuery{
			PID: mockPIDBlocked, User: "app", Database: m.cfg.Database, State: "active", Blocked: true,
			Query: "UPDATE billing.invoices SET amount = amount + 1 WHERE id = 9120", StartedAt: snap.CapturedAt.Add(-mockBlockedDur),
			Duration: mockBlockedDur, WaitEvent: "Lock:transactionid",
		})
	}
	snap.ActiveQueries = append(snap.ActiveQueries, m.trk.active()...)
	return snap, nil
}

// ExecTx accepts the batch without storing anything: the mock keeps no rows.
func (m *mockPG) ExecTx(ctx context.Context, sqls []string, _ [][]any) (int, error) {
	if m.cfg.ReadOnly {
		return -1, ErrReadOnly
	}
	if err := m.Ping(ctx); err != nil {
		return -1, err
	}
	_ = sqls
	return -1, nil
}
