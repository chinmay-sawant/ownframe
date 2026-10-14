package styleinvalid

import "strings"

// GlobalPercentageProperties are CSS properties whose percentage values can
// affect layout beyond one subtree.
var GlobalPercentageProperties = []string{"width", "height", "left", "top", "right", "bottom", "margin", "padding"}

// IsGlobalPercentageProperty reports whether prop is layout-global with %.
func IsGlobalPercentageProperty(prop string) bool {
	prop = strings.ToLower(strings.TrimSpace(prop))
	for _, p := range GlobalPercentageProperties {
		if prop == p || strings.HasPrefix(prop, p+"-") {
			return true
		}
	}
	return false
}

// IsGlobalPercentage reports whether property/value is a global percentage.
func IsGlobalPercentage(property string, value any) bool {
	if !IsGlobalPercentageProperty(property) {
		return false
	}
	s, _ := value.(string)
	return strings.Contains(s, "%")
}

// IsGlobalPercent reports whether a CSS value string uses a global percent.
func IsGlobalPercent(value string) bool { return strings.Contains(strings.TrimSpace(value), "%") }

// GlobalPercentage reports whether property participates in global percentage.
func GlobalPercentage(property string) bool { return IsGlobalPercentageProperty(property) }

// AffectsWholeDocument reports whether a property can affect whole document.
func AffectsWholeDocument(property string) bool { return IsGlobalPercentageProperty(property) }
