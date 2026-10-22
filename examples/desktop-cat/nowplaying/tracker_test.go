package nowplaying

import "testing"

func TestSettledTrackChangesAndPause(t *testing.T) {
	var state tracker
	a := Track{Title: "Song A", App: "Chrome", Playing: true}
	b := Track{Title: "Song B", App: "Chrome", Playing: true}
	if state.next(a) != "" {
		t.Fatal("unsettled title announced")
	}
	key := state.next(a)
	if key == "" {
		t.Fatal("settled title missing")
	}
	state.last = key
	if state.next(a) != "" {
		t.Fatal("duplicate title")
	}
	state.next(Track{})
	if state.next(a) != "" {
		t.Fatal("pause and resume repeated title")
	}
	if state.next(a) != "" {
		t.Fatal("pause and resume repeated title")
	}
	if state.next(b) != "" {
		t.Fatal("new title not settled")
	}
	key = state.next(b)
	if key == "" {
		t.Fatal("song change missing")
	}
	if state.next(b) != key {
		t.Fatal("failed publication could not retry")
	}
	state.last = key
	if state.next(b) != "" {
		t.Fatal("published title repeated")
	}
}
