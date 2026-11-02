package backdrop

import (
	"math"
	"strconv"
	"strings"
)

// Filter is the subset of backdrop-filter operations painted by OwnFrame.
type Filter struct {
	BlurPx       float64
	Saturation   float64
	RefractionPx float64
}

// Painted bounds for the backdrop-filter subset.
const (
	maxBlurPx       = 64
	maxSaturation   = 3
	maxRefractionPx = 16
	percentDivisor  = 100
)

// Parse accepts one each of blur, saturation, and the host refract extension.
func Parse(value string) (Filter, bool) {
	filter := Filter{Saturation: 1}
	value = strings.TrimSpace(value)
	if value == "" || strings.EqualFold(value, "none") {
		return Filter{}, false
	}

	seen := false
	seenBlur, seenSaturation, seenRefraction := false, false, false
	for value != "" {
		value = strings.TrimSpace(value)
		open := strings.IndexByte(value, '(')
		if open <= 0 {
			return Filter{}, false
		}
		close := strings.IndexByte(value[open+1:], ')')
		if close < 0 {
			return Filter{}, false
		}
		close += open + 1
		name := strings.ToLower(strings.TrimSpace(value[:open]))
		argument := strings.TrimSpace(value[open+1 : close])

		switch name {
		case "blur":
			if seenBlur {
				return Filter{}, false
			}
			radius, ok := parseLength(argument)
			if !ok {
				return Filter{}, false
			}
			filter.BlurPx = math.Min(radius, maxBlurPx)
			seenBlur = true
			seen = true
		case "saturate":
			if seenSaturation {
				return Filter{}, false
			}
			saturation, ok := parseAmount(argument)
			if !ok {
				return Filter{}, false
			}
			filter.Saturation = math.Min(saturation, maxSaturation)
			seenSaturation = true
			seen = true
		case "refract":
			if seenRefraction {
				return Filter{}, false
			}
			strength, ok := parseLength(argument)
			if !ok {
				return Filter{}, false
			}
			filter.RefractionPx = math.Min(strength, maxRefractionPx)
			seenRefraction = true
			seen = true
		default:
			return Filter{}, false
		}

		value = value[close+1:]
	}

	return filter, seen && (filter.BlurPx > 0 || filter.Saturation != 1 || filter.RefractionPx > 0)
}

func parseLength(value string) (float64, bool) {
	if value == "0" {
		return 0, true
	}
	if !strings.HasSuffix(value, "px") {
		return 0, false
	}
	return parseNumber(strings.TrimSpace(strings.TrimSuffix(value, "px")))
}

func parseAmount(value string) (float64, bool) {
	if strings.HasSuffix(value, "%") {
		amount, ok := parseNumber(strings.TrimSpace(strings.TrimSuffix(value, "%")))
		return amount / percentDivisor, ok
	}
	return parseNumber(value)
}

func parseNumber(value string) (float64, bool) {
	amount, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	return amount, err == nil && !math.IsNaN(amount) && !math.IsInf(amount, 0) && amount >= 0
}
