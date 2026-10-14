package page_test

import (
	"context"
	"image"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

const mergeW, mergeH = 320, 200

// smallChangePage builds the counter page the diff tests mutate.
func smallChangePage(t *testing.T) (*page.Page, []string) {
	t.Helper()

	screen, err := page.New(page.Config{
		HTML:  `<style>div{height:12px;font-size:10px;line-height:12px}</style><body style="margin:0">{{range $i, $v := .}}<div id="c{{$i}}">{{$v}}</div>{{end}}</body>`,
		Width: mergeW, Height: mergeH, Perf: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	vals := make([]string, 20)
	for i := range vals {
		vals[i] = "0"
	}

	screen.SetData(vals)
	if err = screen.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	screen.TakeDirtyRects()

	return screen, vals
}

// rectUnion covers every rect.
func rectUnion(rects []image.Rectangle) image.Rectangle {
	r := rects[0]
	for _, b := range rects[1:] {
		r = r.Union(b)
	}

	return r
}

// isFullFrame reports a take of exactly the whole frame.
func isFullFrame(rects []image.Rectangle, w, h int) bool {
	return len(rects) == 1 && rects[0] == image.Rect(0, 0, w, h)
}

// Twenty small changes in one column must stay partial instead of falling
// back to the frame on an op count.
func TestDiffManySmallChangesStaysPartial(t *testing.T) {
	t.Parallel()

	screen, vals := smallChangePage(t)

	for i := range vals {
		vals[i] = "12345"
	}

	screen.SetData(vals)
	if err := screen.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	if got := screen.Stats().ChangedOps; got != len(vals) {
		t.Fatalf("changed %d, want %d", got, len(vals))
	}

	rects, ok := screen.TakeDirtyRects()
	if !ok || len(rects) == 0 || len(rects) > 8 {
		t.Fatalf("rects %v ok %v", rects, ok)
	}

	union := rectUnion(rects)
	if union == image.Rect(0, 0, mergeW, mergeH) {
		t.Fatalf("fell back to the frame: %v", rects)
	}

	if union.Dx()*union.Dy() > mergeW*mergeH/3 {
		t.Fatalf("union %v over a third of the frame", union)
	}
}
