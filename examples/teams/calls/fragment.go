package calls

import "embed"

//go:embed calls.html calls.css
var files embed.FS

// HTML returns the menu fragment.
func HTML() string { return file("calls.html") }

// CSS returns the menu stylesheet.
func CSS() string { return file("calls.css") }

// file returns one embedded fragment as text.
func file(name string) string {
	b, err := files.ReadFile(name)
	if err != nil {
		return ""
	}

	return string(b)
}
