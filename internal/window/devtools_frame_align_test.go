package window

import (
	"strings"
	"testing"
)

// TestDevFrameRightAligned checks the value is a right-aligned span that
// ends at the row width.
func TestDevFrameRightAligned(t *testing.T) {
	t.Parallel()

	s := newDevShell(newDevScreen())
	s.dev.stats.Redraws = 12

	for _, row := range s.devFrameRows(320) {
		if !strings.HasPrefix(row.line.plain(), "Redraws ") {
			continue
		}
		last := row.line.spans[len(row.line.spans)-1]
		if last.text != "12" || !last.right || last.w <= 0 {
			t.Fatalf("redraws value span = %+v, want a right-aligned sized span", last)
		}
		if row.w != 320 {
			t.Fatalf("row width = %v, want 320", row.w)
		}

		return
	}
	t.Fatal("no redraws row")
}
