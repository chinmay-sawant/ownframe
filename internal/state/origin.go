// Package state records field-level state patches for reactive updates.
//
// ownframe historically rebuilds a whole View through Page.SetData. This
// package adds the bookkeeping needed to describe a small, origin-tagged
// change instead, so downstream invalidation can act on the changed
// properties rather than the whole document.
package state

// Origin reports where a state patch came from.
type Origin int

const (
	// OriginUnknown is the zero value.
	OriginUnknown Origin = iota
	// OriginUserInput means the patch followed a user action.
	OriginUserInput
	// OriginBackground means the patch came from non-user code.
	OriginBackground
)

// String returns a stable label for the origin.
func (o Origin) String() string {
	switch o {
	case OriginUserInput:
		return "user"
	case OriginBackground:
		return "background"
	default:
		return "unknown"
	}
}
