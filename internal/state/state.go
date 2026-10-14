// Package state holds reactive, field-level updates for a Page data model.
// It records which property changed, its origin, and undo history keyed to
// stable identities rather than rendered box hits.
package state

import "time"

// Change describes one property transition.
type Change struct {
	Path       string
	Old        any
	New        any
	Origin     Origin
	At         time.Time
	Identity   string
	ElementID  string
	StableID   string
	Properties []string
	Changed    []string
	Props      []string
}
