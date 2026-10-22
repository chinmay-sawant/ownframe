package page

import (
	"context"
	"testing"
)

// TestFallbackRebuildsFixedHits checks that the bitmap path builds the hit
// map as well, so a fixed box still wins over the content it covers.
func TestFallbackRebuildsFixedHits(t *testing.T) {
	ctx := context.Background()
	p, err := New(Config{
		HTML: `<style>div { border-radius: 50% / 20%; }</style>` +
			`<div id="head" style="position:fixed;width:100px;height:50px;background:#c00"></div>` +
			`<div id="body" style="width:100px;height:50px;background:#00c"></div>`,
		Width:  200,
		Height: 200,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if p.draw.Display != nil {
		t.Fatal("the elliptical radius replayed, want the bitmap path")
	}
	if len(p.hits) == 0 {
		t.Fatal("the bitmap path left no hit map")
	}

	spot := p.boxAt(10, 10)
	if !spot.ok || spot.id != "head" {
		t.Fatalf("boxAt = %+v, want the fixed header", spot)
	}
}
