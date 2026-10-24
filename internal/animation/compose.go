package animation

import (
	"fmt"
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// compose samples every property an effect animates and writes the result
// onto current.
func compose(target *Target, current cssmotion.Paint, e Effect, at float64) cssmotion.Paint {
	paint, _ := composeSampled(target, current, e, at)

	return paint
}

// restampSample is one target's sampled values that a 2D operation matrix
// cannot carry: a 3D transform, a perspective, or both. The zero sample needs
// no engine pass.
type restampSample struct {
	transform   TransformValue
	perspective PerspectiveValue
	// hasTransform reports that the effect sampled a transform value, whether
	// or not it fit a 2D matrix.
	hasTransform bool
	// transformRestamp reports that the sampled transform needs the engine's
	// re-projection pass.
	transformRestamp bool
	hasPerspective   bool
}

// composeSampled is compose plus the values the 2D matrix path cannot write.
// The returned sample is set only when the effect sampled a transform or a
// perspective; a paint-only effect leaves it zero.
func composeSampled(
	target *Target,
	current cssmotion.Paint,
	e Effect,
	at float64,
) (cssmotion.Paint, restampSample) {
	base := target.From
	mode := e.Timing.Composition
	fields := []struct {
		get  valueOf
		kind int
	}{
		{moveX, fieldX}, {moveY, fieldY}, {opacity, fieldOpacity}, {red, fieldR}, {green, fieldG}, {blue, fieldB},
	}
	fallbacks := missingFallbacks(base, mode)
	transform := hasTransform(e.Frames)
	for i, item := range fields {
		if transform && (item.kind == fieldX || item.kind == fieldY) {
			continue
		}
		v := e.sample(at, fallbacks[i], item.get)
		if !hasValue(e.Frames, item.get) {
			continue
		}
		apply(&current, base, item.kind, v, mode)
	}

	var sample restampSample
	if transform {
		// Bind rejects a mismatched 3D pair, so ok is false only for a frame
		// list that changed after NewComposite. Leaving the underlying paint
		// in place is safer than snapping the transform.
		if value, ok := e.sampleTransform(at); ok {
			sample.hasTransform = true
			sample.transform = value
			sample.transformRestamp = applyTransform(target, &current, value)
		}
	}
	if hasPerspective(e.Frames) {
		if value, ok := e.samplePerspective(at); ok {
			sample.hasPerspective = true
			sample.perspective = value
		}
	}

	return current, sample
}

// hasTransform reports whether any keyframe carries a parsed transform list or
// a perspective declaration. Both sample per target instead of through the
// precomputed paint field table.
func hasTransform(frames []Keyframe) bool {
	for _, frame := range frames {
		if frame.Transform.declared() || frame.Perspective.declared() {
			return true
		}
	}
	return false
}

// missingFallbacks returns the value a keyframe that omits a property
// contributes. A replacing composition shows the underlying paint, so an
// omitted property falls back to the base value. An additive composition
// sums the keyframe onto the paint, so an omitted property must add
// nothing, or the paint is added to itself and saturates.
func missingFallbacks(base cssmotion.Paint, mode Composition) [6]float64 {
	if mode != Replace {
		return [6]float64{}
	}
	return [6]float64{0, 0, base.Opacity, base.R, base.G, base.B}
}

func apply(p *cssmotion.Paint, base cssmotion.Paint, field int, v float64, mode Composition) {
	mask, ok := fieldMask(field)
	if !ok || !p.Writes(mask) {
		return
	}
	add := mode != Replace
	switch field {
	case fieldX:
		if add {
			p.X += v
		} else {
			p.X = base.X + v
		}
	case fieldY:
		if add {
			p.Y += v
		} else {
			p.Y = base.Y + v
		}
	case fieldOpacity:
		if add {
			p.Opacity += v
		} else {
			p.Opacity = v
		}
	case fieldR:
		p.R = color(p.R, v, add)
	case fieldG:
		p.G = color(p.G, v, add)
	case fieldB:
		p.B = color(p.B, v, add)
	}
}

// applyTransform writes a sampled transform through the operation's matrix.
// Legacy X/Y-only keyframes never reach here; apply keeps them on the scalar
// path.
//
// The composed matrix replaces the element's static transform instead of
// stacking on top of it, because the operation's captured matrix already
// carries it. The invariant is:
//
//	M_target = Parent * M_anim * World^-1 * Base
//
// M_anim is the sampled decomposition with the element's transform-origin
// baked in, and Base is the matrix the operation carried at bind. World^-1
// strips the element's static transform from Base, so a static transform is
// replaced, not doubled, while any transform deeper than the element survives.
//
// A 3D keyframe or a 3D element cannot be expressed as a 2D matrix. That
// sample leaves the operation alone, clears any earlier 2D matrix write, and
// reports that the element needs the engine's 3D pass; RestampValues hands the
// serialized list to the page, which calls layout.RestampTransforms.
//
// The caller sets Paint.NeedsRestamp from the collected samples, so a
// perspective sample in the same step is not hidden by a later 2D write.
//
// Every composition mode replaces the matrix. Concatenating additive transform
// lists belongs to the composition phase; a second effect on the same element
// writes last, matching the scalar path's replacing mode.
func applyTransform(target *Target, p *cssmotion.Paint, value TransformValue) bool {
	if !p.Writes(cssmotion.MaskMatrix) {
		return false
	}
	if target.HasTransform3D || value.ThreeD {
		return true
	}
	anim := layout.BakeOrigin(matrixFromTransform(value.Decomposed), target.OriginX, target.OriginY)
	parent, world := targetMatrices(target)
	p.Matrix = parent.Mul(anim).Mul(inverseMatrix2D(world)).Mul(matrixOrIdentity(target.From.Matrix))
	p.HasMatrix = true

	return false
}

// matrixFromTransform builds the sampled decomposition as the CSS product
// T(Tx,Ty) * R(Rotate) * SkewX * S(Sx,Sy). The two-dimensional decomposition
// folds every shear into SkewX, so SkewY stays reserved.
func matrixFromTransform(t Transform) layout.Matrix2D {
	t = withIdentity(t)
	sin, cos := math.Sincos(t.Rotate)
	m := layout.Matrix2D{A: 1, D: 1, E: t.Tx, F: t.Ty}
	m = m.Mul(layout.Matrix2D{A: cos, B: sin, C: -sin, D: cos})
	m = m.Mul(layout.Matrix2D{A: 1, D: 1, C: math.Tan(t.SkewX)})
	return m.Mul(layout.Matrix2D{A: t.Sx, D: t.Sy})
}

// inverseMatrix2D returns the affine inverse of m. A singular or non-finite
// determinant reports the identity, so a degenerate base cannot zero every
// coordinate of the operation.
func inverseMatrix2D(m layout.Matrix2D) layout.Matrix2D {
	det := m.A*m.D - m.B*m.C
	if det == 0 || math.IsNaN(det) || math.IsInf(det, 0) {
		return layout.IdentityMatrix()
	}
	return layout.Matrix2D{
		A: m.D / det,
		B: -m.B / det,
		C: -m.C / det,
		D: m.A / det,
		E: (m.C*m.F - m.D*m.E) / det,
		F: (m.B*m.E - m.A*m.F) / det,
	}
}

// matrixOrIdentity treats the zero matrix as the identity. Targets built
// without engine metadata carry the zero value; the transform path must not
// collapse the operation because of it.
func matrixOrIdentity(m layout.Matrix2D) layout.Matrix2D {
	if m == (layout.Matrix2D{}) {
		return layout.IdentityMatrix()
	}
	return m
}

// targetMatrices returns the parent and world matrices the composition needs.
// A zero World falls back to Parent, which is the engine's invariant for an
// element with no static transform.
func targetMatrices(t *Target) (layout.Matrix2D, layout.Matrix2D) {
	parent := matrixOrIdentity(t.Parent)
	if t.World == (layout.Matrix2D{}) {
		return parent, parent
	}
	return parent, t.World
}

// RestampPending reports whether the latest sample skipped a 2D matrix write
// because the transform needs layout.RestampTransforms. The runtime counts one
// per player and collects one request per element from RestampValues.
func (p *Player) RestampPending() bool {
	if p == nil {
		return false
	}
	for i := range p.current {
		if p.current[i].NeedsRestamp {
			return true
		}
	}
	return false
}

// WantsRestamp reports, before any sample, whether the player can need the
// engine's re-projection pass: a keyframe that carries a 3D transform function
// or a perspective declaration, or a target element that already declares a
// 3D transform. The page turns retained layout state on from this answer,
// because the pass cannot run without it. RestampPending reports whether the
// latest sample actually needed the pass.
func (p *Player) WantsRestamp() bool {
	if p == nil {
		return false
	}
	for i := range p.targets {
		if p.targets[i].HasTransform3D {
			return true
		}
	}
	for i := range p.effects {
		for j := range p.effects[i].Frames {
			if p.effects[i].Frames[j].Transform.ThreeD || p.effects[i].Frames[j].Perspective.declared() {
				return true
			}
		}
	}

	return false
}

// color applies one sRGB channel. Legacy CSS colours interpolate in sRGB, so
// the channel stays in that space and is clamped to [0,1]; oklab and color()
// interpolation is out of scope until the engine exposes it.
func color(current, value float64, add bool) float64 {
	if add {
		current += value
	} else {
		current = value
	}
	return math.Max(0, math.Min(1, current))
}

// fieldMask maps a paint field index to the mask bit that owns it.
func fieldMask(field int) (cssmotion.Mask, bool) {
	switch field {
	case fieldX:
		return cssmotion.MaskX, true
	case fieldY:
		return cssmotion.MaskY, true
	case fieldOpacity:
		return cssmotion.MaskOpacity, true
	case fieldR, fieldG, fieldB:
		return cssmotion.MaskRGB, true
	}
	return 0, false
}

func hasValue(frames []Keyframe, field valueOf) bool {
	for _, frame := range frames {
		if _, ok := field(frame); ok {
			return true
		}
	}
	return false
}

type transformStop struct {
	at    float64
	value TransformValue
}

// rotate3DArgCount is the axis plus angle count of rotate3d().
const rotate3DArgCount = 4

// transformStops collects the keyframes that declare a transform, in sorted
// order.
func transformStops(frames []Keyframe) []transformStop {
	stops := make([]transformStop, 0, len(frames)+extraStops)
	for _, frame := range frames {
		if value, ok := frame.transformValue(); ok {
			stops = append(stops, transformStop{frame.Offset, value})
		}
	}
	return stops
}

// padStops brackets the stops with the underlying none at 0% and 100%. A
// keyframe that omits transform holds the underlying paint, which is the
// identity matrix.
func padStops(stops []transformStop) []transformStop {
	fallback := TransformValue{Decomposed: IdentityTransform()}
	if stops[0].at > 0 {
		stops = append([]transformStop{{0, fallback}}, stops...)
	}
	if stops[len(stops)-1].at < 1 {
		stops = append(stops, transformStop{1, fallback})
	}
	return stops
}

// sampleTransform samples the decomposed transform and, for a matching
// function list, the parsed functions.
func (e Effect) sampleTransform(at float64) (TransformValue, bool) {
	stops := transformStops(e.Frames)
	if len(stops) == 0 {
		return TransformValue{}, false
	}
	stops = padStops(stops)
	if at <= stops[0].at {
		return stops[0].value, true
	}
	for i := 1; i < len(stops); i++ {
		if at > stops[i].at {
			continue
		}
		prev, next := stops[i-1], stops[i]
		span := next.at - prev.at
		if span <= 0 {
			return next.value, true
		}
		part := eased(e.Timing, (at-prev.at)/span)
		value, err := interpolateTransform(prev.value, next.value, part)
		if err != nil {
			return TransformValue{}, false
		}
		return value, true
	}
	return stops[len(stops)-1].value, true
}

func eased(t Timing, part float64) float64 {
	if t.Easing != nil {
		return t.Easing(part)
	}
	return cssmotion.Ease(t.TimingFunction, part)
}

// perspectiveStop is one keyframe's perspective declaration at a normalized
// offset.
type perspectiveStop struct {
	at    float64
	value PerspectiveValue
}

// discreteMidpoint is the CSS threshold inside a keyframe interval that is
// discrete because one endpoint is none.
const discreteMidpoint = 0.5

// perspectiveStops collects the keyframes that declare perspective or
// perspective-origin, in sorted order, together with the implicit endpoint the
// padding uses. keyframes() resolves the element's 50% 50% origin on every
// frame, so the implicit endpoint holds the same default origin.
func perspectiveStops(frames []Keyframe) ([]perspectiveStop, PerspectiveValue) {
	fallback := PerspectiveValue{HasDistance: true}
	if len(frames) > 0 {
		fallback.OriginX = frames[0].Perspective.OriginX
		fallback.OriginY = frames[0].Perspective.OriginY
	}
	stops := make([]perspectiveStop, 0, len(frames)+extraStops)
	for _, frame := range frames {
		if frame.Perspective.declared() {
			stops = append(stops, perspectiveStop{frame.Offset, frame.Perspective})
		}
	}

	return stops, fallback
}

// padPerspectiveStops brackets the stops with the underlying none at 0% and
// 100%. A keyframe that omits perspective holds the element's underlying
// value; the host cannot read the element's static perspective here, so the
// implicit endpoint is none.
func padPerspectiveStops(stops []perspectiveStop, fallback PerspectiveValue) []perspectiveStop {
	if stops[0].at > 0 {
		stops = append([]perspectiveStop{{0, fallback}}, stops...)
	}
	if stops[len(stops)-1].at < 1 {
		stops = append(stops, perspectiveStop{1, fallback})
	}

	return stops
}

// samplePerspective samples the camera distance and vanishing point at at. The
// distance is a length that interpolates linearly; a none endpoint switches at
// the midpoint. The origin always interpolates component-wise.
func (e Effect) samplePerspective(at float64) (PerspectiveValue, bool) {
	stops, fallback := perspectiveStops(e.Frames)
	if len(stops) == 0 {
		return PerspectiveValue{}, false
	}
	stops = padPerspectiveStops(stops, fallback)
	if at <= stops[0].at {
		return stops[0].value, true
	}
	for i := 1; i < len(stops); i++ {
		if at > stops[i].at {
			continue
		}
		prev, next := stops[i-1], stops[i]
		span := next.at - prev.at
		if span <= 0 {
			return next.value, true
		}
		part := eased(e.Timing, (at-prev.at)/span)

		return interpolatePerspective(prev.value, next.value, part), true
	}

	return stops[len(stops)-1].value, true
}

// hasPerspective reports whether any keyframe declares perspective or
// perspective-origin.
func hasPerspective(frames []Keyframe) bool {
	for _, frame := range frames {
		if frame.Perspective.declared() {
			return true
		}
	}

	return false
}

// interpolatePerspective blends two perspective values. The distance
// interpolates linearly when both endpoints carry a length; a none endpoint is
// discrete, so the sample takes one endpoint's whole distance at the midpoint
// and reports Discrete. The origin interpolates component-wise, using each
// endpoint's resolved default when it did not declare one.
func interpolatePerspective(a, b PerspectiveValue, t float64) PerspectiveValue {
	var out PerspectiveValue
	switch {
	case !a.HasDistance && !b.HasDistance:
	case a.HasDistance && b.HasDistance && a.Distance > 0 && b.Distance > 0:
		out.Distance = lerp(a.Distance, b.Distance, t)
		out.HasDistance = true
	default:
		out.Discrete = true
		out.Distance = b.Distance
		if t < discreteMidpoint {
			out.Distance = a.Distance
		}
		out.HasDistance = true
	}
	if a.HasOrigin || b.HasOrigin {
		out.HasOrigin = true
	}
	out.OriginX = lerp(a.OriginX, b.OriginX, t)
	out.OriginY = lerp(a.OriginY, b.OriginY, t)

	return out
}

// interpolateTransform blends two transform values. Matching function lists
// interpolate argument by argument. Mismatched 2D lists interpolate the
// decomposed translation, scale, rotation, and skew, which is the WebKit
// two-dimensional decomposition. A mismatched 3D list returns ErrTransform3D
// rather than snapping to one endpoint.
func interpolateTransform(a, b TransformValue, t float64) (TransformValue, error) {
	if a.ThreeD || b.ThreeD {
		return interpolate3D(a, b, t)
	}
	value := TransformValue{Decomposed: blendTransform(a.Decomposed, b.Decomposed, t)}
	if interpolable(a.Functions, b.Functions) {
		value.Functions = blendFunctions(a.Functions, b.Functions, t)
	}
	return value, nil
}

func interpolate3D(a, b TransformValue, t float64) (TransformValue, error) {
	from, to := a.Functions, b.Functions
	if len(from) == 0 && len(to) > 0 {
		var ok bool
		from, ok = identityFunctions(to)
		if !ok {
			return TransformValue{}, Decompose3DError(from, to)
		}
	}
	if len(to) == 0 && len(from) > 0 {
		var ok bool
		to, ok = identityFunctions(from)
		if !ok {
			return TransformValue{}, Decompose3DError(from, to)
		}
	}
	if !interpolable(from, to) {
		return TransformValue{}, Decompose3DError(from, to)
	}
	return TransformValue{
		Decomposed: blendTransform(a.Decomposed, b.Decomposed, t),
		Functions:  blendFunctions(from, to, t),
		ThreeD:     true,
	}, nil
}

// Decompose3D is the entry point for blending two mismatched 3D function
// lists through the CSS Transforms Level 2 four-by-four decomposition. That
// algorithm is not implemented yet: it reports ErrTransform3D instead of
// snapping the list at one endpoint. NewComposite turns the error into one
// bind failure per effect.
func Decompose3D(from, to []TransformFunction, _ float64) ([]TransformFunction, error) {
	return nil, Decompose3DError(from, to)
}

// Decompose3DError builds the shared mismatch error.
func Decompose3DError(from, to []TransformFunction) error {
	return fmt.Errorf("%w: %s to %s", ErrTransform3D, listNames(from), listNames(to))
}

// interpolable reports whether two lists can blend function by function.
// matrix, matrix3d, and perspective are not interpolable that way; CSS
// decomposes them first.
func interpolable(a, b []TransformFunction) bool {
	if len(a) == 0 || len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Name != b[i].Name || len(a[i].Args) != len(b[i].Args) {
			return false
		}
		switch a[i].Name {
		case Matrix, Matrix3D, Perspective:
			return false
		}
	}
	return true
}

func blendTransform(a, b Transform, t float64) Transform {
	a, b = withIdentity(a), withIdentity(b)
	return Transform{
		Tx:     lerp(a.Tx, b.Tx, t),
		Ty:     lerp(a.Ty, b.Ty, t),
		Sx:     lerp(a.Sx, b.Sx, t),
		Sy:     lerp(a.Sy, b.Sy, t),
		Rotate: lerp(a.Rotate, b.Rotate, t),
		SkewX:  lerp(a.SkewX, b.SkewX, t),
		SkewY:  lerp(a.SkewY, b.SkewY, t),
		Set:    true,
	}
}

func withIdentity(t Transform) Transform {
	if t.Set {
		return t
	}
	return IdentityTransform()
}

func blendFunctions(a, b []TransformFunction, t float64) []TransformFunction {
	out := make([]TransformFunction, len(a))
	for i := range a {
		out[i].Name = a[i].Name
		out[i].Args = make([]float64, len(a[i].Args))
		for j := range a[i].Args {
			out[i].Args[j] = lerp(a[i].Args[j], b[i].Args[j], t)
		}
	}
	return out
}

func lerp(a, b, t float64) float64 { return a + (b-a)*t }

// identityFunctions builds a neutral list with the same function signatures.
// It turns a missing keyframe (the underlying none) into a value that can
// blend with a 3D list. A perspective function has no finite identity, so it
// reports false and the caller returns ErrTransform3D.
func identityFunctions(functions []TransformFunction) ([]TransformFunction, bool) {
	out := make([]TransformFunction, len(functions))
	for i, f := range functions {
		out[i].Name = f.Name
		out[i].Args = make([]float64, len(f.Args))
		switch f.Name {
		case Translate, TranslateX, TranslateY, Translate3D, TranslateZ:
			out[i].Args = make([]float64, len(f.Args))
		case Scale, ScaleX, ScaleY, Scale3D, ScaleZ:
			for j := range out[i].Args {
				out[i].Args[j] = 1
			}
		case Rotate, RotateX, RotateY, RotateZ, Skew, SkewX, SkewY:
			out[i].Args = make([]float64, len(f.Args))
		case Rotate3D:
			if len(f.Args) == rotate3DArgCount {
				out[i].Args[2] = 1
			}
		case Matrix:
			copy(out[i].Args, []float64{1, 0, 0, 1, 0, 0})
		case Matrix3D:
			copy(out[i].Args, []float64{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1})
		default:
			return nil, false
		}
	}
	return out, true
}

func listNames(functions []TransformFunction) string {
	if len(functions) == 0 {
		return "none"
	}
	names := ""
	for i, f := range functions {
		if i > 0 {
			names += " "
		}
		names += string(f.Name)
	}
	return names
}
