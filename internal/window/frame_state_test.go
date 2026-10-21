package window

import (
	"image"
	"slices"
	"testing"
)

// frameRects copies the committed chrome set so later frames cannot change
// the value under a check.
func frameRects(s *shell) []image.Rectangle {
	return slices.Clone(s.frameRestore())
}

// TestFrameBeginClearsCurrent drops a note made before begin: only the rects
// collected inside the open frame reach the commit.
func TestFrameBeginClearsCurrent(t *testing.T) {
	t.Parallel()

	s := &shell{}
	s.frameNote(image.Rect(1, 1, 2, 2))
	s.frameBegin()
	s.frameNote(image.Rect(3, 3, 4, 4))
	s.frameCommit()

	want := []image.Rectangle{image.Rect(3, 3, 4, 4)}
	if got := frameRects(s); !slices.Equal(got, want) {
		t.Fatalf("committed = %v, want %v", got, want)
	}
}

// TestFrameNoteRecordsOnlyPaintedRects keeps the non-empty notes in order
// and drops every empty one.
func TestFrameNoteRecordsOnlyPaintedRects(t *testing.T) {
	t.Parallel()

	s := &shell{}
	s.frameBegin()
	s.frameNote(image.Rect(10, 10, 30, 30))
	s.frameNote(image.Rect(5, 5, 5, 40))
	s.frameNote(image.Rectangle{})
	s.frameNote(image.Rectangle{Min: image.Pt(40, 40), Max: image.Pt(30, 30)})
	s.frameNote(image.Rect(50, 50, 70, 70))
	s.frameCommit()

	want := []image.Rectangle{image.Rect(10, 10, 30, 30), image.Rect(50, 50, 70, 70)}
	if got := frameRects(s); !slices.Equal(got, want) {
		t.Fatalf("committed = %v, want %v", got, want)
	}
}

// TestFrameRestoreKeepsTheCommittedFrame reads restore while the next frame
// is being collected: new notes must not leak into the committed set.
func TestFrameRestoreKeepsTheCommittedFrame(t *testing.T) {
	t.Parallel()

	s := &shell{}
	old := image.Rect(1, 1, 2, 2)
	s.frameBegin()
	s.frameNote(old)
	s.frameCommit()

	s.frameBegin()
	s.frameNote(image.Rect(3, 3, 4, 4))
	s.frameNote(image.Rect(5, 5, 6, 6))

	want := []image.Rectangle{old}
	if got := frameRects(s); !slices.Equal(got, want) {
		t.Fatalf("restore during collection = %v, want %v", got, want)
	}
}
