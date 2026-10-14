package window

import (
	"strings"
	"testing"
)

// devFrameText joins the rows' plain text.
func devFrameText(rows []devRow) string {
	var lines []string
	for _, row := range rows {
		lines = append(lines, row.line.plain())
	}
	return strings.Join(lines, "\n")
}

// TestDevFrameSections checks the tab's four sections.
func TestDevFrameSections(t *testing.T) {
	t.Parallel()

	s := newDevShell(newDevScreen())
	got := devFrameText(s.devFrameRows(320))

	for _, want := range []string{"WINDOW", "Window", "RENDERING", "PIPELINE", "RELOAD", "Redraws", "0.0ms"} {
		if !strings.Contains(got, want) {
			t.Fatalf("frame rows = %q, want %q", got, want)
		}
	}
}

// TestDevFrameReloadErrorInk checks the error ink.
func TestDevFrameReloadErrorInk(t *testing.T) {
	t.Parallel()

	s := newDevShell(newDevScreen())
	s.dev.stats.LastReloadError = "boom"

	for _, row := range s.devFrameRows(320) {
		if !strings.Contains(row.line.plain(), "boom") {
			continue
		}
		for _, span := range row.line.spans {
			if span.ink == devErrInk {
				return
			}
		}
		t.Fatalf("reload error row %q has no error ink", row.line.plain())
	}
	t.Fatal("no reload error row")
}

// TestDevFrameStable checks two calls agree.
func TestDevFrameStable(t *testing.T) {
	t.Parallel()

	s := newDevShell(newDevScreen())
	first := devFrameText(s.devFrameRows(320))
	if second := devFrameText(s.devFrameRows(320)); first != second {
		t.Fatalf("rows changed between calls:\n%q\n%q", first, second)
	}
}
