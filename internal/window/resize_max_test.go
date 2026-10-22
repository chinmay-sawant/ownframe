package window

import (
	"context"
	"testing"
)

// freeScreen clamps to the minimum only, the way page.Page does now, and
// reports the max as a window bound.
type freeScreen struct {
	fakeScreen
}

func (f *freeScreen) Clamp(width, height int) (int, int) {
	if width < f.minW {
		width = f.minW
	}

	if height < f.minH {
		height = f.minH
	}

	return width, height
}

func (f *freeScreen) SetSize(width, height int) {
	f.width, f.height = f.Clamp(width, height)
}

func (f *freeScreen) MaxSize() (int, int) {
	return f.maxW, f.maxH
}

func TestResizeLaysOutPastTheMax(t *testing.T) {
	t.Parallel()

	app := &freeScreen{fakeScreen: fakeScreen{
		width: 480, height: 640,
		minW: 320, minH: 400,
		maxW: 800, maxH: 900,
	}}
	shell := newShell(context.Background(), app)

	shell.Layout(2000, 1500)
	if err := shell.resize(); err != nil {
		t.Fatal(err)
	}

	if app.width != 2000 || app.height != 1500 {
		t.Fatalf("size = %d x %d, want 2000 x 1500", app.width, app.height)
	}

	if shell.stretched() {
		t.Fatal("stretched with the frame at the window size")
	}
}

func TestWindowBoundsReadsMaxSize(t *testing.T) {
	t.Parallel()

	bounded := &freeScreen{fakeScreen: fakeScreen{maxW: 800, maxH: 900}}
	if maxW, maxH := windowBounds(bounded); maxW != 800 || maxH != 900 {
		t.Fatalf("bounds = %d x %d, want 800 x 900", maxW, maxH)
	}

	plain := &fakeScreen{}
	if maxW, maxH := windowBounds(plain); maxW != -1 || maxH != -1 {
		t.Fatalf("plain bounds = %d x %d, want -1 x -1", maxW, maxH)
	}
}
