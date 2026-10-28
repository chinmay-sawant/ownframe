package main

import "testing"

func TestHistoryScrollRows(t *testing.T) {
	t.Parallel()

	var scroll historyScroll

	if start, end := scroll.Rows(0, 10); start != 0 || end != 0 {
		t.Fatalf("empty list rows = %d..%d, want none", start, end)
	}

	if start, end := scroll.Rows(100, 0); start != 0 || end != 0 {
		t.Fatalf("no space rows = %d..%d, want none", start, end)
	}

	if start, end := scroll.Rows(100, 10); start != 0 || end != 10 {
		t.Fatalf("first page rows = %d..%d, want 0..10", start, end)
	}

	scroll.SetOffset(90, 100, 10)
	if start, end := scroll.Rows(100, 10); start != 90 || end != 100 {
		t.Fatalf("last page rows = %d..%d, want 90..100", start, end)
	}

	scroll.SetOffset(200, 100, 10)
	if scroll.Offset != 90 {
		t.Fatalf("offset past the end = %d, want 90", scroll.Offset)
	}

	scroll.SetOffset(-4, 100, 10)
	if scroll.Offset != 0 {
		t.Fatalf("negative offset = %d, want 0", scroll.Offset)
	}

	scroll.SetOffset(30, 100, 10)
	scroll.ScrollBy(-5, 100, 10)
	if scroll.Offset != 25 {
		t.Fatalf("offset after scroll = %d, want 25", scroll.Offset)
	}

	scroll.ScrollBy(500, 100, 10)
	if scroll.Offset != 90 {
		t.Fatalf("offset after a big scroll = %d, want 90", scroll.Offset)
	}

	scroll.ScrollBy(-500, 100, 10)
	if start, end := scroll.Rows(100, 10); start != 0 || end != 10 {
		t.Fatalf("rows after scrolling back = %d..%d, want 0..10", start, end)
	}

	scroll.SetOffset(4, 6, 10)
	if start, end := scroll.Rows(6, 10); start != 0 || end != 6 {
		t.Fatalf("short list rows = %d..%d, want 0..6", start, end)
	}

	// A stale offset is pinned to the history, so a shorter list cannot keep
	// drawing past its end.
	scroll.Offset = 80
	if start, end := scroll.Rows(6, 10); start != 0 || end != 6 {
		t.Fatalf("stale offset rows = %d..%d, want 0..6", start, end)
	}

	if scroll.Offset != 0 {
		t.Fatalf("pinned offset = %d, want 0", scroll.Offset)
	}

	scroll.ScrollBy(-2, 6, 10)
	if scroll.Offset != 0 {
		t.Fatalf("scroll from a stale offset = %d, want 0", scroll.Offset)
	}
}

func TestHistoryScrollPreloadAtHalf(t *testing.T) {
	t.Parallel()

	scroll := historyScroll{Limit: 40}

	if scroll.ShouldLoad(40, 10) {
		t.Fatal("load asked at the top of a full page")
	}

	scroll.SetOffset(9, 40, 10)
	if scroll.ShouldLoad(40, 10) {
		t.Fatal("load asked before half of the loaded rows")
	}

	scroll.SetOffset(10, 40, 10)
	if !scroll.ShouldLoad(40, 10) {
		t.Fatal("no load at exactly half of the loaded rows")
	}

	scroll.StartLoad()

	if !scroll.Loading || scroll.ShouldLoad(40, 10) {
		t.Fatalf("load asked while one is in flight: loading=%v", scroll.Loading)
	}

	scroll.DoneLoad()

	if scroll.Loading || !scroll.ShouldLoad(40, 10) {
		t.Fatalf("load not asked again after the fetch: loading=%v", scroll.Loading)
	}

	if scroll.ShouldLoad(30, 10) {
		t.Fatal("load asked after the history came back short")
	}

	scroll = historyScroll{}
	if scroll.ShouldLoad(40, 10) {
		t.Fatal("load asked with no limit set")
	}

	if scroll.ShouldLoad(0, 10) {
		t.Fatal("load asked for an empty history")
	}

	scroll = historyScroll{Limit: 40}
	if scroll.ShouldLoad(40, 0) {
		t.Fatal("load asked with no rows visible")
	}
}

func TestNextHistoryLimit(t *testing.T) {
	t.Parallel()

	if got := NextHistoryLimit(20, 500); got != 40 {
		t.Fatalf("limit from 20 = %d, want 40", got)
	}

	if got := NextHistoryLimit(40, 500); got != 60 {
		t.Fatalf("limit from 40 = %d, want 60", got)
	}

	if got := NextHistoryLimit(60, 500); got != 90 {
		t.Fatalf("limit from 60 = %d, want 90", got)
	}

	if got := NextHistoryLimit(0, 500); got != 20 {
		t.Fatalf("first limit = %d, want 20", got)
	}

	if got := NextHistoryLimit(-8, 500); got != 20 {
		t.Fatalf("negative limit = %d, want 20", got)
	}

	if got := NextHistoryLimit(200, 250); got != 250 {
		t.Fatalf("limit over the cap = %d, want 250", got)
	}

	if got := NextHistoryLimit(250, 250); got != 250 {
		t.Fatalf("limit at the cap = %d, want 250", got)
	}

	if got := NextHistoryLimit(60, 0); got != 90 {
		t.Fatalf("uncapped limit = %d, want 90", got)
	}
}

func TestHistoryScrollThumb(t *testing.T) {
	t.Parallel()

	const (
		track  = 300
		minH   = 18
		total  = 40
		window = 10
	)

	var scroll historyScroll

	if _, _, show := scroll.Thumb(track, 20, 20, minH); show {
		t.Fatal("thumb shown when every row fits")
	}

	if _, _, show := scroll.Thumb(track, 0, window, minH); show {
		t.Fatal("thumb shown for an empty history")
	}

	if _, _, show := scroll.Thumb(0, total, window, minH); show {
		t.Fatal("thumb shown on a zero-height track")
	}

	top, height, show := scroll.Thumb(track, total, window, minH)
	if !show || top != 0 {
		t.Fatalf("top thumb = %d,%d show=%v, want top 0", top, height, show)
	}

	if want := track * window / total; height != want {
		t.Fatalf("thumb height = %d, want %d", height, want)
	}

	scroll.SetOffset(total-window, total, window)

	bottom, height, show := scroll.Thumb(track, total, window, minH)
	if !show {
		t.Fatal("no thumb with rows below the window")
	}

	if bottom+height != track {
		t.Fatalf("bottom thumb = %d..%d, want it flush with %d", bottom, bottom+height, track)
	}

	scroll.SetOffset(window, total, window)

	middle, _, _ := scroll.Thumb(track, total, window, minH)
	if middle <= 0 || middle >= bottom {
		t.Fatalf("middle thumb top = %d, want between 0 and %d", middle, bottom)
	}

	// A short list still gets a grabbable thumb.
	scroll.SetOffset(90, 100, 5)
	if _, height, show := scroll.Thumb(track, 100, 5, minH); !show || height != minH {
		t.Fatalf("short-list thumb height = %d show=%v, want %d", height, show, minH)
	}

	// A very long history still gets a one pixel thumb.
	if top, height, show := scroll.Thumb(1, 1000, 1, 0); !show || height != 1 || top != 0 {
		t.Fatalf("one pixel thumb = %d..%d show=%v, want a 1 px thumb", top, top+height, show)
	}

	// A stale offset cannot push the thumb past the track.
	scroll.Offset = 500
	if top, height, show := scroll.Thumb(track, total, window, minH); !show || top+height != track {
		t.Fatalf("stale thumb = %d..%d show=%v, want it flush with %d", top, top+height, show, track)
	}
}

func TestHistoryScrollWheel(t *testing.T) {
	t.Parallel()

	var scroll historyScroll

	if rows := scroll.WheelRows(0.5); rows != 0 {
		t.Fatalf("first half row = %d, want 0", rows)
	}

	if rows := scroll.WheelRows(0.5); rows != 1 {
		t.Fatalf("accumulated half rows = %d, want 1", rows)
	}

	if scroll.Accum != 0 {
		t.Fatalf("accum after a whole row = %v, want 0", scroll.Accum)
	}

	if rows := scroll.WheelRows(-1.5); rows != -1 {
		t.Fatalf("reversed wheel = %d, want -1", rows)
	}

	if rows := scroll.WheelRows(-0.5); rows != -1 || scroll.Accum != 0 {
		t.Fatalf("second reversed wheel = %d accum %v, want -1 and 0", rows, scroll.Accum)
	}
}
