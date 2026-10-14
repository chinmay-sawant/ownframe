package page_test

import (
	"context"
	"image"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

func TestPartialVsFullBounds(t *testing.T) {
	t.Parallel()

	t.Run("partial", func(t *testing.T) {
		t.Parallel()
		screen, vals := smallChangePage(t)
		vals[0] = "9"
		screen.SetData(vals)
		if err := screen.Redraw(context.Background()); err != nil {
			t.Fatal(err)
		}
		rects, ok := screen.TakeDirtyRects()
		if !ok {
			t.Fatal("no dirty")
		}
		if isFullFrame(rects, mergeW, mergeH) {
			t.Fatalf("small change became full: %v", rects)
		}
	})

	t.Run("full", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		screen, err := page.New(page.Config{
			HTML:  `<body style="margin:0">{{range .}}<div style="width:320px;height:40px;background:{{.C}}">x</div>{{end}}</body>`,
			Width: 320, Height: 400,
		})
		if err != nil {
			t.Fatal(err)
		}
		rows := make([]struct{ C string }, 10)
		for i := range rows {
			rows[i].C = "#eeeeee"
		}
		screen.SetData(rows)
		if err = screen.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
		screen.TakeDirty()
		for i := range rows {
			rows[i].C = "#111111"
		}
		screen.SetData(rows)
		if err = screen.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
		rect, ok := screen.TakeDirty()
		if !ok || rect != image.Rect(0, 0, 320, 400) {
			t.Fatalf("rect %v ok %v", rect, ok)
		}
	})
}
