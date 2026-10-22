package files

import "embed"

//go:embed files.html files.css
var files embed.FS

// HTML returns the menu fragment.
func HTML() string { return file("files.html") }

// CSS returns the menu stylesheet.
func CSS() string { return file("files.css") }

// file returns one embedded fragment as text.
func file(name string) string {
	b, err := files.ReadFile(name)
	if err != nil {
		return ""
	}

	return string(b)
}
