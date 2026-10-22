package invite

import "embed"

//go:embed invite.html invite.css
var files embed.FS

// HTML returns the page fragment.
func HTML() string { return file("invite.html") }

// CSS returns the page stylesheet.
func CSS() string { return file("invite.css") }

// file returns one embedded fragment as text.
func file(name string) string {
	b, err := files.ReadFile(name)
	if err != nil {
		return ""
	}

	return string(b)
}
