// Package patch edits a retained display and applies element edits.
// It does not import package page. Callers pass the display, the element
// registry, and the edit values in.
package patch

// DisplayPatch reports how one display update changed retained ops.
// ChangedProperties is the canonical property set; use it in new code.
// Properties, Props, and ChangedProps mirror it for compat; do not mutate.
type DisplayPatch struct {
	Changed int
	Added   int
	Removed int
	Regions int
	Full    bool

	// Deprecated: use ChangedProperties.
	Properties []string
	// Deprecated: use ChangedProperties.
	Props             []string
	ChangedProperties []string
	// Deprecated: use ChangedProperties.
	ChangedProps []string
}

// WithProps fills the canonical field once. The deprecated fields share
// that backing array. Treat them as read-only.
func WithProps(p DisplayPatch, props []string) DisplayPatch {
	clone := append([]string(nil), props...)
	p.ChangedProperties = clone
	p.Properties = clone
	p.Props = clone
	p.ChangedProps = clone

	return p
}
