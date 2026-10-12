package studio

import (
	"fmt"
	"html/template"
	"math"
	"regexp"
	"strings"
)

// Token is one custom property from the editor stylesheet. Style is typed
// template.CSS so html/template does not escape it to ZgotmplZ.
type Token struct {
	Name  string
	Value string
	RGB   [3]float64
	OK    bool
	Style template.CSS
}

// maxCSS is the longest stylesheet parseTokens will scan.
const maxCSS = 200000

var (
	commentRe = regexp.MustCompile(`(?s)/\*.*?\*/`)
	tokenRe   = regexp.MustCompile(`(--[A-Za-z0-9_-]+)\s*:\s*([^;]*);?`)
)

// parseTokens returns the custom properties in source order. When a name
// appears more than once, the last value wins and keeps that last position.
func parseTokens(css string) []Token {
	if len(css) > maxCSS {
		return nil
	}

	found := scanTokens(css)

	last := make(map[string]int, len(found))
	for i, tok := range found {
		last[tok.Name] = i
	}

	out := make([]Token, 0, len(last))

	for i, tok := range found {
		if last[tok.Name] == i {
			out = append(out, tok)
		}
	}

	return out
}

// scanTokens reads every --name: value declaration in source order.
func scanTokens(css string) []Token {
	matches := tokenRe.FindAllStringSubmatch(commentRe.ReplaceAllString(css, ""), -1)

	out := make([]Token, 0, len(matches))

	for _, m := range matches {
		out = append(out, newToken(m[1], trimImportant(strings.TrimSpace(m[2]))))
	}

	return out
}

// newToken builds a token and parses its value as a colour.
func newToken(name, value string) Token {
	tok := Token{Name: name, Value: value}

	rgb, ok := parseRGB(value)
	if !ok {
		return tok
	}

	tok.RGB, tok.OK = rgb, true
	tok.Style = template.CSS(fmt.Sprintf("background:rgb(%d,%d,%d)",
		round255(rgb[0]), round255(rgb[1]), round255(rgb[2])))

	return tok
}

// round255 turns a 0..1 channel into its 0..255 int.
func round255(v float64) int { return int(math.Round(v * 255)) }
