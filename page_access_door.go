package ownframe

import "context"

// Access reads the accessibility tree and runs one tree action.
type Access struct{ page *Page }

// Access returns the accessibility snapshot and actions for this page.
func (p *Page) Access() Access { return Access{page: p} }

// AccessibilityTree returns current semantics in document order.
func (a Access) AccessibilityTree() (AccessibilityTree, error) {
	p := innerPage(a.page)
	if p == nil {
		return AccessibilityTree{}, nil
	}

	return p.AccessibilityTree()
}

// AccessibilityPerform applies a tree action.
// The accessibility example drives this from app code.
func (a Access) AccessibilityPerform(ctx context.Context, request AccessibilityRequest) error {
	p := innerPage(a.page)
	if p == nil {
		return nil
	}

	return p.AccessibilityPerform(ctx, request)
}
