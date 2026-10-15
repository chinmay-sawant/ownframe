package page_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

// TestSetImageClonesCallerBytes corrupts the caller's slice after SetImage
// and requires the painted background to stay red.
func TestSetImageClonesCallerBytes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen, err := page.New(page.Config{HTML: backgroundHTML, Width: 40, Height: 40})
	if err != nil {
		t.Fatal(err)
	}

	data := redPNG(t)
	screen.SetImage("bg", data)
	for i := range data {
		data[i] = 0
	}

	if err := screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if got := centerPixel(t, screen.PNG()); got.R < 200 || got.G > 60 || got.B > 60 {
		t.Fatalf("center pixel %+v after caller edit, want red", got)
	}
}
