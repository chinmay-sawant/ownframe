package ui

import (
	"context"
	"fmt"
	"sync"
)

// fingerprintTable remembers the last successfully applied refresh
// fingerprint per View. View's struct lives in mount.go and gains no fields
// here, so the table is keyed by View instead.
type fingerprintTable struct {
	mu     sync.Mutex
	byView map[*View]string
}

func newFingerprintTable() *fingerprintTable {
	return &fingerprintTable{byView: make(map[*View]string)}
}

func (t *fingerprintTable) get(v *View) string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.byView[v]
}

func (t *fingerprintTable) set(v *View, fingerprint string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.byView[v] = fingerprint
}

// appliedFingerprints is the refresh memory for UpdateChanged: one entry per
// View holding the fingerprint of its last successful refresh. Entries live
// as long as their View keeps refreshing; Views are per-page and long-lived,
// so the table stays at one small entry per mounted view.
var appliedFingerprints = newFingerprintTable() //nolint:gochecknoglobals // per-View refresh memory; View's fields are frozen in mount.go, so the last-applied fingerprint is kept in this table rather than a new field.

// UpdateChanged refreshes the view only when fingerprint describes state the
// page does not show yet.
//
// The caller derives fingerprint from every piece of state the built tree
// reads (for example the selected row id plus the active filter). When
// fingerprint equals the last fingerprint this view applied successfully,
// UpdateChanged returns nil without calling build, rendering, diffing, or
// touching the page, so repeating a selection costs nothing. Any other
// fingerprint builds a fresh tree and delegates to Update; the fingerprint
// is recorded only when Update succeeds, so a failed refresh retries on the
// next call instead of pinning a tree the page never showed.
//
// An empty fingerprint never matches: it always rebuilds and is never
// recorded, so callers with nothing to hash cannot accidentally skip the
// first refresh. A cancelled context returns its error and records nothing,
// both before building and before touching the page.
//
// Like Update, call UpdateChanged on the page loop.
//
// Usage sketch (a selection handler; re-clicking the row below is free):
//
//	func (a *App) handleSelect(ctx context.Context, id string) error {
//	    // Re-clicking the selected row keeps fp identical,
//	    // so the repeat below builds and renders nothing.
//	    a.state.selected = id
//	    fp := "sel=" + a.state.selected
//	    fp += "&filter=" + a.state.filter
//	    return a.view.UpdateChanged(ctx, fp, func() *Element {
//	        return renderCompareTable(a.state)
//	    })
//	}
func (v *View) UpdateChanged(ctx context.Context, fingerprint string, build func() *Element) error {
	if ctx != nil && ctx.Err() != nil {
		return ctx.Err()
	}
	if fingerprint != "" && appliedFingerprints.get(v) == fingerprint {
		// True zero-cost repeat: build never runs and the page is untouched.
		return nil
	}
	if build == nil {
		return fmt.Errorf("ui: UpdateChanged requires a non-nil build function")
	}
	root := build()
	if ctx != nil && ctx.Err() != nil {
		return ctx.Err()
	}
	if err := v.Update(ctx, root); err != nil {
		return err
	}
	if fingerprint != "" {
		appliedFingerprints.set(v, fingerprint)
	}
	return nil
}
