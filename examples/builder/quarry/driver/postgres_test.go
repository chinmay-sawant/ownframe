package driver

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

// Captured from PostgreSQL 16: EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) with a
// hash join over a sequential scan and an index scan.
const pgAnalyzeFixture = `[
 {"Plan": {"Node Type": "Hash Join", "Join Type": "Inner", "Startup Cost": 12.10, "Total Cost": 260.55,
   "Plan Rows": 480, "Actual Startup Time": 0.350, "Actual Total Time": 2.410, "Actual Rows": 455, "Actual Loops": 1,
   "Hash Cond": "(o.user_id = u.id)", "Shared Hit Blocks": 41, "Shared Read Blocks": 3,
   "Plans": [
     {"Node Type": "Seq Scan", "Parent Relationship": "Outer", "Relation Name": "orders", "Alias": "o",
      "Startup Cost": 0.00, "Total Cost": 210.00, "Plan Rows": 5000, "Actual Startup Time": 0.010,
      "Actual Total Time": 1.200, "Actual Rows": 5000, "Actual Loops": 1, "Filter": "(total > 10)",
      "Shared Hit Blocks": 30, "Shared Read Blocks": 2},
     {"Node Type": "Hash", "Parent Relationship": "Inner", "Startup Cost": 11.80, "Total Cost": 11.80,
      "Plan Rows": 24, "Actual Total Time": 0.300, "Actual Rows": 20, "Actual Loops": 1,
      "Plans": [
        {"Node Type": "Index Scan", "Parent Relationship": "Outer", "Relation Name": "users", "Alias": "u",
         "Index Name": "users_pkey", "Startup Cost": 0.28, "Total Cost": 11.80, "Plan Rows": 24,
         "Actual Total Time": 0.280, "Actual Rows": 20, "Actual Loops": 1, "Index Cond": "(id < 50)",
         "Shared Hit Blocks": 11, "Shared Read Blocks": 1}]}]},
  "Planning Time": 0.210, "Execution Time": 2.600}
]`

const pgPlainFixture = `[{"Plan": {"Node Type": "Index Only Scan", "Relation Name": "users", "Alias": "users",
 "Index Name": "users_email_key", "Startup Cost": 0.29, "Total Cost": 4.31, "Plan Rows": 1}}]`

func TestParsePGExplainJSON(t *testing.T) {
	p, err := ParsePGExplainJSON([]byte(pgAnalyzeFixture))
	if err != nil {
		t.Fatal(err)
	}
	if !p.Analyzed || p.PlanningMs != 0.21 || p.ExecutionMs != 2.6 || p.Engine != EnginePostgres {
		t.Errorf("plan header = %+v", p)
	}
	r := p.Root
	if r.NodeType != "Hash Join" || r.ID != 1 || r.ParentID != -1 || r.TotalCost != 260.55 || r.StartupCost != 12.10 ||
		r.ActualRows != 455 || r.ActualMs != 2.41 || r.SharedHit != 41 || r.SharedRead != 3 || len(r.Children) != 2 {
		t.Fatalf("root = %+v", r)
	}
	if !strings.Contains(r.Detail, "Inner join") || !strings.Contains(r.Detail, "o.user_id = u.id") {
		t.Errorf("root detail = %q", r.Detail)
	}
	seq := r.Children[0]
	if seq.Scan != ScanFull || seq.Relation != "orders" || seq.ParentID != 1 || seq.PlanRows != 5000 || seq.Detail != "(total > 10)" {
		t.Errorf("seq = %+v", seq)
	}
	idx := r.Children[1].Children[0]
	if idx.Scan != ScanSearch || idx.Index != "users_pkey" || idx.ID != 4 || idx.ParentID != 3 {
		t.Errorf("idx = %+v", idx)
	}
	if p.FullScans != 1 || p.IndexSearches != 1 {
		t.Errorf("scan counts = %d/%d", p.FullScans, p.IndexSearches)
	}

	plain, err := ParsePGExplainJSON([]byte(pgPlainFixture))
	if err != nil || plain.Analyzed || plain.Root.ActualMs != -1 || plain.Root.Scan != ScanSearch {
		t.Errorf("plain = %+v err=%v", plain, err)
	}
	for _, bad := range []string{"", "[]", "{", `{"Plan":1}`} {
		if _, err := ParsePGExplainJSON([]byte(bad)); err == nil {
			t.Errorf("ParsePGExplainJSON(%q) should fail", bad)
		}
	}
}

func TestPostgresDSN(t *testing.T) {
	dsn := PostgresDSN(ConnectionConfig{Host: "db.local", Port: 6543, Database: "app", User: "me", Password: "p@ss/w", SSLMode: "verify-full", ConnectTimeout: 3 * time.Second})
	for _, want := range []string{"postgres://me:p%40ss%2Fw@db.local:6543/app", "sslmode=verify-full", "connect_timeout=3"} {
		if !strings.Contains(dsn, want) {
			t.Errorf("dsn %q missing %q", dsn, want)
		}
	}
	def := PostgresDSN(ConnectionConfig{Database: "x"})
	if !strings.Contains(def, "localhost:5432") || !strings.Contains(def, "sslmode=prefer") {
		t.Errorf("defaults dsn = %q", def)
	}
}

func TestPostgresFKActionAndDDL(t *testing.T) {
	if pgFKAction("c") != "CASCADE" || pgFKAction("n") != "SET NULL" || pgFKAction("a") != "NO ACTION" {
		t.Error("fk actions")
	}
	ddl := BuildCreateTable("public", "t", []ColumnMeta{
		{Name: "id", Type: "bigint", PrimaryKey: true},
		{Name: "n", Type: "text", Nullable: false, DefaultVal: "'x'::text"},
	}, []string{`CONSTRAINT "t_pkey" PRIMARY KEY (id)`}, []string{"CREATE INDEX i ON public.t USING btree (n)"})
	want := "CREATE TABLE \"public\".\"t\" (\n    \"id\" bigint,\n    \"n\" text DEFAULT 'x'::text NOT NULL,\n    CONSTRAINT \"t_pkey\" PRIMARY KEY (id)\n);\n\nCREATE INDEX i ON public.t USING btree (n);\n"
	if ddl != want {
		t.Errorf("ddl =\n%s\nwant\n%s", ddl, want)
	}
}

func TestPostgresConnectFailure(t *testing.T) {
	ctx := context.Background()
	_, err := OpenPostgres(ctx, ConnectionConfig{Engine: EnginePostgres, Host: "127.0.0.1", Port: 1, Database: "x", User: "u", Password: "secret-pw", SSLMode: "disable", ConnectTimeout: time.Second})
	if err == nil {
		t.Fatal("expected connect failure")
	}
	if strings.Contains(err.Error(), "secret-pw") {
		t.Errorf("error leaks password: %v", err)
	}
}

// TestPostgresIntegration runs against a live server when QUARRY_TEST_PG_DSN
// (a postgres:// URL) is set; otherwise it skips.
func TestPostgresIntegration(t *testing.T) {
	dsn := os.Getenv("QUARRY_TEST_PG_DSN")
	if dsn == "" {
		t.Skip("QUARRY_TEST_PG_DSN not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cfg, err := ConfigFromDSN(dsn)
	if err != nil {
		t.Fatal(err)
	}
	d, err := OpenPostgres(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = d.Close() }()
	if v, err := d.Version(ctx); err != nil || !strings.HasPrefix(v, "PostgreSQL") {
		t.Fatalf("version = %q err=%v", v, err)
	}
	if _, err := d.Exec(ctx, "CREATE TEMP TABLE quarry_it(id serial PRIMARY KEY, j jsonb, u uuid, t timestamptz, b bytea, a int[])"); err != nil {
		// temp tables are per connection; use a regular table cleaned afterward
		t.Log("temp table path unavailable:", err)
	}
	_, _ = d.Exec(ctx, "DROP TABLE IF EXISTS quarry_it")
	if _, err := d.Exec(ctx, "CREATE TABLE quarry_it(id serial PRIMARY KEY, j jsonb, u uuid, t timestamptz, b bytea, a int[])"); err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = d.Exec(ctx, "DROP TABLE IF EXISTS quarry_it") }()
	if _, err := d.Exec(ctx, `INSERT INTO quarry_it(j,u,t,b,a) VALUES ('{"k":1}','11111111-2222-4333-8444-555555555555',now(),'\xdeadbeef','{1,2,3}')`); err != nil {
		t.Fatal(err)
	}
	res, err := d.Query(ctx, "SELECT * FROM quarry_it", 10)
	if err != nil || len(res.Rows) != 1 {
		t.Fatalf("query: %v %+v", err, res)
	}
	row := res.Rows[0]
	if row[1].Kind != KindJSON || row[2].S != "11111111-2222-4333-8444-555555555555" || row[3].Kind != KindTime || row[4].Kind != KindBlob {
		t.Errorf("decoded row = %+v", row)
	}
	if row[5].S != "[1,2,3]" {
		t.Errorf("array = %q", row[5].S)
	}
	schemas, err := d.ListSchemas(ctx)
	if err != nil || len(schemas) == 0 {
		t.Fatalf("schemas: %v", err)
	}
	cols, err := d.TableColumns(ctx, "public", "quarry_it")
	if err != nil || len(cols) != 6 || !cols[0].PrimaryKey {
		t.Errorf("cols = %+v err=%v", cols, err)
	}
	if ixs, err := d.TableIndexes(ctx, "public", "quarry_it"); err != nil || len(ixs) != 1 || !ixs[0].Primary {
		t.Errorf("indexes = %+v err=%v", ixs, err)
	}
	if ddl, err := d.TableDDL(ctx, "public", "quarry_it"); err != nil || !strings.Contains(ddl, "CREATE TABLE") {
		t.Errorf("ddl = %q err=%v", ddl, err)
	}
	if p, err := d.Explain(ctx, "SELECT * FROM quarry_it", true); err != nil || !p.Analyzed || p.Root == nil {
		t.Errorf("explain = %+v err=%v", p, err)
	}
	if m, err := d.CollectMetrics(ctx); err != nil || m.DBSizeBytes <= 0 {
		t.Errorf("metrics = %+v err=%v", m, err)
	}
	ro := cfg
	ro.ReadOnly = true
	rd, err := OpenPostgres(ctx, ro)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rd.Close() }()
	if _, err := rd.Exec(ctx, "DELETE FROM quarry_it"); err == nil {
		t.Error("read-only delete should fail")
	}
}
