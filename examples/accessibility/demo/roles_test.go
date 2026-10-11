package demo

import (
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestExampleCoversEveryExportedRole(t *testing.T) {
	tree, err := renderedApp(t).accessible().AccessibilityTree()
	if err != nil {
		t.Fatal(err)
	}
	want := []ownframe.AccessibilityRole{
		ownframe.AccessibilityDocument, ownframe.AccessibilityMain,
		ownframe.AccessibilityNavigation, ownframe.AccessibilityGroup,
		ownframe.AccessibilityHeading, ownframe.AccessibilityParagraph,
		ownframe.AccessibilityText, ownframe.AccessibilityImage,
		ownframe.AccessibilityLink, ownframe.AccessibilityButton,
		ownframe.AccessibilityTextbox, ownframe.AccessibilityCheckbox,
		ownframe.AccessibilityRadio, ownframe.AccessibilityCombobox,
		ownframe.AccessibilityList, ownframe.AccessibilityListItem,
	}
	seen := map[ownframe.AccessibilityRole]bool{}
	for _, node := range tree.Nodes {
		seen[node.Role] = true
	}
	for _, role := range want {
		if !seen[role] {
			t.Errorf("example tree does not contain role %q", role)
		}
	}
}
