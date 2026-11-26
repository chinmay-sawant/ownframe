package driver

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"time"
)

// DemoDBName is the file name of the seeded demo database.
const DemoDBName = "quarry_demo.db"

// Row counts of the seeded demo database.
const (
	SeedUsers      = 500
	SeedProducts   = 120
	SeedOrders     = 2500
	SeedOrderItems = 6000
	SeedAudit      = 5000
)

// Parameters of the deterministic demo data.
const (
	seedDirMode       = 0o750
	seedRNG           = 8
	seedInactiveOdds  = 10
	seedPriceMin      = 500
	seedPriceSpan     = 49500
	seedStockMax      = 400
	seedMaxQty        = 4
	seedOctet         = 255
	seedFailOdds      = 20
	seedLatencyMin    = 5
	seedLatencySpan   = 400
	seedNullUserEvery = 50
	seedTimeStep      = 97 * time.Minute
	seedOrderTimeMul  = 3
)

const seedSchema = `
CREATE TABLE users (
    id         INTEGER PRIMARY KEY,
    email      TEXT    NOT NULL UNIQUE,
    full_name  TEXT    NOT NULL,
    country    TEXT    NOT NULL,
    plan       TEXT    NOT NULL DEFAULT 'free',
    active     BOOLEAN NOT NULL DEFAULT 1,
    created_at TEXT    NOT NULL
);
CREATE TABLE products (
    id          INTEGER PRIMARY KEY,
    sku         TEXT    NOT NULL UNIQUE,
    name        TEXT    NOT NULL,
    category    TEXT    NOT NULL,
    price_cents INTEGER NOT NULL,
    stock       INTEGER NOT NULL DEFAULT 0
);
CREATE TABLE orders (
    id          INTEGER PRIMARY KEY,
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status      TEXT    NOT NULL,
    total_cents INTEGER NOT NULL,
    placed_at   TEXT    NOT NULL
);
CREATE TABLE order_items (
    id         INTEGER PRIMARY KEY,
    order_id   INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES products(id),
    quantity   INTEGER NOT NULL,
    unit_cents INTEGER NOT NULL
);
CREATE TABLE audit_events (
    id          INTEGER PRIMARY KEY,
    user_id     INTEGER REFERENCES users(id) ON DELETE SET NULL,
    action      TEXT NOT NULL,
    payload     JSON,
    occurred_at TEXT NOT NULL
);
CREATE INDEX idx_orders_user      ON orders(user_id);
CREATE INDEX idx_orders_placed_at ON orders(placed_at);
CREATE INDEX idx_items_order      ON order_items(order_id);
CREATE INDEX idx_items_product    ON order_items(product_id);
CREATE INDEX idx_audit_user       ON audit_events(user_id);
CREATE INDEX idx_audit_action     ON audit_events(action, occurred_at);
CREATE VIEW v_revenue_by_day AS
    SELECT substr(placed_at, 1, 10) AS day, COUNT(*) AS orders, SUM(total_cents) / 100.0 AS revenue
    FROM orders WHERE status <> 'cancelled' GROUP BY substr(placed_at, 1, 10);
CREATE VIEW v_top_customers AS
    SELECT u.id, u.full_name, u.email, COUNT(o.id) AS orders, SUM(o.total_cents) / 100.0 AS lifetime_value
    FROM users u JOIN orders o ON o.user_id = u.id WHERE o.status <> 'cancelled'
    GROUP BY u.id ORDER BY lifetime_value DESC LIMIT 50;
`

// SeedDemoDB creates (when absent) the seeded demo database at path and
// returns once it is fully populated. It is idempotent: an existing database
// that already has the users table is left untouched. The data is deterministic.
func SeedDemoDB(ctx context.Context, path string) error {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, seedDirMode); err != nil {
			return err
		}
	}
	db, err := sql.Open("sqlite", sqliteDSN(ConnectionConfig{FilePath: path}))
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	db.SetMaxOpenConns(1)
	return SeedSQLite(ctx, db)
}

// EnsureDemoDB seeds <dir>/quarry_demo.db if needed and returns its path.
func EnsureDemoDB(ctx context.Context, dir string) (string, error) {
	p := filepath.Join(dir, DemoDBName)
	return p, SeedDemoDB(ctx, p)
}

// SeedSQLite populates an open SQLite handle (use MaxOpenConns(1) for
// ":memory:"). It does nothing when the users table already exists.
func SeedSQLite(ctx context.Context, db *sql.DB) error {
	var exists int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM sqlite_master WHERE name = 'users'`).Scan(&exists); err != nil {
		return err
	}
	if exists > 0 {
		return nil
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, seedSchema); err != nil {
		return fmt.Errorf("seed schema: %w", err)
	}
	// Deterministic demo data, not a security context.
	rng := rand.New(rand.NewPCG(seedRNG, seedRNG)) //nolint:gosec // G404: fixed-seed generator makes the demo database reproducible
	base := time.Date(2025, 1, 1, 8, 0, 0, 0, time.UTC)
	ts := func(i int) string { return base.Add(time.Duration(i) * seedTimeStep).Format("2006-01-02 15:04:05") }
	pick := func(xs ...string) string { return xs[rng.IntN(len(xs))] }

	ins := func(query string, n int, args func(i int) []any) error {
		st, err := tx.PrepareContext(ctx, query)
		if err != nil {
			return err
		}
		defer func() { _ = st.Close() }()
		for i := 1; i <= n; i++ {
			if _, err := st.ExecContext(ctx, args(i)...); err != nil {
				return err
			}
		}
		return nil
	}
	first := []string{"Ava", "Liam", "Noah", "Mia", "Zoe", "Ethan", "Ivy", "Leo", "Nora", "Omar", "Priya", "Sven"}
	last := []string{"Stone", "Rivera", "Khan", "Nguyen", "Okafor", "Silva", "Berg", "Tanaka", "Moreau", "Patel"}
	if err := ins(`INSERT INTO users(id, email, full_name, country, plan, active, created_at) VALUES (?,?,?,?,?,?,?)`, SeedUsers, func(i int) []any {
		fn, ln := first[rng.IntN(len(first))], last[rng.IntN(len(last))]
		return []any{i, fmt.Sprintf("%s.%s%d@example.com", fn, ln, i), fn + " " + ln, pick("US", "DE", "IN", "BR", "JP", "GB"), pick("free", "pro", "team"), rng.IntN(seedInactiveOdds) != 0, ts(i)}
	}); err != nil {
		return fmt.Errorf("seed users: %w", err)
	}
	cats := []string{"hardware", "software", "books", "apparel", "office"}
	if err := ins(`INSERT INTO products(id, sku, name, category, price_cents, stock) VALUES (?,?,?,?,?,?)`, SeedProducts, func(i int) []any {
		return []any{i, fmt.Sprintf("SKU-%04d", i), fmt.Sprintf("%s item %d", cats[i%len(cats)], i), cats[i%len(cats)], seedPriceMin + rng.IntN(seedPriceSpan), rng.IntN(seedStockMax)}
	}); err != nil {
		return fmt.Errorf("seed products: %w", err)
	}
	if err := ins(`INSERT INTO orders(id, user_id, status, total_cents, placed_at) VALUES (?,?,?,?,?)`, SeedOrders, func(i int) []any {
		return []any{i, 1 + rng.IntN(SeedUsers), pick("paid", "paid", "paid", "shipped", "pending", "cancelled"), 0, ts(i * seedOrderTimeMul)}
	}); err != nil {
		return fmt.Errorf("seed orders: %w", err)
	}
	if err := ins(`INSERT INTO order_items(id, order_id, product_id, quantity, unit_cents) VALUES (?,?,?,?,?)`, SeedOrderItems, func(i int) []any {
		return []any{i, 1 + (i-1)%SeedOrders, 1 + rng.IntN(SeedProducts), 1 + rng.IntN(seedMaxQty), seedPriceMin + rng.IntN(seedPriceSpan)}
	}); err != nil {
		return fmt.Errorf("seed order_items: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE orders SET total_cents =
        (SELECT COALESCE(SUM(quantity * unit_cents), 0) FROM order_items WHERE order_id = orders.id)`); err != nil {
		return err
	}
	actions := []string{"login", "logout", "order.create", "order.update", "profile.edit", "export", "password.reset"}
	if err := ins(`INSERT INTO audit_events(id, user_id, action, payload, occurred_at) VALUES (?,?,?,?,?)`, SeedAudit, func(i int) []any {
		a := actions[rng.IntN(len(actions))]
		var uid any = 1 + rng.IntN(SeedUsers)
		if i%seedNullUserEvery == 0 {
			uid = nil
		}
		payload := fmt.Sprintf(`{"action":%q,"ip":"10.%d.%d.%d","ok":%t,"meta":{"agent":%q,"latency_ms":%d}}`,
			a, rng.IntN(seedOctet), rng.IntN(seedOctet), rng.IntN(seedOctet), rng.IntN(seedFailOdds) != 0, pick("web", "ios", "android", "cli"), seedLatencyMin+rng.IntN(seedLatencySpan))
		return []any{i, uid, a, payload, ts(i)}
	}); err != nil {
		return fmt.Errorf("seed audit_events: %w", err)
	}
	return tx.Commit()
}
