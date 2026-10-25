package window

import (
	"strings"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

// devPathDisplay is one fill and one corner-shape path, so the Ops tab has a
// path row to label and count.
func devPathDisplay() *layout.Display {
	return &layout.Display{
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 10, Y: 10, W: 40, H: 20},
			{Kind: layout.DisplayOpPath, X: 10, Y: 40, W: 40, H: 20},
		},
		Order:          []int{0, 1},
		Width:          800,
		Height:         600,
		PointsPerPixel: 0.75,
	}
}

// TestDevOpsRowsCountAndLabelAPath checks that a corner-shape path reaches the
// DISPLAY counts, the total, and the paint-order list instead of disappearing
// from the tab.
func TestDevOpsRowsCountAndLabelAPath(t *testing.T) {
	t.Parallel()

	s := newDevShell(newDevScreen())
	s.display = devPathDisplay()
	s.dev.ops = true

	var count, total *devRow
	var painted []devRow

	all := s.devOpsRows()
	for i := range all {
		row := &all[i]
		plain := strings.TrimSpace(row.line.plain())

		switch {
		case row.act == devActOp:
			painted = append(painted, *row)
		case strings.HasPrefix(plain, "path"):
			count = row
		case strings.HasPrefix(plain, "total"):
			total = row
		}
	}

	if count == nil || !strings.HasSuffix(strings.TrimSpace(count.line.plain()), "1") {
		t.Fatalf("path count row = %v, want one path", count)
	}

	if total == nil || strings.TrimSpace(total.line.plain()) != "total 2" {
		t.Fatalf("total row = %v, want total 2", total)
	}

	if len(painted) != 2 {
		t.Fatalf("paint rows = %v, want two", painted)
	}

	if got := painted[1].line.plain(); !strings.Contains(got, "#2 path") {
		t.Fatalf("path row = %q, want #2 path", got)
	}

	if ink, ok := devOpInk(layout.DisplayOpPath); !ok || ink != devFillInk {
		t.Fatalf("path ink = %v (ok %v), want the fill ink", ink, ok)
	}
}
