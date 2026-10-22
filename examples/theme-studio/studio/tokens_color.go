package studio

import (
	"math"
	"strconv"
	"strings"
)

// parseRGB reads #rgb, #rrggbb, #rrggbbaa, rgb(), and rgba() values. Alpha
// is ignored. rgb() channels are 0..255 ints or 0%..100% percentages.
func parseRGB(v string) ([3]float64, bool) {
	var out [3]float64

	v = strings.TrimSpace(v)

	if strings.HasPrefix(v, "#") {
		return parseHex(v[1:])
	}

	open := strings.IndexByte(v, '(')
	if open < 0 || !strings.HasSuffix(v, ")") {
		return out, false
	}

	fn := strings.ToLower(strings.TrimSpace(v[:open]))
	if fn != "rgb" && fn != "rgba" {
		return out, false
	}

	parts := strings.Split(v[open+1:len(v)-1], ",")
	if len(parts) != 3 && len(parts) != 4 {
		return out, false
	}

	for i := range out {
		f, ok := parseChannel(parts[i])
		if !ok {
			return out, false
		}
		out[i] = f
	}

	return out, true
}

// parseChannel reads one channel: an int 0..255 or a percentage.
func parseChannel(s string) (float64, bool) {
	s = strings.TrimSpace(s)

	full, whole := 255.0, true

	if strings.HasSuffix(s, "%") {
		s, full, whole = strings.TrimSpace(s[:len(s)-1]), 100, false
	}

	n, err := strconv.ParseFloat(s, 64)
	if err != nil || n < 0 || n > full || (whole && n != math.Trunc(n)) {
		return 0, false
	}

	return n / full, true
}

// parseHex reads 3, 6, or 8 hex digits. An alpha pair is checked, then
// dropped.
func parseHex(h string) ([3]float64, bool) {
	var out [3]float64

	if len(h) == hexShortLen {
		h = h[:1] + h[:1] + h[1:2] + h[1:2] + h[2:3] + h[2:3]
	}

	if len(h) != 6 && len(h) != 8 {
		return out, false
	}

	for i := 0; i < len(h)/2; i++ {
		n, err := strconv.ParseUint(h[i*2:i*2+2], 16, 8)
		if err != nil {
			return out, false
		}
		if i < hexShortLen {
			out[i] = float64(n) / channelMax
		}
	}

	return out, true
}

// trimImportant drops a trailing !important and the space before it.
func trimImportant(v string) string {
	const mark = "!important"

	if strings.HasSuffix(strings.ToLower(v), mark) {
		v = strings.TrimSpace(v[:len(v)-len(mark)])
	}

	return v
}
