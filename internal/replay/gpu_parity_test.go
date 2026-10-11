package replay

import (
	"context"
	"os"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/page"
	"github.com/hajimehoshi/ebiten/v2"
)

func TestGPUVisibleParity(t *testing.T) {
	if os.Getenv("OWNFRAME_REPLAY_GPU_TEST") != "1" && os.Getenv("GPUI_REPLAY_GPU_TEST") != "1" {
		t.Skip("requires an isolated graphics session")
	}
	source, err := os.ReadFile("../../examples/perf-complex/layout.html")
	if err != nil {
		t.Fatal(err)
	}
	rows := make([]map[string]any, 480)
	for i := range rows {
		rows[i] = map[string]any{"ID": i, "Region": i % 8, "Name": "service"}
	}
	p, err := page.New(page.Config{HTML: string(source), Width: 1908, Height: 999})
	if err != nil {
		t.Fatal(err)
	}
	p.SetData(map[string]any{"Revision": 0, "Top": 0, "Bottom": 0, "Nav": []string{"Overview", "Services", "Regions", "Deployments", "Incidents", "Capacity"}, "Rows": rows})
	if err = p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	d := p.Display()
	d.Order = append(d.Order, len(d.Ops))
	d.Ops = append(d.Ops, layout.DisplayOp{Kind: layout.DisplayOpFillRect, Fixed: true, X: 10, Y: 10, W: 20, H: 20, R: 1})
	g := &parityGame{display: d}
	ebiten.SetWindowSize(1908, 999)
	if err = ebiten.RunGame(g); err != nil {
		t.Fatal(err)
	}
	if g.err != nil {
		t.Fatal(g.err)
	}
}
