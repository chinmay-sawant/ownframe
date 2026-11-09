package main

import (
	"github.com/chinmay-sawant/ownframe/components/apexcharts"
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/components/chartjs"
	"github.com/chinmay-sawant/ownframe/components/echarts"
	"github.com/chinmay-sawant/ownframe/components/nivo"
	"github.com/chinmay-sawant/ownframe/components/recharts"
	"github.com/chinmay-sawant/ownframe/components/shadcnchart"
	"github.com/chinmay-sawant/ownframe/components/tremor"
	"github.com/chinmay-sawant/ownframe/ui/theme"
)

type galleryAppearance string

const (
	galleryLight  galleryAppearance = "light"
	galleryDark   galleryAppearance = "dark"
	galleryStudio galleryAppearance = "studio"
)

type themeState struct {
	Appearance galleryAppearance
	Accent     string
	Radius     string
	Spacing    string
}

func defaultThemeState() themeState { return themeState{Appearance: galleryDark} }

func (s themeState) stylesheet() (string, error) {
	base := chadcn.Dark
	switch s.Appearance {
	case galleryLight:
		base = chadcn.Light
	case galleryStudio:
		custom, err := studioProperties().Stylesheet()
		if err != nil {
			return "", err
		}
		base = chadcn.Light + custom
	}
	overrides := theme.Overrides{}
	for name, value := range map[string]string{"--of-primary": s.Accent, "--of-radius": s.Radius, "--of-spacing": s.Spacing} {
		if value != "" {
			overrides[name] = value
		}
	}
	if s.Accent != "" {
		if s.Appearance == galleryDark {
			overrides["--of-primary"] = darkAccent(s.Accent)
			overrides["--of-on-primary"] = "#172b4d"
		} else {
			overrides["--of-on-primary"] = "#ffffff"
		}
	}
	extra, err := overrides.Stylesheet()
	if err != nil {
		return "", err
	}
	return base + extra + chadcn.Stylesheet +
		chartjs.Stylesheet + shadcnchart.Stylesheet + recharts.Stylesheet +
		tremor.Stylesheet + echarts.Stylesheet + apexcharts.Stylesheet +
		nivo.Stylesheet + builderMotionCSS, nil
}

func studioProperties() theme.Overrides {
	return theme.Overrides{
		"--of-background": "#f4f1fa", "--of-surface": "#ffffff", "--of-control": "#ede7f6",
		"--of-text": "#292338", "--of-muted": "#655d78", "--of-primary": "#6241b5",
		"--of-on-primary": "#ffffff", "--of-border": "#d6cce9", "--of-selected": "#e5daf7",
		"--of-scrollbar": "#8a79a8", "--of-radius": "12px", "--of-spacing": "16px",
	}
}

func projectProperties() theme.Overrides {
	return theme.Overrides{"--project-accent": "#996627", "--of-primary": "#78521d", "--of-on-primary": "#ffffff",
		"--of-surface": "#fff8e8", "--of-text": "#3a2b14", "--of-muted": "#705e42",
		"--of-control": "#f0e4cc", "--of-border": "#d7c49f", "--of-selected": "#ebdbb9", "--of-radius": "14px"}
}

func darkAccent(value string) string {
	switch value {
	case "#21694e":
		return "#88d5b1"
	case "#6241b5":
		return "#cfb5ff"
	case "#a34c20":
		return "#ffbe91"
	default:
		return "#a6c1ff"
	}
}
