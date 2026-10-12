package studio

import (
	"html/template"
	"math"
	"strings"
	"testing"
)

// testName returns the token name without an optional leading dash pair.
func testName(tok Token) string { return strings.TrimPrefix(tok.Name, "--") }

var _ template.CSS = Token{}.Style

func TestParseTokensKeepsSourceOrder(t *testing.T) {
	t.Parallel()

	var names []string

	for _, tok := range parseTokens(":root { --bg: #102030; --fg: #fff; --accent: rgb(1, 2, 3); }") {
		names = append(names, testName(tok))
	}

	if strings.Join(names, ",") != "bg,fg,accent" {
		t.Fatalf("names = %v", names)
	}
}

func TestParseTokensLastDuplicateWins(t *testing.T) {
	t.Parallel()

	var acc []Token

	for _, tok := range parseTokens(":root { --accent: #ff0000; --bg: #fff; --accent: #0000ff; }") {
		if testName(tok) == "accent" {
			acc = append(acc, tok)
		}
	}

	if len(acc) != 1 || acc[0].Value != "#0000ff" || acc[0].RGB != [3]float64{0, 0, 1} {
		t.Fatalf("accent = %+v", acc)
	}
}

func TestParseTokensColours(t *testing.T) {
	t.Parallel()

	byName := map[string]Token{}

	for _, tok := range parseTokens(`:root { --a: #f00; --b: #ff0000; --c: rgb(255, 0, 0);
		--d: rgba(255, 0, 0, 0.5); --e: hsl(0, 100%, 50%); --f: 12px; --g: nope; }`) {
		byName[testName(tok)] = tok
	}

	want := map[string]bool{"a": true, "b": true, "c": true, "d": true, "e": false, "f": false, "g": false}

	for name, ok := range want {
		tok, found := byName[name]

		if !found || tok.OK != ok || (tok.Style != "") != ok || (ok && !red(tok.RGB)) {
			t.Fatalf("%s = %+v", name, tok)
		}
	}
}

// red reports whether c is opaque red.
func red(c [3]float64) bool {
	return math.Abs(c[0]-1) < 0.001 && math.Abs(c[1]) < 0.001 && math.Abs(c[2]) < 0.001
}
