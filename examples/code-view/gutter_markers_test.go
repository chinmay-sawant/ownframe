package main

import (
	"strings"
	"testing"
)

func TestGutterMarkForKinds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		kind inlineKind
		line inlineLine
		want gutterMark
	}{
		{
			name: "added",
			kind: inlineAdded,
			want: gutterMark{Glyph: "+", Class: "gutter-add", Hint: "Added line"},
		},
		{
			name: "modified",
			kind: inlineModified,
			line: inlineLine{OldText: "const total = 1"},
			want: gutterMark{Glyph: "~", Class: "gutter-mod", Hint: "Modified line; old: const total = 1"},
		},
		{
			name: "deleted",
			kind: inlineDeleted,
			line: inlineLine{OldText: "return nil"},
			want: gutterMark{Glyph: "-", Class: "gutter-del", Hint: "Deleted line; old: return nil"},
		},
		{
			name: "none",
			kind: inlineNone,
			want: gutterMark{},
		},
		{
			name: "unknown",
			kind: inlineKind("renamed"),
			line: inlineLine{OldText: "old name"},
			want: gutterMark{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := gutterMarkFor(tt.kind, tt.line); got != tt.want {
				t.Fatalf("gutterMarkFor(%q, %+v) = %+v, want %+v", tt.kind, tt.line, got, tt.want)
			}
		})
	}
}

func TestGutterMarksBuildsMap(t *testing.T) {
	t.Parallel()

	changes := map[int]inlineLine{
		2: {Kind: inlineAdded},
		5: {Kind: inlineModified, OldText: "before"},
		8: {Kind: inlineDeleted, OldText: "removed"},
		9: {Kind: inlineNone},
	}

	marks := gutterMarks(changes)
	if len(marks) != len(changes) {
		t.Fatalf("marks = %d, want %d", len(marks), len(changes))
	}

	want := map[int]gutterMark{
		2: {Glyph: "+", Class: "gutter-add", Hint: "Added line"},
		5: {Glyph: "~", Class: "gutter-mod", Hint: "Modified line; old: before"},
		8: {Glyph: "-", Class: "gutter-del", Hint: "Deleted line; old: removed"},
		9: {},
	}

	for no, wantMark := range want {
		if got := marks[no]; got != wantMark {
			t.Errorf("line %d mark = %+v, want %+v", no, got, wantMark)
		}
	}

	if got := marks[3]; got != (gutterMark{}) {
		t.Errorf("missing line mark = %+v, want zero", got)
	}

	if got := gutterMarks(nil); got != nil {
		t.Errorf("gutterMarks(nil) = %v, want nil", got)
	}
}

func TestGutterHintIsClipped(t *testing.T) {
	t.Parallel()

	const modPrefix = "Modified line; old: "

	long := gutterMarkFor(inlineModified, inlineLine{OldText: strings.Repeat("x", gutterHintLimit+40)})
	if got := len([]rune(long.Hint)); got != gutterHintLimit {
		t.Fatalf("hint runes = %d, want %d", got, gutterHintLimit)
	}

	if !strings.HasPrefix(long.Hint, modPrefix+"xx") {
		t.Fatalf("hint lost its prefix: %q", long.Hint)
	}

	multibyte := gutterMarkFor(inlineDeleted, inlineLine{OldText: strings.Repeat("é", gutterHintLimit)})
	if got := len([]rune(multibyte.Hint)); got != gutterHintLimit {
		t.Fatalf("multibyte hint runes = %d, want %d", got, gutterHintLimit)
	}

	wrapped := gutterMarkFor(inlineModified, inlineLine{OldText: "one\ntwo\r\nthree\rfour"})
	if strings.ContainsAny(wrapped.Hint, "\r\n") {
		t.Fatalf("hint kept a newline: %q", wrapped.Hint)
	}

	if want := modPrefix + "one two three four"; wrapped.Hint != want {
		t.Fatalf("hint = %q, want %q", wrapped.Hint, want)
	}
}
