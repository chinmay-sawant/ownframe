package main

import (
	"context"
	"testing"
	"time"
)

// TestScrollKeepsRunningAnimations checks that a vertical scroll keeps the
// automatic driver going. A row that leaves the window is unbound; when it
// comes back the started class binds a new player and the card moves again.
func TestScrollKeepsRunningAnimations(t *testing.T) {
	names := []string{"-webkit-animation", "animation-name", "view-transition-name"}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			p, d, v := fixturePage(t)
			m, ok := measureCatalog(p)
			if !ok {
				t.Fatal("could not measure the catalog")
			}
			if err := installCatalogWindow(ctx, p, v, d, m); err != nil {
				t.Fatal(err)
			}
			row := catalogRowByName(t, v, name)
			index := rowIndex(v, name)
			scrollTo(t, ctx, p, m, m.rowTop[rowAt(m, index/2)]-200)
			clickID(t, p, row.Button)
			tickPage(t, p)
			first := xformX(targetOp(t, p, row.TargetID), p.Probe().Display().PointsPerPixel)

			far := 0
			if index/2 > 2 {
				far = 0
			} else {
				far = m.contentH
			}
			scrollTo(t, ctx, p, m, far)
			tickPage(t, p)
			if _, found := findBox(p.Boxes(), row.TargetID); found {
				t.Fatal("the row stayed in the window")
			}

			scrollTo(t, ctx, p, m, m.rowTop[rowAt(m, index/2)]-200)
			time.Sleep(5 * time.Millisecond)
			tickPage(t, p)
			if !v.Started[row.Button] {
				t.Fatal("the scroll lost the started state")
			}
			second := xformX(targetOp(t, p, row.TargetID), p.Probe().Display().PointsPerPixel)
			if second == 0 {
				t.Fatalf("%s did not start again after the scroll (first %v)", name, first)
			}
		})
	}
}

// rowIndex returns the catalog index of a property.
func rowIndex(v *view, name string) int {
	for i, row := range v.Catalog {
		if row.Name == name {
			return i
		}
	}
	return 0
}
