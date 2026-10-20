package main

import (
	"strings"
)

func catalogSampleKey(name string) string { return "catalog-" + catalogSlug(name) }

func catalogSlug(name string) string {
	return strings.ReplaceAll(strings.TrimPrefix(name, "-"), "-", "_")
}

func catalogPreviewClass(name string) string {
	return "catalog-preview-" + strings.TrimPrefix(name, "-")
}

func catalogPreviewLabel(name string) string {
	if strings.HasPrefix(name, "-webkit-") {
		return "WebKit " + strings.ReplaceAll(strings.TrimPrefix(name, "-webkit-"), "-", " ") + " sample"
	}
	return strings.ReplaceAll(name, "-", " ") + " sample"
}

func catalogIsAnimation(name string) bool {
	return strings.Contains(name, "animation") || name == "trigger-scope" || strings.HasPrefix(name, "view-transition-")
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
