package main

import (
	"strconv"
	"testing"
)

// TestWalkStatePageCounters checks the positions the click script reads
// against the page the snapshot came from.
func TestWalkStatePageCounters(t *testing.T) {
	p, d, v := fixturePage(t)
	w := &walker{p: p, d: d, v: v}
	w.snapshot()
	fields := walkStateFields(t, w)
	if len(fields) != 20 {
		t.Fatalf("state has %d fields, want 20", len(fields))
	}
	st := p.Stats()
	for _, c := range []struct {
		index int
		name  string
		want  string
	}{
		{0, "next", strconv.Itoa(w.next)},
		{1, "started", strconv.Itoa(w.trace.started)},
		{2, "selected", w.trace.selected},
		{3, "last redraw", strconv.FormatInt(w.trace.last, 10)},
		{4, "redraws", strconv.FormatUint(st.Redraws, 10)},
		{7, "layouts", strconv.FormatUint(st.Layouts, 10)},
		{18, "press timestamp", strconv.FormatInt(w.trace.startedNS, 10)},
		{19, "started sequence", strconv.FormatUint(w.trace.seq, 10)},
	} {
		if fields[c.index] != c.want {
			t.Errorf("field %d (%s) = %q, want %q", c.index, c.name, fields[c.index], c.want)
		}
	}
}
