package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// TestGutterNumbersPaint checks the line-number gutter paints its numbers.
// The .ln rule once carried z-index 2, which the flattened z-band model
// drops below the band-10 editor fills, so the row background covered every
// number. The gutter must show ink in both wrap modes.
func TestGutterNumbersPaint(t *testing.T) {
	for _, wrap := range []bool{true, false} {
		app := newTestApp(t)
		app.openFile("backend/probe.go")

		if app.wrap != wrap {
			app.toggleWrap()
		}

		if err := app.Page().Redraw(context.Background()); err != nil {
			t.Fatal(err)
		}

		var target ownframe.Box
		found := false

		for _, box := range app.Page().Boxes() {
			if len(box.ID) > 3 && box.ID[:3] == "ln-" && box.Text != "" {
				target = box
				found = true

				break
			}
		}

		if !found {
			t.Fatal("no numbered gutter box")
		}

		img := mustDecodePNG(t, app.Page().Probe().PNG())
		bgR, bgG, bgB, _ := img.At(int(target.X)+1, int(target.Y)+1).RGBA()
		ink := 0

		for y := int(target.Y); y < int(target.Y+target.H); y++ {
			for x := int(target.X); x < int(target.X+target.W); x++ {
				r, g, b, _ := img.At(x, y).RGBA()
				dr := int(r>>8) - int(bgR>>8)
				dg := int(g>>8) - int(bgG>>8)
				db := int(b>>8) - int(bgB>>8)

				if dr < 0 {
					dr = -dr
				}

				if dg < 0 {
					dg = -dg
				}

				if db < 0 {
					db = -db
				}

				if dr+dg+db > 30 {
					ink++
				}
			}
		}

		if ink < 5 {
			t.Fatalf("wrap=%v gutter %s shows %d ink pixels, want numbers", wrap, target.ID, ink)
		}
	}
}
