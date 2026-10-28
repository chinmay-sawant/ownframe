package main

import (
	"math"
	"testing"
)

func TestChangesScrollRows(t *testing.T) {
	t.Parallel()

	var s changesScroll

	if start, end := s.Rows(0, 5); start != 0 || end != 0 {
		t.Fatalf("empty list rows = %d..%d, want 0..0", start, end)
	}

	if start, end := s.Rows(4, 5); start != 0 || end != 4 {
		t.Fatalf("short list rows = %d..%d, want 0..4", start, end)
	}

	s.Offset = 3
	if start, end := s.Rows(4, 5); s.Offset != 0 || start != 0 || end != 4 {
		t.Fatalf("fitting list = offset %d rows %d..%d, want offset 0 rows 0..4", s.Offset, start, end)
	}

	s.Offset = 2
	if start, end := s.Rows(10, 3); s.Offset != 2 || start != 2 || end != 5 {
		t.Fatalf("middle window = offset %d rows %d..%d, want offset 2 rows 2..5", s.Offset, start, end)
	}

	s.Offset = 99
	if start, end := s.Rows(10, 3); s.Offset != 7 || start != 7 || end != 10 {
		t.Fatalf("tail window = offset %d rows %d..%d, want offset 7 rows 7..10", s.Offset, start, end)
	}

	s.Offset = -4
	if start, end := s.Rows(10, 3); s.Offset != 0 || start != 0 || end != 3 {
		t.Fatalf("leading window = offset %d rows %d..%d, want offset 0 rows 0..3", s.Offset, start, end)
	}

	s.Offset = 5
	if start, end := s.Rows(10, 0); s.Offset != 0 || start != 0 || end != 0 {
		t.Fatalf("zero-height window = offset %d rows %d..%d, want offset 0 rows 0..0", s.Offset, start, end)
	}
}

func TestChangesScrollWheel(t *testing.T) {
	t.Parallel()

	var s changesScroll

	if rows := s.WheelRows(0); rows != 0 || s.Accum != 0 {
		t.Fatalf("zero delta = %d rows accum %v, want 0 rows accum 0", rows, s.Accum)
	}

	if rows := s.WheelRows(0.4); rows != 0 {
		t.Fatalf("first fraction = %d rows, want 0", rows)
	}

	if rows := s.WheelRows(0.4); rows != 0 {
		t.Fatalf("second fraction = %d rows, want 0", rows)
	}

	if rows := s.WheelRows(0.4); rows != 1 {
		t.Fatalf("third fraction = %d rows, want 1", rows)
	}

	if math.Abs(s.Accum-0.2) > 1e-9 {
		t.Fatalf("remainder = %v, want 0.2", s.Accum)
	}

	if rows := s.WheelRows(4.5); rows != 4 {
		t.Fatalf("large delta = %d rows, want 4", rows)
	}

	if math.Abs(s.Accum-0.7) > 1e-9 {
		t.Fatalf("large remainder = %v, want 0.7", s.Accum)
	}

	s = changesScroll{}
	got := []int{s.WheelRows(-0.4), s.WheelRows(-0.4), s.WheelRows(-0.4)}
	for i, want := range []int{0, 0, -1} {
		if got[i] != want {
			t.Fatalf("negative deltas = %v, want [0 0 -1]", got)
		}
	}

	if math.Abs(s.Accum+0.2) > 1e-9 {
		t.Fatalf("negative remainder = %v, want -0.2", s.Accum)
	}

	if rows := s.WheelRows(-1.8); rows != -2 {
		t.Fatalf("negative whole delta = %d rows, want -2", rows)
	}

	if math.Abs(s.Accum) > 1e-9 {
		t.Fatalf("negative whole remainder = %v, want 0", s.Accum)
	}
}

func TestChangesScrollThumb(t *testing.T) {
	t.Parallel()

	var s changesScroll

	if top, height, show := s.Thumb(200, 10, 10, 24); show || top != 0 || height != 0 {
		t.Fatalf("fitting list thumb = top %d height %d show %v, want none", top, height, show)
	}

	if _, _, show := s.Thumb(0, 100, 10, 24); show {
		t.Fatal("thumb shows without a track")
	}

	if _, _, show := s.Thumb(200, 100, 0, 24); show {
		t.Fatal("thumb shows without rows")
	}

	top, height, show := s.Thumb(200, 100, 10, 0)
	if !show || top != 0 || height != 20 {
		t.Fatalf("proportional thumb = top %d height %d show %v, want top 0 height 20", top, height, show)
	}

	s.SetOffset(45, 100, 10)

	top, height, show = s.Thumb(200, 100, 10, 0)
	if !show || top != 90 || height != 20 {
		t.Fatalf("middle thumb = top %d height %d show %v, want top 90 height 20", top, height, show)
	}

	s.SetOffset(90, 100, 10)

	top, height, show = s.Thumb(200, 100, 10, 0)
	if !show || top != 180 || height != 20 {
		t.Fatalf("end thumb = top %d height %d show %v, want top 180 height 20", top, height, show)
	}

	s = changesScroll{}

	_, height, _ = s.Thumb(200, 1000, 10, 24)
	if height != 24 {
		t.Fatalf("min thumb height = %d, want 24", height)
	}

	top, height, show = s.Thumb(10, 1000, 10, 24)
	if !show || top != 0 || height != 10 {
		t.Fatalf("short track thumb = top %d height %d show %v, want top 0 height 10", top, height, show)
	}

	s.Offset = 500

	top, height, show = s.Thumb(200, 20, 5, 10)
	if !show || top != 150 || height != 50 {
		t.Fatalf("stale offset thumb = top %d height %d show %v, want top 150 height 50", top, height, show)
	}

	if s.Offset != 500 {
		t.Fatalf("thumb wrote offset %d, want 500", s.Offset)
	}
}

func TestChangesScrollClamp(t *testing.T) {
	t.Parallel()

	var s changesScroll

	s.SetOffset(-5, 10, 4)
	if s.Offset != 0 {
		t.Fatalf("negative offset = %d, want 0", s.Offset)
	}

	s.SetOffset(50, 10, 4)
	if s.Offset != 6 {
		t.Fatalf("past-end offset = %d, want 6", s.Offset)
	}

	s.SetOffset(0, 10, 4)
	s.ScrollBy(3, 10, 4)
	if s.Offset != 3 {
		t.Fatalf("scroll down = %d, want 3", s.Offset)
	}

	s.ScrollBy(9, 10, 4)
	if s.Offset != 6 {
		t.Fatalf("scroll past end = %d, want 6", s.Offset)
	}

	s.ScrollBy(-100, 10, 4)
	if s.Offset != 0 {
		t.Fatalf("scroll past start = %d, want 0", s.Offset)
	}

	s.SetOffset(5, 3, 4)
	if s.Offset != 0 {
		t.Fatalf("offset for a fitting list = %d, want 0", s.Offset)
	}

	s.SetOffset(5, 0, 0)
	if s.Offset != 0 {
		t.Fatalf("offset for an empty list = %d, want 0", s.Offset)
	}

	s.SetOffset(5, 10, 0)
	if s.Offset != 0 {
		t.Fatalf("offset for no rows = %d, want 0", s.Offset)
	}

	s.SetOffset(0, 30, 6)
	s.Page(1, 30, 6)
	if s.Offset != 5 {
		t.Fatalf("first page down = %d, want 5", s.Offset)
	}

	s.Page(1, 30, 6)
	if s.Offset != 10 {
		t.Fatalf("second page down = %d, want 10", s.Offset)
	}

	s.Page(-1, 30, 6)
	if s.Offset != 5 {
		t.Fatalf("page up = %d, want 5", s.Offset)
	}

	s.Page(0, 30, 6)
	if s.Offset != 5 {
		t.Fatalf("zero page = %d, want 5", s.Offset)
	}

	s.SetOffset(100, 30, 6)
	s.Page(1, 30, 6)
	if s.Offset != 24 {
		t.Fatalf("page down at the end = %d, want 24", s.Offset)
	}

	s.SetOffset(0, 5, 1)
	s.Page(1, 5, 1)
	if s.Offset != 1 {
		t.Fatalf("one-row page = %d, want 1", s.Offset)
	}
}
