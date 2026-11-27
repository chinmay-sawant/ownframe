package player

import (
	"math"
	"regexp"
	"strconv"
	"strings"
)

const (
	wideLayoutWidth  = 1800
	wideLayoutScale  = 100
	largeLayoutWidth = 2400
	largeLayoutScale = 150
	baseLayoutScale  = 100
)

// responsiveTheme scales the Spotify rules at wide viewport breakpoints.
func responsiveTheme() string {
	var base strings.Builder
	for _, name := range styles {
		base.WriteString(file("components/" + name))
		base.WriteByte('\n')
	}

	css := base.String()
	var theme strings.Builder
	for _, layout := range []struct{ width, scale int }{
		{wideLayoutWidth, wideLayoutScale},
		{largeLayoutWidth, largeLayoutScale},
	} {
		theme.WriteString("@media (min-width: ")
		theme.WriteString(strconv.Itoa(layout.width))
		theme.WriteString("px) {\n")
		theme.WriteString(scaleCSS(css, layout.scale))
		theme.WriteString("\n.page { height: 96.5%; }\n")
		theme.WriteString("\n}\n")
	}

	return theme.String()
}

// scaleCSS scales pixel dimensions while preserving one-pixel hairlines.
func scaleCSS(css string, scale int) string {
	if scale == baseLayoutScale {
		return css
	}

	pxLength := regexp.MustCompile(`(\d+\.\d+|\d+|\.\d+)px`)
	lines := strings.Split(css, "\n")
	for i, line := range lines {
		if !strings.Contains(line, "px") {
			continue
		}
		lines[i] = pxLength.ReplaceAllStringFunc(line, func(value string) string {
			length, err := strconv.ParseFloat(strings.TrimSuffix(value, "px"), 64)
			if err != nil {
				return value
			}
			return scaleCSSLength(length, scale) + "px"
		})
	}

	return strings.Join(lines, "\n")
}

func scaleCSSLength(length float64, scale int) string {
	if length == 1 {
		return "1"
	}
	if length == math.Trunc(length) {
		scaled := (int(length)*scale + baseLayoutScale/2) / baseLayoutScale
		return strconv.Itoa(scaled)
	}

	scaled := math.Round(length*float64(scale)/baseLayoutScale*100) / 100
	return strconv.FormatFloat(scaled, 'f', -1, 64)
}
