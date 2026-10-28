package main

import "strings"

// inlineKind classifies one line of the open file against the saved version.
type inlineKind string

const (
	inlineNone     inlineKind = ""
	inlineAdded    inlineKind = "added"
	inlineModified inlineKind = "modified"
	inlineDeleted  inlineKind = "deleted"
)

// inlineLine is one line in a change map, with the text it replaced.
type inlineLine struct {
	Kind    inlineKind
	OldText string
}

// gutterMark is the marker the editor paints beside one line.
type gutterMark struct {
	Glyph string
	Class string
	Hint  string
}

const gutterHintLimit = 120

// gutterMarkFor maps one changed line to its gutter marker. A kind the viewer
// does not mark, including the zero kind, yields the zero mark.
func gutterMarkFor(kind inlineKind, line inlineLine) gutterMark {
	var mark gutterMark

	switch kind {
	case inlineAdded:
		mark = gutterMark{Glyph: "+", Class: "gutter-add", Hint: "Added line"}
	case inlineModified:
		mark = gutterMark{Glyph: "~", Class: "gutter-mod", Hint: "Modified line; old: " + line.OldText}
	case inlineDeleted:
		mark = gutterMark{Glyph: "-", Class: "gutter-del", Hint: "Deleted line; old: " + line.OldText}
	default:
		return gutterMark{}
	}

	mark.Hint = clipGutterHint(mark.Hint)

	return mark
}

// gutterMarks maps each changed line number to its marker. Every entry in
// changes gets a mark, so a line with no marker has the zero value.
func gutterMarks(changes map[int]inlineLine) map[int]gutterMark {
	if len(changes) == 0 {
		return nil
	}

	marks := make(map[int]gutterMark, len(changes))

	for no, line := range changes {
		marks[no] = gutterMarkFor(line.Kind, line)
	}

	return marks
}

// clipGutterHint flattens newlines and keeps the hint within gutterHintLimit
// runes.
func clipGutterHint(hint string) string {
	hint = strings.ReplaceAll(hint, "\r\n", " ")
	hint = strings.ReplaceAll(hint, "\r", " ")
	hint = strings.ReplaceAll(hint, "\n", " ")

	runes := []rune(hint)
	if len(runes) > gutterHintLimit {
		return string(runes[:gutterHintLimit])
	}

	return hint
}
