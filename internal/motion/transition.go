package motion

import (
	"image"
	"math"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
	"github.com/chinmay-sawant/ownframe/internal/csstransition"
	"github.com/chinmay-sawant/ownframe/internal/page/dirty"
	"github.com/chinmay-sawant/ownframe/internal/transition"
)

// transitionPaintEpsilon is the difference at which two captured paints count
// as a style change. The engine resolves one declaration to the same values on
// both displays, so an unchanged property is bit-identical in practice; the
// epsilon only absorbs a rounding difference in a colour or a matrix.
const transitionPaintEpsilon = 1e-6

// transitionKind selects the registry a capture binds. A paint capture writes
// operation fields, a geometry capture reaches the engine's relayout override,
// and a discrete capture installs a value at a progress boundary.
type transitionKind uint8

const (
	transitionPaint transitionKind = iota
	transitionGeometry
	transitionDiscrete
)

// TransitionCapture is one property change captured at a display commit. The
// page holds it between the commit and the bind, then hands it back unchanged.
// The fields stay unexported so only CaptureTransitions can build one.
type TransitionCapture struct {
	key      transitionKey
	kind     transitionKind
	indices  []int
	from     []cssmotion.Paint
	to       []cssmotion.Paint
	record   layout.Transition
	property string
	// geometry carries the layout-length endpoints of a geometry capture.
	geometry geometryCapture
	// discrete carries the rule a discrete capture installs.
	discrete *discreteCapture
	// cancel reports that the style change landed on the value already on
	// screen but away from the destination a live run is heading for. The
	// committed paint is already the answer, so the live renderer is dropped
	// without starting a replacement.
	cancel bool
}

// transitionKey addresses one element's transition for one property across a
// publish. A repeated element id keeps the ordinal the animation registry
// assigns, so two boxes with the same id do not trade transitions.
type transitionKey struct {
	element  entryKey
	property string
}

// transitionEntry is one live renderer and the clock the runtime keeps for it.
// The renderer holds the target operations; retarget points them at the next
// published list without restarting the run.
type transitionEntry struct {
	key      transitionKey
	renderer *transition.Renderer
	indices  []int
	kinds    []uint8
	// reversingStart is the CSS reversing-adjusted start value: the value a
	// reversal of this run returns to. A normal run starts with the same value
	// as its captured from paint; a reversal takes the destination of the run
	// it replaced.
	reversingStart []cssmotion.Paint
	// reversingFactor is the CSS reversing shortening factor. It is 1 for a
	// normal run and the portion of the old value range a reversal crossed.
	reversingFactor float64
	// to is the destination the run is heading for, and the destination a
	// reverse returns from.
	to     []cssmotion.Paint
	record layout.Transition
	start  time.Time
	// sampled is set by the first Step after a bind, so Active asks for that
	// frame even before the first delay sample.
	sampled bool
	// active is the renderer's answer at the latest Step.
	active bool
	// wrote records that the from paint reached the operations. A delayed
	// transition writes it once and then parks until the delay ends instead of
	// repainting the same value every frame.
	wrote bool
}

// transitionCaptureCounts is the capture-side diagnostic block BindTransitions
// copies into Stats.
type transitionCaptureCounts struct {
	noID    int
	noOld   int
	skipped int
	blocked int
	reason  string
}

// block records one property change the display cannot transition yet. The
// first reason is the one Stats reports.
func (c *transitionCaptureCounts) block(reason string) {
	c.blocked++
	if c.reason == "" {
		c.reason = reason
	}
}

// transitionPropertyList enumerates the transition properties this phase
// captures. The order is deterministic so a commit produces the same capture
// order even when the stylesheet declares transition-property: all.
//
// A property whose value the display exposes is claimed here. The paint
// properties carry operation fields, width/height/top/left carry the element's
// box, and visibility, content-visibility, and z-index carry presence or a
// stacking level. Display, order, and the grid-line integers are listed so an
// all declaration reaches them and their blocked reason is reported; they do
// not produce a renderer.
var transitionProperties = []string{
	"background-color", "accent-color",
	"color", "text-decoration-color", "text-emphasis-color",
	"border-color", "border-top-color", "border-right-color",
	"border-bottom-color", "border-left-color", "border-block-color",
	"border-block-start-color", "border-block-end-color", "border-inline-color",
	"border-inline-start-color", "border-inline-end-color", "outline-color",
	"column-rule-color", "row-rule-color", "rule-color",
	"opacity", "transform",
	"width", "height", "top", "left",
	"position", "visibility", "content-visibility", "z-index",
	"display", "order",
	"grid-column-start", "grid-column-end",
	"grid-row-start", "grid-row-end",
}

func transitionPropertyList() []string {
	return transitionProperties
}

// transitionPaintTarget maps a property to the operations and paint fields a
// paint renderer writes. It is transitionClass plus position, which moves the
// captured operations at the halfway point but must not offer itself to the
// layer path as a subtree transform.
func transitionPaintTarget(property string) (cssmotion.Kind, cssmotion.Mask, bool) {
	if property == "position" {
		return cssmotion.KindSubtree, cssmotion.MaskX | cssmotion.MaskY, true
	}

	return transitionClass(property)
}

// transitionClass maps a transition property to the operations and paint
// fields it writes. It mirrors the animation paint classes: a colour belongs
// to the nearest owner and its kind, while opacity and the matrix move the
// element's whole subtree.
func transitionClass(property string) (cssmotion.Kind, cssmotion.Mask, bool) {
	switch property {
	case "opacity":
		return cssmotion.KindSubtree, cssmotion.MaskOpacity, true
	case "transform":
		return cssmotion.KindSubtree, cssmotion.MaskMatrix, true
	case "background-color", "accent-color":
		return cssmotion.KindFill, cssmotion.MaskRGB, true
	case "color", "text-decoration-color", "text-emphasis-color":
		return cssmotion.KindInk, cssmotion.MaskRGB, true
	case "border-color", "border-top-color", "border-right-color",
		"border-bottom-color", "border-left-color", "border-block-color",
		"border-block-start-color", "border-block-end-color", "border-inline-color",
		"border-inline-start-color", "border-inline-end-color", "outline-color",
		"column-rule-color", "row-rule-color", "rule-color":
		return cssmotion.KindStroke, cssmotion.MaskRGB, true
	}
	return 0, 0, false
}

// transitionMask returns the paint fields a property owns.
func transitionMask(property string) cssmotion.Mask {
	_, mask, _ := transitionPaintTarget(property)
	return mask
}

// transitionSupported reports whether a record for property can reach one of
// the capture paths. A property outside this set is not claimed: the commit
// snaps it and no renderer is built.
func transitionSupported(property string) bool {
	if _, _, ok := transitionPaintTarget(property); ok {
		return true
	}

	return transitionGeometryProperty(property) || transitionDiscreteProperty(property) ||
		transitionReportProperty(property)
}

// clampTransitionProgress keeps a sampled or computed progress inside the
// value range the CSS transitions model uses.
func clampTransitionProgress(value float64) float64 {
	switch {
	case value < 0:
		return 0
	case value > 1:
		return 1
	}
	return value
}

// reverseFactor applies the CSS reversing shortening factor formula. progress
// is the old run's timing-function output at the style change, and previous is
// the old run's own reversing shortening factor. The absolute value and clamp
// absorb a timing function whose handles overshoot the value range.
func reverseFactor(progress, previous float64) float64 {
	return math.Abs(clampTransitionProgress(progress*previous + (1 - previous)))
}

// samePaintSet reports that every captured operation carries the same masked
// value in both sets.
func samePaintSet(a, b []cssmotion.Paint, mask cssmotion.Mask) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if paintChanged(a[i], b[i], mask) {
			return false
		}
	}
	return true
}

// transitionRecords reduces one box's transition list to the winning record
// per property. A later declaration overrides an earlier one, and "all"
// expands to every property this phase can capture. The returned map carries
// the record exactly as declared, so "all" keeps Behavior and timing.
func transitionRecords(box layout.Box) map[string]layout.Transition {
	var out map[string]layout.Transition
	for i := range box.Transitions {
		item := box.Transitions[i]
		if item.Property == "all" {
			for _, property := range transitionPropertyList() {
				if out == nil {
					out = make(map[string]layout.Transition)
				}
				out[property] = item
			}
			continue
		}
		if !transitionSupported(item.Property) {
			continue
		}
		if out == nil {
			out = make(map[string]layout.Transition)
		}
		out[item.Property] = item
	}
	return out
}

// boxRef is one element box and its position in a display.
type boxRef struct {
	box   layout.Box
	index int
}

// indexBoxes maps every id-bearing box to its position. The ordinal counts
// repeated ids in source order, the same rule Bind uses for animation entries.
// An anonymous box has no key because a transition cannot be matched to it.
func indexBoxes(boxes []layout.Box) map[entryKey]boxRef {
	if len(boxes) == 0 {
		return nil
	}
	idCount := 0
	for i := range boxes {
		if boxes[i].ID != "" {
			idCount++
		}
	}
	if idCount == 0 {
		return nil
	}
	out := make(map[entryKey]boxRef, idCount)
	ordinals := make(map[string]int, idCount)
	for i := range boxes {
		box := boxes[i]
		if box.ID == "" {
			continue
		}
		ordinal := ordinals[box.ID]
		ordinals[box.ID] = ordinal + 1
		out[entryKey{id: box.ID, ordinal: ordinal}] = boxRef{box: box, index: i}
	}
	return out
}

// CaptureTransitions reads the style change between two published displays.
// It returns one capture per changed property that the new box declares a
// transition for. The old box is matched by id and ordinal; an element with no
// id is counted and skipped, because a splice can move indices.
//
// The from paint comes from the old display, which the runtime has been
// sampling into, so an interrupted transition starts from the value on screen
// rather than the declared base. The destination comes from the new display.
// A target set whose operation count or kinds changed is skipped and counted:
// a half-interpolated element looks worse than a snap.
func (r *Runtime) CaptureTransitions(previous, next *Display) []TransitionCapture {
	if r == nil {
		return nil
	}
	r.transitionCapture = transitionCaptureCounts{}
	if previous == nil || next == nil {
		return nil
	}

	hasTransitions := false
	for i := range next.Boxes {
		if len(next.Boxes[i].Transitions) != 0 {
			hasTransitions = true
			break
		}
	}
	if !hasTransitions {
		for i := range previous.Boxes {
			if len(previous.Boxes[i].Transitions) != 0 {
				hasTransitions = true
				break
			}
		}
	}
	if !hasTransitions {
		return nil
	}

	oldBoxes := indexBoxes(previous.Boxes)
	var out []TransitionCapture
	ordinals := make(map[string]int, len(oldBoxes))
	// Both displays are resolved through one shared index each, so the
	// per-property selections below share two ownership scans instead of
	// repeating them for every transitioning element. The indices are built
	// only when some box declares a transition, so a transition-free publish
	// pays no scan.
	var prevIndex, nextIndex *cssmotion.Index
	for i := range next.Boxes {
		if len(next.Boxes[i].Transitions) != 0 {
			prevIndex = cssmotion.NewIndex(previous)
			nextIndex = cssmotion.NewIndex(next)
			break
		}
	}
	for i := range next.Boxes {
		box := next.Boxes[i]
		if box.ID == "" {
			if len(box.Transitions) != 0 {
				r.transitionCapture.noID++
			}
			continue
		}
		ordinal := ordinals[box.ID]
		ordinals[box.ID] = ordinal + 1
		if len(box.Transitions) == 0 {
			continue
		}
		key := entryKey{id: box.ID, ordinal: ordinal}
		records := transitionRecords(box)
		old, ok := oldBoxes[key]
		if !ok {
			r.transitionCapture.noOld++
			if _, declared := records["display"]; declared {
				r.transitionCapture.block(displayEntryReason)
			}
			continue
		}
		animated := layoutAnimatedProperties(next, box)
		current := boxRef{box: box, index: i}
		var seen []transitionTarget
		for _, property := range transitionPropertyList() {
			record, ok := records[property]
			if !ok {
				continue
			}
			if transitionGeometryProperty(property) {
				if ordinal > 0 {
					// The engine addresses an element by id, so a repeated id
					// cannot be relaid out as a second box.
					r.transitionCapture.block(duplicateIDReason(property))
					continue
				}
				if layoutAnimationOwns(animated, property) {
					// The animation writes the same layout property later in
					// the frame, so the transition never becomes a renderer.
					r.transitionCapture.skipped++
					r.transitionCapture.block(animationOwnsReason(property))
					continue
				}
				capture, changed := captureGeometryChange(old, next, current, property)
				capture.key = transitionKey{element: key, property: property}
				capture.record = record
				if !changed {
					// The presented box already equals the committed one. It is
					// a real change only when a live run is heading elsewhere.
					live := r.liveGeometry(capture.key)
					if live == nil || math.Abs(live.to-capture.geometry.to) <= transitionPaintEpsilon {
						continue
					}
					capture.cancel = true
					out = append(out, capture)
					continue
				}
				if r.geometryTargetUnchanged(capture.key, capture.geometry.to) {
					// The new base is the destination the live run already has;
					// keep the running clock instead of restarting from the box.
					continue
				}
				out = append(out, capture)
				continue
			}
			if transitionDiscreteProperty(property) {
				capture, ok := r.captureDiscreteChange(previous, old, next, current, property, record,
					prevIndex, nextIndex)
				if !ok {
					continue
				}
				capture.key = transitionKey{element: key, property: property}
				if property == "visibility" || property == "content-visibility" {
					// Both names can describe one presence change. The presence
					// rule writes the same alpha field, so one rule owns it.
					if sameTransitionTarget(seen, capture.indices, cssmotion.MaskAlpha) {
						continue
					}
					seen = append(seen, transitionTarget{indices: capture.indices,
						mask: cssmotion.MaskAlpha})
				}
				if r.discreteTargetUnchanged(capture.key, capture.discrete) {
					continue
				}
				out = append(out, capture)
				continue
			}
			if transitionReportProperty(property) {
				r.noteReportChange(previous, old, next, current, property)
				continue
			}
			_, mask, _ := transitionPaintTarget(property)
			capture, changed, ok := captureTransition(previous, old, next, current,
				property, record, &r.transitionCapture, prevIndex, nextIndex)
			if !ok {
				continue
			}
			capture.key = transitionKey{element: key, property: property}
			capture.kind = transitionPaint
			if !changed {
				// The new computed value is already the paint on screen. That
				// is a real style change only when a live run is heading
				// somewhere else: a delayed run interrupted back to its start
				// value, or a target that happens to equal the presented
				// sample. Both cancel without a replacement.
				live := r.liveTransition(capture.key)
				if live == nil || samePaintSet(capture.to, live.to, mask) {
					continue
				}
				if sameTransitionTarget(seen, capture.indices, mask) {
					continue
				}
				seen = append(seen, transitionTarget{indices: capture.indices, mask: mask})
				capture.cancel = true
				out = append(out, capture)
				continue
			}
			if sameTransitionTarget(seen, capture.indices, mask) {
				// Two property names can paint the same operation field, such
				// as background-color and accent-color on one fill. The paint
				// model cannot split them, so one renderer owns the field.
				continue
			}
			seen = append(seen, transitionTarget{indices: capture.indices, mask: mask})
			if r.transitionTargetUnchanged(capture.key, capture.to) {
				// The new computed value is the destination the live run is
				// already heading for. This publish did not change the style;
				// keep the running clock instead of restarting from the screen.
				continue
			}
			out = append(out, capture)
		}
	}

	r.noteDisplayExits(previous, next)
	return out
}

// transitionTarget identifies one paint field set already captured for an
// element in this commit.
type transitionTarget struct {
	indices []int
	mask    cssmotion.Mask
}

// sameTransitionTarget reports whether a property resolves to a field set an
// earlier property already owns.
func sameTransitionTarget(seen []transitionTarget, indices []int, mask cssmotion.Mask) bool {
	for i := range seen {
		if seen[i].mask != mask || len(seen[i].indices) != len(indices) {
			continue
		}
		same := true
		for k := range indices {
			if seen[i].indices[k] != indices[k] {
				same = false

				break
			}
		}
		if same {
			return true
		}
	}
	return false
}

// captureTransition captures one property's selected operations. changed is
// false when the old and new paints carry the same masked value, which is
// still useful to the caller when a live run is heading elsewhere. ok is false
// when the operation set itself cannot carry a transition.
func captureTransition(previous *Display, old boxRef, next *Display, current boxRef,
	property string, record layout.Transition,
	counts *transitionCaptureCounts, prevIndex, nextIndex *cssmotion.Index) (capture TransitionCapture, changed, ok bool) {
	kind, mask, ok := transitionPaintTarget(property)
	if !ok {
		return TransitionCapture{}, false, false
	}
	if property == "transform" &&
		(old.box.Transform.HasTransform3D || current.box.Transform.HasTransform3D) {
		// Phase 01 carries the two-dimensional matrix. A 3D or perspective
		// element awaits the transition matrix phase and snaps for now.
		counts.skipped++
		return TransitionCapture{}, false, false
	}
	oldIndices := prevIndex.Select(old.box, old.index, kind)
	newIndices := nextIndex.Select(current.box, current.index, kind)
	if len(oldIndices) != len(newIndices) || !sameOpKinds(previous, oldIndices, next, newIndices) {
		if len(oldIndices) > 0 || len(newIndices) > 0 {
			counts.skipped++
		}
		return TransitionCapture{}, false, false
	}
	if len(newIndices) == 0 {
		return TransitionCapture{}, false, false
	}

	matrix := mask == cssmotion.MaskMatrix
	from := readTransitionPaints(previous, oldIndices, mask, matrix)
	to := readTransitionPaints(next, newIndices, mask, matrix)
	return TransitionCapture{indices: newIndices, from: from, to: to, record: record,
		property: property}, paintSetChanged(from, to, mask), true
}

// liveTransition returns the running entry for one element and property, or
// nil. Capture reads it to tell an unchanged paint that is already the
// destination from an unchanged paint that cancels a run heading elsewhere.
func (r *Runtime) liveTransition(key transitionKey) *transitionEntry {
	for i := range r.transitions {
		if r.transitions[i].key == key {
			return &r.transitions[i]
		}
	}
	return nil
}

// transitionTargetUnchanged reports that a live run for this element and
// property already targets the captured destination. An unrelated publish
// rebuilds the static paint from the same stylesheet, so the presented paint
// on the old display differs from the new base even though the computed value
// did not change; the registry must survive that publish.
func (r *Runtime) transitionTargetUnchanged(key transitionKey, to []cssmotion.Paint) bool {
	_, mask, ok := transitionClass(key.property)
	if !ok {
		return false
	}
	for i := range r.transitions {
		entry := &r.transitions[i]
		if entry.key != key {
			continue
		}
		if len(entry.to) != len(to) {
			return false
		}
		for k := range to {
			if paintChanged(entry.to[k], to[k], mask) {
				return false
			}
		}
		return true
	}
	return false
}

// animationOwns reports whether an active keyframe animation on the element
// writes any field in mask. A player that has not sampled yet will write on the
// first tick and a paused player keeps its pose, so both own the field. A
// finished player still owns it when a fill mode keeps its last pose, and an
// untouched trigger player owns nothing until it fires.
func (r *Runtime) animationOwns(key entryKey, mask cssmotion.Mask) bool {
	if mask == 0 {
		return false
	}
	for i := range r.entries {
		e := &r.entries[i]
		if e.key != key || e.layout || e.player == nil || e.paintMask&mask == 0 {
			continue
		}
		if e.player.Active() || e.player.Paused() {
			return true
		}
		if !e.sampled && !e.idle {
			// The first sample writes the declared pose, so the animation
			// already owns the field on the commit that bound it.
			return true
		}
		if e.idle {
			continue
		}
		if r.animationFills(e) {
			return true
		}
	}
	return false
}

// animationFills reports whether an entry's element declares a fill mode that
// keeps the animation's last pose after it finishes.
func (r *Runtime) animationFills(e *entry) bool {
	if r.display == nil || e.boxIndex < 0 || e.boxIndex >= len(r.display.Boxes) {
		return false
	}
	for i := range r.display.Boxes[e.boxIndex].Animations {
		fill := r.display.Boxes[e.boxIndex].Animations[i].FillMode
		if fill == "forwards" || fill == "both" {
			return true
		}
	}
	return false
}

// sameOpKinds compares the operation kinds at matching positions. A changed
// kind means the captured index now describes different paint.
func sameOpKinds(a *Display, aIndices []int, b *Display, bIndices []int) bool {
	if len(aIndices) != len(bIndices) {
		return false
	}
	for i := range aIndices {
		if a.Ops[aIndices[i]].Kind != b.Ops[bIndices[i]].Kind {
			return false
		}
	}
	return true
}

// readTransitionPaints snapshots the selected operations with the property's
// mask. A transform property always carries a matrix: an operation without a
// baked transform reads the identity, so a class that adds the first
// transform can interpolate from where the element stood.
func readTransitionPaints(d *Display, indices []int, mask cssmotion.Mask,
	matrix bool) []cssmotion.Paint {
	paints := make([]cssmotion.Paint, len(indices))
	for i, index := range indices {
		paint := cssmotion.Read(&d.Ops[index])
		paint.Mask = mask
		if matrix && !paint.HasMatrix {
			paint.Matrix, paint.HasMatrix = layout.IdentityMatrix(), true
		}
		paints[i] = paint
	}
	return paints
}

// paintSetChanged reports whether any selected operation's masked paint
// differs between the two captures.
func paintSetChanged(from, to []cssmotion.Paint, mask cssmotion.Mask) bool {
	for i := range from {
		if paintChanged(from[i], to[i], mask) {
			return true
		}
	}
	return false
}

// paintChanged compares the fields a property owns.
func paintChanged(a, b cssmotion.Paint, mask cssmotion.Mask) bool {
	if mask&cssmotion.MaskX != 0 && math.Abs(a.X-b.X) > transitionPaintEpsilon {
		return true
	}
	if mask&cssmotion.MaskY != 0 && math.Abs(a.Y-b.Y) > transitionPaintEpsilon {
		return true
	}
	if mask&cssmotion.MaskW != 0 && math.Abs(a.W-b.W) > transitionPaintEpsilon {
		return true
	}
	if mask&cssmotion.MaskH != 0 && math.Abs(a.H-b.H) > transitionPaintEpsilon {
		return true
	}
	if mask&cssmotion.MaskOpacity != 0 &&
		math.Abs(a.Opacity-b.Opacity) > transitionPaintEpsilon {
		return true
	}
	if mask&cssmotion.MaskRGB != 0 &&
		(math.Abs(a.R-b.R) > transitionPaintEpsilon ||
			math.Abs(a.G-b.G) > transitionPaintEpsilon ||
			math.Abs(a.B-b.B) > transitionPaintEpsilon) {
		return true
	}
	if mask&cssmotion.MaskAlpha != 0 && math.Abs(a.Alpha-b.Alpha) > transitionPaintEpsilon {
		return true
	}
	if mask&cssmotion.MaskMatrix != 0 {
		am, bm := transitionMatrix(a), transitionMatrix(b)
		return math.Abs(am.A-bm.A) > transitionPaintEpsilon ||
			math.Abs(am.B-bm.B) > transitionPaintEpsilon ||
			math.Abs(am.C-bm.C) > transitionPaintEpsilon ||
			math.Abs(am.D-bm.D) > transitionPaintEpsilon ||
			math.Abs(am.E-bm.E) > transitionPaintEpsilon ||
			math.Abs(am.F-bm.F) > transitionPaintEpsilon
	}
	return false
}

// transitionMatrix treats an absent matrix as the identity.
func transitionMatrix(paint cssmotion.Paint) layout.Matrix2D {
	if paint.HasMatrix {
		return paint.Matrix
	}
	return layout.IdentityMatrix()
}

// transitionIndices resolves the operations a property owns on one box.
func transitionIndices(ref boxRef, property string, idx *cssmotion.Index) ([]int, bool) {
	kind, _, ok := transitionPaintTarget(property)
	if !ok {
		return nil, false
	}
	return idx.Select(ref.box, ref.index, kind), true
}

// BindTransitions installs the captures as renderers and keeps every live
// entry the new display did not change. A capture replaces the entry for its
// element and property with a fresh renderer from the presented paint; an
// untouched property keeps its renderer, its clock, and its endpoints, and
// only its operation pointers move to the new list.
//
// A capture that targets the run's reversing-adjusted start value is a
// reversal: it starts from the presented paint and scales the declared
// duration (and a negative delay) by the reversing shortening factor. A third
// value starts from the presented paint with the full declared duration.
//
// A property that leaves the new style, an element that leaves the display, a
// target set whose operations changed, and a field an active animation owns
// are cancelled. The committed display already carries the new paint, so a
// cancel leaves that final value in place exactly once; the runtime never
// writes a stale destination over a fresh commit and never revives the entry.
func (r *Runtime) BindTransitions(next *Display, captures []TransitionCapture, now time.Time) {
	if r == nil {
		return
	}
	r.stats.TransitionsStarted = 0
	r.stats.TransitionsCancelled = 0
	r.stats.TransitionsSkipped = r.transitionCapture.skipped
	r.stats.TransitionsBlocked = r.transitionCapture.blocked
	r.stats.TransitionSkipReason = r.transitionCapture.reason
	r.stats.TransitionsNoID = r.transitionCapture.noID
	r.stats.TransitionsNoOld = r.transitionCapture.noOld
	r.stats.TransitionError = ""

	if next == nil {
		r.releaseTransitionRegistries()
		r.stats.TransitionsBound = 0
		r.stats.TransitionsActive = 0
		r.stats.TransitionsSampled = 0
		r.transitionCapture = transitionCaptureCounts{}

		return
	}
	if len(captures) == 0 && len(r.transitions) == 0 && len(r.geometry) == 0 &&
		len(r.discrete) == 0 {
		// The common page has no transition declarations. Do not build the
		// element and record maps on every redraw for it.
		r.stats.TransitionsBound = 0
		r.stats.TransitionsActive = 0
		r.stats.TransitionsSampled = 0
		r.transitionCapture = transitionCaptureCounts{}

		return
	}

	elements := indexBoxes(next.Boxes)
	records := make(map[entryKey]map[string]layout.Transition, len(elements))
	for key, ref := range elements {
		if len(ref.box.Transitions) != 0 {
			records[key] = transitionRecords(ref.box)
		}
	}
	captured := make(map[transitionKey]struct{}, len(captures))
	for i := range captures {
		captured[captures[i].key] = struct{}{}
	}

	// live keeps the entries a capture may reverse. The keep loop reuses the
	// registry slice, so the structs are copied before the first append.
	live := make(map[transitionKey]transitionEntry, len(r.transitions))
	for i := range r.transitions {
		live[r.transitions[i].key] = r.transitions[i]
	}
	liveGeometry := make(map[transitionKey]geometryEntry, len(r.geometry))
	for i := range r.geometry {
		liveGeometry[r.geometry[i].key] = r.geometry[i]
	}

	r.keepTransitions(next, elements, records, captured)

	for i := range captures {
		capture := captures[i]
		ref, ok := elements[capture.key.element]
		if !ok {
			continue
		}
		record, ok := records[capture.key.element][capture.key.property]
		if !ok {
			continue
		}
		if capture.cancel {
			// The keep loop already dropped the live entry. The committed
			// paint is the value the cancel leaves, written once by the
			// commit, so no renderer replaces it.
			continue
		}
		switch capture.kind {
		case transitionGeometry:
			r.bindGeometry(next, capture, record, ref, liveGeometry, now)
		case transitionDiscrete:
			r.bindDiscrete(next, capture, record, now)
		default:
			r.bindPaint(next, capture, record, ref, live, now)
		}
	}

	r.stats.TransitionsBound = len(r.transitions) + len(r.geometry) + len(r.discrete)
	if r.stats.TransitionsBound == 0 {
		r.stats.TransitionsActive = 0
		r.stats.TransitionsSampled = 0
	}
	r.transitionCapture = transitionCaptureCounts{}
}

// bindPaint installs one paint renderer, applying the reversing shortening
// factor when the capture returns to the start a live run came from.
func (r *Runtime) bindPaint(next *Display, capture TransitionCapture, record layout.Transition,
	ref boxRef, live map[transitionKey]transitionEntry, now time.Time) {
	mask := transitionMask(capture.key.property)
	if r.animationOwns(capture.key.element, mask) {
		// The animation wins the cascade for this field, so the capture
		// never becomes a renderer. Counting it keeps the precedence
		// visible in Probe().Motion. Nothing is stored, so a later commit
		// cannot revive it when the animation finishes.
		r.stats.TransitionsSkipped++
		return
	}
	entry := transitionEntry{
		key:             capture.key,
		indices:         capture.indices,
		kinds:           opKinds(next, capture.indices),
		reversingStart:  capture.from,
		reversingFactor: 1,
		to:              capture.to,
		record:          record,
		start:           now,
	}
	bindRecord := record
	if old, ok := live[capture.key]; ok && old.renderer != nil &&
		samePaintSet(capture.to, old.reversingStart, mask) {
		factor := reverseFactor(old.renderer.Progress(now), old.reversingFactor)
		bindRecord.DurationMS *= factor
		if bindRecord.DelayMS < 0 {
			// A negative delay starts the reverse part-way in, so it is
			// scaled with the duration. A positive delay keeps its length.
			bindRecord.DelayMS *= factor
		}
		entry.reversingStart = old.to
		entry.reversingFactor = factor
	}
	if bindRecord.DurationMS <= 0 {
		// A zero-duration run has no frame to sample. The committed display
		// already carries the destination, so releasing the old entry and
		// leaving that paint is the snap CSS asks for.
		return
	}
	bindBox := ref.box
	bindBox.Transitions = []layout.Transition{bindRecord}
	renderer, err := csstransition.BindTo(next, bindBox, capture.property,
		capture.indices, capture.from, capture.to, now)
	if err != nil {
		r.stats.TransitionError = err.Error()
		r.stats.TransitionsSkipped++
		return
	}
	entry.renderer = renderer
	entry.record = bindRecord
	r.transitions = append(r.transitions, entry)
	r.stats.TransitionsStarted++
}

// bindGeometry installs one layout-length run and queues its element for the
// next relayout pass.
func (r *Runtime) bindGeometry(next *Display, capture TransitionCapture, record layout.Transition,
	ref boxRef, live map[transitionKey]geometryEntry, now time.Time) {
	entry := geometryEntry{
		key:             capture.key,
		from:            capture.geometry.from,
		to:              capture.geometry.to,
		record:          record,
		start:           now,
		animated:        layoutAnimatedProperties(next, ref.box),
		reversingStart:  capture.geometry.from,
		reversingFactor: 1,
	}
	if old, ok := live[capture.key]; ok &&
		math.Abs(old.reversingStart-capture.geometry.to) <= transitionPaintEpsilon {
		progress, _ := transition.ProgressFor(old.record, old.start, now)
		factor := reverseFactor(progress, old.reversingFactor)
		entry.record.DurationMS *= factor
		if entry.record.DelayMS < 0 {
			entry.record.DelayMS *= factor
		}
		entry.reversingStart = old.to
		entry.reversingFactor = factor
	}
	if entry.record.DurationMS <= 0 {
		return
	}
	r.geometry = append(r.geometry, entry)
	r.layout.enqueue(entry.key.element.id)
	r.stats.TransitionsStarted++
}

// bindDiscrete installs one non-interpolated run. The committed display
// carries the destination side, so the from side is written now; the first
// drawn frame then shows the value the run starts from.
func (r *Runtime) bindDiscrete(next *Display, capture TransitionCapture, record layout.Transition,
	now time.Time) {
	if capture.discrete == nil || capture.discrete.rule == nil {
		return
	}
	if record.Behavior != "allow-discrete" {
		// Without allow-discrete a discrete change snaps at the delay end. The
		// committed display already carries it, so there is nothing to bind.
		return
	}
	if record.DurationMS <= 0 {
		return
	}
	entry := discreteEntry{
		key:       capture.key,
		rule:      capture.discrete.rule,
		targetKey: capture.discrete.targetKey,
		record:    record,
		start:     now,
	}
	entry.rule.Step(next, 0)
	entry.wrote = true
	r.discrete = append(r.discrete, entry)
	r.stats.TransitionsStarted++
}

// AdoptTransitions points the live registries at the display a relayout patch
// produced. The patch already carries the sampled values, so no capture runs
// and the clocks, endpoints, and records are kept; only the target operations,
// boxes, and paint order move.
func (r *Runtime) AdoptTransitions(next *Display) {
	if r == nil || next == nil {
		return
	}
	r.keepTransitions(next, indexBoxes(next.Boxes), nil, nil)
	r.stats.TransitionsBound = len(r.transitions) + len(r.geometry) + len(r.discrete)
	if r.stats.TransitionsBound == 0 {
		r.stats.TransitionsActive = 0
		r.stats.TransitionsSampled = 0
	}
}

// keepTransitions drops the live entries the new display no longer supports and
// repoints the rest at it. A nil records map skips the declaration check, which
// is the relayout adoption path: a geometry patch never changes the
// declarations. A nil captured map keeps every entry whose targets still
// resolve.
func (r *Runtime) keepTransitions(next *Display, elements map[entryKey]boxRef,
	records map[entryKey]map[string]layout.Transition, captured map[transitionKey]struct{}) {
	if len(r.transitions) == 0 && len(r.geometry) == 0 && len(r.discrete) == 0 {
		return
	}
	index := cssmotion.NewIndex(next)
	kept := r.transitions[:0]
	for i := range r.transitions {
		entry := r.transitions[i]
		ref, ok := elements[entry.key.element]
		if !ok {
			r.stats.TransitionsCancelled++
			continue
		}
		if records != nil {
			if _, ok := records[entry.key.element][entry.key.property]; !ok {
				r.stats.TransitionsCancelled++
				continue
			}
		}
		if captured != nil {
			if _, ok := captured[entry.key]; ok {
				r.stats.TransitionsCancelled++
				continue
			}
		}
		if r.animationOwns(entry.key.element, transitionMask(entry.key.property)) {
			// An animation writes this field later in the same frame. Drop the
			// run now so it cannot reappear when the animation ends.
			r.stats.TransitionsCancelled++
			continue
		}
		indices, ok := transitionIndices(ref, entry.key.property, index)
		if !ok || len(indices) != len(entry.indices) || !sameKindList(next, indices, entry.kinds) {
			r.stats.TransitionsCancelled++
			continue
		}
		ops := cssmotion.Resolve(next, indices)
		if len(ops) != len(indices) || entry.renderer.Retarget(ops) != nil {
			r.stats.TransitionsCancelled++
			continue
		}
		entry.indices = indices
		kept = append(kept, entry)
	}
	r.transitions = kept

	keptGeometry := r.geometry[:0]
	for i := range r.geometry {
		entry := r.geometry[i]
		ref, ok := elements[entry.key.element]
		if !ok {
			r.stats.TransitionsCancelled++
			continue
		}
		if records != nil {
			if _, ok := records[entry.key.element][entry.key.property]; !ok {
				r.stats.TransitionsCancelled++
				continue
			}
		}
		if captured != nil {
			if _, ok := captured[entry.key]; ok {
				r.stats.TransitionsCancelled++
				continue
			}
		}
		if layoutAnimationOwns(layoutAnimatedProperties(next, ref.box), entry.key.property) {
			r.stats.TransitionsCancelled++
			continue
		}
		entry.animated = layoutAnimatedProperties(next, ref.box)
		keptGeometry = append(keptGeometry, entry)
	}
	r.geometry = keptGeometry

	keptDiscrete := r.discrete[:0]
	for i := range r.discrete {
		entry := r.discrete[i]
		ref, ok := elements[entry.key.element]
		if !ok {
			r.stats.TransitionsCancelled++
			continue
		}
		if records != nil {
			if _, ok := records[entry.key.element][entry.key.property]; !ok {
				r.stats.TransitionsCancelled++
				continue
			}
		}
		if captured != nil {
			if _, ok := captured[entry.key]; ok {
				r.stats.TransitionsCancelled++
				continue
			}
		}
		ops := cssmotion.Resolve(next, index.Select(ref.box, ref.index, cssmotion.KindSubtree))
		if entry.rule.Retarget(next, ops) != nil {
			r.stats.TransitionsCancelled++
			continue
		}
		entry.rule.Reapply(next)
		keptDiscrete = append(keptDiscrete, entry)
	}
	r.discrete = keptDiscrete
}

// sameKindList compares the resolved operation kinds with the kinds an entry
// was bound to.
func sameKindList(d *Display, indices []int, kinds []uint8) bool {
	if len(indices) != len(kinds) {
		return false
	}
	for i, index := range indices {
		if uint8(d.Ops[index].Kind) != kinds[i] {
			return false
		}
	}
	return true
}

// opKinds snapshots the operation kinds a renderer writes.
func opKinds(d *Display, indices []int) []uint8 {
	kinds := make([]uint8, len(indices))
	for i, index := range indices {
		kinds[i] = uint8(d.Ops[index].Kind)
	}
	return kinds
}

// ReleaseTransitions drops every live renderer without writing. The operations
// keep the value the last sample left, so turning the switch off freezes a
// transition at the presented paint instead of snapping to its base or its
// destination.
func (r *Runtime) ReleaseTransitions() {
	if r == nil {
		return
	}
	r.releaseTransitionRegistries()
	r.transitionCapture = transitionCaptureCounts{}
	r.stats.TransitionsBound = 0
	r.stats.TransitionsActive = 0
	r.stats.TransitionsSampled = 0
	r.stats.TransitionsStarted = 0
	r.stats.TransitionsCancelled = 0
	r.stats.TransitionsSkipped = 0
	r.stats.TransitionsBlocked = 0
	r.stats.TransitionSkipReason = ""
	r.stats.TransitionsNoID = 0
	r.stats.TransitionsNoOld = 0
	r.stats.TransitionError = ""
}

// releaseTransitionRegistries drops the live entries of all three transition
// kinds without writing. Geometry runs leave the last patched box on screen and
// discrete runs leave the endpoint they reached, which is the freeze rule the
// switch documents.
func (r *Runtime) releaseTransitionRegistries() {
	r.transitions = r.transitions[:0]
	r.geometry = r.geometry[:0]
	r.discrete = r.discrete[:0]
}

// StepTransitions samples the transition registry alone. The page calls it
// when the animation driver is off but transitions are on, so the two
// switches stay independent. Geometry runs are sampled by StepLayout, because
// they share the relayout budget with layout keyframes.
func (r *Runtime) StepTransitions(d *Display, now time.Time) Damage {
	var damage Damage
	if r == nil {
		return damage
	}
	r.stats.TransitionsActive, r.stats.TransitionsSampled = 0, 0
	if len(r.transitions) == 0 && len(r.discrete) == 0 {
		return damage
	}
	if d == nil {
		r.ReleaseTransitions()
		return damage
	}
	return r.stepTransitions(d, now)
}

// stepTransitions samples every entry in order and removes the ones that
// finished. It runs before the animation pass, so an animation on the same
// field writes second and wins.
func (r *Runtime) stepTransitions(d *Display, now time.Time) Damage {
	var damage Damage
	kept := r.transitions[:0]
	for i := range r.transitions {
		entry := r.transitions[i]
		if entry.renderer == nil || !indicesInRange(d, entry.indices) {
			r.stats.TransitionsCancelled++
			continue
		}
		if r.animationOwns(entry.key.element, transitionMask(entry.key.property)) {
			// The animation pass writes this field second. Cancel the run
			// rather than sampling a value the animation overwrites, and never
			// bring it back when the animation ends.
			r.stats.TransitionsCancelled++
			continue
		}
		if now.Before(transitionDelayEnd(&entry)) && entry.wrote {
			// The delayed from paint is already on screen. Do not rewrite it
			// every frame; NextFrame asks for the delay end instead.
			entry.sampled, entry.active = true, true
			r.stats.TransitionsActive++
			kept = append(kept, entry)
			continue
		}
		before := transitionBounds(d, entry.indices)
		active := entry.renderer.Step(now)
		entry.sampled, entry.wrote, entry.active = true, true, active
		r.stats.TransitionsSampled++
		after := transitionBounds(d, entry.indices)
		damage.Add(before.Union(after))
		if active {
			r.stats.TransitionsActive++
			kept = append(kept, entry)
		}
	}
	r.transitions = kept

	keptDiscrete := r.discrete[:0]
	for i := range r.discrete {
		entry := r.discrete[i]
		if entry.rule == nil {
			r.stats.TransitionsCancelled++
			continue
		}
		if entry.wrote && now.Before(entry.start.Add(milliseconds(entry.record.DelayMS))) {
			entry.sampled, entry.active = true, true
			r.stats.TransitionsActive++
			keptDiscrete = append(keptDiscrete, entry)
			continue
		}
		before := discreteBounds(d, entry.rule)
		progress, active := transition.ProgressFor(entry.record, entry.start, now)
		changed := entry.rule.Step(d, progress)
		entry.sampled, entry.wrote, entry.active = true, true, active
		r.stats.TransitionsSampled++
		if changed {
			damage.Add(before.Union(discreteBounds(d, entry.rule)))
		}
		if active {
			r.stats.TransitionsActive++
			keptDiscrete = append(keptDiscrete, entry)
		}
	}
	r.discrete = keptDiscrete

	return damage
}

// indicesInRange reports that every captured index still resolves.
func indicesInRange(d *Display, indices []int) bool {
	for _, index := range indices {
		if index < 0 || index >= len(d.Ops) {
			return false
		}
	}
	return true
}

// transitionBounds is the union of the entry's painted operation bounds.
func transitionBounds(d *Display, indices []int) (box image.Rectangle) {
	for _, index := range indices {
		if index < 0 || index >= len(d.Ops) {
			continue
		}
		box = box.Union(dirty.OpBounds(&d.Ops[index], d))
	}
	return box
}

// transitionDelayEnd is the clock at which a delayed transition starts
// moving.
func transitionDelayEnd(entry *transitionEntry) time.Time {
	return entry.start.Add(milliseconds(entry.record.DelayMS))
}

// milliseconds converts a CSS millisecond count to a duration.
func milliseconds(value float64) time.Duration {
	return time.Duration(value * float64(time.Millisecond))
}
