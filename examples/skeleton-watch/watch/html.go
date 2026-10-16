package watch

import (
	"embed"
	"strings"
)

//go:embed template/*
var templateFS embed.FS

// buildHTML merges the embedded HTML template with its stylesheet.
func buildHTML() string {
	indexBytes, err := templateFS.ReadFile("template/index.html")
	if err != nil {
		return ""
	}
	cssBytes, err := templateFS.ReadFile("template/styles.css")
	if err != nil {
		return string(indexBytes)
	}
	styleTag := "<style>\n" + string(cssBytes) + "\n</style>"
	return strings.Replace(string(indexBytes), "<!-- styles -->", styleTag, 1)
}
