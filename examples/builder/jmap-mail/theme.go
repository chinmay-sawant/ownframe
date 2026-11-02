package main

import "github.com/chinmay-sawant/ownframe/ui/theme"

type mailTheme string

const (
	mailThemeLight  mailTheme = "light"
	mailThemeDark   mailTheme = "dark"
	mailThemeCustom mailTheme = "custom"
)

func (name mailTheme) stylesheet() (string, error) {
	switch name {
	case mailThemeDark:
		return theme.Dark, nil
	case mailThemeCustom:
		overrides, err := sageProperties().Stylesheet()
		if err != nil {
			return "", err
		}
		return theme.Light + overrides, nil
	default:
		return theme.Light, nil
	}
}

// Sage is local to this application. The generic presets live in ui/theme.
func sageProperties() theme.Overrides {
	return theme.Overrides{
		"--of-background": "#f3f0e4", "--of-surface": "#fffdf6", "--of-control": "#ecefe2",
		"--of-text": "#29372d", "--of-muted": "#5c6d60", "--of-primary": "#21694e",
		"--of-on-primary": "#ffffff", "--of-border": "#c8cfbc", "--of-selected": "#dce8d4",
		"--of-scrollbar": "#75866a", "--of-radius": "12px", "--of-spacing": "18px",
	}
}

// The Compose button keeps its local shape when the application theme changes.
func composeProperties() theme.Overrides {
	return theme.Overrides{"--of-radius": "18px"}
}
