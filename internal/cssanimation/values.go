package cssanimation

import (
	"strconv"
	"strings"
)

func number(value string) (float64, bool) {
	v, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	return v, err == nil
}

func color(value string) (float64, float64, float64, bool) {
	value = strings.TrimPrefix(strings.TrimSpace(value), "#")
	if len(value) == 3 {
		value = string([]byte{value[0], value[0], value[1], value[1], value[2], value[2]})
	}
	if len(value) != 6 {
		return 0, 0, 0, false
	}
	n, err := strconv.ParseUint(value, 16, 24)
	if err != nil {
		return 0, 0, 0, false
	}
	return float64(n>>16) / 255, float64(n>>8&255) / 255, float64(n&255) / 255, true
}

func transformShift(value string) (float64, float64, bool) {
	value = strings.TrimSpace(value)
	for _, name := range []string{"translateX", "translateY", "translate"} {
		prefix := name + "("
		if !strings.HasPrefix(value, prefix) || !strings.HasSuffix(value, ")") {
			continue
		}
		return shift(name, strings.TrimSuffix(strings.TrimPrefix(value, prefix), ")"))
	}
	return 0, 0, false
}

func shift(name, body string) (float64, float64, bool) {
	parts := strings.Split(body, ",")
	x, y := 0.0, 0.0
	first, ok := pixels(parts[0])
	if !ok {
		return 0, 0, false
	}
	if name == "translateY" {
		y = first
	} else {
		x = first
	}
	if len(parts) > 1 {
		y, ok = pixels(parts[1])
	}
	return x, y, ok
}

func pixels(value string) (float64, bool) {
	return number(strings.TrimSuffix(strings.TrimSpace(value), "px"))
}
