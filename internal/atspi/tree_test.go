package atspi

import "github.com/chinmay-sawant/ownframe/internal/host"

const (
	testDoc = 1
	testRev = 2
)

func sampleTree() host.AccessibilityTree {
	return host.AccessibilityTree{
		Document: testDoc,
		Revision: testRev,
		Root:     "document",
		Focused:  "id:name",
		Nodes: []host.AccessibilityNode{
			{
				ID: "document", Role: host.AccessibilityDocument, Name: "Doc",
				Children: []host.AccessibilityNodeID{"id:name", "id:go", "id:pw", "id:on"},
			},
			textboxNode(),
			{
				ID: "id:go", Parent: "document", Role: host.AccessibilityButton,
				Name: "Go", Actions: []host.AccessibilityActionKind{host.AccessibilityActivate},
			},
			passwordNode(),
			checkboxNode(),
		},
	}
}

func textboxNode() host.AccessibilityNode {
	return host.AccessibilityNode{
		ID: "id:name", Parent: "document", Role: host.AccessibilityTextbox,
		Name: "Name", Value: "Ada", Focusable: true, Focused: true,
		Actions: []host.AccessibilityActionKind{host.AccessibilityFocus, host.AccessibilitySetValue},
	}
}

func passwordNode() host.AccessibilityNode {
	return host.AccessibilityNode{
		ID: "id:pw", Parent: "document", Role: host.AccessibilityTextbox,
		Name: "Password", Value: "secret", Password: true,
		Actions: []host.AccessibilityActionKind{host.AccessibilitySetValue},
	}
}

func checkboxNode() host.AccessibilityNode {
	return host.AccessibilityNode{
		ID: "id:on", Parent: "document", Role: host.AccessibilityCheckbox,
		Name: "On", HasChecked: true, Checked: true,
		Actions: []host.AccessibilityActionKind{host.AccessibilityActivate},
	}
}
