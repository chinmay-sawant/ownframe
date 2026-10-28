package main

import (
	"strings"
	"testing"
)

func TestInlineCardForAdded(t *testing.T) {
	t.Parallel()

	const path = "examples/code-view/inline_card.go"

	card := inlineCardFor(240, 900, 700, 42, inlineLine{Kind: inlineAdded}, path, "const inlineCardWidth = 360")

	if !card.Show {
		t.Fatal("card is hidden, want shown")
	}

	if card.Left != 12 || card.Width != 360 {
		t.Fatalf("card box = %d px at left %d, want 360 px at 12", card.Width, card.Left)
	}

	if card.Top != 232 {
		t.Fatalf("card top = %d, want 232", card.Top)
	}

	if card.Title != "Line 42" {
		t.Fatalf("card title = %q, want Line 42", card.Title)
	}

	if card.Path != path {
		t.Fatalf("card path = %q, want %q", card.Path, path)
	}

	if card.Kind != cardKindAdded {
		t.Fatalf("card kind = %q, want %q", card.Kind, cardKindAdded)
	}

	if card.OldText != "" {
		t.Fatalf("old text = %q, want empty for an addition", card.OldText)
	}

	if card.NewText != "const inlineCardWidth = 360" {
		t.Fatalf("new text = %q", card.NewText)
	}

	if card.Hint == "" || len([]rune(card.Hint)) > 24 {
		t.Fatalf("hint = %q, want a short non-empty line", card.Hint)
	}
}

func TestInlineCardKindsAndSides(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		change  inlineLine
		want    string
		wantOld string
		wantNew string
	}{
		{
			name:    "added",
			change:  inlineLine{Kind: inlineAdded},
			want:    cardKindAdded,
			wantOld: "",
			wantNew: "fresh",
		},
		{
			name:    "modified",
			change:  inlineLine{Kind: inlineModified, OldText: "before"},
			want:    cardKindModified,
			wantOld: "before",
			wantNew: "fresh",
		},
		{
			name:    "deleted",
			change:  inlineLine{Kind: inlineDeleted, OldText: "removed"},
			want:    cardKindDeleted,
			wantOld: "removed",
			wantNew: "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			card := inlineCardFor(100, 900, 700, 3, c.change, "main.go", "fresh")
			if card.Kind != c.want {
				t.Fatalf("kind = %q, want %q", card.Kind, c.want)
			}

			if card.OldText != c.wantOld || card.NewText != c.wantNew {
				t.Fatalf("sides = %q / %q, want %q / %q", card.OldText, card.NewText, c.wantOld, c.wantNew)
			}
		})
	}
}

func TestInlineCardClampsTop(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		rowTop int
		pageH  int
		want   int
	}{
		{name: "lifted", rowTop: 300, pageH: 700, want: 292},
		{name: "top edge", rowTop: 4, pageH: 700, want: 8},
		{name: "zero row", rowTop: 0, pageH: 700, want: 8},
		{name: "bottom edge", rowTop: 800, pageH: 700, want: 580},
		{name: "tiny page", rowTop: 40, pageH: 100, want: 8},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			card := inlineCardFor(c.rowTop, 900, c.pageH, 7, inlineLine{}, "", "")
			if card.Top != c.want {
				t.Fatalf("top for row %d on a %d px page = %d, want %d", c.rowTop, c.pageH, card.Top, c.want)
			}

			upper := max(8, c.pageH-120)
			if card.Top < 8 || card.Top > upper {
				t.Fatalf("top %d left the [8,%d] clamp", card.Top, upper)
			}
		})
	}
}

func TestInlineCardTextStripsAndClips(t *testing.T) {
	t.Parallel()

	if got := inlineCardText("one\ntwo\tthree", 40); got != "one two three" {
		t.Fatalf("flattened text = %q, want one two three", got)
	}

	if got := inlineCardText("  padded\r\n", 40); got != "padded" {
		t.Fatalf("padded text = %q, want padded", got)
	}

	if got := inlineCardText("abcdefghijkl", 5); got != "abcde…" {
		t.Fatalf("clipped text = %q, want abcde…", got)
	}

	if got := inlineCardText("abcdefghij", 10); got != "abcdefghij" {
		t.Fatalf("exact-width text = %q, want no ellipsis", got)
	}

	if got := inlineCardText("anything", 0); got != "" {
		t.Fatalf("zero width = %q, want empty", got)
	}

	if got := inlineCardText("a\nb\tc", 40); strings.ContainsAny(got, "\r\n\t") {
		t.Fatalf("control character survived in %q", got)
	}
}
