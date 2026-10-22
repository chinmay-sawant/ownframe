package draw

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

func TestPictureEmpty(t *testing.T) {
	t.Parallel()

	img, err := Picture(nil)
	if err != nil || img != nil {
		t.Fatalf("nil %v %v", img, err)
	}
	img, err = Picture(&layout.Display{})
	if err != nil || img != nil {
		t.Fatalf("empty %v %v", img, err)
	}
}

func TestPaintSource(t *testing.T) {
	t.Parallel()

	img, err := PaintSource(context.Background(), "<p>x</p>", 40, 40, render.State{})
	if err != nil {
		t.Fatal(err)
	}
	if img == nil {
		t.Fatal("no picture")
	}
}

func TestPaintSourceCanceled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	img, err := PaintSource(ctx, "<p>x</p>", 40, 40, render.State{})
	if err == nil && img != nil {
		t.Fatal("canceled")
	}
}
