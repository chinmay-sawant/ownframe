package freemonth

import "embed"

//go:embed freemonth.html freemonth.css
var files embed.FS

// HTML returns the page fragment.
func HTML() string { return file("freemonth.html") }

// CSS returns the page stylesheet.
func CSS() string { return file("freemonth.css") }

// file returns one embedded fragment as text.
func file(name string) string {
	b, err := files.ReadFile(name)
	if err != nil {
		return ""
	}

	return string(b)
}
