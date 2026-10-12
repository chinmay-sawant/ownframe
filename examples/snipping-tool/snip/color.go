package snip

import (
	"image/color"
	"strconv"
	"strings"
)

// parseColor maps "#rrggbb", "#rgb", or a basic name to a color.
// Anything else falls back to opaque red.
func parseColor(s string) color.Color {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "white":
		return color.RGBA{255, 255, 255, 255}
	case "black":
		return color.RGBA{0, 0, 0, 255}
	case "green":
		return color.RGBA{0, 128, 0, 255}
	case "blue":
		return color.RGBA{0, 0, 255, 255}
	case "yellow":
		return color.RGBA{255, 255, 0, 255}
	}
	t := strings.TrimPrefix(strings.TrimSpace(s), "#")
	if len(t) == 3 {
		t = string([]byte{t[0], t[0], t[1], t[1], t[2], t[2]})
	}
	if len(t) == 6 {
		if n, err := strconv.ParseUint(t, 16, 32); err == nil {
			return color.RGBA{uint8(n >> 16), uint8(n >> 8), uint8(n), 255}
		}
	}
	return color.RGBA{255, 0, 0, 255}
}
