package page_test

import (
	"context"
	"image"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

func TestDirtyRectClampsToContent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen, err := page.New(page.Config{
		HTML:   `<body style="margin:0"><p id="t" style="margin-top:80px">{{.V}}</p></body>`,
		Width:  320,
		Height: 100,
	})
	if err != nil {
		t.Fatal(err)
	}

	screen.SetData(struct{ V string }{V: "a"})
	if err = screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	screen.TakeDirty()

	screen.SetData(struct{ V string }{V: "b"})
	if err = screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	rect, ok := screen.TakeDirty()
	// Text can extend the content below the viewport. Keep that ink dirty.
	frame := image.Rect(0, 0, screen.Display().Width, screen.Display().Height)
	if !ok || rect.Empty() || !rect.In(frame) {
		t.Fatalf("rect %v ok %v", rect, ok)
	}

	if rect.Max.Y <= 100 {
		t.Fatalf("rect %v lost ink below the viewport", rect)
	}
}
