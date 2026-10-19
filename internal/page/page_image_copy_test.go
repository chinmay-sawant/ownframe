package page_test

import (
	"bytes"
	"context"
	"reflect"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

// TestPNGReturnsCopy mutates the returned bytes and requires the next PNG
// to still decode the same picture.
func TestPNGReturnsCopy(t *testing.T) {
	t.Parallel()

	screen, err := page.New(page.Config{HTML: `<div id="a">copy</div>`, Width: 120, Height: 80})
	if err != nil {
		t.Fatal(err)
	}

	if err := screen.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	first := screen.PNG()
	if first == nil {
		t.Fatal("no PNG to copy")
	}

	want := append([]byte(nil), first...)
	for i := range first {
		first[i] ^= 0xff
	}

	if got := screen.PNG(); !bytes.Equal(got, want) {
		t.Fatal("PNG cache changed after caller edit")
	}
}

// TestBoxesReturnsCopy mutates the returned slice and requires the page
// boxes to stay unchanged.
func TestBoxesReturnsCopy(t *testing.T) {
	t.Parallel()

	screen, err := page.New(page.Config{HTML: `<div id="a">copy</div>`, Width: 120, Height: 80})
	if err != nil {
		t.Fatal(err)
	}

	if err := screen.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	boxes := screen.Boxes()
	if len(boxes) == 0 {
		t.Fatal("no boxes to copy")
	}

	want := boxes[0]
	boxes[0].X = want.X + 10000
	boxes[0].W = want.W + 10000

	if got := screen.Boxes(); !reflect.DeepEqual(got[0], want) {
		t.Fatalf("boxes changed after caller edit: %+v, want %+v", got[0], want)
	}
}
