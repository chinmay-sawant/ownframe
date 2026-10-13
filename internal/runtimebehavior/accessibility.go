package runtimebehavior

// AccessibilityFallback keeps a fallback accessibility state tied to a
// retained element's identity.
type AccessibilityFallback struct {
	ElementID string
	Retained  bool
	Fallback  bool
}

// Bind ties the fallback to the retained element state.
func (a *AccessibilityFallback) Bind(id string, retained bool) {
	if a == nil {
		return
	}
	a.ElementID = id
	a.Retained = retained
	a.Fallback = !retained
}

// IsFallback reports whether the fallback is active for a removed element.
func (a *AccessibilityFallback) IsFallback() bool {
	if a == nil {
		return false
	}
	return a.Fallback
}

// Clear resets the fallback state.
func (a *AccessibilityFallback) Clear() {
	if a == nil {
		return
	}
	a.ElementID = ""
	a.Retained = false
	a.Fallback = false
}

// Update refreshes retained/fallback state for an element.
func (a *AccessibilityFallback) Update(retained bool) {
	if a == nil {
		return
	}
	a.Retained = retained
	a.Fallback = !retained
}

// Active reports whether any fallback is active.
func (a *AccessibilityFallback) Active() bool {
	if a == nil {
		return false
	}
	return a.Fallback
}

// SetElement stores the element id.
func (a *AccessibilityFallback) SetElement(id string) {
	if a == nil {
		return
	}
	a.ElementID = id
}
