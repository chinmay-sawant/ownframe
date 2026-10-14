package replay

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

// cardPage lays out a grid of cards, each a background, border, title, and
// two body lines, the shape a dashboard repaints on hover.
func cardPage(t testing.TB, cards int) *layout.Display {
	t.Helper()

	var sb strings.Builder
	sb.WriteString(`<html><body style="margin:0;background:#202030">`)

	for i := range cards {
		fmt.Fprintf(&sb, `<div id="card%d" style="margin:8px;width:200px;height:90px;`+
			`background:#eef;border:2px solid #336;border-radius:6px;padding:6px">`+
			`<p style="margin:0">Card %d</p><p style="margin:2px 0">detail line</p></div>`, i, i)
	}

	sb.WriteString(`</body></html>`)

	d, err := render.DisplayList(context.Background(), sb.String(), 480, 640)
	if err != nil {
		t.Fatal(err)
	}

	return d
}

// planStats counts a plan's direct ops, runs, and run members.
func planStats(p *chunkPlan) (direct, runs, members int) {
	for i := range p.items {
		switch {
		case p.items[i].run != nil:
			runs++
			members += len(p.items[i].run.ops)
		case p.items[i].op != nil:
			direct++
		}
	}

	return direct, runs, members
}

func TestChunkPlanFoldsCards(t *testing.T) {
	d := cardPage(t, 6)
	p := buildChunks(d, buildPlan(d))
	direct, runs, members := planStats(p)
	t.Logf("ops=%d boxes=%d direct=%d runs=%d members=%d", len(d.Ops), len(d.Boxes), direct, runs, members)

	if runs == 0 {
		t.Fatal("no card run folded")
	}

	if members < 2 {
		t.Fatalf("runs hold %d ops", members)
	}

	for i := range p.items {
		if p.items[i].run == nil {
			continue
		}

		for _, op := range p.items[i].run.ops {
			if _, ok := opBounds(op, d.PointsPerPixel); !ok {
				t.Fatalf("run holds an unbounded op %v", op.Kind)
			}
		}
	}
}
