package studio

import (
	_ "embed"
	"strings"
)

// editorHTML is the complete studio document. The preview pane is spliced in
// where the placeholder comment sits.
//
//go:embed editor.html
var editorHTML string

// previewHTML is the preview pane fragment the placeholder is replaced with.
//
//go:embed preview.html
var previewHTML string

// studioHTML is the document the page parses.
var studioHTML = strings.Replace(editorHTML, "<!-- preview -->", previewHTML, 1)
