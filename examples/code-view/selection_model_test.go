package main

import "testing"

// selectionTestText is three lines of five, four, and five runes with no
// trailing newline. Its offsets are 0..16.
const selectionTestText = "alpha\nbeta\ngamma"

func TestSelectionRangeClamps(t *testing.T) {
	t.Parallel()

	const limit = 9

	cases := []struct {
		name  string
		sel   textSelection
		n     int
		start int
		end   int
	}{
		{name: "empty buffer", sel: textSelection{}, n: 0, start: 0, end: 0},
		{name: "collapsed", sel: textSelection{Anchor: 4, Caret: 4}, n: limit, start: 4, end: 4},
		{name: "forward", sel: textSelection{Anchor: 1, Caret: 5}, n: limit, start: 1, end: 5},
		{name: "backward", sel: textSelection{Anchor: 7, Caret: 2}, n: limit, start: 2, end: 7},
		{name: "past the end", sel: textSelection{Anchor: 3, Caret: 40}, n: limit, start: 3, end: limit},
		{name: "negative anchor", sel: textSelection{Anchor: -6, Caret: 3}, n: limit, start: 0, end: 3},
		{name: "both outside", sel: textSelection{Anchor: -6, Caret: 40}, n: limit, start: 0, end: limit},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			start, end := tc.sel.Range(tc.n)
			if start != tc.start || end != tc.end {
				t.Fatalf("Range(%d) = (%d, %d), want (%d, %d)", tc.n, start, end, tc.start, tc.end)
			}
		})
	}
}

func TestSelectionFromPoints(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                string
		text                string
		startLine, startCol int
		endLine, endCol     int
		want                textSelection
	}{
		{
			name: "collapsed", text: selectionTestText,
			startLine: 1, startCol: 2, endLine: 1, endCol: 2,
			want: textSelection{Anchor: 8, Caret: 8},
		},
		{
			name: "inside one line", text: selectionTestText,
			startLine: 0, startCol: 0, endLine: 0, endCol: 3,
			want: textSelection{Anchor: 0, Caret: 3},
		},
		{
			name: "forward across lines", text: selectionTestText,
			startLine: 0, startCol: 2, endLine: 2, endCol: 3,
			want: textSelection{Anchor: 2, Caret: 14},
		},
		{
			name: "backward across lines", text: selectionTestText,
			startLine: 2, startCol: 3, endLine: 0, endCol: 2,
			want: textSelection{Anchor: 14, Caret: 2},
		},
		{
			name: "column past line end", text: selectionTestText,
			startLine: 0, startCol: 99, endLine: 1, endCol: 99,
			want: textSelection{Anchor: 5, Caret: 10},
		},
		{
			name: "line past buffer", text: selectionTestText,
			startLine: 9, startCol: 0, endLine: 9, endCol: 9,
			want: textSelection{Anchor: 11, Caret: 16},
		},
		{
			name: "negative point", text: selectionTestText,
			startLine: -2, startCol: -3, endLine: -1, endCol: -1,
			want: textSelection{Anchor: 0, Caret: 0},
		},
		{
			name: "whole buffer", text: selectionTestText,
			startLine: 0, startCol: 0, endLine: 2, endCol: 5,
			want: textSelection{Anchor: 0, Caret: 16},
		},
		{
			name: "empty buffer", text: "",
			startLine: 0, startCol: 0, endLine: 0, endCol: 0,
			want: textSelection{Anchor: 0, Caret: 0},
		},
		{
			name: "crlf pair is one break", text: "a\r\nb",
			startLine: 0, startCol: 1, endLine: 1, endCol: 1,
			want: textSelection{Anchor: 1, Caret: 4},
		},
		{
			name: "lone cr is line text", text: "a\rb",
			startLine: 0, startCol: 3, endLine: 0, endCol: 2,
			want: textSelection{Anchor: 3, Caret: 2},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := selectionFromPoints(tc.text, tc.startLine, tc.startCol, tc.endLine, tc.endCol)
			if got != tc.want {
				t.Fatalf("selectionFromPoints(%q, %d, %d, %d, %d) = %+v, want %+v",
					tc.text, tc.startLine, tc.startCol, tc.endLine, tc.endCol, got, tc.want)
			}
		})
	}
}

func TestSelectionLineRange(t *testing.T) {
	t.Parallel()

	n := len([]rune(selectionTestText))

	cases := []struct {
		name string
		text string
		sel  textSelection
		line int
		from int
		to   int
		ok   bool
	}{
		{name: "whole buffer line 0", text: selectionTestText, sel: textSelection{Anchor: 0, Caret: n}, line: 0, from: 0, to: 5, ok: true},
		{name: "whole buffer line 1", text: selectionTestText, sel: textSelection{Anchor: 0, Caret: n}, line: 1, from: 0, to: 4, ok: true},
		{name: "whole buffer line 2", text: selectionTestText, sel: textSelection{Anchor: 0, Caret: n}, line: 2, from: 0, to: 5, ok: true},
		{name: "first partial", text: selectionTestText, sel: textSelection{Anchor: 0, Caret: 3}, line: 0, from: 0, to: 3, ok: true},
		{name: "stops at line start", text: selectionTestText, sel: textSelection{Anchor: 3, Caret: 6}, line: 0, from: 3, to: 5, ok: true},
		{name: "line after the end", text: selectionTestText, sel: textSelection{Anchor: 3, Caret: 6}, line: 1},
		{name: "into next line", text: selectionTestText, sel: textSelection{Anchor: 3, Caret: 7}, line: 1, from: 0, to: 1, ok: true},
		{name: "reversed", text: selectionTestText, sel: textSelection{Anchor: 10, Caret: 6}, line: 1, from: 0, to: 4, ok: true},
		{name: "collapsed", text: selectionTestText, sel: textSelection{Anchor: 6, Caret: 6}, line: 1},
		{name: "break only", text: selectionTestText, sel: textSelection{Anchor: 5, Caret: 6}, line: 0},
		{name: "clamped", text: selectionTestText, sel: textSelection{Anchor: -5, Caret: n + 20}, line: 2, from: 0, to: 5, ok: true},
		{name: "negative line", text: selectionTestText, sel: textSelection{Anchor: 0, Caret: n}, line: -1},
		{name: "line past buffer", text: selectionTestText, sel: textSelection{Anchor: 0, Caret: n}, line: 3},
		{name: "crlf first line", text: "a\r\nb", sel: textSelection{Anchor: 0, Caret: 4}, line: 0, from: 0, to: 1, ok: true},
		{name: "crlf second line", text: "a\r\nb", sel: textSelection{Anchor: 0, Caret: 4}, line: 1, from: 0, to: 1, ok: true},
		{name: "crlf before the break", text: "a\r\nb", sel: textSelection{Anchor: 0, Caret: 2}, line: 1},
		{name: "trailing empty line", text: "a\n", sel: textSelection{Anchor: 0, Caret: 2}, line: 1},
		{name: "lone cr is line text", text: "a\rb", sel: textSelection{Anchor: 0, Caret: 3}, line: 0, from: 0, to: 3, ok: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			from, to, ok := selectionLineRange(tc.text, tc.sel, tc.line)
			if from != tc.from || to != tc.to || ok != tc.ok {
				t.Fatalf("selectionLineRange(%q, %+v, %d) = (%d, %d, %t), want (%d, %d, %t)",
					tc.text, tc.sel, tc.line, from, to, ok, tc.from, tc.to, tc.ok)
			}
		})
	}
}

func TestSelectionText(t *testing.T) {
	t.Parallel()

	n := len([]rune(selectionTestText))

	cases := []struct {
		name string
		text string
		sel  textSelection
		want string
	}{
		{name: "first line", text: selectionTestText, sel: textSelection{Anchor: 0, Caret: 5}, want: "alpha"},
		{name: "second line", text: selectionTestText, sel: textSelection{Anchor: 6, Caret: 10}, want: "beta"},
		{name: "whole buffer", text: selectionTestText, sel: textSelection{Anchor: 0, Caret: n}, want: selectionTestText},
		{name: "reversed", text: selectionTestText, sel: textSelection{Anchor: 10, Caret: 6}, want: "beta"},
		{name: "clamped", text: selectionTestText, sel: textSelection{Anchor: -4, Caret: n + 30}, want: selectionTestText},
		{name: "collapsed", text: selectionTestText, sel: textSelection{Anchor: 4, Caret: 4}, want: ""},
		{name: "newline only", text: selectionTestText, sel: textSelection{Anchor: 5, Caret: 6}, want: "\n"},
		{name: "crlf kept", text: "a\r\nb", sel: textSelection{Anchor: 0, Caret: 4}, want: "a\r\nb"},
		{name: "empty buffer", text: "", sel: textSelection{Anchor: 0, Caret: 9}, want: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := selectionText(tc.text, tc.sel); got != tc.want {
				t.Fatalf("selectionText(%q, %+v) = %q, want %q", tc.text, tc.sel, got, tc.want)
			}
		})
	}
}

func TestWordAndLineSelection(t *testing.T) {
	t.Parallel()

	t.Run("word", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name   string
			text   string
			offset int
			start  int
			end    int
		}{
			{name: "inside a word", text: selectionTestText, offset: 7, start: 6, end: 10},
			{name: "last word", text: selectionTestText, offset: 11, start: 11, end: 16},
			{name: "first word", text: selectionTestText, offset: 0, start: 0, end: 5},
			{name: "end of buffer", text: selectionTestText, offset: 16, start: 11, end: 16},
			{name: "on a space", text: "alpha beta gamma", offset: 5, start: 5, end: 6},
			{name: "on a tab", text: "a\tb", offset: 1, start: 1, end: 2},
			{name: "underscore keeps the word", text: "foo_bar baz", offset: 4, start: 0, end: 7},
			{name: "punctuation rune", text: "foo.bar", offset: 3, start: 3, end: 4},
			{name: "after punctuation", text: "foo.bar", offset: 4, start: 4, end: 7},
			{name: "empty buffer", text: "", offset: 0, start: 0, end: 0},
			{name: "past buffer", text: "abc", offset: 99, start: 0, end: 3},
			{name: "line break", text: "abc\ndef", offset: 3, start: 0, end: 3},
			{name: "next line", text: "abc\ndef", offset: 4, start: 4, end: 7},
			{name: "blank line", text: "a\n\nb", offset: 2, start: 2, end: 2},
			{name: "crlf looks back", text: "ab\r\ncd", offset: 2, start: 0, end: 2},
			{name: "crlf next line", text: "ab\r\ncd", offset: 4, start: 4, end: 6},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				start, end := wordSelection(tc.text, tc.offset)
				if start != tc.start || end != tc.end {
					t.Fatalf("wordSelection(%q, %d) = (%d, %d), want (%d, %d)",
						tc.text, tc.offset, start, end, tc.start, tc.end)
				}
			})
		}
	})

	t.Run("line", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name   string
			text   string
			offset int
			start  int
			end    int
		}{
			{name: "first line", text: selectionTestText, offset: 0, start: 0, end: 5},
			{name: "at the break", text: selectionTestText, offset: 5, start: 0, end: 5},
			{name: "second line", text: selectionTestText, offset: 6, start: 6, end: 10},
			{name: "second break", text: selectionTestText, offset: 10, start: 6, end: 10},
			{name: "last line", text: selectionTestText, offset: 11, start: 11, end: 16},
			{name: "past buffer", text: selectionTestText, offset: 99, start: 11, end: 16},
			{name: "negative offset", text: selectionTestText, offset: -4, start: 0, end: 5},
			{name: "trailing empty line", text: "alpha\n", offset: 6, start: 6, end: 6},
			{name: "crlf second line", text: "a\r\nb", offset: 4, start: 3, end: 4},
			{name: "crlf first line", text: "a\r\nb", offset: 1, start: 0, end: 1},
			{name: "lone cr is line text", text: "a\rb", offset: 2, start: 0, end: 3},
			{name: "empty buffer", text: "", offset: 0, start: 0, end: 0},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				start, end := lineSelection(tc.text, tc.offset)
				if start != tc.start || end != tc.end {
					t.Fatalf("lineSelection(%q, %d) = (%d, %d), want (%d, %d)",
						tc.text, tc.offset, start, end, tc.start, tc.end)
				}
			})
		}
	})
}

func TestPointForOffset(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		text   string
		offset int
		line   int
		col    int
	}{
		{name: "start", text: "alpha\nbeta", offset: 0, line: 0, col: 0},
		{name: "line end", text: "alpha\nbeta", offset: 5, line: 0, col: 5},
		{name: "next line", text: "alpha\nbeta", offset: 6, line: 1, col: 0},
		{name: "last rune", text: "alpha\nbeta", offset: 10, line: 1, col: 4},
		{name: "past the end", text: "alpha\nbeta", offset: 99, line: 1, col: 4},
		{name: "negative offset", text: "alpha\nbeta", offset: -3, line: 0, col: 0},
		{name: "trailing newline", text: "alpha\n", offset: 6, line: 1, col: 0},
		{name: "empty buffer", text: "", offset: 0, line: 0, col: 0},
		{name: "empty buffer past the end", text: "", offset: 5, line: 0, col: 0},
		{name: "crlf before the pair", text: "a\r\nb", offset: 1, line: 0, col: 1},
		{name: "crlf inside the pair", text: "a\r\nb", offset: 2, line: 1, col: 0},
		{name: "crlf next line", text: "a\r\nb", offset: 3, line: 1, col: 0},
		{name: "crlf last rune", text: "a\r\nb", offset: 4, line: 1, col: 1},
		{name: "lone cr is a column", text: "a\rb", offset: 2, line: 0, col: 2},
		{name: "third line start", text: "one\ntwo\nthree", offset: 8, line: 2, col: 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			line, col := pointForOffset(tc.text, tc.offset)
			if line != tc.line || col != tc.col {
				t.Fatalf("pointForOffset(%q, %d) = (%d, %d), want (%d, %d)",
					tc.text, tc.offset, line, col, tc.line, tc.col)
			}
		})
	}

	t.Run("round trip", func(t *testing.T) {
		t.Parallel()

		text := "one\nbc\nd"
		for offset := 0; offset <= len([]rune(text)); offset++ {
			line, col := pointForOffset(text, offset)
			if got := selectionFromPoints(text, line, col, line, col).Anchor; got != offset {
				t.Fatalf("offset %d became line %d column %d and back to %d", offset, line, col, got)
			}
		}
	})
}
