package replay

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

// TestXformGeomMapsOpOrigin pins the matrix the transformed painter uses: a
// local origin maps to the op's transformed layout origin in canvas pixels.
func TestXformGeomMapsOpOrigin(t *testing.T) {
	t.Parallel()

	display, err := render.DisplayList(context.Background(),
		`<div style="width:60px;height:20px;background:#eee;transform:rotate(30deg)"></div>`,
		200, 120)
	if err != nil {
		t.Fatal(err)
	}

	for i := range display.Ops {
		op := &display.Ops[i]
		if op.Kind != layout.DisplayOpFillRect || !op.XformSet {
			continue
		}

		m := op.Transform()
		wantX, wantY := m.Apply(op.X, op.Y)
		geom := xformGeom(op, 1, 2)
		gotX, gotY := geom.Apply(0, 0)

		if !near(gotX, wantX*pxPerPt+1) || !near(gotY, wantY*pxPerPt+2) {
			t.Fatalf("origin = (%v,%v), want (%v,%v)", gotX, gotY, wantX*pxPerPt+1, wantY*pxPerPt+2)
		}

		return
	}

	t.Fatal("no transformed fill")
}

func near(a, b float64) bool {
	d := a - b

	return d < 1e-9 && d > -1e-9
}
