package main

import (
	"bytes"
	"context"
	"fmt"
	"image/png"
	"strings"
	"testing"
)

// TestDrawerWidthClamp pins the drawer widening to [0, drawerMaxExtra]
// (about 30% beyond the 268 px base).
func TestDrawerWidthClamp(t *testing.T) {
	t.Parallel()

	app := &App{}

	app.drawerExtra = -5
	if got := app.drawerWidth(); got != sideWidth {
		t.Fatalf("negative extra width = %d, want %d", got, sideWidth)
	}

	app.drawerExtra = 0
	if got := app.drawerWidth(); got != sideWidth {
		t.Fatalf("base width = %d, want %d", got, sideWidth)
	}

	app.drawerExtra = 67
	if got, want := app.drawerWidth(), sideWidth+67; got != want {
		t.Fatalf("extra 67 width = %d, want %d", got, want)
	}

	app.drawerExtra = drawerMaxExtra + 20
	if got := app.drawerWidth(); got != drawerMaxW {
		t.Fatalf("over-max width = %d, want %d", got, drawerMaxW)
	}

	if drawerMaxW-sideWidth < 54 || drawerMaxW-sideWidth > 81 {
		t.Fatalf("expansion = %d, want 20-30%% of %d", drawerMaxW-sideWidth, sideWidth)
	}
}

// TestDrawerResizeExpandsSidebar checks the resize handle widens the drawer
// about 25% and shifts the editor, so files and cards gain room.
func TestDrawerResizeExpandsSidebar(t *testing.T) {
	app := newTestApp(t)

	if got := app.View().SideW; got != sideWidth {
		t.Fatalf("SideW = %d, want %d", got, sideWidth)
	}

	baseLeft := app.chromeLeftPx()
	app.setDrawerExtra(67)
	app.refresh()

	if got, want := app.View().SideW, sideWidth+67; got != want {
		t.Fatalf("wide SideW = %d, want %d", got, want)
	}

	if got := app.chromeLeftPx(); got != baseLeft+app.zx(67) {
		t.Fatalf("wide chrome left = %d, want %d", got, baseLeft+app.zx(67))
	}

	app.setDrawerExtra(drawerMaxExtra)
	app.refresh()

	if got := app.View().SideW; got != drawerMaxW {
		t.Fatalf("max SideW = %d, want %d", got, drawerMaxW)
	}

	app.setDrawerExtra(0)
	app.refresh()

	if got := app.View().SideW; got != sideWidth {
		t.Fatalf("restored SideW = %d, want %d", got, sideWidth)
	}
}

// TestCommitHashSubjectGap checks the history rail leaves 1-2 spaces between
// the commit hash and the subject, in both the full rail and the narrow
// sidebar rail.
func TestCommitHashSubjectGap(t *testing.T) {
	t.Parallel()

	html := buildHTML()
	if !strings.Contains(html, ".hist-row") || !strings.Contains(html, "gap: 12px") {
		t.Fatal("full history rail gap is not 12 px")
	}

	if !strings.Contains(html, ".side-history .hist-row { gap: 8px") {
		t.Fatal("sidebar history rail gap is not 8 px")
	}
}

// TestDrawerResizeHandleDrag checks the boundary handle claims the drag and
// widens the drawer, so long file and commit labels gain room. It grabs the
// drawer side of the 6 px strip, clear of the editor edge.
func TestDrawerResizeHandleDrag(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()
	app.refresh()

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	box, ok := boxByID(app.Page().Boxes(), "drawer-resize")
	if !ok {
		t.Fatal("no drawer resize handle")
	}

	if box.W != 6 {
		t.Fatalf("resize width = %.0f, want 6", box.W)
	}

	if want := float64(activityW + app.View().SideW - 3); box.X != want {
		t.Fatalf("resize x = %.0f, want %.0f", box.X, want)
	}

	startX, startY := box.X+1, box.Y+box.H/2

	claimed, err := app.Page().Testing().BeginDrag(ctx, startX, startY)
	if err != nil || !claimed {
		t.Fatalf("resize drag claim = %v %v", claimed, err)
	}

	if !app.drawerDrag {
		t.Fatal("resize handle did not claim the drag")
	}

	if err := app.Page().Testing().MoveDrag(ctx, startX+40, startY); err != nil {
		t.Fatal(err)
	}

	if err := app.Page().Testing().EndDrag(ctx); err != nil {
		t.Fatal(err)
	}

	if app.drawerExtra != 40 {
		t.Fatalf("drawerExtra = %d, want 40", app.drawerExtra)
	}

	if got, want := app.View().SideW, sideWidth+40; got != want {
		t.Fatalf("SideW = %d, want %d", got, want)
	}
}

// TestDrawerResizeWinsOverRail checks the edge handle claims the drag even
// where a vertical scroll rail overlaps it, so Source Control and long
// working-tree lists stay resizable. The rail thumbs keep their own drags
// because their centers sit clear of the strip.
func TestDrawerResizeWinsOverRail(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "README.md", "# Fixture\n")

	for i := 0; i < 60; i++ {
		writeFile(t, dir, fmt.Sprintf("many/file%02d.go", i), "package many\n")
	}

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	app.Page().SetSize(1280, 720)

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	clickID(t, app, "tree-many")

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	thumb, ok := boxByID(app.Page().Boxes(), "tree-thumb")
	if !ok {
		t.Fatal("no tree thumb")
	}

	resize, ok := boxByID(app.Page().Boxes(), "drawer-resize")
	if !ok {
		t.Fatal("no drawer resize handle")
	}

	claimed, err := app.Page().Testing().BeginDrag(ctx, resize.X+1, thumb.Y+thumb.H/2)
	if err != nil || !claimed {
		t.Fatalf("resize drag claim = %v %v", claimed, err)
	}

	if !app.drawerDrag || app.treeDrag {
		t.Fatalf("drawerDrag=%v treeDrag=%v, want resize to win the overlap", app.drawerDrag, app.treeDrag)
	}

	if err := app.Page().Testing().EndDrag(ctx); err != nil {
		t.Fatal(err)
	}

	claimed, err = app.Page().Testing().BeginDrag(ctx, thumb.X+thumb.W/2, thumb.Y+thumb.H/2)
	if err != nil || !claimed {
		t.Fatalf("thumb drag claim = %v %v", claimed, err)
	}

	if !app.treeDrag || app.drawerDrag {
		t.Fatalf("drawerDrag=%v treeDrag=%v, want thumb to keep its drag", app.drawerDrag, app.treeDrag)
	}

	if err := app.Page().Testing().EndDrag(ctx); err != nil {
		t.Fatal(err)
	}
}

// TestDrawerResizeGripPaints guards the visible affordance: the 2 px bar on
// the drawer edge must paint differently from the panel behind it, otherwise
// the user cannot find the handle to expand the drawer.
func TestDrawerResizeGripPaints(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()
	app.refresh()

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	box, ok := boxByID(app.Page().Boxes(), "drawer-resize")
	if !ok {
		t.Fatal("no drawer resize handle")
	}

	img, err := png.Decode(bytes.NewReader(app.Page().Probe().PNG()))
	if err != nil {
		t.Fatal(err)
	}

	barR, barG, barB, _ := img.At(int(box.X+3), int(box.Y+box.H/2)).RGBA()
	bgR, bgG, bgB, _ := img.At(int(box.X-10), int(box.Y+box.H/2)).RGBA()

	if barR == bgR && barG == bgG && barB == bgB {
		t.Fatal("resize grip paints the panel color, handle is invisible")
	}
}
