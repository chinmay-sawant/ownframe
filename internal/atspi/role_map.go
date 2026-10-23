package atspi

import "github.com/chinmay-sawant/ownframe/internal/host"

// MapRole converts a page role to an AT-SPI role. Password textboxes
// become password text so the value path stays distinct.
func MapRole(role host.AccessibilityRole, password bool) Role {
	if password && role == host.AccessibilityTextbox {
		return RolePasswordText
	}
	switch role {
	case host.AccessibilityDocument:
		return RoleDocumentWeb
	case host.AccessibilityMain, host.AccessibilityNavigation:
		return RoleLandmark
	case host.AccessibilityGroup:
		return RolePanel
	case host.AccessibilityHeading:
		return RoleHeading
	case host.AccessibilityParagraph:
		return RoleParagraph
	case host.AccessibilityText:
		return RoleStatic
	case host.AccessibilityImage:
		return RoleImage
	case host.AccessibilityLink:
		return RoleLink
	default:
		return mapFormRole(role)
	}
}

func mapFormRole(role host.AccessibilityRole) Role {
	switch role {
	case host.AccessibilityButton:
		return RolePushButton
	case host.AccessibilityTextbox:
		return RoleEntry
	case host.AccessibilityCheckbox:
		return RoleCheckBox
	case host.AccessibilityRadio:
		return RoleRadioButton
	case host.AccessibilityCombobox:
		return RoleComboBox
	case host.AccessibilityList:
		return RoleList
	case host.AccessibilityListItem:
		return RoleListItem
	default:
		return RoleUnknown
	}
}
