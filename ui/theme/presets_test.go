package theme_test

import (
	"math"
	"regexp"
	"strconv"
	"testing"

	"github.com/chinmay-sawant/ownframe/ui/theme"
)

func TestBuiltInTextContrast(t *testing.T) {
	t.Parallel()
	declaration := regexp.MustCompile(`--of-([a-z-]+):([^;]+);`)
	for name, sheet := range map[string]string{"light": theme.Light, "dark": theme.Dark} {
		colors := map[string]string{}
		for _, match := range declaration.FindAllStringSubmatch(sheet, -1) {
			colors[match[1]] = match[2]
		}
		for _, pair := range [][2]string{{"text", "surface"}, {"muted", "surface"}, {"on-primary", "primary"}, {"primary", "surface"}, {"primary", "selected"}} {
			first, second := luminance(t, colors[pair[0]]), luminance(t, colors[pair[1]])
			contrast := (max(first, second) + 0.05) / (min(first, second) + 0.05)
			if contrast < 4.5 {
				t.Errorf("%s %s on %s contrast %.2f", name, pair[0], pair[1], contrast)
			}
		}
	}
}

func luminance(t *testing.T, hex string) float64 {
	t.Helper()
	rgb, err := strconv.ParseUint(hex[1:], 16, 32)
	if err != nil {
		t.Fatal(err)
	}
	channel := func(value uint64) float64 {
		x := float64(value) / 255
		if x <= 0.04045 {
			return x / 12.92
		}
		return math.Pow((x+0.055)/1.055, 2.4)
	}
	return 0.2126*channel((rgb>>16)&255) + 0.7152*channel((rgb>>8)&255) + 0.0722*channel(rgb&255)
}
