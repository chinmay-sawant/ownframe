// Package animation samples CSS-style keyframes onto retained display ops.
package animation

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// Easing maps progress within one keyframe interval to eased progress.
type Easing func(float64) float64

// Target pairs a display operation with its captured starting paint and the
// element transform metadata the matrix path composes with.
//
// Parent maps the element's parent coordinate space to device space, World is
// Parent composed with the element's own static transform, and OriginX and
// OriginY are the resolved transform-origin. All three come from
// layout.TransformInfo and are in layout points. A caller with no engine
// metadata may leave them zero: a zero Parent behaves as the identity and a
// zero World falls back to Parent, which is the value an element with no
// static transform carries.
type Target struct {
	Op *layout.DisplayOp
	// From is the captured starting paint. Its Mask is the set of fields this
	// target's keyframes may write: a colour target leaves the matrix alone, a
	// transform target leaves the colour alone, and neither writes W, H, or
	// Alpha. A zero mask keeps the full-snapshot meaning of cssmotion.Paint.
	From cssmotion.Paint

	Parent layout.Matrix2D
	World  layout.Matrix2D

	OriginX, OriginY float64

	// svg is the latest sampled inline-SVG paint for this target. It never
	// reaches a display operation: the page turns it into a
	// layout.SVGPaintOverride and asks the engine to restyle the SVG raster.
	svg SVGPaintValue

	// HasElementTransform reports whether the stylesheet declares a static
	// transform on the element. The matrix path replaces that transform rather
	// than stacking onto it.
	HasElementTransform bool
	// HasTransform3D reports that the element participates in a 3D
	// composition. A 2D matrix cannot describe it, so the sample is left for
	// the engine's restamp pass.
	HasTransform3D bool
}

// ErrSingularMatrix reports a four-by-four transform matrix that cannot be
// decomposed, so the pair it belongs to has no matrix interpolation. CSS falls
// back to a discrete step for that interval; a mismatched 3D pair with two
// invertible endpoints no longer fails to bind.
var ErrSingularMatrix = errors.New("animation: singular 3D transform matrix")

// TransformFunc names one CSS transform function.
type TransformFunc string

// The transform functions the parser accepts. rotateZ and translateZ stay
// distinct from rotate and translate3d so a matching-list check compares the
// authored function, not just its geometry.
const (
	Translate   TransformFunc = "translate"
	TranslateX  TransformFunc = "translateX"
	TranslateY  TransformFunc = "translateY"
	Translate3D TransformFunc = "translate3d"
	TranslateZ  TransformFunc = "translateZ"
	Scale       TransformFunc = "scale"
	ScaleX      TransformFunc = "scaleX"
	ScaleY      TransformFunc = "scaleY"
	Scale3D     TransformFunc = "scale3d"
	ScaleZ      TransformFunc = "scaleZ"
	Rotate      TransformFunc = "rotate"
	RotateX     TransformFunc = "rotateX"
	RotateY     TransformFunc = "rotateY"
	RotateZ     TransformFunc = "rotateZ"
	Rotate3D    TransformFunc = "rotate3d"
	Skew        TransformFunc = "skew"
	SkewX       TransformFunc = "skewX"
	SkewY       TransformFunc = "skewY"
	Matrix      TransformFunc = "matrix"
	Matrix3D    TransformFunc = "matrix3d"
	Perspective TransformFunc = "perspective"
)

// TransformFunction is one parsed transform function with its arguments
// already resolved to display units:
//
//   - translate*, matrix e/f, and perspective lengths are points;
//   - rotate*, skew*, and the rotate3d angle are radians;
//   - scale*, matrix, matrix3d, and the rotate3d axis are unitless.
//
// Percentages are resolved against the element's border box before the
// function reaches this struct.
type TransformFunction struct {
	Name TransformFunc
	Args []float64
}

// Transform is the decomposed two-dimensional form of a transform list.
// Translate is in display points, Rotate and SkewX are radians, and scale is
// a multiplier. Scale is 1 for an identity transform; a zero-valued Transform
// is not identity unless Set is false.
type Transform struct {
	Tx, Ty float64
	Sx, Sy float64
	Rotate float64
	SkewX  float64
	SkewY  float64 // reserved; the 2D decomposition folds shear into SkewX
	Set    bool
}

// IdentityTransform returns the neutral decomposed two-dimensional transform.
func IdentityTransform() Transform {
	return Transform{Sx: 1, Sy: 1, Set: true}
}

// TransformValue is one keyframe's transform declaration. Decomposed is the
// fast two-dimensional path and is always meaningful for a 2D list. Functions
// keeps the parsed list so matching function lists interpolate component-wise
// and so a 3D list survives for the restamp path; ThreeD reports that at least
// one function needs the engine's 3D pass.
type TransformValue struct {
	Decomposed Transform
	Functions  []TransformFunction
	ThreeD     bool
}

// declared reports whether the value came from a transform declaration.
// X and Y alone keep the legacy keyframe shape working.
func (v TransformValue) declared() bool {
	return v.ThreeD || v.Decomposed.Set || len(v.Functions) > 0
}

// PerspectiveValue is one keyframe's perspective and perspective-origin
// declaration. Distance is the camera distance in layout points; zero means
// none. OriginX and OriginY are the resolved perspective-origin in layout
// points, as an offset from the element's border-box origin; they always carry
// the resolved point (the element's static vanishing point, or its 50% 50%
// default when the element has no perspective), and HasOrigin reports whether
// the keyframe declared one.
//
// HasDistance distinguishes a declared perspective:none from a property the
// keyframe omits. Discrete reports that the sample picked one endpoint because
// one side was none: perspective is discrete across none, while
// perspective-origin still interpolates component-wise.
type PerspectiveValue struct {
	Distance    float64
	OriginX     float64
	OriginY     float64
	HasDistance bool
	HasOrigin   bool
	Discrete    bool
}

// PerspectiveBase is the element's static perspective and perspective-origin,
// the used value a keyframe that omits either property holds. A keyframe list
// with no perspective or perspective-origin stop at 0% or 100% pads that
// endpoint from this base instead of snapping to none and 50% 50%.
//
// Distance is in layout points. HasDistance is false when the element's static
// perspective is none, so an implicit endpoint keeps that none rather than
// inventing a camera. OriginX and OriginY are always resolved: a declared
// perspective carries the engine's resolved origin, and an element without one
// uses the initial 50% 50% point against its border box. Set reports that the
// binder supplied a base; a keyframe built by hand keeps the legacy endpoint.
type PerspectiveBase struct {
	Distance    float64
	OriginX     float64
	OriginY     float64
	HasDistance bool
	Set         bool
}

// declared reports whether the keyframe mentions perspective or
// perspective-origin.
func (v PerspectiveValue) declared() bool {
	return v.HasDistance || v.HasOrigin
}

// SVGPaintColor is one inline-SVG paint colour in sRGB. None is the CSS none
// keyword: it has no channels, so a pair with none on either side is discrete
// rather than interpolated. A is the colour's own alpha; fill-opacity and
// stroke-opacity are sampled as separate properties.
type SVGPaintColor struct {
	R, G, B, A float64
	None       bool
}

// SVGPaintValue is one element's sampled inline-SVG paint. The Has flags name
// the properties the keyframes declared; a property whose flag is false keeps
// the element's used value. Colours are sRGB in 0..1, opacities are 0..1, and
// lengths are layout points, ready for layout.SVGPaintOverride.
//
// This value is not a display field. A paint change on an inline SVG rebuilds
// its raster, so the page hands the value to layout.RestyleInlineSVG instead of
// writing an operation.
//
// Target indexes the targets the player was built with. Deferred carries the
// reason a declaration cannot be restyled, such as a url(#gradient) paint
// server; the property then keeps its base paint.
type SVGPaintValue struct {
	Target int

	Fill        SVGPaintColor
	FillOpacity float64

	Stroke           SVGPaintColor
	StrokeOpacity    float64
	StrokeWidth      float64
	StrokeDashArray  []float64
	StrokeDashOffset float64

	HasFill             bool
	HasFillOpacity      bool
	HasStroke           bool
	HasStrokeOpacity    bool
	HasStrokeWidth      bool
	HasStrokeDashArray  bool
	HasStrokeDashOffset bool

	Deferred string
}

// declared reports whether the value carries any SVG paint property or a
// deferral. A zero value means the element declared no SVG paint.
func (v SVGPaintValue) declared() bool {
	return v.HasFill || v.HasFillOpacity || v.HasStroke || v.HasStrokeOpacity ||
		v.HasStrokeWidth || v.HasStrokeDashArray || v.HasStrokeDashOffset ||
		v.Deferred != ""
}

// Keyframe stores values declared at Offset. X and Y stay as readable aliases
// of Transform.Decomposed.Tx and .Ty for callers that have not moved to the
// decomposed form; a parsed transform sets both. Err holds a declaration that
// could not be parsed, so NewComposite can report it once at bind.
type Keyframe struct {
	Offset                     float64
	X, Y                       float64
	Opacity                    float64
	R, G, B                    float64
	Move, OpacitySet, ColorSet bool
	// ColorAlpha is the alpha a parsed colour declaration carried. The paint
	// phase writes only the sRGB channels, so ColorAlphaSet reports a
	// translucent keyframed colour the binder counts as a fidelity gap instead
	// of dropping it silently.
	ColorAlpha    float64
	ColorAlphaSet bool
	Transform     TransformValue
	Perspective   PerspectiveValue
	// PerspectiveBase is the element's static camera. The binder fills it so
	// an omitted 0% or 100% perspective or perspective-origin endpoint uses
	// the static value instead of none and the box centre.
	PerspectiveBase PerspectiveBase
	// SVG holds the inline-SVG paint declarations of this keyframe. The paint
	// classes ignore it: the value is collected per target and handed to the
	// engine's RestyleInlineSVG, never written onto an operation.
	SVG SVGPaintValue
	// TimingFunction is an animation-timing-function declared inside the
	// keyframe block. It eases the interval that starts at this keyframe; an
	// empty value uses the effect's own timing function.
	TimingFunction string
	// Layout holds the layout-tier declarations of this keyframe, one value
	// per property. The paint classes ignore it; the layout driver samples it
	// into the engine's override relayout request.
	Layout []LayoutValue
	// ease caches TimingFunction, or the effect's function when the keyframe
	// declares none. The binder fills it once so a sample does not parse a
	// CSS timing function.
	ease easingFn
	Err  error
}

// transformValue returns the keyframe's transform, including the legacy X/Y
// form. A parsed declaration wins over X and Y.
func (k Keyframe) transformValue() (TransformValue, bool) {
	if k.Transform.declared() {
		return k.Transform, true
	}
	if k.Move {
		return TransformValue{Decomposed: Transform{Tx: k.X, Ty: k.Y, Sx: 1, Sy: 1, Set: true}}, true
	}
	return TransformValue{}, false
}

// LayoutTerm is one component of a layout-tier value: a number with its unit,
// or a keyword such as auto or normal. The slash that separates the horizontal
// and vertical halves of border-radius is kept as a keyword term, so a pair of
// matching radius lists interpolates component-wise.
//
// Unit is empty for a unitless number (a flex factor or a line-height
// multiplier). An absolute length is already resolved to px by the keyframe
// parser; a percentage and a font- or viewport-relative unit keep the authored
// unit, because the engine resolves them against the element's containing
// block.
type LayoutTerm struct {
	Number  float64
	Unit    string
	Keyword string
}

// LayoutValue is one layout-tier property value sampled from a keyframe list.
// The property name is the CSS longhand or shorthand the override names, and
// the terms are the sampled value in the form the engine accepts.
type LayoutValue struct {
	Property string
	Terms    []LayoutTerm
}

// CSS renders the value in the form the engine's override relayout accepts.
func (v LayoutValue) CSS() string {
	var b strings.Builder
	for i := range v.Terms {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(v.Terms[i].CSS())
	}
	return b.String()
}

// CSS renders one term. A keyword keeps its spelling; a number keeps its unit
// and prints without trailing zeros, so a live sample costs no padding.
func (t LayoutTerm) CSS() string {
	if t.Keyword != "" {
		return t.Keyword
	}
	return strconv.FormatFloat(t.Number, 'f', -1, 64) + t.Unit
}

// NewLayout builds a player that samples layout-tier keyframes only. The
// effects carry no paint fields, so the single scratch operation the player
// addresses is never written for drawing; it exists because a player samples
// its effects against an operation. The caller starts the player with the
// element's clock, the same clock a paint player on the element uses.
func NewLayout(effects []Effect) (*Player, error) {
	if len(effects) == 0 {
		return nil, fmt.Errorf("animation: a layout player needs at least one effect")
	}

	var scratch layout.DisplayOp
	return NewComposite([]Target{{Op: &scratch}}, effects)
}
