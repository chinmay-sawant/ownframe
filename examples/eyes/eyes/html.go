package eyes

import (
	_ "embed"
	"strings"
)

//go:embed template/index.html
var indexHTML string

//go:embed template/styles.css
var stylesCSS string

func buildHTML() string {
	html := strings.Replace(indexHTML,
		"<!-- stylesheet -->", "<style>"+stylesCSS+"</style>", 1)
	return strings.Replace(html, "<!-- spiral image slots -->", spiralMarkup(), 1)
}

func spiralMarkup() string {
	return `<div class="spiral-layer">` +
		`<div class="spiral" id="spiral-l"></div>` +
		`<div class="spiral spiral-right" id="spiral-r"></div></div>`
}
