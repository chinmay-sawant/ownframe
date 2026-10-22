package platform_test

import (
	"context"
	"image"
	"testing"
)

func TestClicksRepaintCountOnly(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)
	app.Page().TakeDirty()

	count := boxByID(t, app, "count")
	want := image.Rect(
		int(count.X)-8, int(count.Y)-8,
		int(count.X+count.W)+8, int(count.Y+count.H)+8,
	)

	for i := 0; i < 2; i++ {
		click(t, ctx, app, "inc")

		rect, ok := app.Page().TakeDirty()
		if !ok {
			t.Fatalf("click %d: no dirty rect", i)
		}

		if !rect.In(want) {
			t.Fatalf("click %d: rect %v outside #count %v", i, rect, want)
		}
	}
}
