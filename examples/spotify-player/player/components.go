package player

import "embed"

//go:embed components/*.html components/*.css assets/*.svg assets/*.png
var files embed.FS

// file returns one embedded component or asset as text.
func file(name string) string {
	b, err := files.ReadFile(name)
	if err != nil {
		return ""
	}

	return string(b)
}

func asset(name string) []byte {
	b, _ := files.ReadFile(name)
	return b
}
