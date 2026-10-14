package page_test

import (
	"context"
	"image"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

func TestFullFallbackReport(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen, err := page.New(page.Config{
		HTML:   `<style>div { width: 60px; height: 40px; background: #123456; border-radius: 50% / 20%; }</style><div>{{.V}}</div>`,
		Width:  320,
		Height: 400,
	})
	if err != nil {
		t.Fatal(err)
	}

	screen.SetData(struct{ V string }{V: "a"})
	if err := screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if screen.Display() != nil {
		t.Fatal("elliptical border was replayed")
	}
	if screen.FallbackDisplay() == nil {
		t.Fatal("no fallback display")
	}
	if screen.Image() == nil {
		t.Fatal("no fallback bitmap")
	}
	if !screen.LastPatch().Full {
		t.Fatalf("patch = %+v", screen.LastPatch())
	}
	rect, ok := screen.TakeDirty()
	if !ok || rect != image.Rect(0, 0, 320, 400) {
		t.Fatalf("rect %v ok %v", rect, ok)
	}
}
