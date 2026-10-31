package atspi

// AT-SPI role numbers are stable values from AtspiRole in
// atspi-constants.h. Role names remain the source of truth for
// GetRoleName and for logs.
const (
	roleNumUnknown      uint32 = 67  // ATSPI_ROLE_UNKNOWN
	roleNumApplication  uint32 = 75  // ATSPI_ROLE_APPLICATION
	roleNumDocumentWeb  uint32 = 95  // ATSPI_ROLE_DOCUMENT_WEB
	roleNumLandmark     uint32 = 110 // ATSPI_ROLE_LANDMARK
	roleNumPanel        uint32 = 39  // ATSPI_ROLE_PANEL
	roleNumHeading      uint32 = 83  // ATSPI_ROLE_HEADING
	roleNumParagraph    uint32 = 73  // ATSPI_ROLE_PARAGRAPH
	roleNumStatic       uint32 = 116 // ATSPI_ROLE_STATIC
	roleNumImage        uint32 = 27  // ATSPI_ROLE_IMAGE
	roleNumLink         uint32 = 88  // ATSPI_ROLE_LINK
	roleNumPushButton   uint32 = 43  // ATSPI_ROLE_BUTTON
	roleNumEntry        uint32 = 79  // ATSPI_ROLE_ENTRY
	roleNumPasswordText uint32 = 40  // ATSPI_ROLE_PASSWORD_TEXT
	roleNumCheckBox     uint32 = 7   // ATSPI_ROLE_CHECK_BOX
	roleNumRadioButton  uint32 = 44  // ATSPI_ROLE_RADIO_BUTTON
	roleNumComboBox     uint32 = 11  // ATSPI_ROLE_COMBO_BOX
	roleNumList         uint32 = 31  // ATSPI_ROLE_LIST
	roleNumListItem     uint32 = 32  // ATSPI_ROLE_LIST_ITEM
)

// RoleNumber returns the AT-SPI numeric role for a mapped role.
func RoleNumber(role Role) uint32 {
	switch role {
	case RoleApplication:
		return roleNumApplication
	case RoleDocumentWeb:
		return roleNumDocumentWeb
	case RoleLandmark:
		return roleNumLandmark
	case RolePanel:
		return roleNumPanel
	case RoleHeading:
		return roleNumHeading
	case RoleParagraph:
		return roleNumParagraph
	case RoleStatic:
		return roleNumStatic
	case RoleImage:
		return roleNumImage
	case RoleLink:
		return roleNumLink
	case RolePushButton:
		return roleNumPushButton
	case RoleEntry:
		return roleNumEntry
	case RolePasswordText:
		return roleNumPasswordText
	case RoleCheckBox:
		return roleNumCheckBox
	case RoleRadioButton:
		return roleNumRadioButton
	case RoleComboBox:
		return roleNumComboBox
	case RoleList:
		return roleNumList
	case RoleListItem:
		return roleNumListItem
	default:
		return roleNumUnknown
	}
}
