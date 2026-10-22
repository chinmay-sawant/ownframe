package dictionary

import "embed"

//go:embed dictionary.html dictionary.css
var files embed.FS

// HTML returns the page fragment.
func HTML() string { return file("dictionary.html") }

// CSS returns the page stylesheet.
func CSS() string { return file("dictionary.css") }

// file returns one embedded fragment as text.
func file(name string) string {
	b, err := files.ReadFile(name)
	if err != nil {
		return ""
	}

	return string(b)
}
