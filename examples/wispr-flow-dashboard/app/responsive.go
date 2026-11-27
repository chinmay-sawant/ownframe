package app

import (
	"math"
	"regexp"
	"strconv"
	"strings"
)

const (
	wideLayoutWidth  = 1800
	wideLayoutScale  = 125
	largeLayoutWidth = 2400
	largeLayoutScale = 150
	baseLayoutScale  = 100
	cssFractionScale = 100
)

// responsiveTheme scales the same page rules at wide viewport breakpoints.
// The media rules in the source sheets target smaller windows and do not
// apply at these widths, so the wide sheet keeps only the base declarations.
func responsiveTheme() string {
	base := withoutMediaRules(styleSheets())
	var b strings.Builder

	for _, layout := range []struct{ width, scale int }{
		{wideLayoutWidth, wideLayoutScale},
		{largeLayoutWidth, largeLayoutScale},
	} {
		b.WriteString("@media (min-width: ")
		b.WriteString(strconv.Itoa(layout.width))
		b.WriteString("px) {\n")
		b.WriteString(scaleCSS(base, layout.scale))
		b.WriteString("\n}\n")
	}

	return b.String()
}

// withoutMediaRules removes top-level media blocks before the sheet is
// wrapped in wide-screen media queries.
func withoutMediaRules(css string) string {
	var b strings.Builder
	start := 0

	for start < len(css) {
		relative := strings.Index(css[start:], "@media")
		if relative < 0 {
			b.WriteString(css[start:])
			return b.String()
		}

		at := start + relative
		b.WriteString(css[start:at])

		open := strings.IndexByte(css[at:], '{')
		if open < 0 {
			b.WriteString(css[at:])
			return b.String()
		}
		open += at

		depth := 0
		end := -1
		for i := open; i < len(css); i++ {
			switch css[i] {
			case '{':
				depth++
			case '}':
				depth--
				if depth == 0 {
					end = i + 1
				}
			}
			if end >= 0 {
				break
			}
		}
		if end < 0 {
			b.WriteString(css[at:])
			return b.String()
		}

		start = end
	}

	return b.String()
}

// scaleCSS rescales pixel lengths like the spreadsheet example's zoom sheet.
// One-pixel strokes stay hairlines while layout, type, and spacing grow.
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
	scaled := math.Round(length*float64(scale)/baseLayoutScale*cssFractionScale) / cssFractionScale
	return strconv.FormatFloat(scaled, 'f', -1, 64)
}
