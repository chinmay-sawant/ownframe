package main

import "embed"

//go:embed components/*.html components/*.css
var files embed.FS

// file returns one embedded component or stylesheet as text.
func file(name string) string {
	b, err := files.ReadFile(name)
	if err != nil {
		return ""
	}

	return string(b)
}
