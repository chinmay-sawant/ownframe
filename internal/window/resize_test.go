package window

import (
	"context"
	"testing"
)

func TestResizeRedrawsOnTheFirstChangedFrame(t *testing.T) {
	t.Parallel()

	app := &fakeScreen{
		width: 480, height: 640,
		minW: 320, minH: 400,
		maxW: 2560, maxH: 2560,
	}
	game := NewGame(context.Background(), app)
	shell, ok := game.(*shell)
	if !ok {
		t.Fatalf("game = %T", game)
	}

	shell.Layout(600, 720)
	if err := shell.resize(); err != nil {
		t.Fatal(err)
	}

	if app.redraws != 1 {
		t.Fatalf("redraws = %d, want 1", app.redraws)
	}

	if app.width != 600 || app.height != 720 {
		t.Fatalf("size = %d x %d", app.width, app.height)
	}

	if err := shell.resize(); err != nil {
		t.Fatal(err)
	}

	if app.redraws != 1 {
		t.Fatalf("redraws after a settled frame = %d, want 1", app.redraws)
	}
}

func TestResizeClampsBeforeRedraw(t *testing.T) {
	t.Parallel()

	app := &fakeScreen{
		width: 480, height: 640,
		minW: 320, minH: 400,
		maxW: 800, maxH: 900,
	}
	shell := newShell(context.Background(), app)

	shell.Layout(50, 50)
	if err := shell.resize(); err != nil {
		t.Fatal(err)
	}

	if app.width != 320 || app.height != 400 {
		t.Fatalf("clamped size = %d x %d", app.width, app.height)
	}

	if app.redraws != 1 {
		t.Fatalf("redraws = %d, want 1", app.redraws)
	}
}

func TestResizeDefersCommitDuringTitlebarDrag(t *testing.T) {
	t.Parallel()

	app := &fakeScreen{
		width: 480, height: 640,
		minW: 320, minH: 400,
		maxW: 2560, maxH: 2560,
	}
	shell := newShell(context.Background(), app)

	shell.windowDrag.active = true
	shell.Layout(600, 720)
	if err := shell.resize(); err != nil {
		t.Fatal(err)
	}

	if app.redraws != 0 {
		t.Fatalf("redraws during drag = %d, want 0", app.redraws)
	}

	if app.width != 480 || app.height != 640 {
		t.Fatalf("size during drag = %d x %d, want unchanged", app.width, app.height)
	}

	shell.windowDrag.active = false
	if err := shell.resize(); err != nil {
		t.Fatal(err)
	}

	if app.redraws != 1 {
		t.Fatalf("redraws after release = %d, want 1", app.redraws)
	}

	if app.width != 600 || app.height != 720 {
		t.Fatalf("size after release = %d x %d, want 600 x 720", app.width, app.height)
	}
}
