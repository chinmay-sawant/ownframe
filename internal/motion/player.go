package motion

import (
	"errors"
	"fmt"
	"github.com/chinmay-sawant/blinkless/layout"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
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
	From Paint

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

// Player applies animation effects to retained display operations.
type Player struct {
	targets []Target
	effects []Effect
	states  []effectState
	tables  []effectTable
	current []Paint
	start   time.Time
	started bool
	active  bool
	touched bool
	// restampSlots holds the latest sample per target that the 2D matrix path
	// could not write. restamp is the exposed form, one entry per such target,
	// and restampBuf is the serialization scratch shared by every target.
	restampSlots []restampSlot
	restamp      []RestampValue
	restampBuf   []byte
}

// restampSlot is one target's latest sampled values that need the engine's
// re-projection pass. A transform slot is kept even after a 2D matrix write,
// because a perspective request for the same element re-stamps the element's
// own transform too.
type restampSlot struct {
	transform        TransformValue
	perspective      PerspectiveValue
	hasTransform     bool
	transformRestamp bool
	hasPerspective   bool
}

// active reports whether the slot asks for one re-projection request.
func (s *restampSlot) active() bool {
	return s.transformRestamp || s.hasPerspective
}

// reset clears every sampled value before the next Step.
func (s *restampSlot) reset() {
	s.transform = TransformValue{}
	s.perspective = PerspectiveValue{}
	s.hasTransform = false
	s.transformRestamp = false
	s.hasPerspective = false
}

// Start starts every untriggered effect at the same clock time. An effect
// declared animation-play-state:paused starts enabled but frozen: it keeps
// sampling its first pose instead of disappearing until a resume.
func (p *Player) Start(now time.Time) {
	p.ensureBound()
	p.start, p.started = now, true
	p.resetCursors()
	for i := range p.states {
		if p.effects[i].Trigger == "" || p.effects[i].Trigger == "none" {
			s := &p.states[i]
			s.start, s.offset = now, 0
			s.rate, s.finished = 1, false
			s.started, s.enabled = true, true
			s.paused = s.timing.Play == Paused
			s.pausedAt = time.Time{}
			s.hasHold = false
			if s.paused {
				s.pausedAt = now
			}
		}
	}
}

// Pause freezes every running effect at now. The active time an effect has
// accumulated is kept, so a later Resume continues the same phase instead of
// starting over; an effect already paused keeps its earlier freeze, and a
// triggered effect that never started is left for its trigger.
//
// A time-based effect also records the exact position it froze at. The position
// is the sampled phase with the signed start delay folded in, so a Clock taken
// after the pause carries it across a rebind without recomputing it from the
// clock origin.
func (p *Player) Pause(now time.Time) {
	if p == nil {
		return
	}
	for i := range p.states {
		s := &p.states[i]
		if !s.started || s.paused {
			continue
		}
		s.offset = currentOffset(s, now)
		if raw, ok := frozenPosition(s); ok {
			s.holdRaw, s.hasHold = raw, true
			s.lastRaw, s.hasRaw = raw, true
		}
		s.paused, s.pausedAt = true, now
		s.enabled = true
	}
}

// Resume continues every paused effect from the phase it froze at. The time
// spent paused is not counted: the current run starts at now with the frozen
// offset, so a sample after a resume matches an uninterrupted run at the same
// active time.
func (p *Player) Resume(now time.Time) {
	if p == nil {
		return
	}
	for i := range p.states {
		s := &p.states[i]
		if !s.started || !s.paused {
			continue
		}
		s.start, s.paused, s.pausedAt = now, false, time.Time{}
		s.enabled = true
		if timeBased(s.timing) {
			// The frozen sample the pause recorded no longer holds: the next
			// sample advances from the active time it reached.
			s.hasHold = false
		}
		if i < len(p.tables) {
			p.tables[i].resetCursors()
		}
	}
}

// Paused reports whether every started effect is frozen. A player with no
// started effect, or with any running effect, reports false.
func (p *Player) Paused() bool {
	if p == nil {
		return false
	}
	started := false
	for i := range p.states {
		if !p.states[i].started {
			continue
		}
		started = true
		if !p.states[i].paused {
			return false
		}
	}

	return started
}

// Clock is the animation clock a player carries across a rebuild: the clock
// origin of the current run, the active time accumulated before it, and the
// freeze a pause left. A page that rebuilds a player on every redraw restores
// this so the animation keeps its phase instead of starting or snapping.
//
// Position is the exact paused sample: the position in iterations, with the
// signed start delay folded in and a reverse rate already applied. HasPosition
// reports that the paused effect had a sample to carry. A paused rebind
// restores it instead of recomputing the phase from Start and Offset, so a
// negative delay cannot be counted twice and a timing value that changed with
// the new stylesheet cannot move the frozen pose.
type Clock struct {
	Start    time.Time
	Offset   time.Duration
	Paused   bool
	PausedAt time.Time

	Position    float64
	HasPosition bool
}

// Clock reports the clock of the first started effect and false when no effect
// has started. The CSS animations on one element share that clock; an effect
// that a trigger gave its own start or rate is not represented. The clock does
// not fold in the time since Start; a caller that needs the phase at a later
// time adds that elapsed time to Offset.
func (p *Player) Clock() (Clock, bool) {
	if p == nil {
		return Clock{}, false
	}
	for i := range p.states {
		s := &p.states[i]
		if !s.started {
			continue
		}
		clock := Clock{Start: s.start, Offset: s.offset, Paused: s.paused, PausedAt: s.pausedAt}
		if s.paused && timeBased(s.timing) && s.hasHold {
			clock.Position, clock.HasPosition = s.holdRaw, true
		}

		return clock, true
	}

	return Clock{}, false
}

// SetClock puts a freshly started player on a clock another player carried. A
// running clock keeps its origin and offset, so a rebind does not restart the
// phase. A paused clock freezes every effect at the carried offset until
// Resume, which continues from that phase. When the clock carries the paused
// sample itself, the freeze uses it directly: the sampled position then does
// not depend on the new player's start or its timing values.
func (p *Player) SetClock(c Clock) {
	if p == nil {
		return
	}
	for i := range p.states {
		s := &p.states[i]
		if !s.started {
			continue
		}
		s.start, s.offset = c.Start, c.Offset
		s.paused, s.pausedAt = c.Paused, c.PausedAt
		s.enabled = true
		s.finished = false
		if timeBased(s.timing) {
			s.hasHold = c.Paused && c.HasPosition
			if s.hasHold {
				s.holdRaw, s.lastRaw, s.hasRaw = c.Position, c.Position, true
			}
		}
		if i < len(p.tables) {
			p.tables[i].resetCursors()
		}
	}
}

// Step samples the time-based animation frame at now.
func (p *Player) Step(now time.Time) bool { return p.StepWithInput(FrameInput{Now: now}) }

// StepWithInput samples every effect using the supplied host clock and timelines.
// Effects are sampled once and their value sets are applied to every target,
// so a frame does no per-target keyframe work and allocates nothing.
func (p *Player) StepWithInput(input FrameInput) bool {
	if !p.started {
		p.Start(input.Now)
	}
	if input.Now.IsZero() {
		input.Now = p.start
	}
	p.ensureBound()
	p.resetRestamp()
	for i := range p.current {
		p.current[i] = p.targets[i].From
	}
	active, hasAvailable := false, false
	for i := range p.effects {
		if len(p.effects[i].Frames) == 0 {
			continue
		}
		state := &p.states[i]
		if !state.enabled {
			continue
		}
		now := input.Now
		if state.paused && !state.pausedAt.IsZero() {
			now = state.pausedAt
		}
		progress, _, live, apply, sampleAvailable := statePosition(state, withNow(input, now))
		if sampleAvailable {
			state.finished = !live
		}
		active = active || live && !state.paused
		hasAvailable = hasAvailable || sampleAvailable
		table := &p.tables[i]
		if table.wasPaused && !state.paused {
			table.resetCursors()
		}
		table.wasPaused = state.paused
		if !table.lastNow.IsZero() && now.Before(table.lastNow) {
			table.resetCursors()
		}
		table.lastNow = now
		if sampleAvailable && apply {
			if table.transform {
				for target := range p.targets {
					current, sample := composeSampled(&p.targets[target],
						p.current[target], p.effects[i], progress)
					p.current[target] = current
					p.noteRestamp(target, sample)
				}
				continue
			}
			var set valueSet
			table.sample(progress, &set)
			for target := range p.targets {
				set.applyTo(&p.current[target], p.targets[target].From, table.mode)
			}
		}
	}
	p.collectRestamp()
	for target := range p.current {
		p.current[target].NeedsRestamp = p.restampSlots[target].active()
	}
	if hasAvailable || p.touched {
		for target := range p.targets {
			Write(p.targets[target].Op, p.current[target])
		}
	}
	p.touched = p.touched || hasAvailable
	p.active = active
	return active
}

// TargetPaint returns one target's latest sampled paint and the paint the
// target was bound with, in target order. A compositor layer reads the pair:
// current.Matrix composed with the inverse of base.Matrix is the animation
// delta from the captured pose to the pose the display carries now, and
// current.Opacity is the element's group opacity.
//
// Before the first sample both values are the base paint, which is the pose the
// operation already carries. A target index outside the player reports false.
func (p *Player) TargetPaint(target int) (current, base Paint, ok bool) {
	if p == nil || target < 0 || target >= len(p.targets) {
		return Paint{}, Paint{}, false
	}
	base = p.targets[target].From
	if !p.touched || len(p.current) != len(p.targets) {
		return base, base, true
	}

	return p.current[target], base, true
}

// BasePaints appends the paint every target composes from to buf and returns
// the extended slice. A rebind that finds its display already carrying the
// player's own sample reads this from the player it replaces, so the new
// player composes from the declared paint instead of stacking the sample on
// itself.
func (p *Player) BasePaints(buf []Paint) []Paint {
	if p == nil {
		return buf
	}
	for i := range p.targets {
		buf = append(buf, p.targets[i].From)
	}

	return buf
}

// SetBasePaints replaces the paint each target composes from, in target order.
// A shorter list leaves the remaining targets alone. The base is captured at
// bind, so this is for a rebind that has to repair a base another display
// carried; the composed sample of the next Step starts from it.
func (p *Player) SetBasePaints(bases []Paint) {
	if p == nil {
		return
	}
	for i := range p.targets {
		if i >= len(bases) {
			return
		}
		p.targets[i].From = bases[i]
	}
}

func withNow(in FrameInput, now time.Time) FrameInput { in.Now = now; return in }

// RestampValue is one target's latest sampled values that the 2D matrix path
// could not write. Transform is the canonical CSS transform list when the
// sample included a transform; it stays empty for a perspective-only sample.
// Perspective is the camera distance in layout points. PerspectiveOriginX and
// PerspectiveOriginY are the sampled perspective-origin in layout points, as an
// offset from the element's border-box origin; HasPerspectiveOrigin reports
// whether the keyframe declared one. Target indexes the targets the player was
// built with, so a scene maps it to the element and operation it stands for.
type RestampValue struct {
	Target    int
	Transform string

	Perspective          float64
	HasPerspective       bool
	PerspectiveOriginX   float64
	PerspectiveOriginY   float64
	HasPerspectiveOrigin bool
}

// RestampValues returns the samples the latest Step could not write as a 2D
// matrix, one per target. A 2D-only sample returns nil. The slice is reused by
// the next Step, so a caller consumes it before stepping again.
func (p *Player) RestampValues() []RestampValue {
	if p == nil || len(p.restamp) == 0 {
		return nil
	}

	return p.restamp
}

// resetRestamp sizes the per-target slots to the target list and empties the
// exposed slice. Slots are cleared rather than reallocated, so a live 3D or
// perspective sample keeps its backing arrays.
func (p *Player) resetRestamp() {
	if len(p.restampSlots) != len(p.targets) {
		p.restampSlots = make([]restampSlot, len(p.targets))
		p.restamp = make([]RestampValue, 0, len(p.targets))
	}
	for i := range p.restampSlots {
		p.restampSlots[i].reset()
	}
	p.restamp = p.restamp[:0]
}

// noteRestamp records the values sampled for one target. A transform sample
// replaces the previous transform value even when this effect wrote a 2D
// matrix, so a perspective that needs the same element restamped carries the
// animated transform the operation already shows. A later perspective-less
// 2D transform still clears the 3D request because transformRestamp is false.
func (p *Player) noteRestamp(target int, sample restampSample) {
	if target < 0 || target >= len(p.restampSlots) {
		return
	}
	slot := &p.restampSlots[target]
	if sample.hasTransform {
		slot.transform = sample.transform
		slot.hasTransform = true
		slot.transformRestamp = sample.transformRestamp
	}
	if sample.hasPerspective {
		slot.perspective = sample.perspective
		slot.hasPerspective = true
	}
}

// collectRestamp rebuilds the exposed slice from the per-target slots,
// serializing each transform once. A player with no 3D transform and no
// perspective leaves it empty, so RestampValues reports nil.
func (p *Player) collectRestamp() {
	p.restamp = p.restamp[:0]
	for i := range p.restampSlots {
		slot := &p.restampSlots[i]
		if !slot.active() {
			continue
		}
		value := RestampValue{Target: i}
		if slot.hasPerspective {
			value.HasPerspective = slot.perspective.HasDistance
			value.Perspective = slot.perspective.Distance
			value.PerspectiveOriginX = slot.perspective.OriginX
			value.PerspectiveOriginY = slot.perspective.OriginY
			value.HasPerspectiveOrigin = slot.perspective.HasOrigin
		}
		if slot.hasTransform {
			p.restampBuf = AppendTransformCSS(p.restampBuf[:0], slot.transform)
			value.Transform = string(p.restampBuf)
		}
		p.restamp = append(p.restamp, value)
	}
}

// Active reports whether the latest sample leaves work to do.
func (p *Player) Active() bool { return p != nil && p.active }

// newAnimationPlayer validates and copies the targets and keyframes for one animation.
func newAnimationPlayer(targets []Target, frames []Keyframe, timing Timing) (*Player, error) {
	return NewComposite(targets, []Effect{{Frames: frames, Timing: timing}})
}

// NewComposite builds one player that composes several effects per frame.
func NewComposite(targets []Target, effects []Effect) (*Player, error) {
	if len(targets) == 0 {
		return nil, fmt.Errorf("animation needs at least one target")
	}
	if len(effects) == 0 {
		return nil, fmt.Errorf("animation needs at least one effect")
	}
	effects = append([]Effect(nil), effects...)
	for i, target := range targets {
		if target.Op == nil {
			return nil, fmt.Errorf("animation target %d has no operation", i)
		}
	}
	for i := range effects {
		frames := append([]Keyframe(nil), effects[i].Frames...)
		for j, frame := range frames {
			if math.IsNaN(frame.Offset) || frame.Offset < 0 || frame.Offset > 1 {
				return nil, fmt.Errorf("effect %d keyframe %d offset must be between 0 and 1", i, j)
			}
		}
		sort.SliceStable(frames, func(a, b int) bool { return frames[a].Offset < frames[b].Offset })
		for j := range frames {
			if frames[j].Err != nil {
				return nil, fmt.Errorf("effect %d keyframe %d: %w", i, j, frames[j].Err)
			}
		}
		effects[i].Frames = frames
	}
	targets = append([]Target(nil), targets...)
	states := make([]effectState, len(effects))
	for i := range effects {
		states[i].timing = effects[i].Timing
		states[i].rate = 1
	}
	return &Player{targets: targets, effects: effects, states: states}, nil
}

// Direction controls the order in which iterations play.
type Direction uint8

const (
	Normal Direction = iota
	Reverse
	Alternate
	AlternateReverse
)

// FillMode controls paint before and after the active interval.
type FillMode uint8

const (
	FillNone FillMode = iota
	FillForwards
	FillBackwards
	FillBoth
)

// PlayState controls whether Step applies the keyframes.
type PlayState uint8

const (
	Running PlayState = iota
	Paused
)

// Composition controls how an effect combines with the sampled paint.
type Composition uint8

const (
	Replace Composition = iota
	Add
	Accumulate
)

// Range maps a named timeline range to normalized timeline progress.
type Range struct{ Start, End float64 }

// Timing describes the animation clock, timeline, and iteration behavior.
type Timing struct {
	Duration, Delay, EndDelay time.Duration
	DurationAuto              bool
	Iterations                float64
	Infinite                  bool
	Direction                 Direction
	Fill                      FillMode
	Play                      PlayState
	TimingFunction            string
	Easing                    Easing
	Composition               Composition
	Timeline                  string
	RangeStart, RangeEnd      string
	RangeCenter               string
}

// FrameInput carries host clock and timeline progress for one frame.
type FrameInput struct {
	Now                            time.Time
	ScrollProgress, ViewProgress   float64
	SourceProgress, TargetProgress float64
	Timelines                      map[string]float64
	Ranges                         map[string]Range
	RangeEdges                     map[string]float64
}

// Effect is one animation item sharing its target paint with sibling effects.
type Effect struct {
	Frames  []Keyframe
	Timing  Timing
	Trigger string
}

type effectState struct {
	timing           Timing
	start, pausedAt  time.Time
	offset           time.Duration
	rate             float64
	holdRaw, lastRaw float64
	started, enabled bool
	paused, hasHold  bool
	hasRaw           bool
	finished         bool
}

func statePosition(s *effectState, input FrameInput) (float64, int, bool, bool, bool) {
	t, rate := s.timing, s.rate
	if rate == 0 {
		rate = 1
	}
	duration := resolveDuration(t, input)
	if timeBased(t) && duration <= 0 {
		return zeroDurationPosition(t, currentOffset(s, input.Now))
	}
	raw, ok := rawPosition(t, s.start, input)
	if !ok {
		return 0, 0, false, false, false
	}
	switch {
	case s.hasHold:
		raw = s.holdRaw
	case timeBased(t):
		elapsed := float64(s.offset)
		if !s.paused {
			elapsed += float64(input.Now.Sub(s.start)) * rate
		}
		raw = (elapsed - float64(t.Delay)) / float64(duration)
		if s.paused {
			// The first paused sample freezes the exact position it reached,
			// delay included, so a later Clock handoff carries the phase
			// rather than recomputing it from the clock origin.
			s.holdRaw, s.hasHold = raw, true
		}
	case rate < 0:
		raw = t.Iterations - raw
	}
	s.lastRaw, s.hasRaw = raw, true
	progress, cycle, live, apply, available := samplePosition(t, raw)
	if rate < 0 {
		live = raw > 0
	}
	return progress, cycle, live, apply, available
}

// resolveDuration returns the cycle length the position code divides by.
// animation-duration:auto has no intrinsic time for a document-clock
// animation, so it resolves to zero and the animation applies its end value at
// once. A timeline-driven animation keeps its declared duration; resolving auto
// to the timeline's range belongs to the timeline phase.
func resolveDuration(t Timing, _ FrameInput) time.Duration {
	if !timeBased(t) {
		return t.Duration
	}
	if t.DurationAuto || t.Duration < 0 {
		return 0
	}
	return t.Duration
}

func currentOffset(s *effectState, now time.Time) time.Duration {
	offset := s.offset
	if !s.paused && s.started {
		offset += time.Duration(float64(now.Sub(s.start)) * s.rate)
	}
	return offset
}

// frozenPosition returns the position a pause freezes a time-based effect at,
// with the signed start delay folded in and a reverse rate applied. It reads
// the active time the state already carries, so the caller must have updated
// that first. A timeline effect has no document-clock position to freeze and
// reports false.
func frozenPosition(s *effectState) (float64, bool) {
	t := s.timing
	if !timeBased(t) {
		return 0, false
	}
	duration := resolveDuration(t, FrameInput{})
	if duration <= 0 {
		return 0, false
	}
	raw := (float64(s.offset) - float64(t.Delay)) / float64(duration)
	if s.rate < 0 {
		raw = t.Iterations - raw
	}

	return raw, true
}

func rawPosition(t Timing, start time.Time, input FrameInput) (float64, bool) {
	raw, available := timelineProgress(t, input)
	if !available {
		return 0, false
	}
	if t.Timeline == "" || t.Timeline == "auto" {
		if t.Duration <= 0 {
			return 0, false
		}
		raw = float64(input.Now.Sub(start)-t.Delay) / float64(t.Duration)
	} else {
		lo, hi, ok := timelineRange(t, input)
		if !ok || hi <= lo {
			return 0, false
		}
		raw = centerProgress((raw-lo)/(hi-lo), t.RangeCenter, input, lo, hi)
		if t.Iterations > 0 && !t.Infinite {
			raw *= t.Iterations
		}
		if t.Duration > 0 {
			raw -= float64(t.Delay) / float64(t.Duration)
		}
	}
	return raw, true
}

func samplePosition(t Timing, raw float64) (float64, int, bool, bool, bool) {
	if raw < 0 {
		return directed(0, 0, t.Direction), 0, true, t.Fill == FillBackwards || t.Fill == FillBoth, true
	}
	if t.Infinite {
		cycle := math.Floor(raw)
		return directed(raw-cycle, int(cycle), t.Direction), int(cycle), true, true, true
	}
	if t.Iterations <= 0 {
		// Iteration count zero never reaches the active interval; forwards
		// fill still applies the value the animation would have ended on.
		return directed(1, 0, t.Direction), 0, false,
			t.Fill == FillForwards || t.Fill == FillBoth, true
	}
	if raw >= t.Iterations {
		// A fractional iteration count ends inside the last cycle, so last
		// is the fractional tail rather than an implicit whole keyframe.
		cycle := int(math.Ceil(t.Iterations) - 1)
		last := t.Iterations - float64(cycle)
		if last == 0 {
			last = 1
		}
		end := t.Iterations + float64(t.EndDelay)/float64(maxDuration(t.Duration))
		if raw < end {
			// End delay holds the last sampled value for the declared delay,
			// whatever the fill mode decides after it.
			return directed(last, cycle, t.Direction), cycle, true, true, true
		}
		return directed(last, cycle, t.Direction), cycle, false,
			t.Fill == FillForwards || t.Fill == FillBoth, true
	}
	cycle := int(math.Floor(raw))
	return directed(raw-float64(cycle), cycle, t.Direction), cycle, true, true, true
}

// zeroDurationPosition samples a zero-length time-based animation. There is no
// active interval to divide, so the sample is the end value once the start
// delay has passed; the end delay holds it, and the fill mode decides whether
// it survives afterwards. Such an animation has no time left to play, so it is
// live only while an end delay is holding it.
func zeroDurationPosition(t Timing, elapsed time.Duration) (float64, int, bool, bool, bool) {
	if elapsed < t.Delay {
		return directed(0, 0, t.Direction), 0, false,
			t.Fill == FillBackwards || t.Fill == FillBoth, true
	}
	value := directed(1, 0, t.Direction)
	if elapsed < t.Delay+t.EndDelay {
		return value, 0, true, true, true
	}

	return value, 0, false,
		t.Fill == FillForwards || t.Fill == FillBoth, true
}

// timeBased reports whether a timing uses the document clock rather than a
// declared timeline.
func timeBased(t Timing) bool { return t.Timeline == "" || t.Timeline == "auto" }

func directed(progress float64, cycle int, direction Direction) float64 {
	reverse := direction == Reverse || direction == AlternateReverse
	if (direction == Alternate || direction == AlternateReverse) && cycle%2 != 0 {
		reverse = !reverse
	}
	if reverse {
		return 1 - progress
	}
	return progress
}

func maxDuration(d time.Duration) time.Duration {
	if d <= 0 {
		return time.Second
	}
	return d
}
