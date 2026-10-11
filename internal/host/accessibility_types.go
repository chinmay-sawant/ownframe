package host

// AccessibilityBounds are CSS pixels in the page content coordinate space.
type AccessibilityBounds struct{ X, Y, W, H float64 }

// AccessibilityNode is one semantic element in document order.
type AccessibilityNode struct {
	ID, Parent                             AccessibilityNodeID
	Children                               []AccessibilityNodeID
	ElementID                              string
	Role                                   AccessibilityRole
	Name, Description, Value               string
	Bounds                                 AccessibilityBounds
	HasBounds                              bool
	Disabled, Focusable, Focused, Password bool
	Checked, HasChecked                    bool
	Actions                                []AccessibilityActionKind
}

// AccessibilityTree is a detached snapshot. Document changes on navigation;
// Revision follows successful page redraws.
type AccessibilityTree struct {
	Document, Revision uint64
	Root, Focused      AccessibilityNodeID
	Nodes              []AccessibilityNode
}
