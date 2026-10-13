package page

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

// benchDiffCells is how many small text values one generation rewrites.
const benchDiffCells = 20

// benchDiffGens is how many precomputed generations the loop cycles.
const benchDiffGens = 16

// benchDiffPage builds a page of small counter rows in one column and
// returns the page, its display, and a copy a generation can mutate.
func benchDiffPage(b *testing.B) (*Page, *Display, *Display) {
	b.Helper()
	var src strings.Builder
	src.WriteString(`<style>div{height:12px;font-size:10px;line-height:12px}</style><body style="margin:0">`)

	for i := 0; i < benchDiffCells; i++ {
		src.WriteString(`<div>0</div>`)
	}

	src.WriteString(`</body>`)

	p, err := New(Config{HTML: src.String(), Width: 320, Height: 200})
	if err != nil {
		b.Fatal(err)
	}

	if err := p.Redraw(context.Background()); err != nil {
		b.Fatal(err)
	}

	p.TakeDirtyRects()

	base := p.display
	next := *base
	next.Ops = append([]DisplayOp(nil), base.Ops...)

	return p, base, &next
}

// benchDiffValues precomputes every generation's text so the timed loop
// measures the diff and not string building.
func benchDiffValues() []string {
	vals := make([]string, benchDiffCells*benchDiffGens)
	for i := range vals {
		vals[i] = strconv.Itoa(i)
	}

	return vals
}

// mutateBenchText writes one generation's values into the first
// benchDiffCells text ops.
func mutateBenchText(d *Display, vals []string) {
	n := 0

	for i := range d.Ops {
		if d.Ops[i].Kind != layout.DisplayOpText {
			continue
		}

		d.Ops[i].Text = vals[n]
		n++

		if n == benchDiffCells {
			return
		}
	}
}
