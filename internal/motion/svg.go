package motion

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/animation"
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
func svgPaintOverride(v animation.SVGPaintValue, elementID string) (layout.SVGPaintOverride, bool) {
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
func svgPaintColor(c animation.SVGPaintColor) string {
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
