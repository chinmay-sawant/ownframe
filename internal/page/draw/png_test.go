package draw

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

func TestEncodePNGNil(t *testing.T) {
	t.Parallel()

	if EncodePNG(nil) != nil {
		t.Fatal("nil")
	}
}

func TestFallbackDisplayErrorSkipsSource(t *testing.T) {
	t.Parallel()

	f := Frame{Display: &layout.Display{
		Width:  1,
		Height: 1,
		Ops:    []layout.DisplayOp{{Kind: layout.DisplayKind(100)}},
	}}
	if f.Bytes(context.Background(), "<p>x</p>", Viewport{Width: 20, Height: 20}, render.State{}) != nil {
		t.Fatal("source painted")
	}
}

func TestImageBytesCachesPicture(t *testing.T) {
	t.Parallel()

	img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	f := Frame{Img: img}
	b := f.Bytes(context.Background(), "", Viewport{}, render.State{})
	if len(b) == 0 {
		t.Fatal("encode")
	}
	b[0] ^= 0xff
	if next := f.Bytes(context.Background(), "", Viewport{}, render.State{}); len(next) == 0 || next[0] == b[0] {
		t.Fatal("cache")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var empty Frame
	if empty.Bytes(ctx, "<p>x</p>", Viewport{Width: 20, Height: 20}, render.State{}) != nil {
		t.Fatal("canceled")
	}
}

func TestEncodePNGRoundTrip(t *testing.T) {
	t.Parallel()

	img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	b := EncodePNG(img)
	if b == nil {
		t.Fatal("encode")
	}
	if _, err := png.Decode(bytes.NewReader(b)); err != nil {
		t.Fatal(err)
	}
}
