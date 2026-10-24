package motion

import (
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// LayerMode names how a promoted entry composites. It is motion's own string
// type, so the page maps it onto the renderer's mode instead of motion
// exposing a replay type. The zero value is normal source-over.
//
// Motion does not read mix-blend-mode: the display exposes a mode only on the
// blend group nearest an operation, which may belong to an ancestor, and the
// renderer demotes a grouped display wholesale today. Every entry LayerEntries
// reports is therefore LayerModeNormal; a page fills the element's declared
// mode itself once group-contained promotion exists.
type LayerMode string

// LayerModeNormal is source-over compositing of an isolated element. It is the
// zero value, so an entry a page does not inspect maps to normal.
const LayerModeNormal LayerMode = ""

// LayerEntry is one element track the compositor can promote: the element owns
// a subtree whose only moving paint is a 2D transform, an opacity, or both. An
// animation and a transition on the same element share one entry, so the
// element owns one layer, not two. LayerEntries returns the entries that pass
// every promotion check motion can make; the renderer still owns capture, group
// containment, the pixel and count caps, paint-order contiguity, and overlap
// between promoted elements.
//
// Matrix is the motion delta from the pose the operations carried when the
// runtime bound or captured them to the pose they carry now, in display points,
// the same space as DisplayOp.Transform. The runtime writes the sampled matrix
// into the operations rather than publishing the base pose, so an animation's
// Matrix is current * inverse(bindBase) and an element with a static transform
// yields the motion alone, not the static transform doubled. A transition has no
// stable bind pose across a reversal, so its Matrix is the sampled element
// matrix itself and the fixed base is the element-local layout pose. A page
// that maps the entry to a retained layer passes the inverse of Matrix as the
// renderer's Base; the layer texture then holds the fixed pose and compositing
// Matrix reaches the sampled pose. Matrix is the identity for an opacity-only
// entry, and is always invertible on a returned entry.
//
// Alpha is the element's group opacity, 0..1, read from the sampled element
// opacity rather than from the per-operation paint opacity. The renderer
// compensates the group alpha out of the baked operations and applies it once
// at composite, so a promoted opacity entry does not double-apply.
//
// The slice LayerEntries returns belongs to the runtime. It is reused in place
// by the next call, so a caller consumes it before reading again.
type LayerEntry struct {
	// ElementID is the element's box id. It is empty for an anonymous box,
	// where BoxIndex identifies the element across frames.
	ElementID string
	// BoxIndex is the element's index in Display.Boxes, which is also the
	// renderer's promotion key beside the generation.
	BoxIndex int
	// Matrix is the motion matrix in display points. See the type comment.
	Matrix layout.Matrix2D
	// Alpha is the element's group opacity, 0..1, and 1 for an entry whose
	// track writes no opacity field.
	Alpha float64
	// Mode is the compositing mode the entry asks for. See LayerMode.
	Mode LayerMode
	// Has2DTransform reports that the entry writes the operation matrix. The
	// accessor never returns an entry whose transform needs a restamp, so a
	// true here is always representable as a 2D layer matrix.
	Has2DTransform bool
	// HasOpacity reports that the entry writes the element opacity. Alpha
	// carries the sampled value.
	HasOpacity bool
	// HasRestamp reports that an animation player needs the engine's 3D
	// re-projection pass. A layer cannot carry a projective map, so a returned
	// entry always has this false.
	HasRestamp bool
	// HasLayout reports that an animation player also moves the element's
	// layout position, which animation phase 04 owns. A returned entry always
	// has this false.
	HasLayout bool
	// Paused reports that an animation player is frozen. A paused pose does not
	// change per frame, so a returned entry always has this false.
	Paused bool
	// Active reports that the track still has work and is not paused. A
	// returned entry always has this true.
	Active bool
}

// layerDetEpsilon is the smallest determinant a captured base pose may carry
// and still be inverted for the layer delta.
const layerDetEpsilon = 1e-12

// LayerEntries returns the element tracks the compositor may promote on the
// runtime's bound display. It is LayerEntriesFor(r.display); a page whose
// animation driver is off but whose transition driver is on calls
// LayerEntriesFor with the display it is drawing.
func (r *Runtime) LayerEntries() []LayerEntry {
	if r == nil {
		return nil
	}

	return r.LayerEntriesFor(r.display)
}

// LayerEntriesFor returns the element animations and transitions the compositor
// may promote on d this frame, in bind order. It returns only entries that all
// of these hold for:
//
//   - an animation player is active and not paused, so the pose moves between
//     frames;
//   - it writes a 2D transform, an opacity, or both, and no other paint field;
//   - neither the player nor its keyframes can need a restamp pass, because a
//     3D or perspective map is not a layer matrix;
//   - it does not move the element's layout position;
//   - the element box and every target operation are still in the display, and
//     no target is fixed, pinned, or lifted by a non-zero z-index;
//   - the captured base pose and the resulting delta are both invertible, so
//     the page can hand the renderer the inverse delta as Base.
//
// A live transition joins the list under the same rules: transform or opacity
// only, a 2D element, and no fixed, pinned, or non-zero z-index target. A
// transition and an animation on the same element share one entry, because the
// runtime cancels a transition whenever an animation owns the same field. The
// two therefore write disjoint fields and one of the two matrices is the
// identity; where both carry a matrix the documented order is transition
// first, animation second, and the group alphas multiply. A transition whose
// element also runs a layout or restamping animation is not promotable, because
// the layer texture cannot follow the content that pass re-places.
//
// An animation entry is only reported when d is the runtime's bound display,
// because its operation and box indices were resolved against that list. A
// transition entry is resolved against d itself, which is what lets a page with
// the animation driver off still promote transitions.
//
// The display carries the sampled animation and transition matrix, not the
// declared pose. An animation entry's Matrix is current * inverse(bindBase);
// a transition entry's Matrix is the sampled element matrix itself, with the
// element-local layout pose as its fixed base. Either way a page hands the
// renderer the inverse of Matrix (see the LayerEntry comment), and that base
// composed with the operations' matrices is the pose the texture holds for the
// whole run.
//
// The runtime reuses the returned slice, so a steady read allocates nothing; only
// the first read after the entry count grows allocates the larger array. A page
// copies the slice only when the entries must outlive the next call.
func (r *Runtime) LayerEntriesFor(d *Display) []LayerEntry {
	if r == nil {
		return nil
	}
	if d == nil {
		r.layerEntries = r.layerEntries[:0]

		return nil
	}

	entries := r.layerEntries[:0]
	if r.display == d {
		for i := range r.entries {
			entry, ok := r.layerEntry(&r.entries[i], d)
			if !ok {
				continue
			}
			entries = append(entries, entry)
		}
	}
	for i := range r.transitions {
		entry, ok, _ := r.transitionLayerEntry(&r.transitions[i], d)
		if !ok {
			continue
		}
		if at := layerEntryIndex(entries, entry.BoxIndex); at >= 0 {
			entries[at] = mergeLayerEntry(entries[at], entry)
			continue
		}
		entries = append(entries, entry)
	}
	r.layerEntries = entries
	if len(entries) == 0 {
		return nil
	}

	return entries
}

// layerEntryIndex returns the position of the entry for one box index, or -1.
// The transition pass uses it to merge into the animation entry for the same
// element instead of asking the renderer for a second layer.
func layerEntryIndex(entries []LayerEntry, box int) int {
	for i := range entries {
		if entries[i].BoxIndex == box {
			return i
		}
	}

	return -1
}

// mergeLayerEntry folds the next track's entry into the entry already held for
// the same element. Looking from the texture out, the next matrix applies
// first and the held matrix second, so a transition followed by an animation
// composites as animation * transition; the group alphas multiply. The runtime
// cancels a transition when an animation owns the same field, so exactly one of
// the two normally carries a matrix and one of the alphas is 1.
func mergeLayerEntry(held, next LayerEntry) LayerEntry {
	held.Matrix = held.Matrix.Mul(next.Matrix)
	held.Alpha *= next.Alpha
	held.Has2DTransform = held.Has2DTransform || next.Has2DTransform
	held.HasOpacity = held.HasOpacity || next.HasOpacity
	if next.Active {
		held.Active = true
	}

	return held
}

// layerEntry resolves one sampled player to a promotable record, or reports
// false with the reason left implicit: a page that needs the reason reads the
// renderer's demotion counters, which is the single place every rejection is
// counted.
func (r *Runtime) layerEntry(e *entry, d *Display) (LayerEntry, bool) {
	if e.player == nil {
		return LayerEntry{}, false
	}

	transform := e.mask&fieldMatrix != 0
	opacity := e.mask&fieldOpacity != 0
	layoutMotion := e.mask&(fieldX|fieldY) != 0
	if !transform && !opacity {
		return LayerEntry{}, false
	}
	if e.mask&^(fieldMatrix|fieldOpacity) != 0 {
		// A colour or legacy X/Y field re-bakes or re-places content, which a
		// retained texture would fight rather than help.
		return LayerEntry{}, false
	}

	active := e.player.Active()
	paused := e.player.Paused()
	if !active || paused {
		return LayerEntry{}, false
	}

	restamp := e.player.WantsRestamp() || e.player.RestampPending()
	if restamp {
		return LayerEntry{}, false
	}
	if e.boxIndex < 0 || e.boxIndex >= len(d.Boxes) {
		return LayerEntry{}, false
	}

	box := &d.Boxes[e.boxIndex]
	if box.Transform.HasTransform3D || box.Transform.HasPerspective {
		return LayerEntry{}, false
	}

	if !layerOpsAllowed(d, e.indices) {
		return LayerEntry{}, false
	}

	// One player writes one value set to every target, so the first target
	// carries the delta and the element opacity the whole layer needs.
	current, base, ok := e.player.TargetPaint(0)
	if !ok {
		return LayerEntry{}, false
	}

	matrix := layout.IdentityMatrix()
	if transform {
		delta, ok := layerDelta(current, base)
		if !ok {
			return LayerEntry{}, false
		}
		matrix = delta
	}

	alpha := 1.0
	if opacity {
		alpha = layerAlpha(current.Opacity)
	}

	return LayerEntry{
		ElementID:      e.key.id,
		BoxIndex:       e.boxIndex,
		Matrix:         matrix,
		Alpha:          alpha,
		Mode:           LayerModeNormal,
		Has2DTransform: transform,
		HasOpacity:     opacity,
		HasRestamp:     restamp,
		HasLayout:      layoutMotion,
		Paused:         paused,
		Active:         active,
	}, true
}

// layerOpsAllowed checks a target set for the paint-order content a layer
// cannot carry. Fixed, pinned, and z-indexed operations paint in a band a
// retained texture cannot cross, so they keep the element direct. An index
// outside the display means the entry outlived the list it was bound to.
func layerOpsAllowed(d *Display, indices []int) bool {
	for _, index := range indices {
		if index < 0 || index >= len(d.Ops) {
			return false
		}
		op := &d.Ops[index]
		if op.Fixed || op.Pinned || layerZIndexed(op) {
			return false
		}
	}

	return true
}

// layerZIndexed reports an operation that an explicit, non-zero z-index lifts
// into another paint band. ZIndexSet alone is not enough: the engine also marks
// every stacking context it creates, and a CSS transform or an opacity below 1
// creates one at level zero. A transition target carries exactly that marker in
// the committed display, so treating the marker as a band would demote every
// transition layer. Level zero never crosses the positive pin bands the tile
// paint follows.
func layerZIndexed(op *layout.DisplayOp) bool {
	return op.ZIndexSet && op.ZIndex != 0
}

// The reasons a live transition stays direct. LayerEntriesFor drops the entry;
// LayerSkipsFor reports the string so a page or a test can see the cause.
const (
	layerSkipProjection = "3d or perspective transition"
	layerSkipSingular   = "singular transition matrix"
	layerSkipMissingBox = "missing element box"
	layerSkipTargets    = "stale transition operations"
	layerSkipBands      = "fixed, pinned, or z-indexed op"
	layerSkipLayout     = "layout animation on the element"
	layerSkipRestamp    = "restamping animation on the element"
)

// LayerSkip is one live transition the layer path left direct this frame, with
// the reason motion refused it. An animation rejection is not listed: the
// renderer counts those demotions, and motion's animation predicate has no
// separate reason to name. A transition that is not a layer candidate at all,
// such as a colour or background transition, is not a skip: it never asked for
// a texture.
//
// A display or presence cross-fade is not listed either. It needs an outgoing
// layer captured before the new display drops the element's operations, which
// this phase does not build; the transition capture reports it as blocked in
// Stats.TransitionsBlocked and TransitionSkipReason, and it never becomes a
// LayerEntry, so no fade is faked.
type LayerSkip struct {
	// ElementID is the element's box id; it may be empty for an anonymous box.
	ElementID string
	// BoxIndex is the element's index in the display, or -1 when the box is
	// gone.
	BoxIndex int
	// Property is the transition property that asked for the layer.
	Property string
	// Reason names the first condition that refused the promotion.
	Reason string
}

// LayerSkipsFor returns the live transitions LayerEntriesFor left direct, with
// the reason. It resolves every registry entry again rather than reading a
// per-frame cache, so a caller uses it on demand for diagnostics, not once per
// tile frame.
func (r *Runtime) LayerSkipsFor(d *Display) []LayerSkip {
	if r == nil || d == nil || len(r.transitions) == 0 {
		return nil
	}

	var out []LayerSkip
	for i := range r.transitions {
		entry := &r.transitions[i]
		_, ok, reason := r.transitionLayerEntry(entry, d)
		if ok || reason == "" {
			continue
		}
		out = append(out, LayerSkip{
			ElementID: entry.key.element.id,
			BoxIndex:  transitionBoxIndex(d, entry.key.element),
			Property:  entry.key.property,
			Reason:    reason,
		})
	}

	return out
}

// transitionLayerEntry resolves one live transition to a layer record. ok is
// false when the transition cannot promote; reason is non-empty only when it
// asked for a layer and a condition refused it. A colour transition returns
// false with no reason because it never becomes a layer candidate.
//
// A display transition never reaches a live entry: the capture reports it as
// blocked with a reason and the element snaps. The cross-fade's outgoing layer
// is a documented follow-up, not a layer this function invents.
//
// A transition has no bind pose that survives an interruption: a reversal
// replaces the captured start value, so a delta from the from paint would move
// the pose the texture holds. The matrix is therefore the sampled element
// matrix itself, and the page hands the renderer its inverse as Base. The bake
// then holds the element-local layout pose, which is invariant for the whole
// run, so the texture is captured once and a reversal reuses it.
func (r *Runtime) transitionLayerEntry(t *transitionEntry, d *Display) (LayerEntry, bool, string) {
	if t == nil || t.renderer == nil {
		return LayerEntry{}, false, ""
	}
	if t.sampled && !t.active {
		// A finished run is removed by the same Step; this only guards a
		// diagnostic read that lands between the sample and the removal.
		return LayerEntry{}, false, ""
	}
	kind, mask, ok := transitionClass(t.key.property)
	if !ok || kind != cssmotion.KindSubtree {
		return LayerEntry{}, false, ""
	}
	if reason := r.transitionAnimationBlock(t.key.element); reason != "" {
		return LayerEntry{}, false, reason
	}

	boxIndex := transitionBoxIndex(d, t.key.element)
	if boxIndex < 0 {
		return LayerEntry{}, false, layerSkipMissingBox
	}
	box := &d.Boxes[boxIndex]
	if box.Transform.HasTransform3D || box.Transform.HasPerspective {
		// A projective map is not a layer matrix, and a layer flat-composites
		// an element the engine would otherwise re-project.
		return LayerEntry{}, false, layerSkipProjection
	}
	if len(t.indices) == 0 {
		return LayerEntry{}, false, layerSkipTargets
	}
	if !indicesInRange(d, t.indices) || !sameKindList(d, t.indices, t.kinds) {
		return LayerEntry{}, false, layerSkipTargets
	}
	if !layerOpsAllowed(d, t.indices) {
		return LayerEntry{}, false, layerSkipBands
	}

	// One renderer writes one sampled value to every target in the set, so the
	// first target carries the element matrix and the group opacity the layer
	// needs.
	current := cssmotion.Read(&d.Ops[t.indices[0]])

	out := LayerEntry{
		ElementID: t.key.element.id,
		BoxIndex:  boxIndex,
		Matrix:    layout.IdentityMatrix(),
		Alpha:     1,
		Mode:      LayerModeNormal,
		Active:    true,
	}
	if mask&cssmotion.MaskMatrix != 0 {
		matrix := transitionMatrix(current)
		if !layerInvertible(matrix) {
			return LayerEntry{}, false, layerSkipSingular
		}
		out.Matrix = matrix
		out.Has2DTransform = true
	}
	if mask&cssmotion.MaskOpacity != 0 {
		out.Alpha = layerAlpha(current.Opacity)
		out.HasOpacity = true
	}

	return out, true, ""
}

// transitionAnimationBlock reports a live animation on the transition's element
// that a transition layer cannot carry: a layout player moves the content the
// texture baked, and a restamping player needs the engine's 3D pass. A settled
// or triggered player is not live and does not block. A colour player does not
// block either: it changes the layer's signature, so the texture re-bakes once
// per change and otherwise still moves cheaply.
func (r *Runtime) transitionAnimationBlock(key entryKey) string {
	for i := range r.entries {
		e := &r.entries[i]
		if e.key != key || e.player == nil {
			continue
		}
		live := e.player.Active() || e.player.Paused() || (!e.sampled && !e.idle)
		if !live {
			continue
		}
		if e.layout || e.mask&(fieldX|fieldY) != 0 {
			return layerSkipLayout
		}
		if e.player.WantsRestamp() || e.player.RestampPending() {
			return layerSkipRestamp
		}
	}

	return ""
}

// transitionBoxIndex returns the current position of an element key in d, or -1.
// The ordinal counts repeated ids in source order, the same rule Bind and
// CaptureTransitions use, so two boxes with the same id do not trade layers.
func transitionBoxIndex(d *Display, key entryKey) int {
	ordinal := 0
	for i := range d.Boxes {
		if d.Boxes[i].ID != key.id {
			continue
		}
		if ordinal == key.ordinal {
			return i
		}
		ordinal++
	}

	return -1
}

// layerDelta returns the animation delta between a sampled paint and the paint
// the target was bound with. Both matrices live in display point space, so the
// delta is current * inverse(base) and works whether the base pose is the
// identity, a static transform, or an ancestor's transform. It reports false
// when either the base pose or the delta is singular, because a layer cannot
// recover a pose it cannot invert.
func layerDelta(current, base cssmotion.Paint) (layout.Matrix2D, bool) {
	currentMatrix := layout.IdentityMatrix()
	if current.HasMatrix {
		currentMatrix = current.Matrix
	}
	baseMatrix := layout.IdentityMatrix()
	if base.HasMatrix {
		baseMatrix = base.Matrix
	}

	inverse, ok := layerInverse(baseMatrix)
	if !ok {
		return layout.Matrix2D{}, false
	}
	delta := currentMatrix.Mul(inverse)
	if !layerInvertible(delta) {
		return layout.Matrix2D{}, false
	}

	return delta, true
}

// layerInvertible reports whether a 2D affine matrix has an inverse a layer can
// use.
func layerInvertible(m layout.Matrix2D) bool {
	return math.Abs(m.A*m.D-m.B*m.C) >= layerDetEpsilon
}

// layerInverse returns the inverse of a 2D affine matrix, or false for a
// singular base pose.
func layerInverse(m layout.Matrix2D) (layout.Matrix2D, bool) {
	det := m.A*m.D - m.B*m.C
	if math.Abs(det) < layerDetEpsilon {
		return layout.Matrix2D{}, false
	}
	inv := 1 / det

	return layout.Matrix2D{
		A: m.D * inv,
		B: -m.B * inv,
		C: -m.C * inv,
		D: m.A * inv,
		E: (m.C*m.F - m.D*m.E) * inv,
		F: (m.B*m.E - m.A*m.F) * inv,
	}, true
}

// layerAlpha clamps a sampled group opacity to 0..1. NaN reads as opaque, the
// way the renderer's clamp treats it.
func layerAlpha(alpha float64) float64 {
	if math.IsNaN(alpha) {
		return 1
	}
	if alpha < 0 {
		return 0
	}
	if alpha > 1 {
		return 1
	}

	return alpha
}
