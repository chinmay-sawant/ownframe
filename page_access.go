package ownframe

import (
	"context"
	"github.com/chinmay-sawant/ownframe/internal/host"
)

// Access reads the accessibility tree and runs one tree action.
type Access struct{ page *Page }

// Access returns the accessibility snapshot and actions for this page.
func (p *Page) Access() Access { return Access{page: p} }

// AccessibilityTree returns current semantics in document order.
func (a Access) AccessibilityTree() (AccessibilityTree, error) {
	p := innerPage(a.page)
	if p == nil {
		return AccessibilityTree{}, ErrNilPage
	}

	return p.AccessibilityTree()
}

// AccessibilityPerform applies a tree action.
// The accessibility example drives this from app code.
func (a Access) AccessibilityPerform(ctx context.Context, request AccessibilityRequest) error {
	p := innerPage(a.page)
	if p == nil {
		return ErrNilPage
	}

	return p.AccessibilityPerform(ctx, request)
}

type AccessibilityNodeID = host.AccessibilityNodeID
type AccessibilityRole = host.AccessibilityRole
type AccessibilityActionKind = host.AccessibilityActionKind
type AccessibilityBounds = host.AccessibilityBounds
type AccessibilityNode = host.AccessibilityNode
type AccessibilityTree = host.AccessibilityTree
type AccessibilityRequest = host.AccessibilityRequest
type Accessible = host.Accessible

const (
	AccessibilityDocument   = host.AccessibilityDocument
	AccessibilityMain       = host.AccessibilityMain
	AccessibilityNavigation = host.AccessibilityNavigation
	AccessibilityGroup      = host.AccessibilityGroup
	AccessibilityHeading    = host.AccessibilityHeading
	AccessibilityParagraph  = host.AccessibilityParagraph
	AccessibilityText       = host.AccessibilityText
	AccessibilityImage      = host.AccessibilityImage
	AccessibilityLink       = host.AccessibilityLink
	AccessibilityButton     = host.AccessibilityButton
	AccessibilityTextbox    = host.AccessibilityTextbox
	AccessibilityCheckbox   = host.AccessibilityCheckbox
	AccessibilityRadio      = host.AccessibilityRadio
	AccessibilityCombobox   = host.AccessibilityCombobox
	AccessibilityList       = host.AccessibilityList
	AccessibilityListItem   = host.AccessibilityListItem
	AccessibilityFocus      = host.AccessibilityFocus
	AccessibilityActivate   = host.AccessibilityActivate
	AccessibilitySetValue   = host.AccessibilitySetValue
)

var (
	ErrAccessibilityUnavailable = host.ErrAccessibilityUnavailable
	ErrAccessibilityStale       = host.ErrAccessibilityStale
	ErrAccessibilityNode        = host.ErrAccessibilityNode
	ErrAccessibilityAction      = host.ErrAccessibilityAction
	ErrAccessibilityDisabled    = host.ErrAccessibilityDisabled
	ErrAccessibilityRange       = host.ErrAccessibilityRange
)
