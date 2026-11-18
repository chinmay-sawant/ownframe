package main

import (
	"html"
	"strings"
	"testing"
)

func TestSvgBoxesExportTheFiveValues(t *testing.T) {
	for _, c := range [][3]string{
		{"text-rendering", "sv-tr", `"optimizeLegibility"`}, {"lighting-color", "sv-lc", `"#336699"`},
		{"glyph-orientation-vertical", "sv-go", `"90deg"`}, {"baseline-shift", "sv-bs", `"super"`},
		{"marker-side", "sv-ms", `"match-parent"`},
	} {
		var demo string

		for _, p := range svgTab().Props {
			if p.Name == c[0] {
				demo = html.UnescapeString(p.Demo())
			}
		}

		if !strings.Contains(demo, c[0]+": "+c[2]+" on the box") {
			t.Errorf("%s demo lacks the exported value %s", c[0], c[2])
		}
	}

}

func TestSvgTabCounts(t *testing.T) {
	if n := svgTab().Count(); n != 5 {
		t.Fatalf("svg covers %d, want 5", n)
	}
}
