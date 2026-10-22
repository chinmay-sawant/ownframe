package window

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestReloadNotePrintsOncePerDistinctError(t *testing.T) {
	t.Parallel()

	s := &shell{}
	err := errors.New("/tmp/index.html: parse: unexpected EOF")

	line, show := s.reloadNote(err)
	if !show || line != "hot reload: /tmp/index.html: parse: unexpected EOF" {
		t.Fatalf("line = %q, show = %v", line, show)
	}

	if _, show := s.reloadNote(err); show {
		t.Fatal("the same error printed twice")
	}

	s.lastNote = ""

	if _, show := s.reloadNote(err); !show {
		t.Fatal("the error did not print after a success")
	}
}

func TestHotReloadErrorKeepsDrawing(t *testing.T) {
	t.Parallel()

	app := newReloadScreen()
	app.err = errors.New("/tmp/index.html: stat: no such file")

	s := newShell(context.Background(), app)
	if err := s.Update(); err != nil {
		t.Fatal(err)
	}

	if app.polls != 1 || s.display == nil {
		t.Fatalf("polls = %d, display = %v", app.polls, s.display)
	}

	s.lastPoll = time.Now().Add(-time.Second)

	if err := s.Update(); err != nil {
		t.Fatal(err)
	}

	if app.polls != 2 {
		t.Fatalf("polls = %d, want 2", app.polls)
	}

	if app.redraws != 1 {
		t.Fatal("an error reloaded the page")
	}
}

func TestReloadClampsScroll(t *testing.T) {
	t.Parallel()

	app := newReloadScreen()
	app.left = 1
	app.boxes = []layout.Box{{Y: 500, H: 100, W: 10}}

	s := newShell(context.Background(), app)
	s.scrollY = 550

	if err := s.Update(); err != nil {
		t.Fatal(err)
	}

	if s.scrollY != 0 {
		t.Fatalf("scrollY = %d, want 0 after the shorter page", s.scrollY)
	}
}
