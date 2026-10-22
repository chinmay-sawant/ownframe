package ownframe

import "context"

// AccessibilityTree returns current semantics in document order.
func (p *Page) AccessibilityTree() (AccessibilityTree, error) {
	return innerPage(p).AccessibilityTree()
}

// AccessibilityPerform applies a tree action.
// The accessibility example drives this from app code.
func (p *Page) AccessibilityPerform(ctx context.Context, request AccessibilityRequest) error {
	return innerPage(p).AccessibilityPerform(ctx, request)
}
