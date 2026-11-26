package store

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/driver"
)

func open(t *testing.T) *Store {
	t.Helper()
	s, err := Open(Memory)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func pg(name string) SavedConnection {
	return SavedConnection{ConnectionConfig: driver.ConnectionConfig{
		Name: name, Engine: driver.EnginePostgres, EnvTag: driver.EnvProd, ReadOnly: true,
		Host: "db.internal", Port: 5432, Database: "app", User: "me", Password: "hunter2", SSLMode: "require",
	}}
}

func TestConnectionCRUD(t *testing.T) {
	ctx := context.Background()
	s := open(t)
	c, err := s.SaveConnection(ctx, pg("Prod DB"))
	if err != nil || c.ID == "" || c.Password != "" {
		t.Fatalf("save = %+v err=%v", c, err)
	}
	got, err := s.GetConnection(ctx, c.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Password != "" || got.Host != "db.internal" || got.Port != 5432 || !got.ReadOnly || got.EnvTag != "Prod" || got.SSLMode != "require" || got.Engine != driver.EnginePostgres {
		t.Errorf("round trip = %+v", got)
	}

	// opt-in password persistence
	keep := pg("Keeps pw")
	keep.SavePassword = true
	kc, _ := s.SaveConnection(ctx, keep)
	if g, _ := s.GetConnection(ctx, kc.ID); g.Password != "hunter2" || !g.SavePassword {
		t.Errorf("opted-in password lost: %+v", g)
	}

	// update in place
	got.Name = "Renamed"
	got.ReadOnly = false
	if _, err := s.SaveConnection(ctx, got); err != nil {
		t.Fatal(err)
	}
	list, _ := s.ListConnections(ctx)
	if len(list) != 2 {
		t.Fatalf("list = %d", len(list))
	}
	if g, _ := s.GetConnection(ctx, c.ID); g.Name != "Renamed" || g.ReadOnly {
		t.Errorf("update = %+v", g)
	}

	// validation
	for _, bad := range []SavedConnection{
		{},
		{ConnectionConfig: driver.ConnectionConfig{Name: "x", Engine: driver.EngineSQLite}},
		{ConnectionConfig: driver.ConnectionConfig{Name: "x", Engine: driver.EnginePostgres, Host: "h"}},
		{ConnectionConfig: driver.ConnectionConfig{Name: "x", Engine: "oracle"}},
	} {
		if _, err := s.SaveConnection(ctx, bad); err == nil {
			t.Errorf("expected validation error for %+v", bad.ConnectionConfig)
		}
	}
	mock, err := s.SaveConnection(ctx, SavedConnection{ConnectionConfig: driver.ConnectionConfig{Name: "Mock", Engine: driver.EnginePostgres, Mock: true}})
	if err != nil {
		t.Fatal(err)
	}
	if g, _ := s.GetConnection(ctx, mock.ID); !g.Mock || g.EnvTag != driver.EnvLocal {
		t.Errorf("mock = %+v", g)
	}

	if err := s.DeleteConnection(ctx, c.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetConnection(ctx, c.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("after delete err = %v", err)
	}
	if err := s.DeleteConnection(ctx, c.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("double delete err = %v", err)
	}
}

func TestRecentSQLiteFile(t *testing.T) {
	ctx := context.Background()
	s := open(t)
	a, err := s.RecentSQLiteFile(ctx, "/data/shop.db")
	if err != nil || a.Name != "shop.db" || a.LastUsedAt.IsZero() {
		t.Fatalf("recent = %+v err=%v", a, err)
	}
	b, _ := s.RecentSQLiteFile(ctx, "/data/other.sqlite")
	again, _ := s.RecentSQLiteFile(ctx, "/data/shop.db")
	if again.ID != a.ID {
		t.Error("same path should reuse the profile")
	}
	list, _ := s.ListConnections(ctx)
	if len(list) != 2 || list[0].ID != a.ID || list[1].ID != b.ID {
		t.Errorf("recent order = %+v", list)
	}
	if _, err := s.RecentSQLiteFile(ctx, "/data/other.sqlite"); err != nil {
		t.Fatal(err)
	}
	if list, _ = s.ListConnections(ctx); list[0].ID != b.ID {
		t.Error("touching should move to front")
	}
}

func TestHistorySnippetsPrefs(t *testing.T) {
	ctx := context.Background()
	s := open(t)
	for i := 0; i < 5; i++ {
		conn := "a"
		if i%2 == 1 {
			conn = "b"
		}
		if _, err := s.AddHistory(ctx, HistoryEntry{ConnectionID: conn, SQL: "SELECT " + string(rune('0'+i)), Duration: time.Millisecond * time.Duration(i), Rows: int64(i)}); err != nil {
			t.Fatal(err)
		}
	}
	all, _ := s.ListHistory(ctx, "", 0)
	if len(all) != 5 || all[0].SQL != "SELECT 4" {
		t.Fatalf("all = %+v", all)
	}
	onlyB, _ := s.ListHistory(ctx, "b", 10)
	if len(onlyB) != 2 || onlyB[0].Rows != 3 || onlyB[0].Duration != 3*time.Millisecond {
		t.Errorf("b = %+v", onlyB)
	}
	if lim, _ := s.ListHistory(ctx, "", 2); len(lim) != 2 {
		t.Error("limit ignored")
	}
	_ = s.ClearHistory(ctx, "b")
	if rest, _ := s.ListHistory(ctx, "", 0); len(rest) != 3 {
		t.Errorf("after clear b = %d", len(rest))
	}
	_ = s.ClearHistory(ctx, "")
	if rest, _ := s.ListHistory(ctx, "", 0); len(rest) != 0 {
		t.Error("clear all failed")
	}

	sn, err := s.SaveSnippet(ctx, Snippet{Name: "recent orders", SQL: "SELECT * FROM orders LIMIT 10"})
	if err != nil || sn.ID == 0 {
		t.Fatal(err)
	}
	sn.Name = "latest orders"
	if _, err := s.SaveSnippet(ctx, sn); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SaveSnippet(ctx, Snippet{Name: "", SQL: "x"}); err == nil {
		t.Error("empty name should fail")
	}
	if _, err := s.SaveSnippet(ctx, Snippet{ID: 999, Name: "n", SQL: "x"}); !errors.Is(err, ErrNotFound) {
		t.Errorf("update missing err = %v", err)
	}
	sl, _ := s.ListSnippets(ctx)
	if len(sl) != 1 || sl[0].Name != "latest orders" {
		t.Errorf("snippets = %+v", sl)
	}
	if err := s.DeleteSnippet(ctx, sn.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteSnippet(ctx, sn.ID); !errors.Is(err, ErrNotFound) {
		t.Error("double delete")
	}

	if s.Theme(ctx) != "dark" {
		t.Error("default theme must be dark")
	}
	_ = s.SetTheme(ctx, false)
	if s.Theme(ctx) != "light" {
		t.Error("light not stored")
	}
	_ = s.SetPref(ctx, "sidebar_w", "300")
	_ = s.SetPref(ctx, "sidebar_w", "320")
	if s.GetPref(ctx, "sidebar_w", "") != "320" || s.GetPref(ctx, "missing", "d") != "d" {
		t.Error("prefs")
	}
}

func TestHistoryPrunes(t *testing.T) {
	ctx := context.Background()
	s := open(t)
	for i := 0; i < MaxHistory+25; i++ {
		if _, err := s.AddHistory(ctx, HistoryEntry{ConnectionID: "a", SQL: "SELECT 1"}); err != nil {
			t.Fatal(err)
		}
	}
	h, _ := s.ListHistory(ctx, "", MaxHistory*2)
	if len(h) != MaxHistory {
		t.Errorf("history rows = %d, want %d", len(h), MaxHistory)
	}
}

func TestPersistenceAcrossReopenAndMemoryIsolation(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	c, _ := s.SaveConnection(ctx, pg("Persisted"))
	_ = s.SetTheme(ctx, false)
	_, _ = s.AddHistory(ctx, HistoryEntry{ConnectionID: c.ID, SQL: "SELECT 42"})
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s2, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s2.Close() }()
	if g, err := s2.GetConnection(ctx, c.ID); err != nil || g.Name != "Persisted" {
		t.Errorf("reopen conn = %+v err=%v", g, err)
	}
	if s2.Theme(ctx) != "light" {
		t.Error("theme not persisted")
	}
	if h, _ := s2.ListHistory(ctx, c.ID, 5); len(h) != 1 || h[0].SQL != "SELECT 42" {
		t.Errorf("history = %+v", h)
	}

	m1, m2 := open(t), open(t)
	_, _ = m1.SaveConnection(ctx, pg("only in m1"))
	if l, _ := m2.ListConnections(ctx); len(l) != 0 {
		t.Error(":memory: stores must be isolated")
	}
}
