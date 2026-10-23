package draw

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestPaintBitmap(t *testing.T) {
	t.Parallel()

	var paintTime time.Duration
	got, err := PaintBitmap(false, &paintTime, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.Image != nil || got.Changed != 0 || got.Added != 0 || got.Regions != 0 || got.Full {
		t.Fatalf("nil display %#v", got)
	}
	if paintTime != 0 {
		t.Fatalf("paintTime %v", paintTime)
	}

	if _, err = PaintBitmap(true, nil, nil); err != nil {
		t.Fatal(err)
	}

	display := &layout.Display{
		Width:  2,
		Height: 2,
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpNoop},
			{Kind: layout.DisplayOpNoop},
		},
	}
	got, err = PaintBitmap(false, &paintTime, display)
	if err != nil {
		t.Fatal(err)
	}
	if got.Image == nil || got.Changed != 2 || got.Added != 2 || got.Regions != 1 || !got.Full {
		t.Fatalf("counts %#v", got)
	}
	if paintTime != 0 {
		t.Fatalf("paintTime %v", paintTime)
	}

	bad := &layout.Display{
		Width:  10,
		Height: 10,
		Ops:    []layout.DisplayOp{{Kind: layout.DisplayKind(100)}},
		Order:  []int{0},
	}
	got, err = PaintBitmap(true, &paintTime, bad)
	if err == nil {
		t.Fatal("picture error")
	}
	if got.Image != nil || got.Changed != 0 || got.Added != 0 || got.Regions != 0 || got.Full {
		t.Fatalf("error bitmap %#v", got)
	}
	if paintTime != 0 {
		t.Fatalf("paintTime %v", paintTime)
	}
}
