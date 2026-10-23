package draw

import (
	"testing"
	"time"
)

func TestStatsSnapshotPrefersFrameDraw(t *testing.T) {
	t.Parallel()

	got := Snapshot(&Counts{LastDraw: time.Second, FrameDraw: 7 * time.Millisecond, Redraws: 2}, 3, 4)
	if got.LastDraw != 7*time.Millisecond || got.Redraws != 2 || got.Boxes != 3 || got.Ops != 4 {
		t.Fatalf("%+v", got)
	}

	plain := Snapshot(&Counts{LastDraw: time.Second, Parses: 1, LastReloadError: "e"}, 0, 0)
	if plain.LastDraw != time.Second || plain.Parses != 1 || plain.LastReloadError != "e" {
		t.Fatalf("%+v", plain)
	}
}

func TestStatsNoteCommit(t *testing.T) {
	t.Parallel()

	var layouts, repaints uint64
	var last time.Duration
	NoteCommit(&layouts, &repaints, &last, time.Now().Add(-time.Millisecond))
	if layouts != 1 || repaints != 1 || last <= 0 {
		t.Fatalf("%d %d %v", layouts, repaints, last)
	}
}
