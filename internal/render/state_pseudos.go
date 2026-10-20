package render

import "github.com/chinmay-sawant/blinkless/css"

// StatePseudos records which interactive pseudo-classes the rendered sheets
// can match. A false field means recording that state cannot change paint, so
// the page can skip the redraw.
type StatePseudos struct {
	Hover  bool
	Active bool
	Focus  bool
}

// allStatePseudos keeps the old redraw behavior when the engine cannot report
// which states it uses.
var allStatePseudos = StatePseudos{Hover: true, Active: true, Focus: true}

// statePseudoReporter is the engine hook. It is an optional interface so a
// pinned engine without it still builds and gets the conservative default.
type statePseudoReporter interface {
	UsesStatePseudo(name string) bool
}

// statePseudosOf reads the states a styled document can match.
func statePseudosOf(doc *css.Document) StatePseudos {
	reporter, ok := any(doc).(statePseudoReporter)
	if !ok {
		return allStatePseudos
	}

	return StatePseudos{
		Hover:  reporter.UsesStatePseudo("hover"),
		Active: reporter.UsesStatePseudo("active"),
		Focus:  reporter.UsesStatePseudo("focus"),
	}
}
