package page

import (
	"context"
	"testing"
)

// TestFixedBoxKeepsInnerControl checks that a control inside a fixed box
// still wins, because the engine paints the whole subtree of a fixed
// element with it.
func TestFixedBoxKeepsInnerControl(t *testing.T) {
	ctx := context.Background()
	p, err := New(Config{
		HTML: `<div id="head" style="position:fixed;width:100px;height:50px;` +
			`background:#c00"><button id="go" style="width:40px;height:20px">go</button></div>` +
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

	spot := p.boxAt(10, 10)
	if !spot.ok || spot.id != "go" {
		t.Fatalf("boxAt = %+v, want the control inside the fixed box", spot)
	}
}
