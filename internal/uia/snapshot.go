package uia

import (
	"github.com/chinmay-sawant/ownframe/internal/host"
)

// ControlType is a UIA control type identifier (UIA_*ControlTypeId).
type ControlType int32

// UIA control type identifiers used by the mapping.
const (
	ControlButton      ControlType = 50000
	ControlCheckBox    ControlType = 50002
	ControlComboBox    ControlType = 50003
	ControlEdit        ControlType = 50004
	ControlHyperlink   ControlType = 50005
	ControlImage       ControlType = 50006
	ControlListItem    ControlType = 50007
	ControlList        ControlType = 50008
	ControlRadioButton ControlType = 50013
	ControlText        ControlType = 50020
	ControlGroup       ControlType = 50026
	ControlDocument    ControlType = 50030
	ControlPane        ControlType = 50033
)

// ToggleState mirrors the UIA ToggleState enumeration.
type ToggleState int32

const (
	ToggleOff ToggleState = 0
	ToggleOn  ToggleState = 1
)

// Element is one UIA element. It is a value copy; callers may keep it.
type Element struct {
	ID       host.AccessibilityNodeID
	Parent   host.AccessibilityNodeID
	Children []host.AccessibilityNodeID
	// RuntimeID is stable for one node ID within one document.
	RuntimeID int32

	ControlType   ControlType
	LocalizedType string
	Name          string
	HelpText      string
	AutomationID  string
	// Value is empty for password fields.
	Value string

	Bounds    host.AccessibilityBounds
	HasBounds bool

	Enabled   bool
	Focusable bool
	Focused   bool
	Password  bool

	// Pattern support. Value is never offered for a password field.
	Invoke        bool
	ValuePattern  bool
	ReadOnly      bool
	Toggle        bool
	SelectionItem bool
	Toggled       ToggleState
	Selected      bool

	canFocus, canActivate, canSet bool
}

// Snapshot is an immutable UIA view of one accessibility tree.
type Snapshot struct {
	Document, Revision uint64
	Root, Focused      host.AccessibilityNodeID
	order              []host.AccessibilityNodeID
	byID               map[host.AccessibilityNodeID]*Element
}

func controlFor(role host.AccessibilityRole) (ControlType, string) {
	switch role {
	case host.AccessibilityDocument:
		return ControlDocument, ""
	case host.AccessibilityMain:
		return ControlGroup, "main"
	case host.AccessibilityNavigation:
		return ControlGroup, "navigation"
	case host.AccessibilityGroup:
		return ControlGroup, ""
	case host.AccessibilityHeading:
		return ControlText, "heading"
	case host.AccessibilityParagraph, host.AccessibilityText:
		return ControlText, ""
	case host.AccessibilityImage:
		return ControlImage, ""
	case host.AccessibilityLink:
		return ControlHyperlink, ""
	case host.AccessibilityButton:
		return ControlButton, ""
	case host.AccessibilityTextbox:
		return ControlEdit, ""
	case host.AccessibilityCheckbox:
		return ControlCheckBox, ""
	case host.AccessibilityRadio:
		return ControlRadioButton, ""
	case host.AccessibilityCombobox:
		return ControlComboBox, ""
	case host.AccessibilityList:
		return ControlList, ""
	case host.AccessibilityListItem:
		return ControlListItem, ""
	default:
		return ControlPane, ""
	}
}

func hasAction(n host.AccessibilityNode, kind host.AccessibilityActionKind) bool {
	for _, a := range n.Actions {
		if a == kind {
			return true
		}
	}
	return false
}

func mapElement(n host.AccessibilityNode, focused host.AccessibilityNodeID) *Element {
	ct, localized := controlFor(n.Role)
	e := &Element{
		ID:            n.ID,
		Parent:        n.Parent,
		ControlType:   ct,
		LocalizedType: localized,
		Name:          n.Name,
		HelpText:      n.Description,
		AutomationID:  n.ElementID,
		Bounds:        n.Bounds,
		HasBounds:     n.HasBounds,
		Enabled:       !n.Disabled,
		Focusable:     n.Focusable && !n.Disabled,
		Focused:       n.Focused || (focused != "" && n.ID == focused),
		Password:      n.Password,
		canFocus:      hasAction(n, host.AccessibilityFocus),
		canActivate:   hasAction(n, host.AccessibilityActivate),
		canSet:        hasAction(n, host.AccessibilitySetValue) && !n.Password,
	}
	if !n.Password {
		// The tree builder blanks password values already. Not copying
		// the value keeps that guarantee independent of the caller.
		e.Value = n.Value
	}
	switch n.Role {
	case host.AccessibilityButton, host.AccessibilityLink:
		e.Invoke = e.canActivate
	case host.AccessibilityCheckbox:
		e.Toggle = e.canActivate && n.HasChecked
		if n.Checked {
			e.Toggled = ToggleOn
		}
	case host.AccessibilityRadio:
		e.SelectionItem = true
		e.Selected = n.Checked
	case host.AccessibilityTextbox, host.AccessibilityCombobox:
		e.ValuePattern = !n.Password
		e.ReadOnly = !e.canSet
	}
	return e
}

// BuildSnapshot maps tree to UIA elements. Runtime IDs are filled by the
// Provider, which keeps them stable across revisions.
func BuildSnapshot(tree host.AccessibilityTree) *Snapshot {
	s := &Snapshot{
		Document: tree.Document,
		Revision: tree.Revision,
		Root:     tree.Root,
		Focused:  tree.Focused,
		byID:     make(map[host.AccessibilityNodeID]*Element, len(tree.Nodes)),
	}
	kids := make(map[host.AccessibilityNodeID][]host.AccessibilityNodeID, len(tree.Nodes))
	for _, n := range tree.Nodes {
		if _, dup := s.byID[n.ID]; dup || n.ID == "" {
			continue
		}
		s.byID[n.ID] = mapElement(n, tree.Focused)
		s.order = append(s.order, n.ID)
		kids[n.ID] = n.Children
	}
	// Keep only children that exist, in snapshot order, and drop a parent
	// that does not.
	for _, id := range s.order {
		e := s.byID[id]
		for _, c := range kids[id] {
			if _, ok := s.byID[c]; ok {
				e.Children = append(e.Children, c)
			}
		}
		if _, ok := s.byID[e.Parent]; !ok {
			e.Parent = ""
		}
	}
	if _, ok := s.byID[s.Root]; !ok {
		s.Root = ""
	}
	if _, ok := s.byID[s.Focused]; !ok {
		s.Focused = ""
	}
	if s.Focused == "" {
		for _, id := range s.order {
			if s.byID[id].Focused {
				s.Focused = id
				break
			}
		}
	}
	return s
}

// Len reports the number of elements.
func (s *Snapshot) Len() int { return len(s.order) }
