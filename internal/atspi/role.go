package atspi

// Role is an AT-SPI role name as GetRoleName would return.
type Role string

const (
	RoleUnknown      Role = "unknown"
	RoleApplication  Role = "application"
	RoleDocumentWeb  Role = "document web"
	RoleLandmark     Role = "landmark"
	RolePanel        Role = "panel"
	RoleHeading      Role = "heading"
	RoleParagraph    Role = "paragraph"
	RoleStatic       Role = "static"
	RoleImage        Role = "image"
	RoleLink         Role = "link"
	RolePushButton   Role = "push button"
	RoleEntry        Role = "entry"
	RolePasswordText Role = "password text"
	RoleCheckBox     Role = "check box"
	RoleRadioButton  Role = "radio button"
	RoleComboBox     Role = "combo box"
	RoleList         Role = "list"
	RoleListItem     Role = "list item"
)
