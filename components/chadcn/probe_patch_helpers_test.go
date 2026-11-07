// Shared helpers for the created rerender-probe tests (TestProbePatch*).
// These tests prove a chadcn state transition takes the ui.View.Update patch
// path instead of falling back to a full document render.
package chadcn_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

func mountProbeView(t *testing.T, root *ui.Element) (*ui.View, context.Context) {
	t.Helper()
	ctx := context.Background()
	v, err := ui.Mount(ownframe.Config{
		Title: "probe", Width: 960, Height: 720,
		Theme: chadcn.Dark + "\n" + chadcn.Stylesheet, Perf: true,
	}, root)
	if err != nil {
		t.Fatal(err)
	}
	if err := v.Page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	v.Page.Probe().TakeDirtyRects()
	return v, ctx
}

// expectProbePatch fails when the Update does not commit exactly one
// generation, or when it reparses (full-render fallback).
func expectProbePatch(t *testing.T, v *ui.View, ctx context.Context, next *ui.Element) {
	t.Helper()
	gen0 := v.Page.Probe().Generation()
	s0 := v.Page.Probe().Stats()
	if err := v.Update(ctx, next); err != nil {
		t.Fatal(err)
	}
	s1 := v.Page.Probe().Stats()
	if got, want := v.Page.Probe().Generation(), gen0+1; got != want {
		t.Fatalf("generation = %d, want %d", got, want)
	}
	if s1.Parses != s0.Parses {
		t.Fatalf("fell back to full render: parses %d -> %d", s0.Parses, s1.Parses)
	}
}

// expectProbeNoop fails when an identical rebuild does any page work.
func expectProbeNoop(t *testing.T, v *ui.View, ctx context.Context, next *ui.Element) {
	t.Helper()
	gen0 := v.Page.Probe().Generation()
	s0 := v.Page.Probe().Stats()
	if err := v.Update(ctx, next); err != nil {
		t.Fatal(err)
	}
	s1 := v.Page.Probe().Stats()
	if got := v.Page.Probe().Generation(); got != gen0 {
		t.Fatalf("generation = %d, want %d (expected no-op)", got, gen0)
	}
	if s1.Parses != s0.Parses {
		t.Fatalf("no-op reparsed: parses %d -> %d", s0.Parses, s1.Parses)
	}
}
