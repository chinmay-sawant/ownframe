package page_test

import (
	"context"
	"image"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

// A removed op must still dirty its old bounds, even when the same redraw
// also changes another op and the change stays partial.
func TestDiffRemovedOpKeepsOldBounds(t *testing.T) {
	t.Parallel()

	screen, err := page.New(page.Config{
		HTML:  `<body style="margin:0">{{if .Show}}<div id="gone" style="width:60px;height:30px;background:#f00">x</div>{{end}}<p id="stay">{{.V}}</p></body>`,
		Width: mergeW, Height: mergeH,
	})
	if err != nil {
		t.Fatal(err)
	}

	screen.SetData(struct {
		Show bool
		V    string
	}{true, "a"})
	if err = screen.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	screen.TakeDirtyRects()

	gone := boxFor(t, screen, "gone")

	screen.SetData(struct {
		Show bool
		V    string
	}{false, "b"})
	if err = screen.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	rect, ok := screen.TakeDirty()
	if !ok || !boxIn(rect, gone) {
		t.Fatalf("rect %v ok %v does not cover %v", rect, ok, gone)
	}

	if rect == image.Rect(0, 0, mergeW, mergeH) {
		t.Fatal("removal forced the frame")
	}
}
