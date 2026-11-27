package ui

import (
	"context"
	"fmt"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/driver"
	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/monitor"
)

// This file is the small driving surface the root package's pixel and
// benchmark tests use. It adds no behavior the UI itself does not have.

// Settle waits for the in-flight query to finish and applies it.
func (a *App) Settle(ctx context.Context) error {
	deadline := time.Now().Add(kitSettleWait)
	for a.running != nil {
		if time.Now().After(deadline) {
			return fmt.Errorf("quarry: query did not finish")
		}
		time.Sleep(kitPoll)
		if err := a.tickHook(ctx); err != nil {
			return err
		}
	}
	return a.tickHook(ctx)
}

// RunSQL puts sql in the active query tab, runs it, and waits.
func (a *App) RunSQL(ctx context.Context, sql string) error {
	t := a.activeTabPtr()
	if t == nil || t.kind != tabQuery {
		t = a.newQueryTab("")
	}
	t.ed = newEditor(sql)
	if err := a.runFromKey(ctx, false); err != nil {
		return err
	}
	return a.Settle(ctx)
}

// OpenTableNamed opens a table's data tab and waits for the first page.
func (a *App) OpenTableNamed(ctx context.Context, name string) error {
	if err := a.openByName(ctx, name, false); err != nil {
		return err
	}
	return a.Settle(ctx)
}

// OpenMonitor opens the Monitor tab and waits for the first sample.
func (a *App) OpenMonitor(ctx context.Context) error {
	if err := a.railClick(ctx, "monitor"); err != nil {
		return err
	}
	deadline := time.Now().Add(kitSampleWait)
	for {
		_ = a.tickHook(ctx)
		if _, ok := a.latestSample(); ok {
			return a.redraw(ctx)
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("quarry: no monitor sample")
		}
		time.Sleep(kitSamplePoll)
	}
}

// ToggleTheme switches between QuarryDark and QuarryLight.
func (a *App) ToggleTheme(ctx context.Context) error { return a.toggleTheme(ctx) }

// LoadSyntheticRows replaces the active tab's grid with n generated rows.
func (a *App) LoadSyntheticRows(ctx context.Context, n int) error {
	t := a.activeTabPtr()
	cols := []driver.ColumnMeta{{Name: "id", Type: "INTEGER", PrimaryKey: true}, {Name: "name", Type: "TEXT"},
		{Name: "score", Type: "REAL"}, {Name: "active", Type: "BOOLEAN"}, {Name: "at", Type: "TEXT"}}
	rows := make([][]driver.CellValue, n)
	for i := range rows {
		rows[i] = []driver.CellValue{driver.Int(int64(i + 1)), driver.Text(fmt.Sprintf("row %d", i+1)),
			driver.Float(float64(i) / kitScoreDiv), driver.Bool(i%2 == 0), driver.Text("2025-01-01 00:00:00")}
	}
	t.grid = newGridState(driver.QueryResult{Columns: cols, Rows: rows})
	return a.redraw(ctx)
}

// ScrollGridBy scrolls the active grid by px and repaints.
func (a *App) ScrollGridBy(ctx context.Context, px float64) error {
	t := a.activeTabPtr()
	if t == nil || t.grid == nil {
		return nil
	}
	t.grid.scrollY += px
	t.grid.clampScroll(a.gridViewW(t), a.gridViewH(t), extraRows(t))
	return a.redraw(ctx)
}

// PushSyntheticSample feeds one monitor sample, as a drained collector would.
func (a *App) PushSyntheticSample(i int) {
	if a.mon.ring == nil {
		a.mon.ring = monitor.NewRing[sample](monitorHistory)
	}
	smp := sample{At: time.Unix(int64(i), 0), TPS: float64(kitTPSBase + i%kitTPSSpan), Active: float64(kitMinConns + i%kitConnSpan), Idle: kitIdle, HitPct: kitHit, P50ms: kitP50, P95ms: kitP95}
	// Carry the last real snapshot so size and storage stay populated.
	if last, ok := a.mon.ring.Last(); ok && last.HasSnapshot {
		smp.SizeBytes, smp.Snapshot, smp.HasSnapshot, smp.JournalWAL = last.SizeBytes, last.Snapshot, true, last.JournalWAL
	}
	a.mon.ring.Push(smp)
}

// RedrawMonitor rebuilds the tree with the current samples.
func (a *App) RedrawMonitor(ctx context.Context) error { return a.redraw(ctx) }

// ScrollGridColumns moves the active grid n columns sideways (negative goes
// back) and repaints.
func (a *App) ScrollGridColumns(ctx context.Context, n int) error {
	t := a.activeTabPtr()
	if t == nil || t.grid == nil {
		return nil
	}
	t.grid.hcol = max(0, min(t.grid.hcol+n, a.maxHcol(t, t.grid)))
	return a.redraw(ctx)
}

// SortGridBy cycles the sort on 1-based column col of the active result.
func (a *App) SortGridBy(ctx context.Context, col int) error {
	return a.cycleSort(ctx, a.activeTabPtr(), col)
}
