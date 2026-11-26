package ui

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/driver"
	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/monitor"
	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/sqlkit"
	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/store"
)

// initPrefs opens the config store. A failure degrades to an in-memory store
// so the app still starts.
func (a *App) initPrefs(ctx context.Context, dir string) error {
	st, err := store.OpenContext(ctx, dir)
	if err != nil {
		st, err = store.OpenContext(ctx, store.Memory)
		if err != nil {
			return err
		}
	}
	a.prefs = st
	a.slow = monitor.NewSlowLog(monitor.DefaultSlowThreshold, slowLogCapacity)
	a.lat = monitor.NewLatencyTracker(latencyWindow)
	return nil
}

func (a *App) closePrefs() {
	if a.prefs != nil {
		_ = a.prefs.Close()
	}
}

func (a *App) loadPref(key string) (string, bool) {
	if a.prefs == nil {
		return "", false
	}
	v := a.prefs.GetPref(a.ctx, key, "")
	return v, v != ""
}

func (a *App) savePref(key, value string) {
	if a.prefs != nil {
		_ = a.prefs.SetPref(a.ctx, key, value)
	}
}

func (a *App) prefInt(key string, def int) int {
	if v, ok := a.loadPref(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

// restoreLayout loads persisted split sizes, clamped to their limits.
func (a *App) restoreLayout() {
	a.sidebarW = clampInt(a.prefInt("split.sidebar", a.sidebarW), minSidebar, maxSidebar)
	a.editorH = clampInt(a.prefInt("split.editor", a.editorH), minEditorPane, maxEditorPane)
	a.inspectorW = clampInt(a.prefInt("split.inspector", a.inspectorW), minInspector, maxInspector)
}

// restoreSettings loads the persisted preferences.
func (a *App) restoreSettings() {
	a.defLimit = a.prefInt("default.limit", thousandRows)
	if ms := a.prefInt("monitor.poll", int(defaultPollEvery.Milliseconds())); ms >= 0 {
		a.mon.pollEvery = time.Duration(ms) * time.Millisecond
	}
	if v, ok := a.loadPref("editor.wrap"); ok {
		a.wrap = v == "1"
	}
	if v, ok := a.loadPref("confirm.destructive"); ok {
		a.confirmDestructive = v != "0"
	}
	if ms := a.prefInt("slow.ms", int(monitor.DefaultSlowThreshold/time.Millisecond)); ms >= 0 {
		a.slow.SetThreshold(time.Duration(ms) * time.Millisecond)
	}
}

func clampInt(v, lo, hi int) int { return max(lo, min(v, hi)) }

// recordHistory logs an executed statement to the store.
func (a *App) recordHistory(s stmtResult, failed bool) {
	if a.prefs == nil {
		return
	}
	h := store.HistoryEntry{ConnectionID: a.cfg.ID, SQL: s.SQL, RanAt: a.clock(), Duration: s.Dur, Rows: int64(len(s.Res.Rows))}
	if !s.IsRead {
		h.Rows = s.Affected
	}
	if s.Err != nil {
		h.Error = s.Err.Error()
	}
	_ = failed
	_, _ = a.prefs.AddHistory(a.ctx, h)
}

// captureSlow feeds the latency tracker and the slow-query log.
func (a *App) captureSlow(s stmtResult) {
	if s.Err != nil {
		return
	}
	a.lat.Observe(s.Dur)
	n := int64(len(s.Res.Rows))
	if !s.IsRead {
		n = s.Affected
	}
	a.slow.Observe(monitor.SlowQuery{At: a.clock(), Duration: s.Dur, Rows: n, SQL: s.SQL, Source: "quarry"})
}

// ---- connections ----

// openInitial opens the flagged database or the seeded demo database.
func (a *App) openInitial(ctx context.Context, db string) error {
	switch {
	case db == "mock:postgres":
		return a.connect(ctx, driver.ConnectionConfig{
			Name: "Mock Postgres", Engine: driver.EnginePostgres, EnvTag: driver.EnvLocal, Mock: true, Database: "quarry",
		})
	case strings.HasPrefix(db, "postgres://") || strings.HasPrefix(db, "postgresql://"):
		cfg, err := parsePostgresURL(db)
		if err != nil {
			return err
		}
		return a.connect(ctx, cfg)
	case db != "":
		return a.connect(ctx, driver.ConnectionConfig{
			Name: filepath.Base(db), Engine: driver.EngineSQLite, EnvTag: driver.EnvLocal, FilePath: db,
		})
	}
	dir := a.demoDir
	if dir == "" {
		if d, err := store.DefaultDir(); err == nil && a.prefs != nil && a.prefs.Path() != store.Memory {
			dir = d
		} else {
			tmp, terr := os.MkdirTemp("", "quarry-demo-")
			if terr != nil {
				return terr
			}
			dir = tmp
		}
	}
	path, err := driver.EnsureDemoDB(ctx, dir)
	if err != nil {
		return err
	}
	return a.connect(ctx, driver.ConnectionConfig{
		ID: "demo", Name: "Quarry Demo", Engine: driver.EngineSQLite, EnvTag: driver.EnvLocal, FilePath: path,
	})
}

func parsePostgresURL(raw string) (driver.ConnectionConfig, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return driver.ConnectionConfig{}, err
	}
	cfg := driver.ConnectionConfig{
		Name: strings.TrimPrefix(u.Path, "/"), Engine: driver.EnginePostgres, EnvTag: driver.EnvLocal,
		Host: u.Hostname(), Database: strings.TrimPrefix(u.Path, "/"), SSLMode: u.Query().Get("sslmode"),
	}
	if p := u.Port(); p != "" {
		cfg.Port, _ = strconv.Atoi(p)
	}
	if u.User != nil {
		cfg.User = u.User.Username()
		cfg.Password, _ = u.User.Password()
	}
	if cfg.Name == "" {
		cfg.Name = cfg.Host
	}
	return cfg, nil
}

// connect opens cfg, replaces the active driver, and loads the schema tree.
// Prod connections open read-only until the user unlocks writes.
func (a *App) connect(ctx context.Context, cfg driver.ConnectionConfig) error {
	if cfg.EnvTag == driver.EnvProd {
		cfg.ReadOnly = true
	}
	d, err := driver.Open(ctx, cfg)
	if err != nil {
		return err
	}
	a.stopMonitor()
	if a.drv != nil {
		_ = a.drv.Close()
	}
	a.drv, a.cfg, a.connErr, a.prodUnlocked = d, cfg, "", false
	a.version, _ = d.Version(ctx)
	a.refreshSchema(ctx)
	a.lastQuery = lastQueryInfo{}
	for _, t := range a.tabs {
		t.grid, t.plan, t.err, t.cs, t.editing = nil, nil, nil, nil, nil
	}
	a.touchRecent()
	if a.hasMonitorTab() {
		a.startMonitor()
	}
	return nil
}

func (a *App) touchRecent() {
	if a.prefs == nil {
		return
	}
	ctx := a.ctx
	if a.cfg.Engine == driver.EngineSQLite && a.cfg.ID != "demo" && a.cfg.ID == "" {
		if sc, err := a.prefs.RecentSQLiteFile(ctx, a.cfg.FilePath); err == nil {
			a.cfg.ID = sc.ID
		}
	} else if a.cfg.ID != "" && a.cfg.ID != "demo" {
		_ = a.prefs.TouchConnection(ctx, a.cfg.ID)
	}
}

// savedConnections lists persisted profiles (recent first).
func (a *App) savedConnections() []store.SavedConnection {
	if a.prefs == nil {
		return nil
	}
	list, _ := a.prefs.ListConnections(a.ctx)
	return list
}

// refreshSchema reloads the explorer tree and the autocomplete catalog.
func (a *App) refreshSchema(ctx context.Context) {
	if a.drv == nil {
		return
	}
	nodes, err := a.drv.ListSchemas(ctx)
	if err != nil {
		a.connErr = err.Error()
		return
	}
	a.schema = nodes
	var sc sqlkit.Schema
	for _, s := range nodes {
		for _, g := range s.Children {
			if g.Kind != driver.NodeGroup {
				continue
			}
			for _, leaf := range g.Children {
				if leaf.Kind != driver.NodeTable && leaf.Kind != driver.NodeView && leaf.Kind != driver.NodeMatView {
					continue
				}
				cols, _ := a.drv.TableColumns(ctx, leaf.Schema, leaf.Name)
				tbl := sqlkit.Table{Name: leaf.Name}
				if a.cfg.Engine == driver.EnginePostgres && leaf.Schema != "public" {
					tbl.Schema = leaf.Schema
				}
				for _, c := range cols {
					tbl.Columns = append(tbl.Columns, sqlkit.Column{Name: c.Name, Type: c.Type})
				}
				sc.Tables = append(sc.Tables, tbl)
			}
		}
	}
	a.sqlSchema = sc
	if len(a.expanded) == 0 {
		for _, s := range nodes {
			a.expanded["schema:"+s.Name] = true
			a.expanded["group:"+s.Name+":Tables"] = true
		}
	}
}

// reopenWritable reconnects without the read-only lock (Prod unlock).
func (a *App) reopenWritable(ctx context.Context, writable bool) error {
	cfg := a.cfg
	cfg.ReadOnly = !writable
	d, err := driver.Open(ctx, cfg)
	if err != nil {
		return err
	}
	if a.drv != nil {
		_ = a.drv.Close()
	}
	a.drv, a.cfg = d, cfg
	a.prodUnlocked = writable && cfg.EnvTag == driver.EnvProd
	for _, t := range a.tabs {
		if t.kind == tabTable {
			t.writable = writable && len(t.pk) > 0
		}
	}
	return nil
}

func (a *App) connLabel() string {
	if a.drv == nil {
		return "No connection"
	}
	return a.cfg.Name
}
