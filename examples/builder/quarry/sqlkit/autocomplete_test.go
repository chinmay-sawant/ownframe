package sqlkit

import (
	"strings"
	"testing"
)

var testSchema = Schema{Tables: []Table{
	{Name: "orders", Columns: []Column{{"id", "INTEGER"}, {"customer_id", "INTEGER"}, {"total", "REAL"}, {"status", "TEXT"}, {"created_at", "TEXT"}}},
	{Name: "customers", Columns: []Column{{"id", "INTEGER"}, {"name", "TEXT"}, {"email", "TEXT"}}},
	{Name: "order_items", Columns: []Column{{"order_id", "INTEGER"}, {"sku", "TEXT"}}},
	{Schema: "audit", Name: "events", Columns: []Column{{"id", "BIGINT"}, {"payload", "JSONB"}}},
}}

// complete runs Complete with the caret at the "|" marker in sql.
func complete(t *testing.T, sql string, d Dialect) (Completion, string) {
	t.Helper()
	caret := strings.Index(sql, "|")
	if caret < 0 {
		t.Fatal("missing | marker")
	}
	src := sql[:caret] + sql[caret+1:]
	return Complete(src, caret, testSchema, d), src
}

func labels(c Completion) []string {
	var out []string
	for _, it := range c.Items {
		out = append(out, it.Label)
	}
	return out
}

func has(c Completion, label string, kind ItemKind) bool {
	for _, it := range c.Items {
		if it.Label == label && it.Kind == kind {
			return true
		}
	}
	return false
}

func TestCompleteTablesAfterFrom(t *testing.T) {
	for _, sql := range []string{"SELECT * FROM |", "SELECT * FROM o|", "UPDATE |", "INSERT INTO |", "SELECT 1 FROM orders JOIN |"} {
		c, _ := complete(t, sql, SQLite)
		if !has(c, "orders", ItemTable) || !has(c, "customers", ItemTable) && !strings.HasSuffix(sql, "o|") {
			t.Errorf("%q: %v", sql, labels(c))
		}
		for _, it := range c.Items {
			if it.Kind == ItemKeyword || it.Kind == ItemColumn {
				t.Errorf("%q: unexpected %v in table context", sql, it)
			}
		}
	}
}

func TestCompletePrefixFilteringAndRanking(t *testing.T) {
	c, src := complete(t, "SELECT * FROM ord|", SQLite)
	if c.Prefix != "ord" || src[c.ReplaceStart:c.ReplaceEnd] != "ord" {
		t.Fatalf("replace range: %+v", c)
	}
	got := labels(c)
	if len(got) < 2 || got[0] != "orders" || got[1] != "order_items" {
		t.Fatalf("ranking: %v", got)
	}
	for _, l := range got {
		if !strings.Contains(strings.ToLower(l), "o") {
			t.Fatalf("non-matching %q in %v", l, got)
		}
	}
	// Case-insensitive.
	c, _ = complete(t, "SELECT * FROM ORD|", SQLite)
	if len(c.Items) == 0 || c.Items[0].Label != "orders" {
		t.Fatalf("case: %v", labels(c))
	}
}

func TestCompleteAliasColumns(t *testing.T) {
	c, _ := complete(t, "SELECT o.| FROM orders o", SQLite)
	if len(c.Items) != 5 {
		t.Fatalf("want 5 orders columns, got %v", labels(c))
	}
	for _, it := range c.Items {
		if it.Kind != ItemColumn || it.Detail == "" {
			t.Fatalf("bad item %+v", it)
		}
	}
	c, _ = complete(t, "SELECT o.cr| FROM orders AS o JOIN customers c ON c.id = o.customer_id", SQLite)
	if len(c.Items) == 0 || c.Items[0].Label != "created_at" {
		t.Fatalf("AS alias prefix: %v", labels(c))
	}
	c, _ = complete(t, "SELECT c.| FROM orders o JOIN customers c ON c.id = o.customer_id", Postgres)
	if !has(c, "email", ItemColumn) || has(c, "total", ItemColumn) {
		t.Fatalf("second alias: %v", labels(c))
	}
	// Comma-separated FROM list.
	c, _ = complete(t, "SELECT b.| FROM orders a, customers b", SQLite)
	if !has(c, "name", ItemColumn) || has(c, "status", ItemColumn) {
		t.Fatalf("comma list: %v", labels(c))
	}
	// Unaliased table name qualifier.
	c, _ = complete(t, "SELECT customers.| FROM customers", SQLite)
	if !has(c, "email", ItemColumn) {
		t.Fatalf("table qualifier: %v", labels(c))
	}
}

func TestCompleteSchemaQualifier(t *testing.T) {
	c, _ := complete(t, "SELECT * FROM audit.|", Postgres)
	if len(c.Items) != 1 || c.Items[0].Label != "events" || c.Items[0].Kind != ItemTable {
		t.Fatalf("%v", labels(c))
	}
	c, _ = complete(t, "SELECT e.| FROM audit.events e", Postgres)
	if !has(c, "payload", ItemColumn) {
		t.Fatalf("%v", labels(c))
	}
}

func TestCompleteColumnsWhereAndOrderBy(t *testing.T) {
	for _, sql := range []string{
		"SELECT * FROM orders WHERE |",
		"SELECT * FROM orders WHERE total > 1 AND |",
		"SELECT * FROM orders ORDER BY |",
		"SELECT | FROM orders",
		"SELECT id, | FROM orders",
	} {
		c, _ := complete(t, sql, SQLite)
		if !has(c, "customer_id", ItemColumn) {
			t.Errorf("%q missing columns: %v", sql, labels(c))
		}
		// Scoped columns outrank keywords and functions.
		if c.Items[0].Kind != ItemColumn {
			t.Errorf("%q: first is %+v", sql, c.Items[0])
		}
	}
	// Columns of unreferenced tables are not offered when a table is in scope.
	c, _ := complete(t, "SELECT * FROM orders WHERE |", SQLite)
	if has(c, "email", ItemColumn) {
		t.Fatal("email belongs to customers, which is not in scope")
	}
	// With no FROM at all, every table's columns are candidates.
	c, _ = complete(t, "SELECT em|", SQLite)
	if !has(c, "email", ItemColumn) {
		t.Fatalf("no-scope columns: %v", labels(c))
	}
}

func TestCompleteKeywordsAndFunctions(t *testing.T) {
	c, _ := complete(t, "SEL|", SQLite)
	if len(c.Items) == 0 || c.Items[0].Label != "SELECT" || c.Items[0].Kind != ItemKeyword {
		t.Fatalf("%v", labels(c))
	}
	c, _ = complete(t, "SELECT cou| FROM orders", SQLite)
	if !has(c, "COUNT", ItemFunction) {
		t.Fatalf("%v", labels(c))
	}
	c, _ = complete(t, "SELECT date_tr| FROM orders", Postgres)
	if !has(c, "DATE_TRUNC", ItemFunction) {
		t.Fatalf("pg function: %v", labels(c))
	}
	c, _ = complete(t, "SELECT date_tr| FROM orders", SQLite)
	if has(c, "DATE_TRUNC", ItemFunction) {
		t.Fatal("DATE_TRUNC is Postgres only")
	}
	c, _ = complete(t, "PRAG|", SQLite)
	if !has(c, "PRAGMA", ItemKeyword) {
		t.Fatalf("pragma: %v", labels(c))
	}
	c, _ = complete(t, "PRAG|", Postgres)
	if has(c, "PRAGMA", ItemKeyword) {
		t.Fatal("PRAGMA is SQLite only")
	}
}

func TestCompleteStaysInCurrentStatement(t *testing.T) {
	// The alias from the first statement must not leak into the second.
	c, _ := complete(t, "SELECT 1 FROM orders o; SELECT o.| FROM customers", SQLite)
	if has(c, "total", ItemColumn) {
		t.Fatalf("leaked scope: %v", labels(c))
	}
}

func TestCompleteQuotesAwkwardNames(t *testing.T) {
	s := Schema{Tables: []Table{{Name: "Order Items", Columns: []Column{{"Select", "TEXT"}}}}}
	c := Complete("SELECT * FROM ", len("SELECT * FROM "), s, SQLite)
	if len(c.Items) != 1 || c.Items[0].Insert != `"Order Items"` {
		t.Fatalf("%+v", c.Items)
	}
}

func TestApply(t *testing.T) {
	c, src := complete(t, "SELECT * FROM ord|", SQLite)
	out, caret := Apply(src, c, c.Items[0])
	if out != "SELECT * FROM orders" || caret != len(out) {
		t.Fatalf("%q %d", out, caret)
	}
	c, src = complete(t, "SELECT o.cr| FROM orders o", SQLite)
	out, caret = Apply(src, c, c.Items[0])
	if out != "SELECT o.created_at FROM orders o" || out[:caret] != "SELECT o.created_at" {
		t.Fatalf("%q %d", out, caret)
	}
}

func TestCompleteCapsResults(t *testing.T) {
	c, _ := complete(t, "|", SQLite)
	if len(c.Items) != maxItems {
		t.Fatalf("len %d", len(c.Items))
	}
}
