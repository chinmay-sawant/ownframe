package main

import (
	"slices"
	"testing"
)

func TestEditBufferRoundTrip(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		raw   string
		text  string
		lines []string
	}{
		{
			name:  "crlf with final newline",
			raw:   "alpha\r\nbeta\r\n",
			text:  "alpha\nbeta\n",
			lines: []string{"alpha", "beta"},
		},
		{
			name:  "crlf without final newline",
			raw:   "alpha\r\nbeta",
			text:  "alpha\nbeta\n",
			lines: []string{"alpha", "beta"},
		},
		{
			name:  "lf with final newline",
			raw:   "one\ntwo\n",
			text:  "one\ntwo\n",
			lines: []string{"one", "two"},
		},
		{
			name:  "empty file",
			raw:   "",
			text:  "",
			lines: nil,
		},
		{
			name:  "one blank line",
			raw:   "\n",
			text:  "\n",
			lines: []string{""},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			buf := newEditBuffer("main.go", []byte(tc.raw))
			if buf.Path != "main.go" {
				t.Errorf("path = %q", buf.Path)
			}

			if buf.Dirty {
				t.Error("new buffer is dirty")
			}

			if got := buf.Text(); got != tc.text {
				t.Errorf("Text() = %q, want %q", got, tc.text)
			}

			if !slices.Equal(buf.Lines, tc.lines) {
				t.Errorf("Lines = %q, want %q", buf.Lines, tc.lines)
			}

			if buf.Original != tc.text {
				t.Errorf("Original = %q, want %q", buf.Original, tc.text)
			}

			if got := string(buf.Save()); got != tc.text {
				t.Errorf("Save() = %q, want %q", got, tc.text)
			}
		})
	}
}

func TestEditBufferDirtyTracking(t *testing.T) {
	t.Parallel()

	buf := newEditBuffer("main.go", []byte("alpha\nbeta\n"))
	if buf.Dirty {
		t.Fatal("new buffer is dirty")
	}

	buf.SetText("alpha\nbeta\n")
	if buf.Dirty {
		t.Fatal("identical text marked dirty")
	}

	buf.SetText("alpha\r\nbeta\r\n")
	if buf.Dirty {
		t.Fatal("CRLF text marked dirty")
	}

	buf.SetText("alpha\nbeta")
	if buf.Dirty {
		t.Fatal("text without the final newline marked dirty")
	}

	buf.SetText("alpha\ngamma\n")
	if !buf.Dirty {
		t.Fatal("changed text is not dirty")
	}

	if got := buf.LineCount(); got != 2 {
		t.Fatalf("line count = %d", got)
	}

	if got := buf.Text(); got != "alpha\ngamma\n" {
		t.Fatalf("text = %q", got)
	}

	buf.MarkSaved()
	if buf.Dirty {
		t.Fatal("saved buffer is dirty")
	}

	if buf.Original != "alpha\ngamma\n" {
		t.Fatalf("original = %q", buf.Original)
	}

	buf.SetText("alpha\ngamma\n")
	if buf.Dirty {
		t.Fatal("text equal to the saved text marked dirty")
	}

	buf.SetText("alpha\nbeta\n")
	if !buf.Dirty {
		t.Fatal("text different from the saved text is not dirty")
	}
}

func TestEditBufferLineCount(t *testing.T) {
	t.Parallel()

	cases := map[string]int{
		"":           0,
		"one":        1,
		"one\n":      1,
		"one\ntwo":   2,
		"one\ntwo\n": 2,
		"one\n\n":    2,
		"\n":         1,
	}

	for raw, want := range cases {
		buf := newEditBuffer("x.go", []byte(raw))
		if got := buf.LineCount(); got != want {
			t.Errorf("LineCount(%q) = %d, want %d", raw, got, want)
		}
	}
}
