package motion

import (
	"errors"
	"fmt"
	"github.com/chinmay-sawant/blinkless/layout"
	"sort"
	"strconv"
	"strings"
	"time"
)

// defaultLayoutBudget is the number of element relayouts one frame may apply.
// A relayout is tens to hundreds of microseconds, and a page can declare many
// geometry animations, so the runtime drains one element per frame and queues
// the rest. The engine pass lives on the page, so the budget is enforced here,
// where the queue is owned.
const defaultLayoutBudget = 1

// LayoutOverride is one layout-tier property value for the engine's
// RelayoutElementWithOverrides call. Property is a CSS property name and Value
// is a CSS value the engine accepts.
type LayoutOverride struct {
	Property string
	Value    string
}

// LayoutRequest is one element's sampled layout-tier values for a frame. Every
// layout property the element animates is in one request, so a card that
// animates width and height reaches the engine's relayout path once.
type LayoutRequest struct {
	ElementID string
	Overrides []LayoutOverride
}

// layoutState is the runtime's layout-tier driver state: the element queue, the
// per-frame budget, and the buffers one sample reuses. A page with no layout
// keyframes leaves it empty and pays nothing.
type layoutState struct {
	// queue is the element ids waiting for a relayout, in the order they will
	// run. An id stays in the queue while its animation is live, so a busy
	// element shares the budget with its neighbours instead of starving them.
	queue []string
	// queued mirrors queue for the membership test. A rebind rebuilds the
	// entries but keeps the queue, so a queued element is not lost.
	queued map[string]struct{}
	// budget overrides defaultLayoutBudget when a caller sets it. Zero means
	// the default.
	budget int
	// stamp counts StepLayout calls. served records the call that last sampled
	// an element, so the reported queue depth counts what actually waited.
	stamp  uint64
	served map[string]uint64
	// requests is the slice StepLayout returns, reused by the next call.
	requests []LayoutRequest
	// values is the sample buffer one element uses. It is handed back to
	// animation.LayoutAt, which appends the sampled values to it.
	values []LayoutValue
}

// reset drops every queued element and the sampled buffers. A bind that
// releases the runtime calls it, so a page with animations off pays nothing.
func (l *layoutState) reset() {
	l.queue = l.queue[:0]
	clear(l.queued)
	clear(l.served)
	l.stamp = 0
	l.requests = l.requests[:0]
	l.values = l.values[:0]
}

// enqueue adds an element to the tail of the queue when it is not already
// queued.
func (l *layoutState) enqueue(id string) {
	if id == "" || l.has(id) {
		return
	}
	if l.queued == nil {
		l.queued = make(map[string]struct{})
	}
	l.queue = append(l.queue, id)
	l.queued[id] = struct{}{}
}

// enqueueFront puts an element back at the head of the queue after a frame that
// could not apply its sample, so the element runs first on the next frame. An
// element that was requeued at the tail is moved, not duplicated.
func (l *layoutState) enqueueFront(id string) {
	if id == "" {
		return
	}
	l.remove(id)
	if l.queued == nil {
		l.queued = make(map[string]struct{})
	}
	l.queue = append(l.queue, "")
	copy(l.queue[1:], l.queue)
	l.queue[0] = id
	l.queued[id] = struct{}{}
}

// remove drops an id from the queue when it is present.
func (l *layoutState) remove(id string) {
	for i, candidate := range l.queue {
		if candidate == id {
			l.queue = append(l.queue[:i], l.queue[i+1:]...)

			break
		}
	}
	delete(l.queued, id)
}

// take removes the head of the queue.
func (l *layoutState) take() (string, bool) {
	if len(l.queue) == 0 {
		return "", false
	}
	id := l.queue[0]
	l.queue = l.queue[1:]
	delete(l.queued, id)

	return id, true
}

// has reports whether an element is queued.
func (l *layoutState) has(id string) bool {
	_, ok := l.queued[id]

	return ok
}

// queuedDepth returns the number of queued elements this frame did not sample.
// A requeued element that ran this frame is not waiting; a neighbour that did
// not run is.
func (l *layoutState) queuedDepth() int {
	depth := 0
	for _, id := range l.queue {
		if id == "" {
			continue
		}
		if l.stamp == 0 || l.served[id] != l.stamp {
			depth++
		}
	}

	return depth
}

// syncLayoutQueue adds the layout entries a bind introduced. An entry that was
// sampled before the rebind keeps its place in the queue, so a page that
// republishes mid-animation does not restart the rotation.
func (r *Runtime) syncLayoutQueue() {
	var missing []string
	bound := 0
	for i := range r.entries {
		entry := &r.entries[i]
		if !entry.layout || entry.key.id == "" {
			continue
		}
		bound++
		if entry.sampled || r.layout.has(entry.key.id) {
			continue
		}
		missing = append(missing, entry.key.id)
	}

	sort.Strings(missing)
	for i := 0; i < len(missing); i++ {
		if i > 0 && missing[i] == missing[i-1] {
			continue
		}
		r.layout.enqueue(missing[i])
	}
	r.stats.LayoutBound = bound
	r.stats.LayoutQueued = len(r.layout.queue)
}

// layoutEntry returns the entry that samples one element's layout keyframes.
func (r *Runtime) layoutEntry(id string) *entry {
	if id == "" {
		return nil
	}
	for i := range r.entries {
		entry := &r.entries[i]
		if entry.layout && entry.key.id == id {
			return entry
		}
	}

	return nil
}

// StepLayout samples the layout-tier animations and geometry transitions that
// are due this frame and returns at most the per-frame budget as engine
// override requests. Every property one element writes is in one request.
//
// An element that does not fit the budget stays queued, and it samples at the
// frame that runs it rather than at the frame it was queued, so its value
// matches the clock of the frame that applies it. A paused element samples once
// and leaves the queue; a live element goes to the back, so the budget rotates
// through the page instead of starving the later ids. A geometry transition
// joins the queue at bind and every live entry is re-added from the registry
// before sampling, so it never spins the budget loop on a sub-pixel sample.
func (r *Runtime) StepLayout(now time.Time) []LayoutRequest {
	if r == nil {
		return nil
	}

	r.layout.requests = r.layout.requests[:0]
	r.stats.LayoutRelayouts = 0
	r.layout.stamp++
	r.pruneGeometry()
	r.syncGeometryQueue()

	budget := r.layout.budget
	if budget <= 0 {
		budget = defaultLayoutBudget
	}
	attempts := len(r.layout.queue)
	for budget > 0 && attempts > 0 {
		id, ok := r.layout.take()
		if !ok {
			break
		}
		attempts--

		overrides := r.geometryOverrides(id, now, nil)
		values, animationLive := r.sampleLayoutAnimation(id, now)
		overrides = mergeLayoutOverrides(overrides, values)
		if len(overrides) == 0 {
			if animationLive {
				r.layout.enqueue(id)
			}

			continue
		}

		if r.layout.served == nil {
			r.layout.served = make(map[string]uint64)
		}
		r.layout.served[id] = r.layout.stamp
		r.layout.requests = append(r.layout.requests,
			LayoutRequest{ElementID: id, Overrides: overrides})
		budget--
		if animationLive {
			r.layout.enqueue(id)
		}
	}

	r.stats.LayoutQueued = r.layout.queuedDepth()
	r.stats.LayoutBudgetHit = r.stats.LayoutQueued > 0

	return r.layout.requests
}

// sampleLayoutAnimation samples one element's layout keyframe player and
// reports whether it still has work. The transition geometry pass calls it
// after its own sample so one element reaches the engine once, and the
// animation value wins when both write the same property.
func (r *Runtime) sampleLayoutAnimation(id string, now time.Time) ([]LayoutValue, bool) {
	entry := r.layoutEntry(id)
	if entry == nil {
		return nil, false
	}
	if entry.sampled && entry.player.Paused() {
		// The frozen pose was applied on the frame it was sampled.
		return nil, false
	}

	input, _ := r.input(entry, now)
	values, live := entry.player.LayoutAt(input, r.layout.values[:0])
	r.layout.values = values[:0]
	entry.sampled, entry.sampledAt = true, now

	return values, live && !entry.player.Paused()
}

// mergeLayoutOverrides appends the animation values a transition does not
// already own. An animation wins the cascade on the same property, so its
// value replaces the transition's when the two collide.
func mergeLayoutOverrides(overrides []LayoutOverride,
	values []LayoutValue) []LayoutOverride {
	for i := range values {
		property := values[i].Property
		value := values[i].CSS()
		replaced := false
		for j := range overrides {
			if overrides[j].Property == property {
				overrides[j].Value = value
				replaced = true

				break
			}
		}
		if !replaced {
			overrides = append(overrides, LayoutOverride{Property: property, Value: value})
		}
	}

	return overrides
}

// ReturnLayout puts the elements of a sample back at the head of the queue. The
// page calls it when the display it would patch does not carry retained layout
// state yet, so the element samples again on the frame that can apply it.
func (r *Runtime) ReturnLayout(requests []LayoutRequest) {
	if r == nil {
		return
	}
	for i := len(requests) - 1; i >= 0; i-- {
		id := requests[i].ElementID
		if id == "" {
			continue
		}
		if entry := r.layoutEntry(id); entry != nil {
			entry.sampled = false
		}
		r.returnGeometry(id)
		delete(r.layout.served, id)
		r.layout.enqueueFront(id)
	}
	r.stats.LayoutQueued = r.layout.queuedDepth()
}

// RequeueLayout marks every bound layout animation due for another sample. A
// publish rebuilds the display from the static styles, so a paused or finished
// layout animation has to run again to put its pose on the new operations; the
// next StepLayout drains the queue under the same budget.
func (r *Runtime) RequeueLayout() {
	if r == nil {
		return
	}
	for i := range r.entries {
		entry := &r.entries[i]
		if !entry.layout || entry.key.id == "" {
			continue
		}
		entry.sampled = false
		r.layout.enqueue(entry.key.id)
	}
	r.stats.LayoutQueued = len(r.layout.queue)
}

// NoteLayout records the outcome of the page's relayout pass. applied is the
// number of elements the engine patched; a failure keeps the display the page
// already draws and is reported through Stats.LayoutError.
func (r *Runtime) NoteLayout(applied int, err error) {
	if r == nil {
		return
	}

	r.stats.LayoutRelayouts = applied
	if err != nil {
		r.stats.LayoutError = err.Error()

		return
	}

	r.stats.LayoutError = ""
}

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
func LayoutKeyframes(rules []layout.KeyframesRule, name string) ([]Keyframe, error) {
	source := ruleFor(rules, name)
	frames := make([]Keyframe, len(source))
	for i, item := range source {
		frame := Keyframe{Offset: item.Offset}
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
func BindLayout(d *layout.Display, box layout.Box, start time.Time) (*Player, error) {
	if d == nil || len(box.Animations) == 0 {
		return nil, nil
	}

	effects := make([]Effect, 0, len(box.Animations))
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
		effects = append(effects, Effect{
			Frames: frames, Timing: timing(item), Trigger: item.Trigger,
		})
	}
	if len(effects) == 0 {
		return nil, nil
	}

	player, err := NewLayout(effects)
	if err != nil {
		return nil, err
	}
	player.Start(start)

	return player, nil
}

// hasLayoutFrames reports whether a parsed rule declares a layout-tier
// property.
func hasLayoutFrames(frames []Keyframe) bool {
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

// layoutGrammar maps a property to its value grammar and maximum arity. The
// tier is one flat property table.
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
func parseLayoutValue(property, raw string) (LayoutValue, error) {
	kind, arity, ok := layoutGrammar(property)
	if !ok {
		return LayoutValue{}, layoutErr("%q is not a layout-tier property", property)
	}

	terms, err := parseLayoutTerms(kind, arity, raw)
	if err != nil {
		return LayoutValue{}, layoutErr("%s: %s", property, err.Error())
	}

	return LayoutValue{Property: property, Terms: terms}, nil
}

// parseLayoutTerms splits one declaration into its terms. A radius declaration
// carries an optional slash, so its two halves parse with the same grammar and
// keep the separator as a keyword term.
func parseLayoutTerms(kind layoutKind, arity layoutArity, value string) ([]LayoutTerm, error) {
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

	terms := make([]LayoutTerm, len(fields))
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
func parseRadiusTerms(value string) ([]LayoutTerm, error) {
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

	return append(append(terms, LayoutTerm{Keyword: "/"}), verticalTerms...), nil
}

// parseLayoutTerm parses one whitespace-separated term: a number with a unit,
// a unitless number where the grammar allows one, or a keyword.
func parseLayoutTerm(kind layoutKind, token string) (LayoutTerm, error) {
	number, unit, numeric := splitLayoutNumber(token)
	if !numeric {
		keyword, ok := layoutKeyword(kind, token)
		if !ok {
			return LayoutTerm{}, layoutErr("%q is not a valid value", token)
		}

		return LayoutTerm{Keyword: keyword}, nil
	}

	return layoutNumericTerm(kind, number, unit, token)
}

// layoutNumericTerm validates a numeric term against its property grammar and
// resolves an absolute unit to px.
func layoutNumericTerm(kind layoutKind, number float64, unit, token string) (LayoutTerm, error) {
	unit = strings.ToLower(unit)
	if unit == "" && kind != layoutFlexFactor && kind != layoutLineHeight {
		if number != 0 {
			return LayoutTerm{}, layoutErr("%q needs a unit", token)
		}
		unit = "px"
	}

	switch kind {
	case layoutFlexFactor:
		if unit != "" {
			return LayoutTerm{}, layoutErr("%q must be a unitless number", token)
		}
		if number < 0 {
			return LayoutTerm{}, layoutErr("%q must not be negative", token)
		}

		return LayoutTerm{Number: number}, nil
	case layoutLineHeight:
		if unit == "" {
			if number < 0 {
				return LayoutTerm{}, layoutErr("%q must not be negative", token)
			}

			return LayoutTerm{Number: number}, nil
		}
		if !layoutAllowsPercent(kind) && unit == "%" {
			return LayoutTerm{}, layoutErr("%q does not take a percentage", token)
		}
	}

	switch unit {
	case "%":
		if !layoutAllowsPercent(kind) {
			return LayoutTerm{}, layoutErr("%q does not take a percentage", token)
		}
	case "px", "pt", "pc", "in", "cm", "mm":
		number, unit = resolveAbsoluteLength(number, unit), "px"
	case "em", "rem", "ex", "ch", "vw", "vh", "vmin", "vmax":
	case "":
		// A zero length carried no unit; it was normalized to px above.
	default:
		return LayoutTerm{}, layoutErr("%q has an unsupported unit", token)
	}

	if number < 0 && !layoutAllowsNegative(kind) {
		return LayoutTerm{}, layoutErr("%q must not be negative", token)
	}

	return LayoutTerm{Number: number, Unit: unit}, nil
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
