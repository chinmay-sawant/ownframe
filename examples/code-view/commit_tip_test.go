package main

import (
	"strings"
	"testing"
)

func TestCommitTipLayoutRightOfSidebar(t *testing.T) {
	t.Parallel()

	got := layoutCommitTip(1500, 900, 46, 268, 300, 42, 4, 1)

	if !got.Show {
		t.Fatal("tip is hidden on a hovered row")
	}

	if want := 46 + 268 + commitTipGap; got.Left != want {
		t.Fatalf("left = %d, want %d", got.Left, want)
	}

	if got.Width != commitTipWidth {
		t.Fatalf("width = %d, want %d", got.Width, commitTipWidth)
	}

	if want := 300 - commitTipLift; got.Top != want {
		t.Fatalf("top = %d, want %d", got.Top, want)
	}

	if want := commitTipPadH + 4*commitTipLineH; got.Height != want {
		t.Fatalf("height = %d, want %d", got.Height, want)
	}

	if got.Left+got.Width+commitTipGap > 1500 {
		t.Fatalf("tip overflows the page: left %d width %d", got.Left, got.Width)
	}

	if got := layoutCommitTip(1500, 900, 46, 268, 0, 42, 1, 1); got.Top != commitTipEdge {
		t.Fatalf("top = %d, want the %d px floor", got.Top, commitTipEdge)
	}
}

func TestCommitTipClampsBottom(t *testing.T) {
	t.Parallel()

	const pageH = 320

	got := layoutCommitTip(1500, pageH, 46, 268, 300, 42, 9, 1)
	if want := commitTipPadH + 9*commitTipLineH; got.Height != want {
		t.Fatalf("height = %d, want %d", got.Height, want)
	}

	if want := pageH - got.Height - commitTipEdge; got.Top != want {
		t.Fatalf("top = %d, want %d", got.Top, want)
	}

	if got.Top+got.Height > pageH-commitTipEdge {
		t.Fatalf("tip bottom = %d, want at most %d", got.Top+got.Height, pageH-commitTipEdge)
	}

	if got := layoutCommitTip(1500, pageH, 46, 268, 10, 42, 40, 1); got.Height != commitTipMaxH {
		t.Fatalf("height = %d, want the %d px cap", got.Height, commitTipMaxH)
	}
}

func TestCommitTipFallsBackInsideWhenNarrow(t *testing.T) {
	t.Parallel()

	const (
		pageW    = 620
		sideLeft = 46
		sideW    = 268
	)

	got := layoutCommitTip(pageW, 900, sideLeft, sideW, 300, 42, 3, 1)
	if !got.Show {
		t.Fatal("tip is hidden on a narrow page")
	}

	if want := sideLeft + commitTipInset; got.Left != want {
		t.Fatalf("left = %d, want %d inside the sidebar", got.Left, want)
	}

	if got.Left < sideLeft || got.Left >= sideLeft+sideW {
		t.Fatalf("left = %d, want the sidebar's left edge at %d", got.Left, sideLeft)
	}

	if got.Left+got.Width+commitTipGap > pageW {
		t.Fatalf("tip overflows the page: left %d width %d", got.Left, got.Width)
	}

	if want := 300 - commitTipLift; got.Top != want {
		t.Fatalf("top = %d, want the row band at %d", got.Top, want)
	}

	if got := layoutCommitTip(0, 900, sideLeft, sideW, 300, 42, 3, 1); got.Show {
		t.Fatal("tip shows on an empty page")
	}

	if got := layoutCommitTip(pageW, 900, sideLeft, sideW, 300, 42, 0, 1); got.Show {
		t.Fatal("tip shows without text")
	}
}

func TestCommitTipTextWraps(t *testing.T) {
	t.Parallel()

	const width = 24

	subject := "Parse the commit body into wrapped lines for the hover card"
	body := strings.Repeat("The body text keeps a bounded number of lines. ", 12)
	meta := "abc1234 · Ada Lovelace · Tue Oct 27 2026"

	lines := commitTipText(subject, body, meta, width)
	if len(lines) == 0 {
		t.Fatal("no lines for a full card")
	}

	if most := commitTipSubjectLines + commitTipBodyLines + commitTipMetaLines; len(lines) > most {
		t.Fatalf("lines = %d, want at most %d", len(lines), most)
	}

	for i, line := range lines {
		if n := len([]rune(line)); n > width {
			t.Fatalf("line %d = %q is %d runes, want at most %d", i, line, n, width)
		}

		if strings.TrimSpace(line) == "" {
			t.Fatalf("line %d is blank", i)
		}
	}

	if lines[0] == subject {
		t.Fatal("subject did not wrap")
	}

	if !strings.HasSuffix(lines[len(lines)-1], "…") {
		t.Fatalf("meta line = %q, want a cut", lines[len(lines)-1])
	}

	if got := commitTipText("", "", "", width); got != nil {
		t.Fatalf("empty card = %+v, want none", got)
	}

	if got := commitTipText(subject, body, meta, 0); got != nil {
		t.Fatalf("zero width = %+v, want none", got)
	}

	long := commitTipText(strings.Repeat("x", 35), "", "", 10)
	if len(long) != commitTipSubjectLines {
		t.Fatalf("hard-break lines = %d, want %d", len(long), commitTipSubjectLines)
	}

	for i, line := range long {
		if n := len([]rune(line)); n > 10 {
			t.Fatalf("hard-break line %d = %q is %d runes", i, line, n)
		}
	}
}

func TestHistorySubjectKeepsWordsOnTwoLines(t *testing.T) {
	t.Parallel()

	subject := "Fix the history subject wrapping so a long commit message uses the full row width"
	lines := historySubjectLines(subject)
	if len(lines) != 2 {
		t.Fatalf("lines = %#v, want two", lines)
	}

	words := strings.Fields(subject)
	seen := 0
	for i, line := range lines {
		if n := len([]rune(line)); n > historySubjectCols+1 {
			t.Fatalf("line %d = %q is %d runes", i, line, n)
		}
		for _, tok := range strings.Fields(strings.TrimSuffix(line, "…")) {
			if seen >= len(words) || tok != words[seen] {
				t.Fatalf("line %d token %q, want %q", i, tok, words[seen])
			}
			seen++
		}
	}
	if seen < 8 {
		t.Fatalf("only %d words wrapped from %q", seen, lines)
	}
}

func TestCommitTipColsFillTheCard(t *testing.T) {
	t.Parallel()

	cols := commitTipCols(commitTipWidth, commitTipCell, commitTipPadX)
	if cols < 60 {
		t.Fatalf("cols = %d, want the hover card filled", cols)
	}

	lines := commitTipText(
		"Fix the history subject wrapping so a long commit message uses the full row width",
		"The hover card should wrap on word boundaries and sit against the text instead of leaving a wide empty gutter.",
		"abc1234 · Ada Lovelace · Tue Oct 27",
		cols,
	)
	if len(lines) < 2 {
		t.Fatalf("lines = %#v", lines)
	}
	for i, line := range lines {
		if n := len([]rune(line)); n > cols {
			t.Fatalf("line %d = %q is %d runes, want at most %d", i, line, n, cols)
		}
	}
}
