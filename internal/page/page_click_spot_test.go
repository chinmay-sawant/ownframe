package page

import (
	"context"
	"testing"
)

// TestClickSingleSpotResolvesAnonymousChild pins the one-spot click path: the
// spot scanned for the animation-trigger pass is the spot the sequence
// resolves, so a click on an anonymous child still reaches the nearest
// id-bearing ancestor and its data-action.
func TestClickSingleSpotResolvesAnonymousChild(t *testing.T) {
	ctx := context.Background()
	p := overlapPage(t)
	p.setBoxes([]Box{
		{ID: "card", Action: "open", X: 0, Y: 0, W: 100, H: 100},
		{ID: "", X: 10, Y: 10, W: 20, H: 20},
	})

	spot := p.boxAt(15, 15)
	if !spot.ok || spot.box.ID != "" || spot.id != "card" {
		t.Fatalf("spot = %+v, want the anonymous child under card", spot)
	}

	var clicked Box
	p.Handle(Handlers{Click: func(_ context.Context, b Box) error {
		clicked = b

		return nil
	}})

	if err := p.Click(ctx, 15, 15); err != nil {
		t.Fatal(err)
	}
	if clicked.ID != "card" || clicked.Action != "open" {
		t.Fatalf("clicked = %+v, want the card control through its child", clicked)
	}
}
