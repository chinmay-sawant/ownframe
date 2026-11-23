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

// allStatePseudos is the conservative answer for a cache with no styled
// document yet: every state may change paint.
var allStatePseudos = StatePseudos{Hover: true, Active: true, Focus: true}

// statePseudosOf reads the states a styled document can match.
func statePseudosOf(doc *css.Document) StatePseudos {
	return StatePseudos{
		Hover:  doc.UsesStatePseudo("hover"),
		Active: doc.UsesStatePseudo("active"),
		Focus:  doc.UsesStatePseudo("focus"),
	}
}
