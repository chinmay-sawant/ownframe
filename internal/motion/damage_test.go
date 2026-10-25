package motion

import (
	"image"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page/dirty"
)

func TestDamageCoalescesNeighbours(t *testing.T) {
	var d Damage

	d.Add(image.Rect(0, 0, 10, 10))
	d.Add(image.Rect(12, 0, 22, 10))

	if got := d.N; got != 1 {
		t.Fatalf("rect count = %d, want 1 after coalescing", got)
	}

	if got, want := d.Bounds(), image.Rect(0, 0, 22, 10); got != want {
		t.Fatalf("bounds = %v, want %v", got, want)
	}
}

func TestDamageOverflowCollapsesToBounds(t *testing.T) {
	var d Damage

	for i := range dirty.MaxRects + 1 {
		x := i * 100
		d.Add(image.Rect(x, 0, x+10, 10))
	}

	if !d.Overflow {
		t.Fatal("damage past the cap did not report overflow")
	}

	if got := d.N; got != 1 {
		t.Fatalf("rect count = %d, want the collapsed bounds", got)
	}

	if got, want := d.Bounds(), image.Rect(0, 0, dirty.MaxRects*100+10, 10); got != want {
		t.Fatalf("bounds = %v, want %v", got, want)
	}
}

func TestDamageIgnoresEmptyRects(t *testing.T) {
	var d Damage

	d.Add(image.Rectangle{})
	if !d.Empty() {
		t.Fatalf("empty rect produced damage %v", d.Bounds())
	}
}
