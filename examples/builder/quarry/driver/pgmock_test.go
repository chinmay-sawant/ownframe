package driver

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func mockDriver(t *testing.T, readOnly bool) Driver {
	t.Helper()
	d, err := Open(context.Background(), ConnectionConfig{Engine: EnginePostgres, Mock: true, ReadOnly: readOnly})
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// TestPGMockEveryMethod exercises every Driver method on the mock.
func TestPGMockEveryMethod(t *testing.T) {
	ctx := context.Background()
	d := mockDriver(t, false)
	if d.Engine() != EnginePostgres || !d.Config().Mock {
		t.Error("engine/config")
	}
	if err := d.Ping(ctx); err != nil {
		t.Fatal(err)
	}
	if v, _ := d.Version(ctx); !strings.HasPrefix(v, "PostgreSQL 16") {
		t.Errorf("version = %q", v)
	}
	schemas, err := d.ListSchemas(ctx)
	if err != nil || len(schemas) != 3 {
		t.Fatalf("schemas = %d err=%v", len(schemas), err)
	}
	names := []string{schemas[0].Name, schemas[1].Name, schemas[2].Name}
	if strings.Join(names, ",") != "public,analytics,billing" {
		t.Errorf("schema order = %v", names)
	}
	if g := findNode(schemas[1].Children, "Materialized Views"); g == nil || len(g.Children) != 1 || g.Children[0].Kind != NodeMatView {
		t.Errorf("analytics matviews = %+v", g)
	}
	if g := findNode(schemas[2].Children, "Tables"); g == nil || len(g.Children) != 2 {
		t.Errorf("billing tables = %+v", g)
	}
	cols, err := d.TableColumns(ctx, "public", "orders")
	if err != nil || len(cols) != 5 || !cols[0].PrimaryKey || !cols[1].ForeignKey {
		t.Errorf("cols = %+v err=%v", cols, err)
	}
	ixs, _ := d.TableIndexes(ctx, "public", "orders")
	if len(ixs) != 2 || !ixs[0].Primary {
		t.Errorf("indexes = %+v", ixs)
	}
	fks, _ := d.TableForeignKeys(ctx, "billing", "payments")
	if len(fks) != 1 || fks[0].RefTable != "invoices" {
		t.Errorf("fks = %+v", fks)
	}
	ddl, err := d.TableDDL(ctx, "public", "orders")
	if err != nil || !strings.Contains(ddl, `CREATE TABLE "public"."orders"`) || !strings.Contains(ddl, "FOREIGN KEY") || !strings.Contains(ddl, "CREATE INDEX orders_user_idx") {
		t.Errorf("ddl = %q err=%v", ddl, err)
	}
	if _, err := d.TableDDL(ctx, "analytics", "v_funnel"); err != nil {
		t.Error(err)
	}
	if _, err := d.TableColumns(ctx, "public", "nope"); err == nil {
		t.Error("missing table should error")
	}

	res, err := d.Query(ctx, "SELECT * FROM public.users LIMIT 25", 10)
	if err != nil || len(res.Rows) != 10 || !res.Truncated || len(res.Columns) != 5 {
		t.Fatalf("query rows=%d trunc=%v err=%v", len(res.Rows), res.Truncated, err)
	}
	again, _ := d.Query(ctx, "SELECT * FROM public.users LIMIT 25", 10)
	for i := range res.Rows {
		for j := range res.Rows[i] {
			if !res.Rows[i][j].Equal(again.Rows[i][j]) {
				t.Fatal("mock rows are not deterministic")
			}
		}
	}
	cnt, _ := d.Query(ctx, "select count(*) from orders", 0)
	if cnt.Rows[0][0].I != 58210 {
		t.Errorf("count = %v", cnt.Rows[0][0])
	}
	pj, _ := d.Query(ctx, "SELECT * FROM products LIMIT 1", 0)
	if pj.Rows[0][4].Kind != KindJSON {
		t.Errorf("jsonb kind = %v", pj.Rows[0][4].Kind)
	}
	if n, err := d.Exec(ctx, "UPDATE users SET active = false"); err != nil || n != 1 {
		t.Errorf("exec = %d %v", n, err)
	}

	plan, err := d.Explain(ctx, "SELECT * FROM orders", false)
	if err != nil || plan.Root.Scan != ScanFull || plan.FullScans != 1 || plan.Analyzed {
		t.Errorf("seq plan = %+v err=%v", plan, err)
	}
	plan, err = d.Explain(ctx, "SELECT * FROM orders WHERE id = 42", true)
	if err != nil || !plan.Analyzed || plan.IndexSearches != 1 || plan.ExecutionMs <= 0 {
		t.Fatalf("index plan = %+v err=%v", plan, err)
	}
	if leaf := plan.Root.Children[0]; leaf.Index != "orders_pkey" || leaf.ActualRows != 1 || leaf.SharedHit != 4 {
		t.Errorf("leaf = %+v", leaf)
	}

	s1, err := d.CollectMetrics(ctx)
	if err != nil || s1.TPS <= 0 || s1.CacheHitRatio <= 0.9 || len(s1.ActiveQueries) < 2 || len(s1.Storage) == 0 {
		t.Fatalf("metrics = %+v err=%v", s1, err)
	}
	s2, _ := d.CollectMetrics(ctx)
	if s2.CommitTotal <= s1.CommitTotal || !s2.CapturedAt.After(s1.CapturedAt) {
		t.Error("counters must advance")
	}
	if err := d.CancelQuery(ctx, 4101); err == nil {
		t.Error("cancelling a non-running pid should error")
	}
	if err := d.Close(); err != nil {
		t.Fatal(err)
	}
	if err := d.Ping(ctx); err == nil {
		t.Error("ping after close should fail")
	}
}

func TestPGMockReadOnlyAndCancel(t *testing.T) {
	ctx := context.Background()
	ro := mockDriver(t, true)
	if _, err := ro.Exec(ctx, "DELETE FROM users"); !errors.Is(err, ErrReadOnly) {
		t.Errorf("exec err = %v", err)
	}
	if _, err := ro.Query(ctx, "DROP TABLE users", 0); !errors.Is(err, ErrReadOnly) {
		t.Errorf("query err = %v", err)
	}
	if _, err := ro.Explain(ctx, "DELETE FROM users", false); !errors.Is(err, ErrReadOnly) {
		t.Errorf("explain err = %v", err)
	}

	d := mockDriver(t, false)
	done := make(chan error, 1)
	go func() { _, err := d.Query(ctx, "SELECT pg_sleep(30)", 0); done <- err }()
	var pid int
	for i := 0; i < 200 && pid == 0; i++ {
		snap, _ := d.CollectMetrics(ctx)
		for _, q := range snap.ActiveQueries {
			if strings.Contains(q.Query, "pg_sleep") {
				pid = q.PID
			}
		}
		time.Sleep(5 * time.Millisecond)
	}
	if pid == 0 {
		t.Fatal("pg_sleep not visible")
	}
	if err := d.CancelQuery(ctx, pid); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("err = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("not cancelled")
	}
}

func TestPGMockMetricsDeterministic(t *testing.T) {
	ctx := t.Context()
	a, err := Open(ctx, ConnectionConfig{Engine: EnginePostgres, Mock: true})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = a.Close() }()
	b, err := Open(ctx, ConnectionConfig{Engine: EnginePostgres, Mock: true})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = b.Close() }()
	for range 3 {
		sa, err := a.CollectMetrics(ctx)
		if err != nil {
			t.Fatal(err)
		}
		sb, err := b.CollectMetrics(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if sa.TPS != sb.TPS || sa.CommitTotal != sb.CommitTotal || sa.DBSizeBytes != sb.DBSizeBytes {
			t.Fatalf("mock metrics not deterministic: %+v vs %+v", sa, sb)
		}
		if sa.CacheHitRatio <= 0 || sa.CacheHitRatio > 1 || sa.DBSizeBytes <= 0 || len(sa.Storage) == 0 {
			t.Fatalf("implausible snapshot: %+v", sa)
		}
	}
}
