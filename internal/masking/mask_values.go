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
		return "", "", fmt.Errorf("expected auto, cover, contain, or one or two lengths")
	}
	if len(parts) == 1 {
		low := strings.ToLower(parts[0])
		if low == "cover" || low == "contain" {
			return low, "", nil
		}
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

type maskAxisPos struct {
	edge    string // "left", "right", "top", "bottom", "center", or ""
	value   float64
	percent bool
	hasOff  bool
}

const (
	maskPosPairTokens     = 2
	maskPosMinMultiTokens = 3
	maskPosMaxMultiTokens = 4
)

func maskPosition(value string) (maskAxisPos, maskAxisPos, error) {
	parts := strings.Fields(strings.TrimSpace(value))
	switch len(parts) {
	case 1:
		return parseMaskPosPair(parts[0], "50%")
	case maskPosPairTokens:
		return parseMaskPosPair(parts[0], parts[1])
	case maskPosMinMultiTokens, maskPosMaxMultiTokens:
		return parseMaskPosMulti(parts)
	default:
		return maskAxisPos{}, maskAxisPos{}, fmt.Errorf("expected one to four position tokens")
	}
}

func parseMaskPosPair(xTok, yTok string) (maskAxisPos, maskAxisPos, error) {
	xv, xp, err := cssPosition(xTok)
	if err != nil {
		return maskAxisPos{}, maskAxisPos{}, err
	}
	yv, yp, err := cssPosition(yTok)
	if err != nil {
		return maskAxisPos{}, maskAxisPos{}, err
	}
	return maskAxisPos{value: xv, percent: xp}, maskAxisPos{value: yv, percent: yp}, nil
}

func parseMaskPosMulti(parts []string) (maskAxisPos, maskAxisPos, error) {
	xAxis := maskAxisPos{value: centerPosition, percent: true}
	yAxis := maskAxisPos{value: centerPosition, percent: true}
	xSet, ySet := false, false

	for i := 0; i < len(parts); {
		tok := strings.ToLower(parts[i])
		switch tok {
		case "left", "right":
			if xSet {
				return maskAxisPos{}, maskAxisPos{}, fmt.Errorf("duplicate horizontal position %q", tok)
			}
			xSet = true
			xAxis = maskAxisPos{edge: tok}
			if tok == "right" {
				xAxis.value, xAxis.percent = 1, true
			} else {
				xAxis.value, xAxis.percent = 0, true
			}
			i++
			if i < len(parts) && !isPosEdgeKeyword(parts[i]) {
				v, pct, err := cssLength(parts[i])
				if err != nil {
					return maskAxisPos{}, maskAxisPos{}, err
				}
				xAxis.value, xAxis.percent, xAxis.hasOff = v, pct, true
				i++
			}
		case "top", "bottom":
			if ySet {
				return maskAxisPos{}, maskAxisPos{}, fmt.Errorf("duplicate vertical position %q", tok)
			}
			ySet = true
			yAxis = maskAxisPos{edge: tok}
			if tok == "bottom" {
				yAxis.value, yAxis.percent = 1, true
			} else {
				yAxis.value, yAxis.percent = 0, true
			}
			i++
			if i < len(parts) && !isPosEdgeKeyword(parts[i]) {
				v, pct, err := cssLength(parts[i])
				if err != nil {
					return maskAxisPos{}, maskAxisPos{}, err
				}
				yAxis.value, yAxis.percent, yAxis.hasOff = v, pct, true
				i++
			}
		case "center":
			switch {
			case !xSet:
				xSet = true
				xAxis = maskAxisPos{edge: "center", value: centerPosition, percent: true}
			case !ySet:
				ySet = true
				yAxis = maskAxisPos{edge: "center", value: centerPosition, percent: true}
			default:
				return maskAxisPos{}, maskAxisPos{}, fmt.Errorf("unexpected center token")
			}
			i++
		default:
			return maskAxisPos{}, maskAxisPos{}, fmt.Errorf("unexpected position token %q", parts[i])
		}
	}

	return xAxis, yAxis, nil
}

func isPosEdgeKeyword(tok string) bool {
	switch strings.ToLower(tok) {
	case "left", "right", "top", "bottom", "center":
		return true
	default:
		return false
	}
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
