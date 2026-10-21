package masking

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func maskSize(value string) (string, string, error) {
	parts := strings.Fields(strings.TrimSpace(value))
	if len(parts) == 0 || len(parts) > 2 {
		return "", "", fmt.Errorf("expected auto or one or two lengths")
	}
	if len(parts) == 1 {
		parts = append(parts, "auto")
	}
	for _, part := range parts {
		if part != "auto" {
			value, _, err := cssLength(part)
			if err != nil || math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
				return "", "", fmt.Errorf("invalid length %q", part)
			}
		}
	}
	return parts[0], parts[1], nil
}

func maskPosition(value string) (string, string, error) {
	parts := strings.Fields(strings.TrimSpace(value))
	if len(parts) == 0 || len(parts) > 2 {
		return "", "", fmt.Errorf("expected one or two positions")
	}
	if len(parts) == 1 {
		parts = append(parts, "50%")
	}
	if _, _, err := cssPosition(parts[0]); err != nil {
		return "", "", err
	}
	if _, _, err := cssPosition(parts[1]); err != nil {
		return "", "", err
	}
	return parts[0], parts[1], nil
}

func cssLength(value string) (float64, bool, error) {
	if strings.HasSuffix(value, "%") {
		n, err := strconv.ParseFloat(strings.TrimSuffix(value, "%"), 64)
		return n / percentDivisor, true, err
	}
	return cssNumber(value)
}

func cssPosition(value string) (float64, bool, error) {
	switch value {
	case "left", "top":
		return 0, true, nil
	case "center":
		return centerPosition, true, nil
	case "right", "bottom":
		return 1, true, nil
	}
	return cssLength(value)
}

func cssNumber(value string) (float64, bool, error) {
	value = strings.TrimSuffix(value, "px")
	n, err := strconv.ParseFloat(value, 64)
	if err == nil && (math.IsNaN(n) || math.IsInf(n, 0)) {
		err = fmt.Errorf("invalid number %q", value)
	}
	return n, false, err
}
