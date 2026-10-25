package cssanimation

import (
	"errors"
	"fmt"
	"strings"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/animation"
)

// ErrSVGPaint reports an inline-SVG paint declaration a keyframe cannot parse.
var ErrSVGPaint = errors.New("cssanimation: invalid svg paint")

// ErrPaintServerDeferred reports a fill or stroke that references a paint
// server, such as url(#gradient). The engine stores paint as sRGB, so a paint
// server cannot be restyled from a sampled colour. The declaration is not a
// bind error: the element keeps its base paint and the sample carries the
// reason, rather than the host guessing a colour.
var ErrPaintServerDeferred = errors.New("cssanimation: paint server is deferred")

// isSVGPaintProperty reports one keyframe property sampled for an inline SVG.
// These live outside paintClass because they target no display operation:
// changing them rebuilds the SVG raster through layout.RestyleInlineSVG.
func isSVGPaintProperty(property string) bool {
	switch property {
	case "fill", "fill-opacity", "stroke", "stroke-opacity", "stroke-width",
		"stroke-dasharray", "stroke-dashoffset":
		return true
	}

	return false
}

// parseSVGPaintDeclaration writes one inline-SVG paint declaration into a
// frame. A paint server and a colour space this package cannot represent are
// deferred with a reason. A malformed value fails the bind, like every other
// paint declaration.
func parseSVGPaintDeclaration(frame *animation.Keyframe, declaration layout.KeyframeDeclaration,
	size transformSize, current Color, hasCurrent bool) {
	switch declaration.Property {
	case "fill":
		value, err := svgColor(declaration.Value, current, hasCurrent)
		if err != nil {
			noteSVGPaintError(frame, declaration.Property, err)
			return
		}
		frame.SVG.Fill, frame.SVG.HasFill = value, true
	case "fill-opacity":
		value, err := svgOpacity(declaration.Value)
		if err != nil {
			noteSVGPaintError(frame, declaration.Property, err)
			return
		}
		frame.SVG.FillOpacity, frame.SVG.HasFillOpacity = value, true
	case "stroke":
		value, err := svgColor(declaration.Value, current, hasCurrent)
		if err != nil {
			noteSVGPaintError(frame, declaration.Property, err)
			return
		}
		frame.SVG.Stroke, frame.SVG.HasStroke = value, true
	case "stroke-opacity":
		value, err := svgOpacity(declaration.Value)
		if err != nil {
			noteSVGPaintError(frame, declaration.Property, err)
			return
		}
		frame.SVG.StrokeOpacity, frame.SVG.HasStrokeOpacity = value, true
	case "stroke-width":
		value, err := svgPaintLength(declaration.Value, size)
		if err != nil {
			noteSVGPaintError(frame, declaration.Property, err)
			return
		}
		if value < 0 {
			noteSVGPaintError(frame, declaration.Property,
				fmt.Errorf("%w: stroke-width %q is negative", ErrSVGPaint, declaration.Value))
			return
		}
		frame.SVG.StrokeWidth, frame.SVG.HasStrokeWidth = value, true
	case "stroke-dasharray":
		value, err := svgDashArray(declaration.Value, size)
		if err != nil {
			noteSVGPaintError(frame, declaration.Property, err)
			return
		}
		frame.SVG.StrokeDashArray, frame.SVG.HasStrokeDashArray = value, true
	case "stroke-dashoffset":
		value, err := svgPaintLength(declaration.Value, size)
		if err != nil {
			noteSVGPaintError(frame, declaration.Property, err)
			return
		}
		frame.SVG.StrokeDashOffset, frame.SVG.HasStrokeDashOffset = value, true
	}
}

// noteSVGPaintError records a declaration this phase cannot turn into a used
// value. A paint server or a colour space the package does not model is
// deferred with a reason; every other error fails the bind.
func noteSVGPaintError(frame *animation.Keyframe, property string, err error) {
	if errors.Is(err, ErrPaintServerDeferred) || errors.Is(err, ErrColorSpace) {
		frame.SVG.Deferred = property + ": " + err.Error()
		return
	}
	frame.Err = fmt.Errorf("cssanimation: %s: %w", property, err)
}

// svgColor parses one fill or stroke value. none is the no-paint keyword, and
// every other form is the colour grammar the paint phase already accepts. A
// paint server reference is deferred: the engine has no RGB value for it.
func svgColor(value string, current Color, hasCurrent bool) (animation.SVGPaintColor, error) {
	text := strings.ToLower(strings.TrimSpace(value))
	if text == "none" {
		return animation.SVGPaintColor{None: true, A: maxChannel}, nil
	}
	if strings.HasPrefix(text, "url(") {
		return animation.SVGPaintColor{}, fmt.Errorf("%w: %s",
			ErrPaintServerDeferred, strings.TrimSpace(value))
	}
	parsed, err := resolveColor(value, current, hasCurrent)
	if err != nil {
		return animation.SVGPaintColor{}, err
	}

	return animation.SVGPaintColor{R: parsed.R, G: parsed.G, B: parsed.B, A: parsed.A}, nil
}

// svgOpacity parses fill-opacity or stroke-opacity as a number or a
// percentage and clamps it to the 0..1 range CSS uses.
func svgOpacity(value string) (float64, error) {
	text := strings.TrimSpace(value)
	if digits, ok := strings.CutSuffix(text, "%"); ok {
		parsed, err := finite(digits)
		if err != nil {
			return 0, fmt.Errorf("%w: opacity %q", ErrSVGPaint, value)
		}

		return clamp01(parsed / percentScale), nil
	}
	parsed, err := finite(text)
	if err != nil {
		return 0, fmt.Errorf("%w: opacity %q", ErrSVGPaint, value)
	}

	return clamp01(parsed), nil
}

// svgPaintLength parses one stroke length as layout points. A bare number and
// px are the same length, and pt is already a layout point. A unit that needs
// font or viewport context is rejected rather than resolved to a guess.
func svgPaintLength(value string, size transformSize) (float64, error) {
	text := strings.ToLower(strings.TrimSpace(value))
	if text == "" {
		return 0, fmt.Errorf("%w: empty length", ErrSVGPaint)
	}
	unit := ""
	switch {
	case strings.HasSuffix(text, "px"):
		unit = "px"
	case strings.HasSuffix(text, "pt"):
		unit = "pt"
	}
	number, err := finite(text[:len(text)-len(unit)])
	if err != nil {
		return 0, fmt.Errorf("%w: %q is not a length in px or pt", ErrSVGPaint, value)
	}
	if unit == "pt" {
		return number, nil
	}
	scale := size.pointsPerPixel
	if scale <= 0 {
		scale = 1
	}

	return number * scale, nil
}

// svgDashArray parses the CSS stroke-dasharray list as lengths in layout
// points. none is the empty list; a negative dash length is invalid.
func svgDashArray(value string, size transformSize) ([]float64, error) {
	text := strings.TrimSpace(value)
	if strings.EqualFold(text, "none") {
		return nil, nil
	}
	fields := strings.Fields(strings.ReplaceAll(text, ",", " "))
	if len(fields) == 0 {
		return nil, fmt.Errorf("%w: empty dash list", ErrSVGPaint)
	}
	out := make([]float64, len(fields))
	for i, field := range fields {
		length, err := svgPaintLength(field, size)
		if err != nil {
			return nil, err
		}
		if length < 0 {
			return nil, fmt.Errorf("%w: dash %q is negative", ErrSVGPaint, field)
		}
		out[i] = length
	}

	return out, nil
}
