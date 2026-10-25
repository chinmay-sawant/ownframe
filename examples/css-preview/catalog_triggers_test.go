package main

import (
	"math"
	"testing"
	"time"
)

// TestCatalogStartRunsTriggerCards covers the three cards whose effect waits
// for an animation-trigger. The manual path in manual_motion.go dispatches
// after the click's redraw; these cards are the reason that file exists.
func TestCatalogStartRunsTriggerCards(t *testing.T) {
	for _, name := range []string{"animation", "animation-trigger", "trigger-scope"} {
		t.Run(name, func(t *testing.T) {
			p, _, v := fixturePage(t)
			row := catalogRowByName(t, v, name)
			clickID(t, p, row.Button)
			time.Sleep(200 * time.Millisecond)
			tickPage(t, p)
			op := targetOp(t, p, row.TargetID)
			if got := xformX(op, p.Probe().Display().PointsPerPixel); math.Abs(got) < .1 {
				t.Fatalf("%s did not move after Start", name)
			}
		})
	}
}

// catalogRowByName returns the row with the given property name.
func catalogRowByName(t *testing.T, v *view, name string) catalogRow {
	t.Helper()
	for _, row := range v.Catalog {
		if row.Name == name {
			return row
		}
	}
	t.Fatalf("missing property %q", name)
	return catalogRow{}
}
