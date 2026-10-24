package cssanimation

import (
	"math"
	"strings"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

// perspectiveTestSize is a 200x100 CSS-pixel box at the display's 0.75
// points-per-pixel.
func perspectiveTestSize() transformSize {
	return transformSize{width: 200, height: 100, pointsPerPixel: 0.75, hasBox: true}
}

func near(a, b float64) bool { return math.Abs(a-b) <= 1e-9 }

func TestParsePerspectiveLengths(t *testing.T) {
	size := perspectiveTestSize()

	for _, test := range []struct {
		name  string
		value string
		want  float64
		fails bool
	}{
		{name: "none", value: "none", want: 0},
		{name: "NONE", value: "NONE", want: 0},
		{name: "zero", value: "0", want: 0},
		{name: "px", value: "500px", want: 375},
		{name: "fractional px", value: "2.5px", want: 1.875},
		{name: "negative", value: "-10px", fails: true},
		{name: "percentage", value: "10%", fails: true},
		{name: "unitless", value: "10", fails: true},
		{name: "garbage", value: "wide", fails: true},
		{name: "empty", value: "", fails: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := parsePerspective(test.value, size)
			if test.fails {
				if err == nil {
					t.Fatalf("parsePerspective(%q) = %v, want an error", test.value, got)
				}

				return
			}
			if err != nil {
				t.Fatalf("parsePerspective(%q): %v", test.value, err)
			}
			if !near(got, test.want) {
				t.Fatalf("parsePerspective(%q) = %v, want %v", test.value, got, test.want)
			}
		})
	}
}

func TestParsePerspectiveOriginPositions(t *testing.T) {
	size := perspectiveTestSize()

	for _, test := range []struct {
		name  string
		value string
		x, y  float64
		fails bool
	}{
		{name: "percent pair", value: "50% 50%", x: 75, y: 37.5},
		{name: "quarter", value: "25% 75%", x: 37.5, y: 56.25},
		{name: "keywords", value: "top left", x: 0, y: 0},
		{name: "keywords reversed", value: "left top", x: 0, y: 0},
		{name: "bottom right", value: "right bottom", x: 150, y: 75},
		{name: "center top", value: "center top", x: 75, y: 0},
		{name: "top center", value: "top center", x: 75, y: 0},
		{name: "px pair", value: "20px 10px", x: 15, y: 7.5},
		{name: "unitless zeros", value: "0 0", x: 0, y: 0},
		{name: "single left", value: "left", x: 0, y: 37.5},
		{name: "single bottom", value: "bottom", x: 75, y: 75},
		{name: "single center", value: "center", x: 75, y: 37.5},
		{name: "single percent", value: "10%", x: 15, y: 37.5},
		{name: "length and keyword", value: "10px bottom", x: 7.5, y: 75},
		{name: "keyword and length", value: "right 10px", x: 150, y: 7.5},
		{name: "two horizontal keywords", value: "left right", fails: true},
		{name: "two vertical keywords", value: "top bottom", fails: true},
		{name: "three values", value: "left 10px top", x: 7.5, y: 0},
		{name: "empty", value: "", fails: true},
		{name: "unitless", value: "10 20", fails: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			x, y, err := parsePerspectiveOrigin(test.value, size)
			if test.fails {
				if err == nil {
					t.Fatalf("parsePerspectiveOrigin(%q) = %v,%v, want an error", test.value, x, y)
				}

				return
			}
			if err != nil {
				t.Fatalf("parsePerspectiveOrigin(%q): %v", test.value, err)
			}
			if !near(x, test.x) || !near(y, test.y) {
				t.Fatalf("parsePerspectiveOrigin(%q) = %v,%v, want %v,%v", test.value, x, y, test.x, test.y)
			}
		})
	}
}

func TestKeyframesReadPerspectiveDeclarations(t *testing.T) {
	rules := []layout.KeyframesRule{{
		Name: "camera",
		Keyframes: []layout.Keyframe{
			{Offset: 0, Declarations: []layout.KeyframeDeclaration{
				{Property: "perspective", Value: "none"},
			}},
			{Offset: 0.5, Declarations: []layout.KeyframeDeclaration{
				{Property: "perspective", Value: "500px"},
				{Property: "perspective-origin", Value: "top left"},
			}},
			{Offset: 1, Declarations: []layout.KeyframeDeclaration{
				{Property: "perspective", Value: "1000px"},
			}},
		},
	}}

	frames := keyframes(rules, "camera", 0.75, 200, 100)
	if len(frames) != 3 {
		t.Fatalf("frames = %d, want 3", len(frames))
	}

	first := frames[0].Perspective
	if !first.HasDistance || first.Distance != 0 {
		t.Fatalf("first perspective = %+v, want a declared none", first)
	}
	if first.HasOrigin {
		t.Fatalf("first perspective origin = %+v, want undeclared", first)
	}
	if !near(first.OriginX, 75) || !near(first.OriginY, 37.5) {
		t.Fatalf("first default origin = %v,%v, want 75,37.5", first.OriginX, first.OriginY)
	}

	middle := frames[1].Perspective
	if !middle.HasDistance || !near(middle.Distance, 375) {
		t.Fatalf("middle perspective = %+v, want 500px as 375 points", middle)
	}
	if !middle.HasOrigin || middle.OriginX != 0 || middle.OriginY != 0 {
		t.Fatalf("middle origin = %+v, want the declared top left", middle)
	}

	last := frames[2].Perspective
	if !last.HasDistance || !near(last.Distance, 750) || last.HasOrigin {
		t.Fatalf("last perspective = %+v, want 1000px and the undeclared default origin", last)
	}
}

func TestKeyframesReportInvalidPerspective(t *testing.T) {
	for _, declaration := range []layout.KeyframeDeclaration{
		{Property: "perspective", Value: "-10px"},
		{Property: "perspective-origin", Value: "left right"},
	} {
		rules := []layout.KeyframesRule{{Name: "bad", Keyframes: []layout.Keyframe{{
			Offset:       0,
			Declarations: []layout.KeyframeDeclaration{declaration},
		}}}}
		frames := keyframes(rules, "bad", 0.75, 200, 100)
		if frames[0].Err == nil {
			t.Fatalf("%s: %q parsed without an error", declaration.Property, declaration.Value)
		}
		if !strings.Contains(frames[0].Err.Error(), "perspective") {
			t.Fatalf("%s error = %v, want the perspective context", declaration.Property, frames[0].Err)
		}
	}
}
