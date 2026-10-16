package synth

import (
	_ "embed"
	"strings"
)

//go:embed template/index.html
var indexHTML string

//go:embed template/styles.css
var stylesCSS string

// buildHTML assembles the page template by inlining styles.css.
func buildHTML() string {
	return strings.Replace(indexHTML, "<!-- stylesheet -->", "<style>"+stylesCSS+"</style>", 1)
}
