package platform_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/platform/platform"
)

// BenchmarkClickCount times one #inc click through the handler, the template
// execute, the parse cache, the relayout, and the dirty rect.
func BenchmarkClickCount(b *testing.B) {
	ctx := context.Background()

	app, err := platform.New()
	if err != nil {
		b.Fatal(err)
	}

	if err := app.Page().Redraw(ctx); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		x, y := incCenter(b, app)

		if err := app.Page().Click(ctx, x, y); err != nil {
			b.Fatal(err)
		}
	}
}

// incCenter returns the center of the #inc box for the current layout.
func incCenter(b *testing.B, app *platform.App) (float64, float64) {
	b.Helper()

	for _, box := range app.Page().Boxes() {
		if box.ID == "inc" {
			return box.X + box.W/2, box.Y + box.H/2
		}
	}

	b.Fatal("#inc not found")

	return 0, 0
}
