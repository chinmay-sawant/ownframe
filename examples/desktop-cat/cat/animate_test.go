package cat

import (
	"context"
	"io/fs"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/desktop-cat/assets"
)

func testAnimation(t *testing.T, cycle bool) *animation {
	t.Helper()
	page, err := NewVariant(1)
	if err != nil {
		t.Fatal(err)
	}
	if err := page.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	files, err := fs.Glob(assets.Cats, "cat_images/*.png")
	if err != nil {
		t.Fatal(err)
	}
	return &animation{page: page, files: files, cycle: cycle}
}

func TestImageAnimationKeepsLayoutBetweenExpressions(t *testing.T) {
	a := testAnimation(t, false)
	ctx := context.Background()
	if err := a.paint(ctx, 0); err != nil {
		t.Fatal(err)
	}
	before := a.page.Stats()
	y := a.image.Y
	if err := a.paint(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if a.image.Y == y {
		t.Fatal("cat did not move")
	}
	if after := a.page.Stats(); after.Parses != before.Parses || after.Layouts != before.Layouts {
		t.Fatal("frame animation parsed or laid out the page")
	}
	if a.image.W != 200*a.display.PointsPerPixel || a.image.H != 200*a.display.PointsPerPixel {
		t.Fatal("cat was not scaled to 200 CSS pixels")
	}
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.paint(ctx, 2); err != nil {
		t.Fatal(err)
	}
	if a.display != a.page.Display() {
		t.Fatal("animation kept an obsolete display")
	}
}
