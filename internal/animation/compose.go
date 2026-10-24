package animation

import (
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
	// hasTransform also reports inline-SVG paint, which shares this per-target
	// path; a paint-only effect simply has no transform stops to sample.
	perTarget := hasTransform(e.Frames)
	for i, item := range fields {
		if perTarget && (item.kind == fieldX || item.kind == fieldY) {
			continue
		}
		v := e.sample(at, fallbacks[i], item.get)
		if !hasValue(e.Frames, item.get) {
			continue
		}
		apply(&current, base, item.kind, v, mode)
	}

	var sample restampSample
	if perTarget {
		// A singular endpoint makes the interval discrete, and a frame list
		// that changed after NewComposite can still fail here. Both leave the
		// underlying paint in place rather than snapping the transform.
		if value, ok := e.sampleTransform(at); ok {
			composed, restamp := composeTransform(target, &current, value, mode)
			sample.hasTransform = true
			sample.transform = composed
			sample.transformRestamp = restamp
		}
	}
	if hasPerspective(e.Frames) {
		if value, ok := e.samplePerspective(at); ok {
			sample.hasPerspective = true
			sample.perspective = value
		}
	}
	if hasSVGPaint(e.Frames) {
		// Inline-SVG paint is not a display field, so the only per-target work
		// is collecting the sample the page hands to RestyleInlineSVG.
		target.svg = e.sampleSVGPaint(at)
	}

	return current, sample
}

// hasTransform reports whether any keyframe carries a parsed transform list, a
// perspective declaration, or inline-SVG paint. All three sample per target
// instead of through the precomputed paint field table: a transform and a
// perspective compose against the target's captured matrix, and SVG paint is
// collected for the target's element.
func hasTransform(frames []Keyframe) bool {
	for _, frame := range frames {
		if frame.Transform.declared() || frame.Perspective.declared() || frame.SVG.declared() {
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

// composeTransform folds one effect's sampled transform onto the transform the
// target already carries and reports the value the restamp path should see,
// plus whether the element needs the engine's re-projection pass. Legacy X/Y
// keyframes never reach here; apply keeps them on the scalar path.
//
// The composition follows animation-composition in declaration order:
//
//   - replace: the sample replaces the transform the target carries;
//   - add: the two transform lists concatenate, which is the ordered matrix
//     product of the underlying transform and the sample;
//   - accumulate: corresponding components add, with the one-based scale
//     components combining as a+b-1, which is not the same transform as add.
//
// The underlying transform is recovered from the paint instead of kept in
// per-target state. A paint always holds parent * anim * world^-1 * base, so
// inverting that expression yields the animation transform earlier effects
// installed, and a target with no earlier transform yields the element's own
// static transform. That makes an effect stack fold without a reset marker,
// and a later effect cannot read a previous frame's value.
//
// The composed matrix replaces the element's static transform instead of
// stacking on top of it, because the operation's captured matrix already
// carries it. The invariant is:
//
//	M_target = Parent * M_anim * World^-1 * Base
//
// M_anim is the two-dimensional composition with the element's
// transform-origin baked in, and Base is the matrix the operation carried at
// bind. World^-1 strips the element's static transform from Base, so a static
// transform is replaced, not doubled, while any transform deeper than the
// element survives.
//
// A 3D keyframe or a 3D element cannot be expressed as a 2D matrix. That
// sample leaves the operation alone and reports that the element needs the
// engine's 3D pass; RestampValues hands the serialized list to the page, which
// calls layout.RestampTransforms. Composition for that path happens in 4x4
// matrices, because the recovered underlying transform is a matrix, not the
// authored function list.
//
// The caller sets Paint.NeedsRestamp from the collected samples, so a
// perspective sample in the same step is not hidden by a later 2D write.
func composeTransform(
	target *Target,
	p *cssmotion.Paint,
	value TransformValue,
	mode Composition,
) (TransformValue, bool) {
	if !p.Writes(cssmotion.MaskMatrix) {
		return value, false
	}
	restamp := target.HasTransform3D || value.ThreeD
	if mode == Replace {
		// Replace keeps the authored list so a single effect serializes
		// exactly as it was declared.
		if !restamp {
			writeTransform(target, p, value.Decomposed)
		}

		return value, restamp
	}
	underlying := underlyingTransform(target, p)
	// CSS converts a missing underlying value to the matching identity list,
	// so an additive or accumulated effect over none is just the sample. Keep
	// the authored list in that case; a matrix round trip would serialize a 3D
	// sample as matrix3d() for no gain. A 3D element is excluded because its
	// static transform is not recoverable from the flattened paint matrix.
	if !target.HasTransform3D && matrixFromTransform(underlying.Decomposed).IsIdentity() {
		if !restamp {
			writeTransform(target, p, deltaTransform(value))
		}

		return value, restamp
	}
	if restamp {
		return composeTransform3D(underlying, value, mode), true
	}
	composed := composeTransform2D(underlying.Decomposed, value, mode)
	writeTransform(target, p, composed)

	return TransformValue{Decomposed: composed}, false
}

// writeTransform installs one two-dimensional animation transform through the
// operation matrix, baking the element's transform-origin.
func writeTransform(target *Target, p *cssmotion.Paint, value Transform) {
	anim := layout.BakeOrigin(matrixFromTransform(value), target.OriginX, target.OriginY)
	parent, world := targetMatrices(target)
	p.Matrix = parent.Mul(anim).Mul(inverseMatrix2D(world)).Mul(matrixOrIdentity(target.From.Matrix))
	p.HasMatrix = true
}

// underlyingTransform recovers the animation transform the paint carries
// before this effect, in the element's own coordinate space:
//
//	A = T(-origin) * Parent^-1 * M * Base^-1 * World * T(origin)
//
// The paint matrix is Parent * BakeOrigin(A) * World^-1 * Base, so this
// inverts that construction exactly. With no effect applied yet, M is Base and
// A is the element's static transform, which is the underlying value CSS
// composes an additive or accumulated effect onto. The recovered value is a
// matrix, so its function list is empty.
//
// A 3D sample writes no matrix, so a second 3D effect after one cannot see the
// first one's value here: the running 3D transform lives in the player's
// restamp slot, which the sampler does not read. That stack composes from the
// element's own transform instead. See the phase notes; wiring it needs the
// render loop to hand the slot back to the sampler.
func underlyingTransform(target *Target, p *cssmotion.Paint) TransformValue {
	parent, world := targetMatrices(target)
	paint := layout.IdentityMatrix()
	if p.HasMatrix {
		paint = matrixOrIdentity(p.Matrix)
	}
	base := matrixOrIdentity(target.From.Matrix)
	anim := inverseMatrix2D(parent).Mul(paint).Mul(inverseMatrix2D(base)).Mul(world)
	anim = unbakeOrigin(anim, target.OriginX, target.OriginY)

	return TransformValue{Decomposed: decomposeMatrix2D(anim)}
}

// composeTransform2D folds a two-dimensional sample onto the underlying
// decomposition. Add multiplies the matrices in declaration order, which is the
// same transform as concatenating the two lists; accumulate adds the decomposed
// components instead.
func composeTransform2D(underlying Transform, value TransformValue, mode Composition) Transform {
	delta := deltaTransform(value)
	if mode == Accumulate {
		return accumulateTransform(underlying, delta)
	}

	return decomposeMatrix2D(matrixFromTransform(underlying).Mul(matrixFromTransform(delta)))
}

// composeTransform3D folds a sample that needs the restamp pass onto the
// underlying transform. Both sides become four-by-four matrices: the underlying
// animation is a matrix, and CSS converts both lists to matrices when they do
// not match function by function. Add multiplies them in declaration order;
// accumulate adds the decomposed components.
func composeTransform3D(underlying TransformValue, value TransformValue, mode Composition) TransformValue {
	if mode == Accumulate {
		if m, ok := accumulateMatrix4(matrix4FromTransformValue(underlying), matrix4FromTransformValue(value)); ok {
			return transformValueFromMatrix4(m)
		}

		return value
	}

	return transformValueFromMatrix4(matrix4FromTransformValue(underlying).Mul(matrix4FromTransformValue(value)))
}

// deltaTransform returns a sample's two-dimensional decomposition. A keyframe
// built by the parser carries both the decomposition and the list; a value with
// only a list still decomposes through its matrix.
func deltaTransform(value TransformValue) Transform {
	if value.Decomposed.Set {
		return value.Decomposed
	}
	if len(value.Functions) > 0 {
		return decomposeMatrix2D(matrix2DFromMatrix4(Matrix4FromFunctions(value.Functions)))
	}

	return IdentityTransform()
}

// accumulateTransform is the CSS accumulation procedure for two decomposed
// two-dimensional transforms: translation, rotation, and skew add, while every
// one-based scale component combines as a+b-1. A keyframed zero scale therefore
// contributes -1, exactly as the spec's one-based rule says.
func accumulateTransform(a, b Transform) Transform {
	a, b = withIdentity(a), withIdentity(b)

	return Transform{
		Tx:     a.Tx + b.Tx,
		Ty:     a.Ty + b.Ty,
		Sx:     a.Sx + b.Sx - 1,
		Sy:     a.Sy + b.Sy - 1,
		Rotate: a.Rotate + b.Rotate,
		SkewX:  a.SkewX + b.SkewX,
		SkewY:  a.SkewY + b.SkewY,
		Set:    true,
	}
}

// accumulateMatrix4 is the CSS accumulation procedure for two matrices: both
// are decomposed, the translation, skew, and perspective components add, every
// one-based scale component combines as a+b-1, the perspective w component is
// one-based, and the rotations compose through their quaternions. A singular
// endpoint has no decomposition, so the caller keeps the sampled value.
func accumulateMatrix4(a, b Matrix4) (Matrix4, bool) {
	from, err := Decompose3D(a)
	if err != nil {
		return Matrix4{}, false
	}
	to, err := Decompose3D(b)
	if err != nil {
		return Matrix4{}, false
	}
	for i := range from.Translate {
		from.Translate[i] += to.Translate[i]
		from.Scale[i] += to.Scale[i] - 1
		from.Skew[i] += to.Skew[i]
	}
	for i := range from.Perspective {
		from.Perspective[i] += to.Perspective[i]
	}
	from.Perspective[perspectiveWIndex] -= 1
	// Decompose3D reads the matrix columns as basis vectors, so the left
	// operand's rotation is the product's right quaternion.
	from.Quaternion = multiplyQuaternions(to.Quaternion, from.Quaternion)

	return Compose3D(from), true
}

// perspectiveWIndex is the homogeneous w component of Decomposition3D.Perspective.
const perspectiveWIndex = 3

// multiplyQuaternions returns the Hamilton product a*b, which composes their
// rotations. It matches the quaternion convention Decompose3D and Compose3D
// share.
func multiplyQuaternions(a, b [4]float64) [4]float64 {
	ax, ay, az, aw := a[0], a[1], a[2], a[3]
	bx, by, bz, bw := b[0], b[1], b[2], b[3]

	return [4]float64{
		aw*bx + ax*bw + ay*bz - az*by,
		aw*by - ax*bz + ay*bw + az*bx,
		aw*bz + ax*by - ay*bx + az*bw,
		aw*bw - ax*bx - ay*by - az*bz,
	}
}

// matrix4FromTransformValue lifts a transform value into four-by-four space.
// The parsed function list wins because it is the authored geometry; a legacy
// or composed value falls back to its two-dimensional decomposition.
func matrix4FromTransformValue(value TransformValue) Matrix4 {
	if len(value.Functions) > 0 {
		return Matrix4FromFunctions(value.Functions)
	}

	return matrix4FromMatrix2D(matrixFromTransform(deltaTransform(value)))
}

// matrix4FromMatrix2D lifts a two-dimensional affine matrix into the
// four-by-four form with z untouched.
func matrix4FromMatrix2D(m layout.Matrix2D) Matrix4 {
	return Matrix4{
		m.A, m.B, 0, 0,
		m.C, m.D, 0, 0,
		0, 0, 1, 0,
		m.E, m.F, 0, 1,
	}
}

// matrix2DFromMatrix4 drops the z row and column. A four-by-four matrix that
// came from two-dimensional inputs is exactly representable this way.
func matrix2DFromMatrix4(m Matrix4) layout.Matrix2D {
	return layout.Matrix2D{
		A: m[0], B: m[1],
		C: m[4], D: m[5],
		E: m[indexM41], F: m[indexM42],
	}
}

// transformValueFromMatrix4 is the value the restamp path serializes as one
// matrix3d().
func transformValueFromMatrix4(m Matrix4) TransformValue {
	args := make([]float64, len(m))
	copy(args, m[:])

	return TransformValue{
		Decomposed: Transform{Tx: m[indexM41], Ty: m[indexM42], Sx: 1, Sy: 1, Set: true},
		Functions:  []TransformFunction{{Name: Matrix3D, Args: args}},
		ThreeD:     true,
	}
}

// unbakeOrigin converts a transform baked about the element's origin back to
// the element's own coordinate space.
func unbakeOrigin(m layout.Matrix2D, ox, oy float64) layout.Matrix2D {
	return translateMatrix(-ox, -oy).Mul(m).Mul(translateMatrix(ox, oy))
}

// translateMatrix builds a pure translation matrix. The public layout package
// re-exports the matrix type but not its constructors.
func translateMatrix(x, y float64) layout.Matrix2D {
	return layout.Matrix2D{A: 1, D: 1, E: x, F: y}
}

// decomposeMatrix2D is the WebKit two-dimensional unmatrix: translation,
// scale, rotation, and one XY shear. ScaleX is signed positive and ScaleY
// carries the determinant sign, so matrixFromTransform reproduces any
// invertible matrix. The decomposition mirrors the one
// internal/cssanimation performs on a parsed list; it lives here because
// internal/cssanimation imports this package for the keyframe model.
func decomposeMatrix2D(m layout.Matrix2D) Transform {
	row0x, row0y := m.A, m.B
	scaleX := math.Hypot(row0x, row0y)
	if scaleX != 0 {
		row0x, row0y = row0x/scaleX, row0y/scaleX
	} else {
		row0x, row0y = 1, 0
	}
	// Project the second column onto the rotated axes. The perpendicular
	// component is the signed Y scale; the parallel component is the shear.
	scaleY := -row0y*m.C + row0x*m.D
	skew := 0.0
	if scaleY != 0 {
		skew = (row0x*m.C + row0y*m.D) / scaleY
	}

	return Transform{
		Tx: m.E, Ty: m.F, Sx: scaleX, Sy: scaleY,
		Rotate: math.Atan2(row0y, row0x), SkewX: math.Atan(skew), Set: true,
	}
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

// LayoutAt samples the player's layout-tier effects at the instant input
// names and appends one value per property to buf. It reports whether any
// effect is still live, so the driver knows whether the element needs another
// sample on a later frame.
//
// The player clock is not advanced: the sample is a read of the state the
// element's Start and SetClock calls installed, which is the same clock the
// paint players on the element use. A paused effect samples at its freeze time,
// so a paused layout animation holds the pose it stopped at. The returned slice
// may alias buf; the caller consumes it before the next sample.
func (p *Player) LayoutAt(input FrameInput, buf []LayoutValue) ([]LayoutValue, bool) {
	if p == nil {
		return buf, false
	}

	p.ensureBound()
	live := false
	for i := range p.effects {
		state := p.states[i]
		if !state.enabled || !state.started {
			continue
		}
		now := input.Now
		if state.paused && !state.pausedAt.IsZero() {
			now = state.pausedAt
		}
		progress, _, effectLive, apply, available := statePosition(&state, withNow(input, now))
		if !available {
			continue
		}
		live = live || effectLive
		if !apply {
			continue
		}
		buf = mergeLayout(buf, p.effects[i].sampleLayout(progress))
	}

	return buf, live
}

// mergeLayout writes one effect's sampled values over the running set. A later
// effect in the element's animation list replaces an earlier value for the same
// property, which is the cascade's declaration order.
func mergeLayout(buf []LayoutValue, values []LayoutValue) []LayoutValue {
	for _, value := range values {
		replaced := false
		for i := range buf {
			if buf[i].Property == value.Property {
				buf[i], replaced = value, true

				break
			}
		}
		if !replaced {
			buf = append(buf, value)
		}
	}

	return buf
}

// layoutStop is one keyframe's value of one layout property at a normalized
// offset. ease is the timing function declared on this keyframe, which eases
// the interval that starts here.
type layoutStop struct {
	at    float64
	value LayoutValue
	ease  easingFn
}

// sampleLayout samples every layout-tier property the effect declares, in the
// order the properties first appear. A property omitted from a keyframe holds
// the nearest declared value rather than the element's base style: the driver
// cannot read a computed base value for a property it only sees inside a
// keyframe. The interval from keyframe i to i+1 uses keyframe i's timing
// function.
func (e Effect) sampleLayout(at float64) []LayoutValue {
	properties := make([]string, 0, len(e.Frames))
	for i := range e.Frames {
		for j := range e.Frames[i].Layout {
			property := e.Frames[i].Layout[j].Property
			if !containsProperty(properties, property) {
				properties = append(properties, property)
			}
		}
	}
	if len(properties) == 0 {
		return nil
	}

	values := make([]LayoutValue, 0, len(properties))
	for _, property := range properties {
		stops := layoutStops(e.Frames, property)
		if len(stops) == 0 {
			continue
		}
		values = append(values, sampleLayoutStops(stops, e.Timing, at))
	}

	return values
}

func containsProperty(properties []string, property string) bool {
	for _, candidate := range properties {
		if candidate == property {
			return true
		}
	}

	return false
}

// layoutStops collects the keyframes that declare one layout property, in
// sorted frame order.
func layoutStops(frames []Keyframe, property string) []layoutStop {
	stops := make([]layoutStop, 0, len(frames))
	for i := range frames {
		for j := range frames[i].Layout {
			value := frames[i].Layout[j]
			if value.Property != property {
				continue
			}
			stops = append(stops, layoutStop{
				at: frames[i].Offset, value: value, ease: frames[i].frameEasing(nil),
			})

			break
		}
	}

	return stops
}

// sampleLayoutStops samples one property's stops at at. A sample before the
// first stop or after the last holds that endpoint; the driver has no base
// value to fall back to.
func sampleLayoutStops(stops []layoutStop, timing Timing, at float64) LayoutValue {
	if at <= stops[0].at {
		return stops[0].value
	}
	for i := 1; i < len(stops); i++ {
		if at > stops[i].at {
			continue
		}
		prev, next := stops[i-1], stops[i]
		span := next.at - prev.at
		if span <= 0 {
			return next.value
		}
		part := segmentEase(prev.ease, timing, (at-prev.at)/span)

		return interpolateLayout(prev.value, next.value, part)
	}

	return stops[len(stops)-1].value
}

// interpolateLayout blends two values of one property. Matching numeric lists
// with compatible units interpolate term by term; a keyword pair that differs,
// a list with a different shape, or a pair of incompatible units is discrete
// and steps at the midpoint, the way CSS treats a value form with no
// continuous interpolation.
func interpolateLayout(a, b LayoutValue, t float64) LayoutValue {
	discrete := a
	if t >= discreteMidpoint {
		discrete = b
	}
	if a.Property != b.Property || len(a.Terms) != len(b.Terms) {
		return discrete
	}

	terms := make([]LayoutTerm, len(a.Terms))
	for i := range a.Terms {
		term, ok := interpolateLayoutTerm(a.Terms[i], b.Terms[i], t)
		if !ok {
			return discrete
		}
		terms[i] = term
	}

	return LayoutValue{Property: a.Property, Terms: terms}
}

// interpolateLayoutTerm blends one term pair. Two numbers with the same unit
// interpolate linearly, and a unitless zero takes the other side's unit so a
// common 0 to 10px keyframe pair animates. A keyword pair matches only when
// both sides spell the same keyword.
func interpolateLayoutTerm(a, b LayoutTerm, t float64) (LayoutTerm, bool) {
	if a.Keyword != "" || b.Keyword != "" {
		if a.Keyword == "" || a.Keyword != b.Keyword {
			return LayoutTerm{}, false
		}

		return a, true
	}
	switch {
	case a.Unit == b.Unit:
		return LayoutTerm{Number: lerp(a.Number, b.Number, t), Unit: a.Unit}, true
	case a.Unit == "" && a.Number == 0:
		return LayoutTerm{Number: lerp(0, b.Number, t), Unit: b.Unit}, true
	case b.Unit == "" && b.Number == 0:
		return LayoutTerm{Number: lerp(a.Number, 0, t), Unit: a.Unit}, true
	}

	return LayoutTerm{}, false
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

// transformStop is one keyframe's transform at a normalized offset. ease is the
// timing function declared on this keyframe, which eases the interval that
// starts here; nil means the effect's own function, so an unbound Effect does
// not parse a CSS string per sample.
type transformStop struct {
	at    float64
	value TransformValue
	ease  easingFn
}

// rotate3DArgCount is the axis plus angle count of rotate3d().
const rotate3DArgCount = 4

// transformStops collects the keyframes that declare a transform, in sorted
// order, with the easing that applies from each one.
func transformStops(frames []Keyframe) []transformStop {
	stops := make([]transformStop, 0, len(frames)+extraStops)
	for _, frame := range frames {
		if value, ok := frame.transformValue(); ok {
			stops = append(stops, transformStop{
				at: frame.Offset, value: value, ease: frame.frameEasing(nil),
			})
		}
	}

	return stops
}

// padStops brackets the stops with the underlying none at 0% and 100%. A
// keyframe that omits transform holds the underlying paint, which is the
// identity matrix. A synthetic endpoint uses the effect's own easing.
func padStops(stops []transformStop) []transformStop {
	fallback := TransformValue{Decomposed: IdentityTransform()}
	if stops[0].at > 0 {
		stops = append([]transformStop{{at: 0, value: fallback}}, stops...)
	}
	if stops[len(stops)-1].at < 1 {
		stops = append(stops, transformStop{at: 1, value: fallback})
	}

	return stops
}

// sampleTransform samples the decomposed transform and, for a matching
// function list, the parsed functions. A singular matrix endpoint has no
// interpolation, so the interval steps to one endpoint at its midpoint: CSS
// falls back to a discrete animation when a matrix cannot be decomposed. The
// interval from keyframe i to i+1 uses keyframe i's timing function.
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
		part := segmentEase(prev.ease, e.Timing, (at-prev.at)/span)
		value, err := interpolateTransform(prev.value, next.value, part)
		if err != nil {
			if part < discreteMidpoint {
				return prev.value, true
			}
			return next.value, true
		}
		return value, true
	}
	return stops[len(stops)-1].value, true
}

// segmentEase applies the easing of the keyframe that starts an interval, or
// the effect's own function when the keyframe declares none.
func segmentEase(ease easingFn, t Timing, part float64) float64 {
	if ease != nil {
		return ease(part)
	}

	return eased(t, part)
}

// eased applies the effect-level timing to a segment fraction.
func eased(t Timing, part float64) float64 {
	if t.Easing != nil {
		return t.Easing(part)
	}

	return cssmotion.Ease(t.TimingFunction, part)
}

// perspectiveStop is one keyframe's perspective declaration at a normalized
// offset. ease is the timing function declared on this keyframe, which eases
// the interval that starts here.
type perspectiveStop struct {
	at    float64
	value PerspectiveValue
	ease  easingFn
}

// discreteMidpoint is the CSS threshold inside a keyframe interval that is
// discrete because one endpoint is none.
const discreteMidpoint = 0.5

// perspectiveStops collects the keyframes that declare perspective or
// perspective-origin, in sorted order, together with the implicit endpoint the
// padding uses. keyframes() resolves the element's static vanishing point on
// every frame, so the implicit endpoint holds that point; a frame that carried a
// static perspective base also supplies its distance, so a partial list starts
// and ends at the element's camera rather than at none.
//
// When the binder supplied a base, every stop also carries the static distance
// for a keyframe that omits perspective and the static origin for a keyframe
// that omits perspective-origin. CSS holds the underlying value at such a
// keyframe, so an origin-only stop in the middle of a distance animation does
// not pull the distance to zero. The Has flags on the fallback stay limited to
// the properties the list actually animates, so an origin-only list still
// leaves the static distance to the engine.
func perspectiveStops(frames []Keyframe) ([]perspectiveStop, PerspectiveValue) {
	fallback := PerspectiveValue{}
	var base PerspectiveBase
	hasBase := false
	if len(frames) > 0 {
		fallback.OriginX = frames[0].Perspective.OriginX
		fallback.OriginY = frames[0].Perspective.OriginY
		for i := range frames {
			if !hasBase && frames[i].PerspectiveBase.Set {
				base, hasBase = frames[i].PerspectiveBase, true
			}
			fallback.HasDistance = fallback.HasDistance || frames[i].Perspective.HasDistance
			fallback.HasOrigin = fallback.HasOrigin || frames[i].Perspective.HasOrigin
		}
	}
	baseDistance := 0.0
	if hasBase {
		baseDistance = base.Distance
		if !base.HasDistance {
			baseDistance = 0
		}
		fallback.Distance = baseDistance
		fallback.OriginX, fallback.OriginY = base.OriginX, base.OriginY
	}
	stops := make([]perspectiveStop, 0, len(frames)+extraStops)
	for _, frame := range frames {
		if !frame.Perspective.declared() {
			continue
		}
		value := frame.Perspective
		if hasBase {
			if !value.HasDistance {
				value.Distance = baseDistance
				value.HasDistance = fallback.HasDistance
			}
			if !value.HasOrigin {
				value.OriginX, value.OriginY = base.OriginX, base.OriginY
			}
		}
		stops = append(stops, perspectiveStop{
			at: frame.Offset, value: value, ease: frame.frameEasing(nil),
		})
	}

	return stops, fallback
}

// padPerspectiveStops brackets the stops with the underlying value at 0% and
// 100%. A keyframe that omits perspective holds the element's underlying value;
// the fallback carries the element's static camera when the binder supplied it,
// and none otherwise. A synthetic endpoint uses the effect's own easing.
func padPerspectiveStops(stops []perspectiveStop, fallback PerspectiveValue) []perspectiveStop {
	if stops[0].at > 0 {
		stops = append([]perspectiveStop{{at: 0, value: fallback}}, stops...)
	}
	if stops[len(stops)-1].at < 1 {
		stops = append(stops, perspectiveStop{at: 1, value: fallback})
	}

	return stops
}

// samplePerspective samples the camera distance and vanishing point at at. The
// distance is a length that interpolates linearly; a none endpoint switches at
// the midpoint. The origin always interpolates component-wise. The interval
// from keyframe i to i+1 uses keyframe i's timing function.
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
		part := segmentEase(prev.ease, e.Timing, (at-prev.at)/span)

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
// two-dimensional decomposition. Any pair with a 3D function on either side,
// matched or not, goes through the CSS Transforms Level 2 four-by-four
// decomposition; only a singular endpoint reports an error, and the sampler
// falls back to a discrete step for it.
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

// interpolate3D blends a pair that includes at least one 3D function. A
// missing list is the underlying none, which CSS replaces with the identity
// list of the other side, so none to rotateY still interpolates the angle.
// Lists that do not match function by function, and the matrix, matrix3d, and
// perspective functions CSS converts first, blend through the four-by-four
// decomposition and collapse to one matrix3d() for the restamp pass.
func interpolate3D(a, b TransformValue, t float64) (TransformValue, error) {
	from, to := a.Functions, b.Functions
	if len(from) == 0 && len(to) > 0 {
		if identity, ok := identityFunctions(to); ok {
			from = identity
		}
	}
	if len(to) == 0 && len(from) > 0 {
		if identity, ok := identityFunctions(from); ok {
			to = identity
		}
	}
	if interpolable(from, to) {
		return TransformValue{
			Decomposed: blendTransform(a.Decomposed, b.Decomposed, t),
			Functions:  blendFunctions(from, to, t),
			ThreeD:     true,
		}, nil
	}
	return interpolateMatrix3D(from, to, t)
}

// interpolateMatrix3D is the CSS Transforms Level 2 matrix interpolation: both
// lists become four-by-four matrices, each is decomposed into translation,
// scale, skew, perspective, and a rotation quaternion, the components are
// interpolated, and the result is recomposed. The returned list is the single
// matrix3d() the engine restamps.
func interpolateMatrix3D(from, to []TransformFunction, t float64) (TransformValue, error) {
	decomposedFrom, err := Decompose3D(Matrix4FromFunctions(from))
	if err != nil {
		return TransformValue{}, err
	}
	decomposedTo, err := Decompose3D(Matrix4FromFunctions(to))
	if err != nil {
		return TransformValue{}, err
	}
	m := Compose3D(interpolateDecomposition(decomposedFrom, decomposedTo, t))
	return TransformValue{
		Decomposed: Transform{Tx: m[indexM41], Ty: m[indexM42], Sx: 1, Sy: 1, Set: true},
		Functions:  []TransformFunction{{Name: Matrix3D, Args: m[:]}},
		ThreeD:     true,
	}, nil
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
// blend with a 3D list. A perspective function has no finite identity list, so
// it reports false and the caller falls through to the matrix decomposition,
// where none is the identity matrix.
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
