package page_test

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

// benchInvalidateRows is the box count for BenchmarkInvalidateMany: enough
// that a per-lookup walk over every box would dominate the loop.
const benchInvalidateRows = 3000

func invalidateFixture(b *testing.B) (*page.Page, []string) {
	b.Helper()

	var src strings.Builder
	src.WriteString(`<style>div{height:12px}</style>`)

	ids := make([]string, 0, benchInvalidateRows)
	for i := 0; i < benchInvalidateRows; i++ {
		id := "row-" + strconv.Itoa(i)
		ids = append(ids, id)
		src.WriteString(`<div id="` + id + `"></div>`)
	}

	screen, err := page.New(page.Config{HTML: src.String(), Width: 640, Height: 480})
	if err != nil {
		b.Fatal(err)
	}

	if err := screen.Redraw(context.Background()); err != nil {
		b.Fatal(err)
	}

	if got := len(screen.Boxes()); got < benchInvalidateRows {
		b.Fatalf("boxes = %d, want at least %d", got, benchInvalidateRows)
	}

	return screen, ids
}

// BenchmarkInvalidateMany marks every id on a page with thousands of boxes
// once per iteration and consumes the dirty state the way a frame does.
func BenchmarkInvalidateMany(b *testing.B) {
	screen, ids := invalidateFixture(b)

	for b.Loop() {
		for _, id := range ids {
			screen.Invalidate(id)
		}

		screen.TakeDirty()
	}
}
