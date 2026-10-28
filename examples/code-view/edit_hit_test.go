package main

import "testing"

func TestVisualRowAt(t *testing.T) {
	t.Parallel()

	const bodyTop = 110

	cases := []struct {
		name    string
		y       int
		scrollY int
		rowH    int
		want    int
	}{
		{name: "top of body", y: bodyTop, rowH: rowH, want: 0},
		{name: "last pixel of first row", y: bodyTop + rowH - 1, rowH: rowH, want: 0},
		{name: "second row", y: bodyTop + rowH, rowH: rowH, want: 1},
		{name: "scrolled rows", y: bodyTop, scrollY: 3 * rowH, rowH: rowH, want: 3},
		{name: "scrolled partial row", y: bodyTop, scrollY: rowH + 4, rowH: rowH, want: 1},
		{name: "above body", y: bodyTop - 1, rowH: rowH, want: 0},
		{name: "negative y", y: -40, rowH: rowH, want: 0},
		{name: "scroll above content", y: bodyTop, scrollY: -rowH, rowH: rowH, want: 0},
		{name: "zero row height", y: bodyTop, rowH: 0, want: 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := visualRowAt(tc.y, bodyTop, tc.scrollY, tc.rowH); got != tc.want {
				t.Fatalf("visualRowAt(%d, %d, %d, %d) = %d, want %d",
					tc.y, bodyTop, tc.scrollY, tc.rowH, got, tc.want)
			}
		})
	}
}

// hitTestDoc builds three source lines where the middle line wraps onto two
// visual rows at ten columns.
func hitTestDoc() *fileDoc {
	return &fileDoc{Tokens: tokenize(langPlain, []string{"alpha", "one two three four", "omega"})}
}

func TestSourceLineAtWrappedLine(t *testing.T) {
	t.Parallel()

	const cols = 10

	doc := hitTestDoc()
	if rows := len(doc.wrapped(cols).Rows); rows != 4 {
		t.Fatalf("wrapped rows = %d, want 4", rows)
	}

	cases := []struct {
		name string
		row  int
		want int
	}{
		{name: "first line", row: 0, want: 0},
		{name: "wrapped line start", row: 1, want: 1},
		{name: "wrapped continuation", row: 2, want: 1},
		{name: "last line", row: 3, want: 2},
		{name: "one past plan", row: 4, want: 2},
		{name: "far past plan", row: 99, want: 2},
		{name: "negative row", row: -1, want: 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := sourceLineAt(doc, cols, tc.row); got != tc.want {
				t.Fatalf("sourceLineAt(row %d) = %d, want %d", tc.row, got, tc.want)
			}
		})
	}
}

func TestEditScrollForLine(t *testing.T) {
	t.Parallel()

	const cols = 10

	doc := hitTestDoc()
	rows := len(doc.wrapped(cols).Rows)

	cases := []struct {
		name string
		line int
		want int
	}{
		{name: "first line", line: 0, want: 0},
		{name: "wrapped line", line: 1, want: rowH},
		{name: "last line", line: 2, want: 3 * rowH},
		{name: "past last line", line: 99, want: rows * rowH},
		{name: "negative line", line: -3, want: 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := editScrollForLine(doc, cols, tc.line); got != tc.want {
				t.Fatalf("editScrollForLine(line %d) = %d, want %d", tc.line, got, tc.want)
			}
		})
	}
}

func TestEditHitNilDoc(t *testing.T) {
	t.Parallel()

	if got := sourceLineAt(nil, 80, 12); got != 0 {
		t.Fatalf("sourceLineAt(nil) = %d, want 0", got)
	}

	if got := editScrollForLine(nil, 80, 12); got != 0 {
		t.Fatalf("editScrollForLine(nil) = %d, want 0", got)
	}
}
