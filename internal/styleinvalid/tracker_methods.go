package styleinvalid

// SkipUnmodifiedSelectors method form.
func (t *Tracker) SkipUnmodifiedSelectors(selectors []string, touched any) []string {
	return skipUnmodifiedSelectors(selectors, touched)
}

// SkipUnmodified method form.
func (t *Tracker) SkipUnmodified(selectors []string, touched any) []string {
	return skipUnmodifiedSelectors(selectors, touched)
}

// SkipUnmodifiedMatchedSelectors method form.
func (t *Tracker) SkipUnmodifiedMatchedSelectors(selectors []string, touched any) []string {
	return skipUnmodifiedSelectors(selectors, touched)
}

// TouchStylePath is a method alias for TouchPath.
func (t *Tracker) TouchStylePath(path StylePath) { t.TouchPath(path) }

// IsStyleDirty is an alias for StyleDirty.
func (t *Tracker) IsStyleDirty(path StylePath) bool { return t.StyleDirty(path) }
