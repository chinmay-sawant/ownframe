package cssanimation

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/animation"
)

// ErrLayout reports a layout-tier keyframe value the parser rejects. The binder
// keeps the first one as the bind error for the element, so a malformed rule
// is counted once at bind rather than dropped per frame.
var ErrLayout = errors.New("cssanimation: invalid layout value")

// layoutKind is the value grammar one layout-tier property follows. The
// grammar decides which keywords are legal, whether a percentage resolves, and
// whether a negative value is valid CSS.
type layoutKind uint8

const (
	layoutExtent layoutKind = iota
	layoutMinExtent
	layoutMaxExtent
	layoutInset
	layoutMargin
	layoutPadding
	layoutGap
	layoutFlexBasis
	layoutFlexFactor
	layoutBorderWidth
	layoutRadius
	layoutFontSize
	layoutLineHeight
	layoutLetterSpacing
	layoutWordSpacing
)

// layoutArity is the maximum number of terms a declaration takes. A shorthand
// takes one to four, a pair takes one or two, a longhand takes one, and a
// radius takes one to four on each side of its slash.
type layoutArity uint8

const (
	arityOne    layoutArity = 1
	arityTwo    layoutArity = 2
	arityFour   layoutArity = 4
	arityRadius layoutArity = 5
)

// Absolute length conversions the parser resolves to CSS pixels. A relative
// unit is left for the engine, which resolves it against the element's font and
// containing block.
const (
	pxPerInch  = 96
	pxPerPoint = pxPerInch / 72
	pxPerPica  = pxPerInch / 6
	pxPerCm    = pxPerInch / 2.54
	pxPerMm    = pxPerInch / 25.4
)

// LayoutKeyframes parses the layout-tier declarations of one keyframe rule.
// Paint declarations are ignored; the paint classes parse those. The returned
// frames keep the rule's timing functions and any declaration error, so the
// binder reports a malformed value once instead of animating a wrong length.
func LayoutKeyframes(rules []layout.KeyframesRule, name string) ([]animation.Keyframe, error) {
	source := ruleFor(rules, name)
	frames := make([]animation.Keyframe, len(source))
	for i, item := range source {
		frame := animation.Keyframe{Offset: item.Offset}
		for _, declaration := range item.Declarations {
			if keyframeTimingFunction(declaration) {
				frame.TimingFunction = strings.TrimSpace(declaration.Value)

				continue
			}
			property := strings.ToLower(strings.TrimSpace(declaration.Property))
			if !layoutProperty(property) {
				continue
			}
			value, err := parseLayoutValue(property, declaration.Value)
			if err != nil {
				return nil, err
			}
			frame.Layout = append(frame.Layout, value)
		}
		frames[i] = frame
	}
	sort.SliceStable(frames, func(i, j int) bool { return frames[i].Offset < frames[j].Offset })

	return frames, nil
}

// BindLayout builds one player for an element's layout-tier keyframes, or nil
// when the element declares none. One player carries every layout animation on
// the element, so a card that animates width, height, and border-radius reaches
// the engine's relayout path once per frame with one override list.
//
// The caller starts the player at the element's clock, which is the clock the
// element's paint players use. Layout samples stay on that clock when a patch
// rebinds them.
func BindLayout(d *layout.Display, box layout.Box, start time.Time) (*animation.Player, error) {
	if d == nil || len(box.Animations) == 0 {
		return nil, nil
	}

	effects := make([]animation.Effect, 0, len(box.Animations))
	for i := range box.Animations {
		item := box.Animations[i]
		if item.Name == "" || item.Name == "none" {
			continue
		}
		frames, err := LayoutKeyframes(d.Keyframes, item.Name)
		if err != nil {
			return nil, fmt.Errorf("keyframe %s: %w", item.Name, err)
		}
		if !hasLayoutFrames(frames) {
			continue
		}
		effects = append(effects, animation.Effect{
			Frames: frames, Timing: timing(item), Trigger: item.Trigger,
		})
	}
	if len(effects) == 0 {
		return nil, nil
	}

	player, err := animation.NewLayout(effects)
	if err != nil {
		return nil, err
	}
	player.Start(start)

	return player, nil
}

// hasLayoutFrames reports whether a parsed rule declares a layout-tier
// property.
func hasLayoutFrames(frames []animation.Keyframe) bool {
	for i := range frames {
		if len(frames[i].Layout) > 0 {
			return true
		}
	}

	return false
}

// layoutProperty reports a property this phase animates through the engine's
// override relayout path.
func layoutProperty(property string) bool {
	_, _, ok := layoutGrammar(property)

	return ok
}

// deferredLayoutProperty reports a layout-tier property this phase does not
// parse yet. The binder counts it as a phase 04 gap instead of dropping it
// silently. The tier is wider than the override path this phase wires: grid
// tracks, aspect ratio, shape and contain values, shadow and background paint,
// and the rule and column geometry all still need a driver.
func deferredLayoutProperty(property string) bool {
	switch property {
	case "box-shadow", "text-shadow", "background-image", "background-position",
		"background-size", "outline-width", "outline-offset",
		"column-rule-width", "row-rule-width", "rule-width",
		"aspect-ratio", "grid-template-columns", "grid-template-rows",
		"grid-auto-columns", "grid-auto-rows", "border-spacing",
		"column-width", "column-height", "text-indent", "tab-size", "vertical-align",
		"text-decoration-thickness", "text-underline-offset",
		"object-position", "object-view-box",
		"shape-outside", "shape-margin", "shape-padding", "shape-inside", "float-offset",
		"contain-intrinsic-size", "contain-intrinsic-width", "contain-intrinsic-height",
		"contain-intrinsic-inline-size", "contain-intrinsic-block-size",
		"overflow-clip-margin", "overflow-clip-margin-inline", "overflow-clip-margin-block",
		"clip":
		return true
	}

	return strings.HasPrefix(property, "border-image") ||
		strings.HasPrefix(property, "column-rule-inset") ||
		strings.HasPrefix(property, "row-rule-inset") ||
		strings.HasPrefix(property, "rule-inset")
}

// layoutGrammar maps a property to its value grammar and maximum arity.
//
//nolint:cyclop,funlen // the tier is one flat property table
func layoutGrammar(property string) (layoutKind, layoutArity, bool) {
	switch property {
	case "width", "height", "inline-size", "block-size":
		return layoutExtent, arityOne, true
	case "min-width", "min-height", "min-inline-size", "min-block-size":
		return layoutMinExtent, arityOne, true
	case "max-width", "max-height", "max-inline-size", "max-block-size":
		return layoutMaxExtent, arityOne, true
	case "top", "right", "bottom", "left",
		"inset-block-start", "inset-block-end", "inset-inline-start", "inset-inline-end":
		return layoutInset, arityOne, true
	case "inset":
		return layoutInset, arityFour, true
	case "inset-block", "inset-inline":
		return layoutInset, arityTwo, true
	case "margin-top", "margin-right", "margin-bottom", "margin-left",
		"margin-block-start", "margin-block-end", "margin-inline-start", "margin-inline-end":
		return layoutMargin, arityOne, true
	case "margin":
		return layoutMargin, arityFour, true
	case "margin-block", "margin-inline":
		return layoutMargin, arityTwo, true
	case "padding-top", "padding-right", "padding-bottom", "padding-left",
		"padding-block-start", "padding-block-end", "padding-inline-start", "padding-inline-end":
		return layoutPadding, arityOne, true
	case "padding":
		return layoutPadding, arityFour, true
	case "padding-block", "padding-inline":
		return layoutPadding, arityTwo, true
	case "gap":
		return layoutGap, arityTwo, true
	case "row-gap", "column-gap":
		return layoutGap, arityOne, true
	case "flex-basis":
		return layoutFlexBasis, arityOne, true
	case "flex-grow", "flex-shrink":
		return layoutFlexFactor, arityOne, true
	case "border-width":
		return layoutBorderWidth, arityFour, true
	case "border-top-width", "border-right-width", "border-bottom-width", "border-left-width",
		"border-block-start-width", "border-block-end-width",
		"border-inline-start-width", "border-inline-end-width":
		return layoutBorderWidth, arityOne, true
	case "border-block-width", "border-inline-width":
		return layoutBorderWidth, arityTwo, true
	case "border-radius":
		return layoutRadius, arityRadius, true
	case "border-top-left-radius", "border-top-right-radius",
		"border-bottom-right-radius", "border-bottom-left-radius",
		"border-start-start-radius", "border-start-end-radius",
		"border-end-start-radius", "border-end-end-radius":
		return layoutRadius, arityOne, true
	case "font-size":
		return layoutFontSize, arityOne, true
	case "line-height":
		return layoutLineHeight, arityOne, true
	case "letter-spacing":
		return layoutLetterSpacing, arityOne, true
	case "word-spacing":
		return layoutWordSpacing, arityOne, true
	}

	return 0, 0, false
}

// parseLayoutValue parses one declaration into the sampled-value form.
func parseLayoutValue(property, raw string) (animation.LayoutValue, error) {
	kind, arity, ok := layoutGrammar(property)
	if !ok {
		return animation.LayoutValue{}, layoutErr("%q is not a layout-tier property", property)
	}

	terms, err := parseLayoutTerms(kind, arity, raw)
	if err != nil {
		return animation.LayoutValue{}, layoutErr("%s: %s", property, err.Error())
	}

	return animation.LayoutValue{Property: property, Terms: terms}, nil
}

// parseLayoutTerms splits one declaration into its terms. A radius declaration
// carries an optional slash, so its two halves parse with the same grammar and
// keep the separator as a keyword term.
func parseLayoutTerms(kind layoutKind, arity layoutArity, value string) ([]animation.LayoutTerm, error) {
	if arity == arityRadius {
		return parseRadiusTerms(value)
	}

	fields := strings.Fields(strings.TrimSpace(value))
	if len(fields) == 0 {
		return nil, layoutErr("empty value")
	}
	if len(fields) > int(arity) {
		return nil, layoutErr("%d values where at most %d are allowed", len(fields), arity)
	}

	terms := make([]animation.LayoutTerm, len(fields))
	for i, field := range fields {
		term, err := parseLayoutTerm(kind, field)
		if err != nil {
			return nil, err
		}
		terms[i] = term
	}

	return terms, nil
}

// parseRadiusTerms parses the 1-4 horizontal and optional 1-4 vertical radii of
// border-radius. The slash stays a term, so two matching lists interpolate
// their numeric terms across it.
func parseRadiusTerms(value string) ([]animation.LayoutTerm, error) {
	horizontal, vertical, hasVertical := strings.Cut(strings.TrimSpace(value), "/")
	terms, err := parseLayoutTerms(layoutRadius, arityFour, horizontal)
	if err != nil {
		return nil, err
	}
	if !hasVertical {
		return terms, nil
	}

	verticalTerms, err := parseLayoutTerms(layoutRadius, arityFour, vertical)
	if err != nil {
		return nil, err
	}

	return append(append(terms, animation.LayoutTerm{Keyword: "/"}), verticalTerms...), nil
}

// parseLayoutTerm parses one whitespace-separated term: a number with a unit,
// a unitless number where the grammar allows one, or a keyword.
func parseLayoutTerm(kind layoutKind, token string) (animation.LayoutTerm, error) {
	number, unit, numeric := splitLayoutNumber(token)
	if !numeric {
		keyword, ok := layoutKeyword(kind, token)
		if !ok {
			return animation.LayoutTerm{}, layoutErr("%q is not a valid value", token)
		}

		return animation.LayoutTerm{Keyword: keyword}, nil
	}

	return layoutNumericTerm(kind, number, unit, token)
}

// layoutNumericTerm validates a numeric term against its property grammar and
// resolves an absolute unit to px.
func layoutNumericTerm(kind layoutKind, number float64, unit, token string) (animation.LayoutTerm, error) {
	unit = strings.ToLower(unit)
	if unit == "" && kind != layoutFlexFactor && kind != layoutLineHeight {
		if number != 0 {
			return animation.LayoutTerm{}, layoutErr("%q needs a unit", token)
		}
		unit = "px"
	}

	switch kind {
	case layoutFlexFactor:
		if unit != "" {
			return animation.LayoutTerm{}, layoutErr("%q must be a unitless number", token)
		}
		if number < 0 {
			return animation.LayoutTerm{}, layoutErr("%q must not be negative", token)
		}

		return animation.LayoutTerm{Number: number}, nil
	case layoutLineHeight:
		if unit == "" {
			if number < 0 {
				return animation.LayoutTerm{}, layoutErr("%q must not be negative", token)
			}

			return animation.LayoutTerm{Number: number}, nil
		}
		if !layoutAllowsPercent(kind) && unit == "%" {
			return animation.LayoutTerm{}, layoutErr("%q does not take a percentage", token)
		}
	}

	switch unit {
	case "%":
		if !layoutAllowsPercent(kind) {
			return animation.LayoutTerm{}, layoutErr("%q does not take a percentage", token)
		}
	case "px", "pt", "pc", "in", "cm", "mm":
		number, unit = resolveAbsoluteLength(number, unit), "px"
	case "em", "rem", "ex", "ch", "vw", "vh", "vmin", "vmax":
	case "":
		// A zero length carried no unit; it was normalized to px above.
	default:
		return animation.LayoutTerm{}, layoutErr("%q has an unsupported unit", token)
	}

	if number < 0 && !layoutAllowsNegative(kind) {
		return animation.LayoutTerm{}, layoutErr("%q must not be negative", token)
	}

	return animation.LayoutTerm{Number: number, Unit: unit}, nil
}

// splitLayoutNumber splits a number from a unit suffix. A unitless token
// reports numeric with an empty unit; a token that is not a number reports
// false so the caller can try the keyword grammar.
func splitLayoutNumber(token string) (float64, string, bool) {
	lower := strings.ToLower(strings.TrimSpace(token))
	unit := ""
	for _, candidate := range []string{
		"vmin", "vmax", "rem", "px", "pt", "pc", "in", "cm", "mm",
		"em", "ex", "ch", "vw", "vh", "%",
	} {
		if strings.HasSuffix(lower, candidate) {
			unit = candidate

			break
		}
	}
	prefix := strings.TrimSpace(lower[:len(lower)-len(unit)])
	number, err := strconv.ParseFloat(prefix, 64)
	if err != nil {
		return 0, "", false
	}

	return number, unit, true
}

// resolveAbsoluteLength converts one absolute CSS length to px. A relative
// length stays in its authored unit because the engine resolves it against the
// element's font and containing block.
func resolveAbsoluteLength(number float64, unit string) float64 {
	switch unit {
	case "pt":
		return number * pxPerPoint
	case "pc":
		return number * pxPerPica
	case "in":
		return number * pxPerInch
	case "cm":
		return number * pxPerCm
	case "mm":
		return number * pxPerMm
	default:
		return number
	}
}

// layoutAllowsPercent reports whether the property accepts a percentage. A
// border width and letter-spacing are lengths only.
func layoutAllowsPercent(kind layoutKind) bool {
	switch kind {
	case layoutBorderWidth, layoutLetterSpacing, layoutFlexFactor:
		return false
	}

	return true
}

// layoutAllowsNegative reports whether the state a property writes may be
// negative. A margin, an inset, and letter- or word-spacing may; a size, a
// padding, a gap, a radius, and a text metric may not.
func layoutAllowsNegative(kind layoutKind) bool {
	switch kind {
	case layoutInset, layoutMargin, layoutLetterSpacing, layoutWordSpacing:
		return true
	}

	return false
}

// layoutKeyword reports whether token is a keyword the property accepts, and
// its lower-cased form.
func layoutKeyword(kind layoutKind, token string) (string, bool) {
	keyword := strings.ToLower(strings.TrimSpace(token))
	switch kind {
	case layoutExtent:
		return keywordInSet(keyword, "auto", "none", "min-content", "max-content",
			"fit-content", "stretch")
	case layoutMinExtent:
		return keywordInSet(keyword, "auto", "min-content", "max-content",
			"fit-content", "stretch")
	case layoutMaxExtent:
		return keywordInSet(keyword, "none", "auto", "min-content", "max-content",
			"fit-content", "stretch")
	case layoutInset, layoutMargin:
		return keywordInSet(keyword, "auto")
	case layoutPadding:
		return "", false
	case layoutGap:
		return keywordInSet(keyword, "normal")
	case layoutFlexBasis:
		return keywordInSet(keyword, "auto", "content", "min-content", "max-content",
			"fit-content", "stretch")
	case layoutFlexFactor:
		return "", false
	case layoutBorderWidth:
		return keywordInSet(keyword, "thin", "medium", "thick")
	case layoutRadius:
		return "", false
	case layoutFontSize:
		return keywordInSet(keyword, "xx-small", "x-small", "small", "medium", "large",
			"x-large", "xx-large", "xxx-large", "smaller", "larger", "math")
	case layoutLineHeight:
		return keywordInSet(keyword, "normal")
	case layoutLetterSpacing, layoutWordSpacing:
		return keywordInSet(keyword, "normal")
	}

	return "", false
}

// keywordInSet reports whether keyword is one of set.
func keywordInSet(keyword string, set ...string) (string, bool) {
	for _, candidate := range set {
		if keyword == candidate {
			return keyword, true
		}
	}

	return "", false
}

func layoutErr(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrLayout, fmt.Sprintf(format, args...))
}
