package atspi

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

func TestMapRoleCoversExportedRoles(t *testing.T) {
	cases := []struct {
		role host.AccessibilityRole
		want Role
	}{
		{host.AccessibilityDocument, RoleDocumentWeb},
		{host.AccessibilityMain, RoleLandmark},
		{host.AccessibilityNavigation, RoleLandmark},
		{host.AccessibilityGroup, RolePanel},
		{host.AccessibilityHeading, RoleHeading},
		{host.AccessibilityParagraph, RoleParagraph},
		{host.AccessibilityText, RoleStatic},
		{host.AccessibilityImage, RoleImage},
		{host.AccessibilityLink, RoleLink},
		{host.AccessibilityButton, RolePushButton},
		{host.AccessibilityTextbox, RoleEntry},
		{host.AccessibilityCheckbox, RoleCheckBox},
		{host.AccessibilityRadio, RoleRadioButton},
		{host.AccessibilityCombobox, RoleComboBox},
		{host.AccessibilityList, RoleList},
		{host.AccessibilityListItem, RoleListItem},
	}
	for _, c := range cases {
		if got := MapRole(c.role, false); got != c.want {
			t.Fatalf("%s: got %s want %s", c.role, got, c.want)
		}
	}
	if MapRole(host.AccessibilityTextbox, true) != RolePasswordText {
		t.Fatal("password textbox")
	}
	if MapRole("nope", false) != RoleUnknown {
		t.Fatal("unknown role")
	}
}
