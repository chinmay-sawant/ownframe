package cssanimation

import (
	"errors"
	"math"
	"testing"
)

// tolerance is the room a colour conversion and the 8-bit source share.
const tolerance = 1e-6

func closeColor(got, want Color) bool {
	return math.Abs(got.R-want.R) < tolerance && math.Abs(got.G-want.G) < tolerance &&
		math.Abs(got.B-want.B) < tolerance && math.Abs(got.A-want.A) < tolerance &&
		got.Current == want.Current
}

func TestParseColorForms(t *testing.T) {
	tests := []struct {
		value string
		want  Color
	}{
		{"#f00", Color{R: 1, A: 1}},
		{"#0f08", Color{G: 1, A: 0x88 / 255.0}},
		{"#ff0000", Color{R: 1, A: 1}},
		{"#00ff0080", Color{G: 1, A: 0x80 / 255.0}},
		{"rgb(255, 0, 0)", Color{R: 1, A: 1}},
		{"rgba(255, 0, 0, 0.5)", Color{R: 1, A: .5}},
		{"rgb(100% 0% 0% / 50%)", Color{R: 1, A: .5}},
		{"hsl(0, 100%, 50%)", Color{R: 1, A: 1}},
		{"hsl(120deg 100% 50%)", Color{G: 1, A: 1}},
		{"hsla(240, 100%, 50%, .25)", Color{B: 1, A: .25}},
		{"hsl(0.5turn, 100%, 50%)", Color{G: 1, B: 1, A: 1}},
		{"red", Color{R: 1, A: 1}},
		{"rebeccapurple", Color{R: 0x66 / 255.0, G: 0x33 / 255.0, B: 0x99 / 255.0, A: 1}},
		{"CURRENTCOLOR", Color{Current: true, A: 1}},
		{"transparent", Color{A: 0}},
	}
	for _, test := range tests {
		got, err := parseColor(test.value)
		if err != nil {
			t.Errorf("parseColor(%q) failed: %v", test.value, err)
			continue
		}
		if !closeColor(got, test.want) {
			t.Errorf("parseColor(%q) = %+v, want %+v", test.value, got, test.want)
		}
	}
}

func TestParseColorRejectsMalformedValues(t *testing.T) {
	for _, value := range []string{
		"", "#", "#12", "#12345", "#ggg", "rgb(1, 2)", "rgb(1 2 3 4)",
		"rgb(1, 2, 3, 4, 5)", "hsl(120, 50, 50%)", "hsl(120, 50%, 50)", "nonsense",
		"inherit", "initial", "unset",
	} {
		if _, err := parseColor(value); !errors.Is(err, ErrColor) {
			t.Errorf("parseColor(%q) error = %v, want ErrColor", value, err)
		}
	}
}

// TestParseColorClampsChannels pins the CSS rule that a numeric channel outside
// 0..255 and an alpha outside 0..1 clamp instead of failing.
func TestParseColorClampsChannels(t *testing.T) {
	got, err := parseColor("rgb(300, -20, 0, 1.5)")
	if err != nil {
		t.Fatalf("parseColor failed: %v", err)
	}
	want := Color{R: 1, A: 1}
	if !closeColor(got, want) {
		t.Fatalf("clamped colour = %+v, want %+v", got, want)
	}
}

// TestParseColorDefersAnUnsupportedSpace pins that a modern colour space is
// reported as deferred rather than malformed, so the binder can leave the
// declaration at base paint.
func TestParseColorDefersAnUnsupportedSpace(t *testing.T) {
	for _, value := range []string{"oklab(0.5 0 0)", "color(display-p3 1 0 0)", "lab(50% 0 0)"} {
		if _, err := parseColor(value); !errors.Is(err, ErrColorSpace) {
			t.Errorf("parseColor(%q) error = %v, want ErrColorSpace", value, err)
		}
	}
}
