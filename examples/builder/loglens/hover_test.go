package main

import (
	"bytes"
	"context"
	"fmt"
	"slices"
	"testing"
)

func TestSidebarHoverPreservesPaintContext(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	a.page.Window().SetTransitions(false)
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	for _, action := range []string{"nav:overview", "nav:logs", "nav:overview"} {
		var x, y float64
		found := false
		for _, box := range a.page.Boxes() {
			if box.Action == action {
				x, y, found = box.X+box.W/2, box.Y+box.H/2, true
				break
			}
		}
		if !found {
			t.Fatalf("sidebar control %q disappeared from the hit boxes", action)
		}
		if err := a.page.Edit().Hover(ctx, x, y); err != nil {
			t.Fatal(err)
		}
		patched := sidebarPaintContext(a)
		patchedOrder := sidebarPaintOrder(a)
		patchedPNG := a.page.Probe().PNG()
		if err := a.page.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
		full := sidebarPaintContext(a)
		if !slices.Equal(patchedOrder, sidebarPaintOrder(a)) {
			t.Fatalf("sidebar paint order after %s differs:\npatched=%v\nfull=%v", action, patchedOrder, sidebarPaintOrder(a))
		}
		if !bytes.Equal(patchedPNG, a.page.Probe().PNG()) {
			t.Fatalf("hover %s differs from a full render", action)
		}
		for text, want := range full {
			if got, ok := patched[text]; !ok || got != want {
				t.Fatalf("%s after hover %s: patched=%+v full=%+v", text, action, got, want)
			}
		}
		if len(patched) != 2 || len(full) != 2 {
			t.Fatalf("missing sidebar text: patched=%v full=%v", patched, full)
		}
	}
}

func sidebarPaintOrder(a *app) []string {
	order := []string{}
	d := a.page.Probe().Display()
	for _, i := range d.Order {
		op := d.Ops[i]
		if op.Fixed && op.X < 180 && op.Y >= 45 && op.Y < 270 {
			order = append(order, fmt.Sprintf("%v %q %.2f,%.2f %.2fx%.2f %.2f,%.2f,%.2f,%.2f", op.Kind, op.Text, op.X, op.Y, op.W, op.H, op.R, op.G, op.B, op.Alpha))
		}
	}
	return order
}

type sidebarContext struct {
	fixed bool
	zSet  bool
	z     int
}

func sidebarPaintContext(a *app) map[string]sidebarContext {
	contexts := map[string]sidebarContext{}
	for _, op := range a.page.Probe().Display().Ops {
		if op.X < 180 && (op.Text == "Overview" || op.Text == "Logs") {
			contexts[op.Text] = sidebarContext{fixed: op.Fixed, zSet: op.ZIndexSet, z: op.ZIndex}
		}
	}
	return contexts
}
