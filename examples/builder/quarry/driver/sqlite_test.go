package driver

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func seeded(t *testing.T, readOnly bool) Driver {
	t.Helper()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), DemoDBName)
	if err := SeedDemoDB(ctx, path); err != nil {
		t.Fatal(err)
	}
	d, err := Open(ctx, ConnectionConfig{Engine: EngineSQLite, FilePath: path, ReadOnly: readOnly, Name: "demo"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = d.Close() })
	return d
}

func TestSeedDemoDB(t *testing.T) {
	ctx := context.Background()
	d := seeded(t, false)
	want := map[string]int64{"users": SeedUsers, "products": SeedProducts, "orders": SeedOrders, "order_items": SeedOrderItems, "audit_events": SeedAudit}
	for table, n := range want {
		res, err := d.Query(ctx, "SELECT count(*) FROM "+table, 0)
		if err != nil {
			t.Fatal(err)
		}
		if got := res.Rows[0][0].I; got != n {
			t.Errorf("%s rows = %d, want %d", table, got, n)
		}
	}
	for _, v := range []string{"v_revenue_by_day", "v_top_customers"} {
		res, err := d.Query(ctx, "SELECT * FROM "+v, 0)
		if err != nil || len(res.Rows) == 0 {
			t.Errorf("view %s: rows=%d err=%v", v, len(res.Rows), err)
		}
	}
	fks, err := d.TableForeignKeys(ctx, "main", "order_items")
	if err != nil || len(fks) != 2 {
		t.Errorf("order_items fks = %v err=%v", fks, err)
	}
	// idempotent: reseeding must not duplicate.
	if err := SeedDemoDB(ctx, d.Config().FilePath); err != nil {
		t.Fatal(err)
	}
	res, _ := d.Query(ctx, "SELECT count(*) FROM users", 0)
	if res.Rows[0][0].I != SeedUsers {
		t.Error("reseed duplicated rows")
	}
}

func TestSQLiteQueryTruncationAndTypes(t *testing.T) {
	ctx := context.Background()
	d := seeded(t, false)
	res, err := d.Query(ctx, "SELECT id, user_id, action, payload FROM audit_events ORDER BY id", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Rows) != 10 || !res.Truncated {
		t.Fatalf("rows=%d truncated=%v", len(res.Rows), res.Truncated)
	}
	res, err = d.Query(ctx, "SELECT id, payload FROM audit_events WHERE id = 1", 10)
	if err != nil || res.Truncated || len(res.Rows) != 1 {
		t.Fatalf("exact fit: %v %+v", err, res)
	}
	if res.Rows[0][0].Kind != KindInt || res.Rows[0][1].Kind != KindJSON || !strings.HasPrefix(res.Rows[0][1].S, "{") {
		t.Errorf("types = %v %v", res.Rows[0][0].Kind, res.Rows[0][1].Kind)
	}
	res, _ = d.Query(ctx, "SELECT active, NULL, x'0102', 1.5 FROM users LIMIT 1", 0)
	row := res.Rows[0]
	if row[0].Kind != KindBool || !row[1].IsNull() || row[2].Kind != KindBlob || row[3].Kind != KindFloat {
		t.Errorf("kinds = %v %v %v %v", row[0].Kind, row[1].Kind, row[2].Kind, row[3].Kind)
	}
	if _, err := d.Query(ctx, "SELECT * FROM nope", 0); err == nil {
		t.Error("expected error for missing table")
	}
}

func TestSQLiteCancellation(t *testing.T) {
	d := seeded(t, false)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err := d.Query(ctx, `WITH RECURSIVE c(x) AS (SELECT 1 UNION ALL SELECT x+1 FROM c) SELECT count(*) FROM c`, 0)
	if err == nil {
		t.Fatal("expected cancellation error")
	}
	if time.Since(start) > 5*time.Second {
		t.Errorf("cancellation took %v", time.Since(start))
	}
}

func TestSQLiteCancelQueryByPID(t *testing.T) {
	d := seeded(t, false)
	done := make(chan error, 1)
	go func() {
		_, err := d.Query(context.Background(), `WITH RECURSIVE c(x) AS (SELECT 1 UNION ALL SELECT x+1 FROM c) SELECT count(*) FROM c`, 0)
		done <- err
	}()
	var pid int
	for i := 0; i < 200 && pid == 0; i++ {
		snap, _ := d.CollectMetrics(context.Background())
		if len(snap.ActiveQueries) > 0 {
			pid = snap.ActiveQueries[0].PID
		}
		time.Sleep(10 * time.Millisecond)
	}
	if pid == 0 {
		t.Fatal("running statement never appeared in ActiveQueries")
	}
	if err := d.CancelQuery(context.Background(), pid); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if err == nil {
			t.Error("cancelled query returned no error")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("query not cancelled")
	}
	if err := d.CancelQuery(context.Background(), 9999); err == nil {
		t.Error("unknown pid should error")
	}
}

func TestSQLiteReadOnly(t *testing.T) {
	ctx := context.Background()
	d := seeded(t, true)
	if _, err := d.Query(ctx, "SELECT count(*) FROM users", 0); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Query(ctx, "DELETE FROM users", 0); !errors.Is(err, ErrReadOnly) {
		t.Errorf("Query delete err = %v", err)
	}
	if _, err := d.Exec(ctx, "DELETE FROM users"); !errors.Is(err, ErrReadOnly) {
		t.Errorf("Exec delete err = %v", err)
	}
	res, _ := d.Query(ctx, "SELECT count(*) FROM users", 0)
	if res.Rows[0][0].I != SeedUsers {
		t.Error("read-only connection lost rows")
	}
}

func TestSQLiteExecAndMemory(t *testing.T) {
	ctx := context.Background()
	d, err := Open(ctx, ConnectionConfig{Engine: EngineSQLite, FilePath: ":memory:"})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = d.Close() }()
	if _, err := d.Exec(ctx, "CREATE TABLE t(a INTEGER PRIMARY KEY, b TEXT)"); err != nil {
		t.Fatal(err)
	}
	n, err := d.Exec(ctx, "INSERT INTO t(b) VALUES ('x'), ('y')")
	if err != nil || n != 2 {
		t.Fatalf("insert n=%d err=%v", n, err)
	}
	v, err := d.Version(ctx)
	if err != nil || !strings.HasPrefix(v, "SQLite 3") {
		t.Errorf("version = %q err=%v", v, err)
	}
	if _, err := Open(ctx, ConnectionConfig{Engine: EngineSQLite}); err == nil {
		t.Error("empty path should fail")
	}
}

func findNode(nodes []SchemaNode, name string) *SchemaNode {
	for i := range nodes {
		if nodes[i].Name == name {
			return &nodes[i]
		}
	}
	return nil
}

func TestSQLiteIntrospection(t *testing.T) {
	ctx := context.Background()
	d := seeded(t, false)
	schemas, err := d.ListSchemas(ctx)
	if err != nil || len(schemas) != 1 || schemas[0].Name != "main" {
		t.Fatalf("schemas = %+v err=%v", schemas, err)
	}
	tables := findNode(schemas[0].Children, "Tables")
	views := findNode(schemas[0].Children, "Views")
	idxs := findNode(schemas[0].Children, "Indexes")
	if tables == nil || views == nil || idxs == nil {
		t.Fatalf("missing groups: %+v", schemas[0].Children)
	}
	if o := findNode(tables.Children, "orders"); o == nil || o.RowCount != SeedOrders || o.Kind != NodeTable {
		t.Errorf("orders node = %+v", o)
	}
	if findNode(views.Children, "v_top_customers") == nil {
		t.Error("view missing")
	}
	if u := findNode(idxs.Children, "idx_orders_user"); u == nil || u.Table != "orders" {
		t.Errorf("index node = %+v", u)
	}

	cols, err := d.TableColumns(ctx, "main", "orders")
	if err != nil || len(cols) != 5 {
		t.Fatalf("cols = %+v err=%v", cols, err)
	}
	if !cols[0].PrimaryKey || cols[0].Nullable || cols[0].Name != "id" {
		t.Errorf("id col = %+v", cols[0])
	}
	if !cols[1].ForeignKey || cols[1].Nullable {
		t.Errorf("user_id col = %+v", cols[1])
	}
	ucols, _ := d.TableColumns(ctx, "", "users")
	if ucols[4].DefaultVal != "'free'" || ucols[5].DefaultVal != "1" {
		t.Errorf("defaults = %q %q", ucols[4].DefaultVal, ucols[5].DefaultVal)
	}
	if _, err := d.TableColumns(ctx, "main", "missing"); err == nil {
		t.Error("missing table should error")
	}

	fks, err := d.TableForeignKeys(ctx, "main", "orders")
	if err != nil || len(fks) != 1 {
		t.Fatalf("fks = %+v err=%v", fks, err)
	}
	if fk := fks[0]; fk.RefTable != "users" || fk.Columns[0] != "user_id" || fk.RefColumns[0] != "id" || fk.OnDelete != "CASCADE" {
		t.Errorf("fk = %+v", fk)
	}

	ixs, err := d.TableIndexes(ctx, "main", "orders")
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]IndexMeta{}
	for _, ix := range ixs {
		byName[ix.Name] = ix
	}
	if ix := byName["idx_orders_user"]; len(ix.Columns) != 1 || ix.Columns[0] != "user_id" || ix.Unique || !strings.Contains(ix.Definition, "CREATE INDEX") {
		t.Errorf("idx_orders_user = %+v", ix)
	}
	if pk := byName["PRIMARY KEY"]; !pk.Primary || pk.Columns[0] != "id" {
		t.Errorf("primary key entry = %+v", pk)
	}
	uix, _ := d.TableIndexes(ctx, "main", "users")
	foundUnique := false
	for _, ix := range uix {
		if ix.Unique && len(ix.Columns) == 1 && ix.Columns[0] == "email" {
			foundUnique = true
		}
	}
	if !foundUnique {
		t.Errorf("users unique email index missing: %+v", uix)
	}

	ddl, err := d.TableDDL(ctx, "main", "orders")
	if err != nil || !strings.Contains(ddl, "CREATE TABLE orders") || !strings.Contains(ddl, "CREATE INDEX idx_orders_user") {
		t.Errorf("ddl = %q err=%v", ddl, err)
	}
	vddl, err := d.TableDDL(ctx, "main", "v_revenue_by_day")
	if err != nil || !strings.Contains(vddl, "CREATE VIEW") {
		t.Errorf("view ddl = %q err=%v", vddl, err)
	}
	if _, err := d.TableDDL(ctx, "main", "missing"); err == nil {
		t.Error("missing ddl should error")
	}
}

func TestSQLiteExplain(t *testing.T) {
	ctx := context.Background()
	d := seeded(t, false)
	full, err := d.Explain(ctx, "SELECT * FROM orders WHERE total_cents > 100", false)
	if err != nil {
		t.Fatal(err)
	}
	if full.FullScans != 1 || full.IndexSearches != 0 || full.Root == nil || len(full.Root.Children) != 1 {
		t.Fatalf("full-scan plan = %+v", full)
	}
	if n := full.Root.Children[0]; n.Scan != ScanFull || n.Relation != "orders" {
		t.Errorf("scan node = %+v", n)
	}
	idx, err := d.Explain(ctx, "SELECT * FROM orders WHERE user_id = 7", false)
	if err != nil {
		t.Fatal(err)
	}
	n := idx.Root.Children[0]
	if idx.IndexSearches != 1 || n.Scan != ScanSearch || n.Index != "idx_orders_user" || n.Relation != "orders" {
		t.Errorf("index plan = %+v node=%+v", idx, n)
	}
	pk, _ := d.Explain(ctx, "SELECT * FROM orders WHERE id = 7", false)
	if pk.Root.Children[0].Index != "PRIMARY KEY" {
		t.Errorf("pk plan node = %+v", pk.Root.Children[0])
	}
	an, err := d.Explain(ctx, "SELECT count(*) FROM audit_events", true)
	if err != nil || !an.Analyzed || an.ExecutionMs <= 0 {
		t.Errorf("analyze = %+v err=%v", an, err)
	}
	ro := seeded(t, true)
	if _, err := ro.Explain(ctx, "DELETE FROM orders", false); !errors.Is(err, ErrReadOnly) {
		t.Errorf("read-only explain err = %v", err)
	}
	// nested plan: subquery yields a child step
	nest, err := d.Explain(ctx, "SELECT * FROM users WHERE id IN (SELECT user_id FROM orders)", false)
	if err != nil || nest.Root == nil || len(nest.Root.Children) == 0 {
		t.Errorf("nested plan = %+v err=%v", nest, err)
	}
}

func TestSQLiteMetrics(t *testing.T) {
	ctx := context.Background()
	d := seeded(t, false)
	_, _ = d.Query(ctx, "SELECT 1", 0)
	snap, err := d.CollectMetrics(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if snap.DBSizeBytes <= 0 || snap.Extra["page_count"] <= 0 || snap.CommitTotal < 1 || snap.CacheHitRatio >= 0 {
		t.Errorf("snapshot = %+v", snap)
	}
}

func TestSQLiteExecTxRollsBackAndReportsIndex(t *testing.T) {
	ctx := context.Background()
	d, err := OpenSQLite(ctx, ConnectionConfig{Name: "tx", Engine: EngineSQLite, FilePath: filepath.Join(t.TempDir(), "tx.db")})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = d.Close() }()
	tx, ok := d.(TxExecer)
	if !ok {
		t.Fatal("sqlite driver must implement TxExecer")
	}
	if _, err := d.Exec(ctx, "CREATE TABLE t (id INTEGER PRIMARY KEY, v TEXT UNIQUE)"); err != nil {
		t.Fatal(err)
	}
	failed, err := tx.ExecTx(ctx,
		[]string{"INSERT INTO t(id, v) VALUES (?, ?)", "INSERT INTO t(id, v) VALUES (?, ?)", "INSERT INTO t(id, v) VALUES (?, ?)"},
		[][]any{{1, "a"}, {2, "b"}, {3, "a"}})
	if err == nil || failed != 2 {
		t.Fatalf("failed=%d err=%v, want index 2 and an error", failed, err)
	}
	res, _ := d.Query(ctx, "SELECT COUNT(*) FROM t", 1)
	if res.Rows[0][0].I != 0 {
		t.Fatalf("rows left after rollback: %d", res.Rows[0][0].I)
	}
	if failed, err := tx.ExecTx(ctx, []string{"INSERT INTO t(id, v) VALUES (?, ?)"}, [][]any{{1, "a"}}); err != nil || failed != -1 {
		t.Fatalf("commit: %d %v", failed, err)
	}
}
