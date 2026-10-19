package csspreview

import (
	"strconv"
	"strings"
)

func number(s string) (float64, bool) {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return v, err == nil
}

func color(s string) (float64, float64, float64, bool) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "#")
	if len(s) == 3 {
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	}
	if len(s) != 6 {
		return 0, 0, 0, false
	}
	n, err := strconv.ParseUint(s, 16, 24)
	if err != nil {
		return 0, 0, 0, false
	}
	return float64(n>>16) / 255, float64(n>>8&255) / 255, float64(n&255) / 255, true
}

func transformShift(s string) (float64, float64, bool) {
	s = strings.TrimSpace(s)
	for _, name := range []string{"translateX", "translateY", "translate"} {
		prefix := name + "("
		if !strings.HasPrefix(s, prefix) || !strings.HasSuffix(s, ")") {
			continue
		}
		parts := strings.Split(strings.TrimSuffix(strings.TrimPrefix(s, prefix), ")"), ",")
		x, y := 0.0, 0.0
		if v, ok := pixels(parts[0]); ok {
			if name == "translateY" {
				y = v
			} else {
				x = v
			}
		} else {
			return 0, 0, false
		}
		if len(parts) > 1 {
			if v, ok := pixels(parts[1]); ok {
				y = v
			} else {
				return 0, 0, false
			}
		}
		return x, y, true
	}
	return 0, 0, false
}

func pixels(s string) (float64, bool) {
	s = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(s), "px"))
	return number(s)
}
