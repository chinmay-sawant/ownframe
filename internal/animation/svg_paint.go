package animation

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
