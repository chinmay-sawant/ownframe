package textrun

import (
	"strings"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

// longLine builds one baseline of eight 448-rune runs for the caret path.
func longLine() line {
	seg := strings.Repeat("abcdef ", 64)
	runs := make([]run, 0, 8)
	for range 8 {
		runs = append(runs, run{op: &layout.DisplayOp{Text: seg}})
	}
	return line{baseline: 10, runs: runs}
}

// BenchmarkCaretLongRun times the caret path on a 3584-rune line: the
// lineText join and the offsetIn fallback. No font means faceFor is nil,
// so offsetIn measures the rune-slice fallback, not shaped widths.
func BenchmarkCaretLongRun(b *testing.B) {
	ln := longLine()
	op := ln.runs[0].op
	b.ReportAllocs()
	b.Run("lineText", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = lineText(ln)
		}
	})
	b.Run("offsetIn", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = offsetIn(op, 10, 1)
		}
	})
}
