package ownframe

import "github.com/chinmay-sawant/ownframe/internal/host"

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
