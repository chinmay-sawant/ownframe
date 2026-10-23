package draw

import (
	"image"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestCommitListDropsBitmap(t *testing.T) {
	t.Parallel()

	d := &layout.Display{Width: 1, Height: 1}
	var f Frame
	f.Img = image.NewNRGBA(image.Rect(0, 0, 1, 1))
	f.PNG = []byte{1}
	f.FallbackDisplay = d
	f.CommitList(d)
	if f.Display != d || f.Img != nil || f.FallbackDisplay != nil || f.PNG != nil {
		t.Fatalf("list %+v", f)
	}
}

func TestCommitBitmapDropsList(t *testing.T) {
	t.Parallel()

	d := &layout.Display{Width: 1, Height: 1}
	img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	var f Frame
	f.Display = d
	f.PNG = []byte{1}
	f.CommitBitmap(d, img)
	if f.Display != nil || f.Img != img || f.FallbackDisplay != d || f.PNG != nil {
		t.Fatalf("bitmap %+v", f)
	}
}

func TestSwapListKeepsPrevAndDropsPNG(t *testing.T) {
	t.Parallel()

	old := &layout.Display{Width: 1, Height: 1}
	next := &layout.Display{Width: 2, Height: 2}
	f := Frame{Display: old, PNG: []byte{1}}
	if prev := f.SwapList(next); prev != old || f.Display != next || f.PNG != nil {
		t.Fatalf("swap %+v prev %v", f, prev)
	}
}

func TestCopyPNGDoesNotAlias(t *testing.T) {
	t.Parallel()

	f := Frame{PNG: []byte{1, 2, 3}}
	got := f.CopyPNG()
	got[0] = 9
	if f.PNG[0] != 1 {
		t.Fatal("alias")
	}
	if f.CopyPNG() == nil {
		t.Fatal("empty")
	}
	f.DropPNG()
	if f.CopyPNG() != nil {
		t.Fatal("drop")
	}
}
