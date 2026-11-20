package motion

import (
	"errors"
	"fmt"
	"github.com/chinmay-sawant/blinkless/layout"
	"math"
	"strconv"
	"strings"
)

// Inline-SVG paint requests.
//
// An inline SVG is serialized and rasterized by the engine, so fill and stroke
// are not display fields. A player built for that paint samples the values per
// element (animation.SVGPaintValues); the runtime turns the sample into one
// layout.SVGPaintOverride and the page restyles the raster through
// layout.RestyleInlineSVG. Nothing here writes an operation.

// defaultSVGBudget is the number of SVG restyles one frame may apply. A restyle
// serializes and rasterizes the element, so a page with many animated icons
// drains one per frame instead of paying for all of them at once.
const defaultSVGBudget = 1

// svgChannelMax is the channel scale of a CSS rgb() colour.
const svgChannelMax = 255

// SVGPaintRequest is one element's inline-SVG paint for one frame, ready for
// layout.RestyleInlineSVG. Overrides holds one root override; the engine parses
// the colour and dash strings, so the runtime never re-implements them.
//
// The identity fields stay unexported: a page hands the whole request back to
// NoteSVGPaint, which is how the runtime remembers the paint it has already
// applied.
type SVGPaintRequest struct {
	ElementID string
	Overrides []layout.SVGPaintOverride

	key       entryKey
	signature string
}

// svgState is the runtime's inline-SVG driver state: the paint already applied
// per element, the requests the latest Step collected, and the one-per-frame
// budget. A page with no SVG paint leaves it empty and pays nothing.
type svgState struct {
	// applied keys the serialized paint last applied for an element, so a
	// steady, paused, or finished sample does not restyle the same raster
	// every tick. A bind clears it: a publish rebuilds the display from the
	// declared paint.
	applied map[entryKey]string
	// requests is the latest Step's collected paint, in bind order. Step
	// rebuilds it every frame.
	requests []SVGPaintRequest
	// taken is the slice TakeSVGPaint hands the page, reused between calls.
	taken []SVGPaintRequest
	// budget overrides defaultSVGBudget in tests. Zero means the default.
	budget int
	// next rotates the start of the next take, so an element whose paint
	// changes every frame shares the budget with its neighbours instead of
	// starving the later ones.
	next int
	// keepBind asks the next BindAfter to preserve the applied signatures.
	// The page sets it when it adopts an engine patch that already carries
	// the restyled raster, so the bind does not make the element request the
	// same paint again.
	keepBind bool
}

// forget drops the applied signatures and the pending requests. BindAfter calls
// it for every bind the page runs, so a publish that rebuilt the display from
// the declared paint requests another restyle. A page that adopts an engine
// patch calls NoteSVGPaint after the bind, which records the paint it just
// installed; the in-flight take is kept, because a bind happens inside that
// adoption.
func (s *svgState) forget() {
	clear(s.applied)
	s.requests = s.requests[:0]
}

// MarkSVGPaintKept asks the next BindAfter to keep the applied signatures. The
// page calls it when it binds a display the engine patched from an applied
// restyle, because that display already carries the raster and a forgotten
// signature would restyle the same paint again on the next tick.
func (r *Runtime) MarkSVGPaintKept() {
	if r == nil {
		return
	}
	r.svg.keepBind = true
}

// reset drops the driver state for a released runtime.
func (s *svgState) reset() {
	s.forget()
	s.taken = s.taken[:0]
	s.budget = 0
	s.next = 0
	s.keepBind = false
}

// collectSVGPaint turns one sampled player into an element request. It runs
// after the player's sample, so the value belongs to the frame being built.
// A paint server is counted and leaves the base paint alone; a sample with no
// restylable property produces no request.
func (r *Runtime) collectSVGPaint(e *entry) {
	values := e.player.SVGPaintValues()
	if len(values) == 0 {
		return
	}
	// Every target of one player carries the same sample; the element only
	// needs one override, not one per operation.
	value := values[0]
	if value.Deferred != "" {
		r.stats.SVGPaintDeferred++
		r.stats.SVGPaintDeferredReason = value.Deferred
	}
	override, ok := svgPaintOverride(value, e.key.id)
	if !ok {
		// Only a deferred declaration: the element keeps its base paint, and
		// a restyle this element requested earlier must not be skipped when a
		// usable sample arrives later.
		delete(r.svg.applied, e.key)

		return
	}
	signature := svgOverrideSignature(override)
	if r.svg.applied[e.key] == signature {
		return
	}
	r.svg.requests = append(r.svg.requests, SVGPaintRequest{
		ElementID: e.key.id,
		Overrides: []layout.SVGPaintOverride{override},
		key:       e.key,
		signature: signature,
	})
	r.stats.SVGPaintRequests++
}

// TakeSVGPaint returns at most the per-frame budget of the inline-SVG restyles
// the last Step collected. A page calls it once per tick and passes the result
// to layout.RestyleInlineSVG. Requests past the budget are not lost: since the
// applied signature is recorded only by NoteSVGPaint, the next Step collects
// them again.
//
// The slice belongs to the runtime and is reused by the next call.
func (r *Runtime) TakeSVGPaint() []SVGPaintRequest {
	if r == nil || len(r.svg.requests) == 0 {
		return nil
	}
	n := len(r.svg.requests)
	budget := r.svg.budget
	if budget <= 0 {
		budget = defaultSVGBudget
	}
	take := min(budget, n)
	start := 0
	if r.svg.next > 0 && r.svg.next < n {
		start = r.svg.next
	}
	if start+take > n {
		start = 0
	}

	r.svg.taken = r.svg.taken[:0]
	for i := 0; i < take; i++ {
		r.svg.taken = append(r.svg.taken, r.svg.requests[(start+i)%n])
	}
	r.svg.next = (start + take) % n

	r.stats.SVGPaintQueued = n - take
	r.stats.SVGPaintBudgetHit = take < n

	return r.svg.taken
}

// NoteSVGPaint records the outcome of the page's restyle pass. applied is the
// number of leading requests from TakeSVGPaint the page installed; a failure
// keeps the display the page already draws and is reported through
// Stats.SVGPaintError, and the request is collected again on the next tick.
func (r *Runtime) NoteSVGPaint(applied int, err error) {
	if r == nil {
		return
	}
	applied = max(0, min(applied, len(r.svg.taken)))
	if applied > 0 && r.svg.applied == nil {
		r.svg.applied = make(map[entryKey]string)
	}
	for i := 0; i < applied; i++ {
		request := r.svg.taken[i]
		r.svg.applied[request.key] = request.signature
	}
	r.stats.SVGPaintRestyles = applied
	if err != nil {
		r.stats.SVGPaintError = err.Error()

		return
	}
	r.stats.SVGPaintError = ""
}

// svgPaintOverride converts one sampled value into the engine's string form.
// It reports false when the sample sets no property, which is the fully
// deferred paint-server case. none is written as the CSS keyword, never as an
// empty colour, and a colour with its own alpha becomes rgba().
func svgPaintOverride(v SVGPaintValue, elementID string) (layout.SVGPaintOverride, bool) {
	override := layout.SVGPaintOverride{ElementID: elementID}
	if v.HasFill {
		override.HasFill, override.Fill = true, svgPaintColor(v.Fill)
	}
	if v.HasFillOpacity {
		override.HasFillOpacity, override.FillOpacity = true, svgClamp01(v.FillOpacity)
	}
	if v.HasStroke {
		override.HasStroke, override.Stroke = true, svgPaintColor(v.Stroke)
	}
	if v.HasStrokeOpacity {
		override.HasStrokeOpacity, override.StrokeOpacity = true, svgClamp01(v.StrokeOpacity)
	}
	if v.HasStrokeWidth {
		override.HasStrokeWidth, override.StrokeWidth = true, v.StrokeWidth
	}
	if v.HasStrokeDashArray {
		override.HasStrokeDashArray = true
		override.StrokeDashArray = svgDashList(v.StrokeDashArray)
	}
	if v.HasStrokeDashOffset {
		override.HasStrokeDashOffset, override.StrokeDashOffset = true, v.StrokeDashOffset
	}
	if !override.HasFill && !override.HasFillOpacity && !override.HasStroke &&
		!override.HasStrokeOpacity && !override.HasStrokeWidth &&
		!override.HasStrokeDashArray && !override.HasStrokeDashOffset {
		return layout.SVGPaintOverride{}, false
	}

	return override, true
}

// svgPaintColor writes one sampled colour. none keeps the CSS keyword, an
// opaque colour is rgb(), and a translucent colour keeps its alpha in rgba()
// even though fill-opacity is a separate property.
func svgPaintColor(c SVGPaintColor) string {
	if c.None {
		return "none"
	}
	red, green, blue := svgChannel(c.R), svgChannel(c.G), svgChannel(c.B)
	if c.A < 1 {
		return fmt.Sprintf("rgba(%d, %d, %d, %s)", red, green, blue, svgNumber(svgClamp01(c.A)))
	}

	return fmt.Sprintf("rgb(%d, %d, %d)", red, green, blue)
}

// svgChannel rounds one sampled channel to the 0..255 scale CSS writes.
func svgChannel(value float64) int {
	if math.IsNaN(value) {
		return 0
	}

	return int(math.Round(svgClamp01(value) * svgChannelMax))
}

// svgClamp01 folds a sampled opacity or channel into the CSS range. NaN reads
// as the low end, the same clamp the renderer applies to a missing value.
func svgClamp01(value float64) float64 {
	if math.IsNaN(value) || value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}

	return value
}

// svgDashList writes a sampled dash list as CSS lengths in points. An empty
// list is the CSS none keyword, not an empty string, so the engine clears the
// authored dashes instead of rejecting the value.
func svgDashList(values []float64) string {
	if len(values) == 0 {
		return "none"
	}
	parts := make([]string, len(values))
	for i := range values {
		parts[i] = svgNumber(values[i]) + "pt"
	}

	return strings.Join(parts, " ")
}

// svgNumber formats one used value with no trailing zeros, so the same sample
// always produces the same string and the applied-signature check is stable.
func svgNumber(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

// svgOverrideSignature is the serialized identity of one override, used to
// skip a restyle whose paint has not changed. It covers exactly the fields the
// Has flags mark, so an override that leaves a property alone keeps the same
// signature as a sample that never declared it.
func svgOverrideSignature(o layout.SVGPaintOverride) string {
	var b strings.Builder
	if o.HasFill {
		b.WriteString("fill=")
		b.WriteString(o.Fill)
		b.WriteByte(';')
	}
	if o.HasFillOpacity {
		b.WriteString("fill-opacity=")
		b.WriteString(svgNumber(o.FillOpacity))
		b.WriteByte(';')
	}
	if o.HasStroke {
		b.WriteString("stroke=")
		b.WriteString(o.Stroke)
		b.WriteByte(';')
	}
	if o.HasStrokeOpacity {
		b.WriteString("stroke-opacity=")
		b.WriteString(svgNumber(o.StrokeOpacity))
		b.WriteByte(';')
	}
	if o.HasStrokeWidth {
		b.WriteString("stroke-width=")
		b.WriteString(svgNumber(o.StrokeWidth))
		b.WriteByte(';')
	}
	if o.HasStrokeDashArray {
		b.WriteString("stroke-dasharray=")
		b.WriteString(o.StrokeDashArray)
		b.WriteByte(';')
	}
	if o.HasStrokeDashOffset {
		b.WriteString("stroke-dashoffset=")
		b.WriteString(svgNumber(o.StrokeDashOffset))
		b.WriteByte(';')
	}

	return b.String()
}

// Inline-SVG paint sampling.
//
// An inline SVG is serialized and rasterized by the engine, so its fill and
// stroke are not fields of a display operation. A keyframe on those properties
// is sampled here as a used value, collected per target, and exposed through
// SVGPaintValues so the page can build a layout.SVGPaintOverride. Nothing in
// this file writes an operation.

// svgProperty names one paint property carried by SVGPaintValue. The order is
// the declaration order of the properties in a rule.
type svgProperty int

const (
	svgFill svgProperty = iota
	svgFillOpacity
	svgStroke
	svgStrokeOpacity
	svgStrokeWidth
	svgStrokeDashArray
	svgStrokeDashOffset
	svgPropertyCount
)

// declaredIn reports whether v carries this property.
func (p svgProperty) declaredIn(v SVGPaintValue) bool {
	switch p {
	case svgFill:
		return v.HasFill
	case svgFillOpacity:
		return v.HasFillOpacity
	case svgStroke:
		return v.HasStroke
	case svgStrokeOpacity:
		return v.HasStrokeOpacity
	case svgStrokeWidth:
		return v.HasStrokeWidth
	case svgStrokeDashArray:
		return v.HasStrokeDashArray
	case svgStrokeDashOffset:
		return v.HasStrokeDashOffset
	}

	return false
}

// pick returns a value carrying only this property, so an interpolation pair
// has one meaning per property rather than a full value with unrelated fields.
func (p svgProperty) pick(v SVGPaintValue) SVGPaintValue {
	var out SVGPaintValue
	switch p {
	case svgFill:
		out.Fill, out.HasFill = v.Fill, true
	case svgFillOpacity:
		out.FillOpacity, out.HasFillOpacity = v.FillOpacity, true
	case svgStroke:
		out.Stroke, out.HasStroke = v.Stroke, true
	case svgStrokeOpacity:
		out.StrokeOpacity, out.HasStrokeOpacity = v.StrokeOpacity, true
	case svgStrokeWidth:
		out.StrokeWidth, out.HasStrokeWidth = v.StrokeWidth, true
	case svgStrokeDashArray:
		out.StrokeDashArray, out.HasStrokeDashArray = v.StrokeDashArray, true
	case svgStrokeDashOffset:
		out.StrokeDashOffset, out.HasStrokeDashOffset = v.StrokeDashOffset, true
	}

	return out
}

// set writes one property's sampled value into out.
func (v *SVGPaintValue) set(p svgProperty, value SVGPaintValue) {
	switch p {
	case svgFill:
		v.Fill, v.HasFill = value.Fill, true
	case svgFillOpacity:
		v.FillOpacity, v.HasFillOpacity = value.FillOpacity, true
	case svgStroke:
		v.Stroke, v.HasStroke = value.Stroke, true
	case svgStrokeOpacity:
		v.StrokeOpacity, v.HasStrokeOpacity = value.StrokeOpacity, true
	case svgStrokeWidth:
		v.StrokeWidth, v.HasStrokeWidth = value.StrokeWidth, true
	case svgStrokeDashArray:
		// Copy the list so the exposed sample never aliases keyframe data.
		v.StrokeDashArray = append([]float64(nil), value.StrokeDashArray...)
		v.HasStrokeDashArray = true
	case svgStrokeDashOffset:
		v.StrokeDashOffset, v.HasStrokeDashOffset = value.StrokeDashOffset, true
	}
}

// hasSVGPaint reports whether any keyframe declares inline-SVG paint. Such an
// effect samples through composeSampled so each target can collect its own
// value, the same per-target path a transform takes.
func hasSVGPaint(frames []Keyframe) bool {
	for i := range frames {
		if frames[i].SVG.declared() {
			return true
		}
	}

	return false
}

// sampleSVGPaint samples every SVG paint property the effect declares. A
// property is sampled from the keyframes that declare it, in sorted order, so a
// keyframe that omits it does not interrupt the interpolation. The interval
// that starts at keyframe i uses that keyframe's timing function.
func (e Effect) sampleSVGPaint(at float64) SVGPaintValue {
	var out SVGPaintValue
	for property := svgFill; property < svgPropertyCount; property++ {
		stops := svgStops(e.Frames, property)
		if len(stops) == 0 {
			continue
		}
		if value, ok := sampleSVGStops(stops, property, e.Timing, at); ok {
			out.set(property, value)
		}
	}
	out.Deferred = svgDeferred(e.Frames)

	return out
}

// svgStop is one keyframe's value of one SVG paint property at a normalized
// offset. ease is the timing function declared on this keyframe, which eases
// the interval that starts here.
type svgStop struct {
	at    float64
	value SVGPaintValue
	ease  easingFn
}

// svgStops collects the keyframes that declare one SVG paint property, in
// sorted frame order.
func svgStops(frames []Keyframe, property svgProperty) []svgStop {
	stops := make([]svgStop, 0, len(frames))
	for i := range frames {
		if !property.declaredIn(frames[i].SVG) {
			continue
		}
		stops = append(stops, svgStop{
			at:    frames[i].Offset,
			value: property.pick(frames[i].SVG),
			ease:  frames[i].frameEasing(nil),
		})
	}

	return stops
}

// sampleSVGStops samples one property at at. A sample before the first stop or
// after the last holds that endpoint: the host has no computed base paint to
// fall back to.
func sampleSVGStops(stops []svgStop, property svgProperty, timing Timing, at float64) (SVGPaintValue, bool) {
	if len(stops) == 0 {
		return SVGPaintValue{}, false
	}
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
		part := segmentEase(prev.ease, timing, (at-prev.at)/span)

		return interpolateSVGProperty(property, prev.value, next.value, part), true
	}

	return stops[len(stops)-1].value, true
}

// interpolateSVGProperty blends two samples of one property. Colours
// interpolate in sRGB, opacity, width, and dash offset interpolate linearly,
// and a matching dash list interpolates component-wise. A none colour
// endpoint and a dash list with a different count are discrete: the sample
// takes one whole endpoint at the midpoint.
func interpolateSVGProperty(property svgProperty, a, b SVGPaintValue, t float64) SVGPaintValue {
	var out SVGPaintValue
	switch property {
	case svgFill:
		out.Fill, out.HasFill = blendSVGColor(a.Fill, b.Fill, t), true
	case svgFillOpacity:
		out.FillOpacity, out.HasFillOpacity = lerp(a.FillOpacity, b.FillOpacity, t), true
	case svgStroke:
		out.Stroke, out.HasStroke = blendSVGColor(a.Stroke, b.Stroke, t), true
	case svgStrokeOpacity:
		out.StrokeOpacity, out.HasStrokeOpacity = lerp(a.StrokeOpacity, b.StrokeOpacity, t), true
	case svgStrokeWidth:
		out.StrokeWidth, out.HasStrokeWidth = lerp(a.StrokeWidth, b.StrokeWidth, t), true
	case svgStrokeDashArray:
		out.StrokeDashArray, out.HasStrokeDashArray = blendSVGDashArray(a.StrokeDashArray, b.StrokeDashArray, t), true
	case svgStrokeDashOffset:
		out.StrokeDashOffset, out.HasStrokeDashOffset = lerp(a.StrokeDashOffset, b.StrokeDashOffset, t), true
	}

	return out
}

// blendSVGColor interpolates two sRGB colours and their alpha. none has no
// channels, so a pair with none on either side switches endpoint at the
// midpoint.
func blendSVGColor(a, b SVGPaintColor, t float64) SVGPaintColor {
	if a.None || b.None {
		if t < discreteMidpoint {
			return a
		}

		return b
	}

	return SVGPaintColor{
		R: lerp(a.R, b.R, t),
		G: lerp(a.G, b.G, t),
		B: lerp(a.B, b.B, t),
		A: lerp(a.A, b.A, t),
	}
}

// blendSVGDashArray interpolates two dash lists component-wise when they have
// the same count. A mismatch, including none against a list, is discrete at the
// midpoint: CSS has no component pairing for it.
func blendSVGDashArray(a, b []float64, t float64) []float64 {
	if len(a) != len(b) {
		if t < discreteMidpoint {
			return a
		}

		return b
	}
	if len(a) == 0 {
		return nil
	}
	out := make([]float64, len(a))
	for i := range a {
		out[i] = lerp(a[i], b[i], t)
	}

	return out
}

// svgDeferred returns the first declaration reason a paint server left behind.
// The declaration stays out of the sampled properties, so the element keeps its
// base paint and the page reports the reason instead of guessing a colour.
func svgDeferred(frames []Keyframe) string {
	for i := range frames {
		if frames[i].SVG.Deferred != "" {
			return frames[i].SVG.Deferred
		}
	}

	return ""
}

// SVGPaintValues returns the inline-SVG paint sampled for each target whose
// latest step produced one, in target order. A plain animation, or a step that
// has not run yet, returns nil. The slice is newly allocated per call; a page
// consumes it before stepping again.
//
// A player that stops sampling holds the last pose it wrote, the same way a
// finished forwards fill holds its end value. The page drops the value when it
// removes the animation and restores the element's base paint.
func (p *Player) SVGPaintValues() []SVGPaintValue {
	if p == nil {
		return nil
	}
	var values []SVGPaintValue
	for i := range p.targets {
		if !p.targets[i].svg.declared() {
			continue
		}
		value := p.targets[i].svg
		value.Target = i
		values = append(values, value)
	}

	return values
}

// SVGPaintMask marks a bound player whose only animated property is inline-SVG
// paint. An inline SVG is rasterized by the engine, so its paint is not a
// display field: the player targets the SVG root's image operation but writes
// no operation value. The bit sits above every cssmotion write field, so
// cssmotion.Read and cssmotion.Write leave the operation at its captured paint
// and the page restyles the raster through layout.RestyleInlineSVG instead.
const SVGPaintMask Mask = 1 << 8

// svgPaintClass is the pseudo target class for an animation whose declarations
// are all inline-SVG paint. The kind is unused: the class selects the element's
// image operation directly instead of going through cssmotion.Select.
func svgPaintClass() paintClass {
	return paintClass{kind: KindFill, mask: SVGPaintMask}
}

// declaresSVGPaint reports whether any keyframe declares a property that only
// the inline-SVG restyle path can apply. A rule that also declares a paint class
// keeps the SVG sample on that player; this predicate only decides whether an
// otherwise classless animation needs a player of its own.
func declaresSVGPaint(source []layout.Keyframe) bool {
	for i := range source {
		for _, declaration := range source[i].Declarations {
			if isSVGPaintProperty(declaration.Property) {
				return true
			}
		}
	}

	return false
}

// svgImageIndices returns the operation index of the inline SVG's raster. The
// SVG root owns exactly one image operation; a background or border the element
// declares paints before it, and the engine bakes the last image of the box as
// the raster. An empty slice means the element did not paint an SVG. A shared
// index resolves the lookup without scanning the display again.
func svgImageIndices(d *layout.Display, box layout.Box, boxIndex int, idx *Index) []int {
	if d == nil {
		return nil
	}
	if idx != nil {
		return idx.ImageIndex(box, boxIndex)
	}
	if boxIndex >= 0 && len(d.OpOwners) == len(d.Ops) {
		last := -1
		for i := range d.Ops {
			if d.OpOwners[i] == boxIndex && d.Ops[i].Kind == layout.DisplayOpImage {
				last = i
			}
		}
		if last >= 0 {
			return []int{last}
		}
	}

	// No trusted owner index, or the owner map is missing: match the box's
	// operation ids. A box that predates operation ids falls back to the same
	// geometric containment the paint classes use.
	ids := make(map[uint64]struct{}, len(box.OperationIDs))
	for _, id := range box.OperationIDs {
		ids[id] = struct{}{}
	}
	last := -1
	for i := range d.Ops {
		if d.Ops[i].Kind != layout.DisplayOpImage {
			continue
		}
		if len(ids) > 0 {
			if _, ok := ids[d.Ops[i].ID]; !ok {
				continue
			}
		} else if !Contains(d, box, d.Ops[i]) {
			continue
		}
		last = i
	}
	if last >= 0 {
		return []int{last}
	}

	return nil
}

// reportSVGPaintDeferred counts the declarations a paint server left at the
// element's base paint. The sample carries the reason on every frame; the bind
// records it once per element, next to the other fidelity gaps.
func reportSVGPaintDeferred(report *gapList, box layout.Box, frames []Keyframe) {
	for i := range frames {
		if frames[i].SVG.Deferred == "" {
			continue
		}
		report.add(fmt.Sprintf("box %q: %s", box.ID, frames[i].SVG.Deferred))

		return
	}
}

// ErrSVGPaint reports an inline-SVG paint declaration a keyframe cannot parse.
var ErrSVGPaint = errors.New("cssanimation: invalid svg paint")

// ErrPaintServerDeferred reports a fill or stroke that references a paint
// server, such as url(#gradient). The engine stores paint as sRGB, so a paint
// server cannot be restyled from a sampled colour. The declaration is not a
// bind error: the element keeps its base paint and the sample carries the
// reason, rather than the host guessing a colour.
var ErrPaintServerDeferred = errors.New("cssanimation: paint server is deferred")

// isSVGPaintProperty reports one keyframe property sampled for an inline SVG.
// These live outside paintClass because they target no display operation:
// changing them rebuilds the SVG raster through layout.RestyleInlineSVG.
func isSVGPaintProperty(property string) bool {
	switch property {
	case "fill", "fill-opacity", "stroke", "stroke-opacity", "stroke-width",
		"stroke-dasharray", "stroke-dashoffset":
		return true
	}

	return false
}

// parseSVGPaintDeclaration writes one inline-SVG paint declaration into a
// frame. A paint server and a colour space this package cannot represent are
// deferred with a reason. A malformed value fails the bind, like every other
// paint declaration.
func parseSVGPaintDeclaration(frame *Keyframe, declaration layout.KeyframeDeclaration,
	size transformSize, current Color, hasCurrent bool) {
	switch declaration.Property {
	case "fill":
		value, err := svgColor(declaration.Value, current, hasCurrent)
		if err != nil {
			noteSVGPaintError(frame, declaration.Property, err)
			return
		}
		frame.SVG.Fill, frame.SVG.HasFill = value, true
	case "fill-opacity":
		value, err := svgOpacity(declaration.Value)
		if err != nil {
			noteSVGPaintError(frame, declaration.Property, err)
			return
		}
		frame.SVG.FillOpacity, frame.SVG.HasFillOpacity = value, true
	case "stroke":
		value, err := svgColor(declaration.Value, current, hasCurrent)
		if err != nil {
			noteSVGPaintError(frame, declaration.Property, err)
			return
		}
		frame.SVG.Stroke, frame.SVG.HasStroke = value, true
	case "stroke-opacity":
		value, err := svgOpacity(declaration.Value)
		if err != nil {
			noteSVGPaintError(frame, declaration.Property, err)
			return
		}
		frame.SVG.StrokeOpacity, frame.SVG.HasStrokeOpacity = value, true
	case "stroke-width":
		value, err := svgPaintLength(declaration.Value, size)
		if err != nil {
			noteSVGPaintError(frame, declaration.Property, err)
			return
		}
		if value < 0 {
			noteSVGPaintError(frame, declaration.Property,
				fmt.Errorf("%w: stroke-width %q is negative", ErrSVGPaint, declaration.Value))
			return
		}
		frame.SVG.StrokeWidth, frame.SVG.HasStrokeWidth = value, true
	case "stroke-dasharray":
		value, err := svgDashArray(declaration.Value, size)
		if err != nil {
			noteSVGPaintError(frame, declaration.Property, err)
			return
		}
		frame.SVG.StrokeDashArray, frame.SVG.HasStrokeDashArray = value, true
	case "stroke-dashoffset":
		value, err := svgPaintLength(declaration.Value, size)
		if err != nil {
			noteSVGPaintError(frame, declaration.Property, err)
			return
		}
		frame.SVG.StrokeDashOffset, frame.SVG.HasStrokeDashOffset = value, true
	}
}

// noteSVGPaintError records a declaration this phase cannot turn into a used
// value. A paint server or a colour space the package does not model is
// deferred with a reason; every other error fails the bind.
func noteSVGPaintError(frame *Keyframe, property string, err error) {
	if errors.Is(err, ErrPaintServerDeferred) || errors.Is(err, ErrColorSpace) {
		frame.SVG.Deferred = property + ": " + err.Error()
		return
	}
	frame.Err = fmt.Errorf("cssanimation: %s: %w", property, err)
}

// svgColor parses one fill or stroke value. none is the no-paint keyword, and
// every other form is the colour grammar the paint phase already accepts. A
// paint server reference is deferred: the engine has no RGB value for it.
func svgColor(value string, current Color, hasCurrent bool) (SVGPaintColor, error) {
	text := strings.ToLower(strings.TrimSpace(value))
	if text == "none" {
		return SVGPaintColor{None: true, A: maxChannel}, nil
	}
	if strings.HasPrefix(text, "url(") {
		return SVGPaintColor{}, fmt.Errorf("%w: %s",
			ErrPaintServerDeferred, strings.TrimSpace(value))
	}
	parsed, err := resolveColor(value, current, hasCurrent)
	if err != nil {
		return SVGPaintColor{}, err
	}

	return SVGPaintColor{R: parsed.R, G: parsed.G, B: parsed.B, A: parsed.A}, nil
}

// svgOpacity parses fill-opacity or stroke-opacity as a number or a
// percentage and clamps it to the 0..1 range CSS uses.
func svgOpacity(value string) (float64, error) {
	text := strings.TrimSpace(value)
	if digits, ok := strings.CutSuffix(text, "%"); ok {
		parsed, err := finite(digits)
		if err != nil {
			return 0, fmt.Errorf("%w: opacity %q", ErrSVGPaint, value)
		}

		return clamp01(parsed / percentScale), nil
	}
	parsed, err := finite(text)
	if err != nil {
		return 0, fmt.Errorf("%w: opacity %q", ErrSVGPaint, value)
	}

	return clamp01(parsed), nil
}

// svgPaintLength parses one stroke length as layout points. A bare number and
// px are the same length, and pt is already a layout point. A unit that needs
// font or viewport context is rejected rather than resolved to a guess.
func svgPaintLength(value string, size transformSize) (float64, error) {
	text := strings.ToLower(strings.TrimSpace(value))
	if text == "" {
		return 0, fmt.Errorf("%w: empty length", ErrSVGPaint)
	}
	unit := ""
	switch {
	case strings.HasSuffix(text, "px"):
		unit = "px"
	case strings.HasSuffix(text, "pt"):
		unit = "pt"
	}
	number, err := finite(text[:len(text)-len(unit)])
	if err != nil {
		return 0, fmt.Errorf("%w: %q is not a length in px or pt", ErrSVGPaint, value)
	}
	if unit == "pt" {
		return number, nil
	}
	scale := size.pointsPerPixel
	if scale <= 0 {
		scale = 1
	}

	return number * scale, nil
}

// svgDashArray parses the CSS stroke-dasharray list as lengths in layout
// points. none is the empty list; a negative dash length is invalid.
func svgDashArray(value string, size transformSize) ([]float64, error) {
	text := strings.TrimSpace(value)
	if strings.EqualFold(text, "none") {
		return nil, nil
	}
	fields := strings.Fields(strings.ReplaceAll(text, ",", " "))
	if len(fields) == 0 {
		return nil, fmt.Errorf("%w: empty dash list", ErrSVGPaint)
	}
	out := make([]float64, len(fields))
	for i, field := range fields {
		length, err := svgPaintLength(field, size)
		if err != nil {
			return nil, err
		}
		if length < 0 {
			return nil, fmt.Errorf("%w: dash %q is negative", ErrSVGPaint, field)
		}
		out[i] = length
	}

	return out, nil
}
