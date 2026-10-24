// Package animation samples CSS-style keyframes onto retained display ops.
package animation

import (
	"errors"

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

	// HasElementTransform reports whether the stylesheet declares a static
	// transform on the element. The matrix path replaces that transform rather
	// than stacking onto it.
	HasElementTransform bool
	// HasTransform3D reports that the element participates in a 3D
	// composition. A 2D matrix cannot describe it, so the sample is left for
	// the engine's restamp pass.
	HasTransform3D bool
}

// ErrTransform3D reports two mismatched transform lists that need the CSS
// Transforms Level 2 four-by-four decomposition. That algorithm is not
// implemented yet; the bind fails with this error instead of snapping the
// transform at the midpoint.
var ErrTransform3D = errors.New("animation: mismatched 3D transform lists need 4x4 decomposition")

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
// the resolved point (the element's 50% 50% default when the keyframe does not
// declare an origin), and HasOrigin reports whether the keyframe declared one.
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

// declared reports whether the keyframe mentions perspective or
// perspective-origin.
func (v PerspectiveValue) declared() bool {
	return v.HasDistance || v.HasOrigin
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
	Err           error
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
