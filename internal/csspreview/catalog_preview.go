package csspreview

import (
	"strings"
)

func catalogSampleKey(name string) string { return "catalog-" + catalogSlug(name) }

func catalogSlug(name string) string {
	return strings.ReplaceAll(strings.TrimPrefix(name, "-"), "-", "_")
}

func catalogPreviewClass(name string) string {
	if name == "trigger-scope" || strings.HasPrefix(name, "view-transition") {
		return "catalog-animation-standard"
	}
	if name == "transition-behavior" {
		return "catalog-transition-standard-background"
	}
	if strings.Contains(name, "animation") {
		if strings.HasPrefix(name, "-webkit-") {
			return "catalog-animation-webkit"
		}
		return "catalog-animation-standard"
	}
	engine := "standard"
	if strings.HasPrefix(name, "-webkit-") {
		engine = "webkit"
	}
	property := strings.ReplaceAll(transitionPreviewProperty(name), "-color", "")
	return "catalog-transition-" + engine + "-" + property
}

func catalogPreviewLabel(name string) string {
	class := catalogPreviewClass(name)
	switch class {
	case "catalog-animation-standard":
		return "Standard animation"
	case "catalog-animation-webkit":
		return "WebKit animation"
	}
	engine := "Standard"
	if strings.Contains(class, "-webkit-") {
		engine = "WebKit"
	}
	property := strings.TrimPrefix(class, "catalog-transition-standard-")
	property = strings.TrimPrefix(property, "catalog-transition-webkit-")
	return engine + " " + property + " transition"
}

func transitionPreviewProperty(name string) string {
	switch {
	case name == "transition-behavior":
		return "background-color"
	case strings.Contains(name, "property"):
		return "width"
	case strings.Contains(name, "duration"):
		return "opacity"
	case strings.Contains(name, "delay"):
		return "background-color"
	default:
		return "left"
	}
}
