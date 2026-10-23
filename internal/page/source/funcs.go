package source

import (
	"html/template"
	"strings"
)

// SafeFuncs returns helpers for page templates. Every helper returns
// a plain string so html/template keeps escaping it. No helper may
// return template.HTML, template.JS, template.URL, or template.CSS.
// A typed-safe value needs a reviewed constructor in the app package
// with a comment naming the safe inputs.
func SafeFuncs() template.FuncMap {
	return template.FuncMap{"upper": strings.ToUpper}
}
