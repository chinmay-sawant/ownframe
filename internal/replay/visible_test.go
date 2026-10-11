package replay

import (
	"image"
	"reflect"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestVisiblePreservesOrderAndUnboundedInk(t *testing.T) {
	d := &layout.Display{
		PointsPerPixel: 0.75,
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 20, H: 20},
			{Kind: layout.DisplayOpFillRect, X: 0, Y: 1500, W: 20, H: 20},
			{Kind: layout.DisplayOpFillRect, X: 0, Y: 750, W: 20, H: 20},
			{Kind: layout.DisplayOpText, Y: 9000, RotateDeg: 20},
		},
		Order: []int{3, 2, 1, 0, -1, 99},
	}
	rect := visibleRect(image.Rect(0, 0, 100, 100), 0, -1000)
	got := []int{}
	visitVisible(d, rect, func(i int) { got = append(got, i) })
	if !reflect.DeepEqual(got, []int{3, 2}) {
		t.Fatalf("visible paint order = %v, want [3 2]", got)
	}
}

func TestVisibleRectRoundsFractionalScrollOutward(t *testing.T) {
	got := visibleRect(image.Rect(5, 10, 105, 110), -0.5, -20.25)
	if want := image.Rect(5, 30, 106, 131); got != want {
		t.Fatalf("rect = %v, want %v", got, want)
	}
}
