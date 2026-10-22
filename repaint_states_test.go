package ownframe

import (
	"context"
	"testing"
)

const statesRepaintHTML = `<html><head><style>
#hover { width:200px;height:60px;margin:20px;background:#e5e7eb }
#hover:hover { background:#1a56db }
</style></head><body style="margin:0;background:#f4f1ea">
<div id="hover"><span>Hover me</span></div></body></html>`

// statesCase moves the pointer onto the hover target.
func statesCase() repaintCase {
	return repaintCase{
		name: "states",
		newPage: func(t *testing.T) *Page {
			return newPage(t, Config{Title: "States", HTML: statesRepaintHTML, Width: 320, Height: 160}, nil)
		},
		interact: func(t *testing.T, p *Page, ctx context.Context) {
			x, y := boxCenter(t, p, "hover")
			if err := p.Testing().Hover(ctx, x, y); err != nil {
				t.Fatal(err)
			}
		},
	}
}
