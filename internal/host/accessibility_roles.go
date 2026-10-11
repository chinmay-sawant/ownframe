package host

type AccessibilityNodeID string
type AccessibilityRole string
type AccessibilityActionKind uint8

const (
	AccessibilityDocument   AccessibilityRole = "document"
	AccessibilityMain       AccessibilityRole = "main"
	AccessibilityNavigation AccessibilityRole = "navigation"
	AccessibilityGroup      AccessibilityRole = "group"
	AccessibilityHeading    AccessibilityRole = "heading"
	AccessibilityParagraph  AccessibilityRole = "paragraph"
	AccessibilityText       AccessibilityRole = "text"
	AccessibilityImage      AccessibilityRole = "image"
	AccessibilityLink       AccessibilityRole = "link"
	AccessibilityButton     AccessibilityRole = "button"
	AccessibilityTextbox    AccessibilityRole = "textbox"
	AccessibilityCheckbox   AccessibilityRole = "checkbox"
	AccessibilityRadio      AccessibilityRole = "radio"
	AccessibilityCombobox   AccessibilityRole = "combobox"
	AccessibilityList       AccessibilityRole = "list"
	AccessibilityListItem   AccessibilityRole = "listitem"
)

const (
	AccessibilityFocus AccessibilityActionKind = iota + 1
	AccessibilityActivate
	AccessibilitySetValue
)
