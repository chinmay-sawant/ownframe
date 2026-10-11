package host

import "context"

// Accessible exposes a detached semantic tree and actions for it.
type Accessible interface {
	AccessibilityTree() (AccessibilityTree, error)
	AccessibilityPerform(ctx context.Context, request AccessibilityRequest) error
}

// AccessibilityRequest identifies one supported action on a tree node.
type AccessibilityRequest struct {
	Document uint64
	Node     AccessibilityNodeID
	Kind     AccessibilityActionKind
	Value    string
}
