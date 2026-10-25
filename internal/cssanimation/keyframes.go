package cssanimation

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/animation"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// paintClass is one property family an effect writes: the operations the
// property targets and the paint fields it fills. Two properties with the same
// target set but a different field mask stay separate classes, and two colour
// properties with different target kinds (a background fill against text ink)
// stay separate even though both write RGB.
type paintClass struct {
	kind cssmotion.Kind
	mask cssmotion.Mask
}

// propertyClass maps a keyframed property to the operations and paint fields
// it writes. ok is false for every property this phase leaves alone, including
// the ones deferredPhase names.
func propertyClass(property string) (paintClass, bool) {
	switch property {
	case "transform", "perspective", "perspective-origin":
		return paintClass{kind: cssmotion.KindSubtree, mask: cssmotion.MaskMatrix}, true
	case "opacity":
		return paintClass{kind: cssmotion.KindSubtree, mask: cssmotion.MaskOpacity}, true
	case "background-color", "accent-color":
		return paintClass{kind: cssmotion.KindFill, mask: cssmotion.MaskRGB}, true
	case "color", "text-decoration-color", "text-emphasis-color":
		return paintClass{kind: cssmotion.KindInk, mask: cssmotion.MaskRGB}, true
	case "border-color", "border-top-color", "border-right-color",
		"border-bottom-color", "border-left-color", "border-block-color",
		"border-block-start-color", "border-block-end-color", "border-inline-color",
		"border-inline-start-color", "border-inline-end-color", "outline-color",
		"column-rule-color", "row-rule-color", "rule-color":
		return paintClass{kind: cssmotion.KindStroke, mask: cssmotion.MaskRGB}, true
	}
	return paintClass{}, false
}

// deferredPhase reports the phase that owns a keyframe property this phase
// leaves at base paint. The binder counts one gap per element and property, so
// a deferred declaration is reported rather than dropped silently.
// Deferred property phases. The number is the phase that owns the property, so
// a gap report can point a reader at the right plan.
const (
	phaseTransform = 2
	phaseLayout    = 4
	phaseFilter    = 8
)

// deferredPhase names the phase that owns a property this binder leaves at the
// element's base paint. Inline-SVG paint is not here: the parser samples it in
// svg_values.go and BindBox routes the sample to layout.RestyleInlineSVG.
func deferredPhase(property string) (int, bool) {
	switch property {
	case "filter", "backdrop-filter", "clip-path", "mask-position", "mask-size",
		"mix-blend-mode", "isolation":
		return phaseFilter, true
	case "translate", "rotate", "scale":
		return phaseTransform, true
	}
	if deferredLayoutProperty(property) {
		return phaseLayout, true
	}
	return 0, false
}

// keyframes parses one rule's declarations for every paint class. The binder
// parses one class at a time through keyframesForBase; this form keeps the
// per-declaration view the transform and perspective tests use.
func keyframes(rules []layout.KeyframesRule, name string, pointsPerPixel, width, height float64) []animation.Keyframe {
	size := transformSize{width: width, height: height, pointsPerPixel: pointsPerPixel, hasBox: true}
	return parseKeyframes(rules, name, size, nil, Color{}, false, animation.PerspectiveBase{})
}

// keyframesFor parses one rule's declarations for one paint class. A
// declaration of another class is skipped, so the frames carry only the fields
// this class writes: a background class never sees the transform, and a text
// class never sees the background colour.
func keyframesFor(source []layout.Keyframe, size transformSize,
	class paintClass, current Color, hasCurrent bool) []animation.Keyframe {
	return parseFrames(source, size, &class, current, hasCurrent, animation.PerspectiveBase{})
}

// keyframesForBase is keyframesFor with the element's static perspective and
// vanishing point, so an implicit endpoint holds the element's static camera.
func keyframesForBase(source []layout.Keyframe, size transformSize,
	class paintClass, current Color, hasCurrent bool, base animation.PerspectiveBase) []animation.Keyframe {
	return parseFrames(source, size, &class, current, hasCurrent, base)
}

// parseKeyframes parses every declaration of one rule. A nil class keeps every
// property, which is the unfiltered view the legacy bind uses.
func parseKeyframes(rules []layout.KeyframesRule, name string, size transformSize,
	class *paintClass, current Color, hasCurrent bool,
	base animation.PerspectiveBase) []animation.Keyframe {
	return parseFrames(ruleFor(rules, name), size, class, current, hasCurrent, base)
}

func parseFrames(source []layout.Keyframe, size transformSize,
	class *paintClass, current Color, hasCurrent bool,
	base animation.PerspectiveBase) []animation.Keyframe {
	frames := make([]animation.Keyframe, len(source))
	for i, item := range source {
		frame := animation.Keyframe{
			Offset:          item.Offset,
			Perspective:     framePerspective(size, base),
			PerspectiveBase: base,
		}
		for _, declaration := range item.Declarations {
			// A keyframe timing function is not a paint property, so it is
			// read before the paint-class filter: every paint class of the
			// rule eases its segment with the same declared function.
			if keyframeTimingFunction(declaration) {
				frame.TimingFunction = strings.TrimSpace(declaration.Value)
				continue
			}
			// Inline-SVG paint targets no display operation, so it is read
			// before the paint-class filter: every class of the rule carries
			// the same sampled paint for the element's SVG box.
			if isSVGPaintProperty(declaration.Property) {
				parseSVGPaintDeclaration(&frame, declaration, size, current, hasCurrent)
				continue
			}
			pc, ok := propertyClass(declaration.Property)
			if !ok || (class != nil && pc != *class) {
				continue
			}
			parseDeclaration(&frame, declaration, size, current, hasCurrent)
		}
		frames[i] = frame
	}
	sort.SliceStable(frames, func(i, j int) bool { return frames[i].Offset < frames[j].Offset })
	return frames
}

// keyframeTimingFunction reports a declaration that eases the interval starting
// at this keyframe. It belongs to no paint class, so the parser handles it for
// every class. The vendor-prefixed form is the same longhand alias the engine
// accepts outside a keyframe block.
func keyframeTimingFunction(declaration layout.KeyframeDeclaration) bool {
	property := strings.TrimSpace(declaration.Property)
	return strings.EqualFold(property, "animation-timing-function") ||
		strings.EqualFold(property, "-webkit-animation-timing-function")
}

// parseDeclaration writes one declaration into a frame. A value this phase
// cannot parse is kept as a bind error; a colour space a later phase owns is
// left at the element's base paint and counted by the binder.
func parseDeclaration(frame *animation.Keyframe, declaration layout.KeyframeDeclaration,
	size transformSize, current Color, hasCurrent bool) {
	switch declaration.Property {
	case "transform":
		value, err := parseTransform(declaration.Value, size)
		if err != nil {
			frame.Err = err
			return
		}
		frame.Transform = value
		frame.X, frame.Y = value.Decomposed.Tx, value.Decomposed.Ty
		frame.Move = true
	case "perspective":
		distance, err := parsePerspective(declaration.Value, size)
		if err != nil {
			frame.Err = err
			return
		}
		frame.Perspective.Distance = distance
		frame.Perspective.HasDistance = true
	case "perspective-origin":
		x, y, err := parsePerspectiveOrigin(declaration.Value, size)
		if err != nil {
			frame.Err = err
			return
		}
		frame.Perspective.OriginX, frame.Perspective.OriginY = x, y
		frame.Perspective.HasOrigin = true
	case "opacity":
		value, ok := number(declaration.Value)
		if !ok {
			frame.Err = fmt.Errorf("%w: opacity %q", ErrNumber, declaration.Value)
			return
		}
		frame.Opacity, frame.OpacitySet = value, true
	default:
		value, err := resolveColor(declaration.Value, current, hasCurrent)
		if err != nil {
			if !errors.Is(err, ErrColorSpace) {
				frame.Err = err
			}
			return
		}
		frame.R, frame.G, frame.B, frame.ColorSet = value.R, value.G, value.B, true
		if value.A < maxChannel {
			frame.ColorAlpha, frame.ColorAlphaSet = value.A, true
		}
	}
}

// name returns the property family a class stands for, for error text.
func (c paintClass) name() string {
	if c.mask == cssmotion.MaskOpacity {
		return "opacity"
	}
	switch c.kind {
	case cssmotion.KindFill:
		return "fill colour"
	case cssmotion.KindInk:
		return "text colour"
	case cssmotion.KindStroke:
		return "stroke colour"
	}
	return "transform"
}

// frameError returns the first declaration error in a parsed frame list. The
// binder reports it even when no operation would be targeted, so a bad
// declaration is never dropped because the element paints the wrong kind.
func frameError(frames []animation.Keyframe) error {
	for i := range frames {
		if frames[i].Err != nil {
			return fmt.Errorf("frame %d: %w", i, frames[i].Err)
		}
	}
	return nil
}

// resolveColor parses one colour declaration and resolves currentColor against
// the element's own ink colour.
func resolveColor(value string, current Color, hasCurrent bool) (Color, error) {
	parsed, err := parseColor(value)
	if err != nil {
		return Color{}, err
	}
	if !parsed.Current {
		return parsed, nil
	}
	if !hasCurrent {
		return Color{}, ErrCurrentColor
	}
	return current, nil
}

// ruleFor returns the keyframe declarations of the last rule with this name,
// which is the rule CSS selects when a name repeats.
func ruleFor(rules []layout.KeyframesRule, name string) []layout.Keyframe {
	for i := len(rules) - 1; i >= 0; i-- {
		if rules[i].Name == name {
			return rules[i].Keyframes
		}
	}
	return nil
}

// effectClasses returns the distinct paint classes the declarations of one
// effect use, in declaration order.
func effectClasses(source []layout.Keyframe) []paintClass {
	var classes []paintClass
	for i := range source {
		for _, declaration := range source[i].Declarations {
			class, ok := propertyClass(declaration.Property)
			if !ok || containsClass(classes, class) {
				continue
			}
			classes = append(classes, class)
		}
	}
	return classes
}

func containsClass(classes []paintClass, class paintClass) bool {
	for _, candidate := range classes {
		if candidate == class {
			return true
		}
	}
	return false
}
