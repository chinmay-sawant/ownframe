package main

import (
	"strings"
	"testing"
)

func TestPreviewTaskLists(t *testing.T) {
	t.Parallel()

	lines := []string{"- [ ] todo", "- [x] done", "- [X] finished", "- plain"}
	rows, _ := previewRows(lines, previewDefaultCols)

	// The renderer pads the block with a leading and a trailing gap row, so
	// the four list items sit between them.
	var items []previewRow

	for _, row := range rows {
		if row.Kind == "li" {
			items = append(items, row)
		}
	}

	if len(items) != 4 {
		t.Fatalf("task list items = %d, want 4", len(items))
	}

	if items[0].Marker != "☐" || lineText(items[0].Segments) != "todo" {
		t.Errorf("unchecked task = %q %q, want ☐ todo", items[0].Marker, lineText(items[0].Segments))
	}

	if items[1].Marker != "☑" || lineText(items[1].Segments) != "done" {
		t.Errorf("checked task = %q %q, want ☑ done", items[1].Marker, lineText(items[1].Segments))
	}

	if items[2].Marker != "☑" || lineText(items[2].Segments) != "finished" {
		t.Errorf("capital task = %q %q, want ☑ finished", items[2].Marker, lineText(items[2].Segments))
	}

	if items[3].Marker != "•" {
		t.Errorf("plain item marker = %q, want •", items[3].Marker)
	}

	for _, row := range rows {
		if strings.Contains(lineText(row.Segments), "[ ]") || strings.Contains(lineText(row.Segments), "[x]") {
			t.Errorf("row kept checkbox brackets: %q", lineText(row.Segments))
		}
	}
}

func TestPreviewHiddenSpan(t *testing.T) {
	t.Parallel()

	lines := []string{
		"Keep <span hidden>secret</span> this.",
		`<span style="display: none">gone</span>shown`,
		`<div hidden>buried</div>`,
		"After.",
	}
	rows, _ := previewRows(lines, previewDefaultCols)
	all := previewText(rows)

	for _, hidden := range []string{"secret", "gone", "buried"} {
		if strings.Contains(all, hidden) {
			t.Errorf("preview leaked hidden text %q: %q", hidden, all)
		}
	}

	for _, want := range []string{"Keep", "this.", "shown", "After."} {
		if !strings.Contains(all, want) {
			t.Errorf("preview lost visible text %q: %q", want, all)
		}
	}
}

func TestPreviewGapAfterList(t *testing.T) {
	t.Parallel()

	lines := []string{"- a", "- b", "para"}
	rows, anchors := previewRows(lines, previewDefaultCols)

	// The block is padded with a leading and a trailing gap; the transition
	// out of the list must add exactly one more between "b" and "para".
	var kinds []string

	for _, row := range rows {
		kinds = append(kinds, row.Kind)
	}

	want := []string{"gap", "li", "li", "gap", "p", "gap"}
	if len(kinds) != len(want) {
		t.Fatalf("row kinds = %v, want %v", kinds, want)
	}

	for i := range want {
		if kinds[i] != want[i] {
			t.Fatalf("row kinds = %v, want %v", kinds, want)
		}
	}

	if got := rows[anchors[2]].Kind; got != "p" {
		t.Fatalf("paragraph anchor kind = %q, want p", got)
	}
}
