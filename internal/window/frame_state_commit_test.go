package window

import (
	"image"
	"slices"
	"testing"
)

// TestFrameCommitSwapsTheSets checks seen and that each commit replaces the
// restore set with the rects of this frame alone.
func TestFrameCommitSwapsTheSets(t *testing.T) {
	t.Parallel()

	s := &shell{}
	if s.frameSeen() {
		t.Fatal("frameSeen before the first commit")
	}

	first := image.Rect(1, 1, 2, 2)
	s.frameBegin()
	s.frameNote(first)
	s.frameCommit()

	if !s.frameSeen() {
		t.Fatal("frameCommit left seen false")
	}

	if got := frameRects(s); !slices.Equal(got, []image.Rectangle{first}) {
		t.Fatalf("first commit = %v, want %v", got, first)
	}

	second := image.Rect(3, 3, 4, 4)
	s.frameBegin()
	s.frameNote(second)
	s.frameCommit()

	if got := frameRects(s); !slices.Equal(got, []image.Rectangle{second}) {
		t.Fatalf("second commit = %v, want only %v", got, second)
	}
}

// TestFrameSteadyStateAllocatesNothing pins the slice swap: a frame that
// notes up to two rects reuses both backing arrays.
func TestFrameSteadyStateAllocatesNothing(t *testing.T) {
	s := &shell{}
	rects := []image.Rectangle{image.Rect(1, 1, 2, 2), image.Rect(3, 3, 4, 4)}
	frame := func() {
		s.frameBegin()
		for _, r := range rects {
			s.frameNote(r)
		}
		s.frameCommit()
	}
	frame()

	if n := testing.AllocsPerRun(100, frame); n != 0 {
		t.Fatalf("allocs per frame = %v, want 0", n)
	}
}
