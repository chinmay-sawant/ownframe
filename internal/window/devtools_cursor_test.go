package window

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/hajimehoshi/ebiten/v2"
)

type devCursorScreen struct {
	*devScreen
	shape host.Shape
}

func (d *devCursorScreen) CursorShape() host.Shape { return d.shape }

func TestDevToolsNormalHoverUpdatesCursor(t *testing.T) {
	d := &devCursorScreen{devScreen: newDevScreen(), shape: host.ShapePointer}
	s := newDevShell(d.devScreen)
	s.app = d
	s.devLayout()
	// The routing under test runs after startup, so commit a frame: the
	// first update defers the page hover until the first paint.
	s.frameCommit()
	s.cursor = ebiten.CursorShapeText
	s.setCursor = func(ebiten.CursorShapeType) {}
	if err := s.devPointer(20, 20, 20, 20); err != nil {
		t.Fatal(err)
	}
	if s.cursor != ebiten.CursorShapePointer || d.hovers != 1 {
		t.Fatalf("normal content hover kept stale cursor: shape=%v hovers=%d", s.cursor, d.hovers)
	}

	if err := s.devPointer(int(s.dev.panel.X)+50, 10, 0, 0); err != nil {
		t.Fatal(err)
	}
	if s.cursor != ebiten.CursorShapeDefault {
		t.Fatal("dock chrome retained the page cursor")
	}
	if err := s.devPointer(int(s.dev.panel.X), 100, 0, 0); err != nil {
		t.Fatal(err)
	}
	if s.cursor != ebiten.CursorShapeEWResize {
		t.Fatal("dock resize edge did not show its resize cursor")
	}
	s.devSetPicker(true)
	if err := s.devPointer(20, 20, 20, 20); err != nil {
		t.Fatal(err)
	}
	if s.cursor != ebiten.CursorShapeDefault || !s.dev.haveHov || d.hovers != 1 {
		t.Fatal("picker did not use its own hover without forwarding to the page")
	}
	s.devSetPicker(false)
	d.shape = host.ShapeDefault
	if err := s.devPointer(21, 20, 21, 20); err != nil {
		t.Fatal(err)
	}
	if s.cursor != ebiten.CursorShapeDefault || d.hovers != 2 {
		t.Fatal("disabling picker did not restore normal page hover")
	}
}
