package transitionpreview

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
	"github.com/chinmay-sawant/ownframe/internal/previewwindow"
)

func paint(ops []*ownframe.DisplayOp) []cssmotion.Paint {
	out := make([]cssmotion.Paint, len(ops))
	for i, op := range ops {
		out[i] = cssmotion.Read(op)
	}
	return out
}

func clickStart(t *testing.T, ctx context.Context, page *ownframe.Page, id string) {
	t.Helper()
	button, ok := findBox(page.Boxes(), id)
	if !ok {
		t.Fatalf("missing Start button %q", id)
	}
	if err := page.Click(ctx, button.X+button.W/2, button.Y+button.H/2); err != nil {
		t.Fatal(err)
	}
}

func testPage(t *testing.T, data *view) *ownframe.Page {
	t.Helper()
	page, err := previewwindow.New(context.Background(), previewwindow.Config{
		FixtureDir: filepath.Join("..", "..", "testdata", "transition"), Data: data, Width: 960, Height: 860})
	if err != nil {
		t.Fatal(err)
	}
	data.Readout = describe(page)
	if err := previewwindow.Redraw(context.Background(), page, data); err != nil {
		t.Fatal(err)
	}
	return page
}
