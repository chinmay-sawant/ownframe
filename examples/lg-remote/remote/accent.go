package remote

import "strings"

type accent struct {
	ID, Label string
	// Dark and light hold the base and highlight colours, then the glow.
	Dark, DarkHi, Light, LightHi string
}

var accents = []accent{
	{"blue", "Blue", "#2f6fe4", "#5b93ff", "#1f56c4", "#4a7fe0"},
	{"pink", "Pink", "#e6125d", "#ff3d7f", "#c8104f", "#e8396f"},
	{"red", "Red", "#d8321f", "#ff5a45", "#b72614", "#e0503c"},
	{"orange", "Orange", "#e8710a", "#ff9a3d", "#c25a00", "#e68025"},
	{"green", "Green", "#1f9d55", "#3fcf7c", "#157a42", "#2fae66"},
	{"purple", "Purple", "#8a4fe0", "#ab7bff", "#6d37bd", "#8b5fd6"},
}

func findAccent(id string) accent {
	for _, a := range accents {
		if a.ID == id {
			return a
		}
	}

	return accents[0]
}

// themeCSS is the palette plus the chosen accent.
func (a *App) themeCSS() string {
	pick := findAccent(a.view.Accent)
	base, hi := pick.Dark, pick.DarkHi
	css := darkTheme
	if !a.dark {
		base, hi = pick.Light, pick.LightHi
		css = lightTheme
	}

	var b strings.Builder
	b.WriteString(css)
	b.WriteString(":root { --accent: " + base + "; --accent-hi: " + hi + "; }\n")

	return b.String()
}
