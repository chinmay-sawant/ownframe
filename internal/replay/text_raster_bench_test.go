package replay

import (
	"context"
	"fmt"
	"image"
	"strings"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

// textBenchDisplay builds a text-heavy page: 48 short label rows, the shape a
// scrolled list or a card grid repaints.
func textBenchDisplay(b *testing.B) *layout.Display {
	var sb strings.Builder
	sb.WriteString(`<html><body style="margin:0;font-size:13px">`)
	for i := range 48 {
		fmt.Fprintf(&sb, `<p style="margin:2px">Row %02d service west-1 healthy 12 ms</p>`, i)
	}
	sb.WriteString(`</body></html>`)

	d, err := render.DisplayList(context.Background(), sb.String(), 520, 900)
	if err != nil {
		b.Fatal(err)
	}

	return d
}

// BenchmarkTextDamageRedraw times a damage repaint over text rows: the same
// display list drawn again, so the second draw finds the runs already baked.
// warm redraws at one position, miss flushes the cache first, and scroll
// moves the page one whole pixel per step, which the key's sub-pixel grid has
// to absorb.
func BenchmarkTextDamageRedraw(b *testing.B) {
	d := textBenchDisplay(b)
	dst := ebiten.NewImage(d.Width, d.Height)
	defer dst.Dispose()
	rect := image.Rect(0, 0, d.Width, d.Height)

	b.ReportAllocs()
	b.Run("warm", func(b *testing.B) {
		DrawRect(dst, d, rect, 0, 0)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			DrawRect(dst, d, rect, 0, 0)
		}
	})
	b.Run("miss", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			textRuns.flush()
			DrawRect(dst, d, rect, 0, 0)
		}
	})
	b.Run("scroll", func(b *testing.B) {
		DrawRect(dst, d, rect, 0, 0)

		for i := 0; i < b.N; i++ {
			DrawRect(dst, d, rect, 0, -float64(i%800))
		}
	})
}
