package main

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe"
)

type paint struct {
	X, Y, W, H     float64
	R, G, B, Alpha float64
	Opacity        float64
	Matrix         layout.Matrix2D
	HasMatrix      bool
}

func capture(d *ownframe.Display, box ownframe.Box) []paint {
	var out []paint
	for _, index := range operationIndices(d, box) {
		out = append(out, readPaint(&d.Ops[index]))
	}
	return out
}

func operationIndices(d *ownframe.Display, box ownframe.Box) []int {
	if d == nil {
		return nil
	}
	if box.OperationIDs == nil {
		var out []int
		for i := range d.Ops {
			if containsOp(d, box, d.Ops[i]) {
				out = append(out, i)
			}
		}
		return out
	}
	ids := make(map[uint64]struct{}, len(box.OperationIDs))
	for _, id := range box.OperationIDs {
		ids[id] = struct{}{}
	}
	var out []int
	for i := range d.Ops {
		if _, ok := ids[d.Ops[i].ID]; ok {
			out = append(out, i)
		}
	}
	return out
}

func resolveOps(d *ownframe.Display, indices []int) []*ownframe.DisplayOp {
	var out []*ownframe.DisplayOp
	if d == nil {
		return out
	}
	for _, index := range indices {
		if index >= 0 && index < len(d.Ops) {
			out = append(out, &d.Ops[index])
		}
	}
	return out
}

const (
	halfScalar         = 0.5
	easeCtrlX1         = 0.25
	easeCtrlY1         = 0.1
	easeCtrlX2         = 0.25
	easeInCtrlX1       = 0.42
	easeOutCtrlX2      = 0.58
	cubicBezierArgLen  = 4
	bezierBisectIters  = 24
	hexRGBStringLen    = 6
	hexChannelShiftR   = 16
	hexChannelShiftG   = 8
	hexChannelByteMask = 0xff
	rgbChannelMax      = 255
	colorRoundScale    = 1000
)

func containsOp(d *ownframe.Display, box ownframe.Box, op ownframe.DisplayOp) bool {
	x, y := op.X+op.W*halfScalar, op.Y+op.H*halfScalar
	if op.Kind == ownframe.DisplayOpText {
		y = op.Y
	}
	p := d.PointsPerPixel
	return x >= box.X*p && x <= (box.X+box.W)*p &&
		y >= box.Y*p && y <= (box.Y+box.H)*p
}

func readPaint(op *ownframe.DisplayOp) paint {
	opacity := op.Opacity()
	if op.Alpha > 0 && op.Alpha < 1 && opacity > 0 {
		opacity /= op.Alpha
	}
	p := paint{
		X: op.X, Y: op.Y, W: op.W, H: op.H,
		R: op.R, G: op.G, B: op.B, Alpha: op.Alpha,
		Opacity: opacity,
	}
	if op.XformSet {
		p.Matrix = op.Transform()
		p.HasMatrix = true
	} else {
		p.Matrix = layout.Matrix2D{A: 1, D: 1}
	}
	return p
}

func writePaint(op *ownframe.DisplayOp, p paint) {
	op.X, op.Y, op.W, op.H = p.X, p.Y, p.W, p.H
	op.R, op.G, op.B, op.Alpha = p.R, p.G, p.B, p.Alpha
	op.SetPaintOpacity(p.Opacity)
	if p.HasMatrix {
		op.SetXform(p.Matrix)
	}
}

func blendPaint(a, b paint, t float64) paint {
	lerp := func(x, y float64) float64 { return x + (y-x)*t }
	out := paint{
		X: lerp(a.X, b.X), Y: lerp(a.Y, b.Y),
		W: lerp(a.W, b.W), H: lerp(a.H, b.H),
		R: lerp(a.R, b.R), G: lerp(a.G, b.G),
		B: lerp(a.B, b.B), Alpha: lerp(a.Alpha, b.Alpha),
		Opacity: lerp(a.Opacity, b.Opacity),
	}
	switch {
	case a.HasMatrix && b.HasMatrix:
		out.HasMatrix = true
		out.Matrix = layout.Matrix2D{
			A: lerp(a.Matrix.A, b.Matrix.A),
			B: lerp(a.Matrix.B, b.Matrix.B),
			C: lerp(a.Matrix.C, b.Matrix.C),
			D: lerp(a.Matrix.D, b.Matrix.D),
			E: lerp(a.Matrix.E, b.Matrix.E),
			F: lerp(a.Matrix.F, b.Matrix.F),
		}
	case b.HasMatrix:
		out.HasMatrix = true
		out.Matrix = b.Matrix
	default:
		out.Matrix = layout.Matrix2D{A: 1, D: 1}
	}
	return out
}

func easeProgress(value string, t float64) float64 {
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "linear":
		return t
	case "ease":
		return cubicEase(t, easeCtrlX1, easeCtrlY1, easeCtrlX2, 1)
	case "ease-in":
		return cubicEase(t, easeInCtrlX1, 0, 1, 1)
	case "ease-out":
		return cubicEase(t, 0, 0, easeOutCtrlX2, 1)
	case "ease-in-out":
		return cubicEase(t, easeInCtrlX1, 0, easeOutCtrlX2, 1)
	case "step-start":
		return 1
	case "step-end":
		return stepEase(t, 1, "end")
	}
	s := strings.TrimSpace(value)
	i := strings.IndexByte(s, '(')
	if i < 1 || !strings.HasSuffix(s, ")") {
		return t
	}
	name := strings.ToLower(strings.TrimSpace(s[:i]))
	body := strings.TrimSuffix(strings.TrimSpace(s[i+1:]), ")")
	switch name {
	case "cubic-bezier":
		parts := strings.Split(body, ",")
		if len(parts) == cubicBezierArgLen {
			var v [cubicBezierArgLen]float64
			for idx, part := range parts {
				val, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
				if err != nil {
					return t
				}
				v[idx] = val
			}
			return cubicEase(t, v[0], v[1], v[2], v[3])
		}
	case "steps":
		parts := strings.Split(body, ",")
		n, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err == nil && n > 0 {
			pos := "end"
			if len(parts) > 1 {
				pos = strings.TrimSpace(parts[1])
			}
			res := stepEase(t, n, pos)
			if res < 0 {
				return 0
			}
			if res > 1 {
				return 1
			}
			return res
		}
	}
	return t
}

func cubicEase(t, x1, y1, x2, y2 float64) float64 {
	lo, hi, u := 0.0, 1.0, t
	bezier := func(p, a, b float64) float64 {
		q := 1 - p
		return 3*q*q*p*a + 3*q*p*p*b + p*p*p
	}
	for i := 0; i < bezierBisectIters; i++ {
		if bezier(u, x1, x2) < t {
			lo = u
		} else {
			hi = u
		}
		u = (lo + hi) * halfScalar
	}
	return bezier(u, y1, y2)
}

func stepEase(t float64, count int, position string) float64 {
	if count <= 1 && position == "jump-none" {
		return t
	}
	switch position {
	case "start", "jump-start":
		return float64(int(t*float64(count))+1) / float64(count)
	case "jump-both":
		return float64(int(t*float64(count))+1) / float64(count+1)
	case "jump-none":
		return float64(int(t*float64(count))) / float64(count-1)
	default:
		return float64(int(t*float64(count))) / float64(count)
	}
}

type transitionTrack struct {
	op              *ownframe.DisplayOp
	from, to        paint
	duration, delay time.Duration
	easing          string
	discrete        bool
}

type localTransition struct {
	tracks []transitionTrack
	start  time.Time
}

func (r *localTransition) Step(now time.Time) bool {
	if r == nil {
		return false
	}
	active := false
	for _, track := range r.tracks {
		if track.duration == 0 {
			writePaint(track.op, track.to)
			continue
		}
		progress := float64(now.Sub(r.start)-track.delay) / float64(track.duration)
		if progress < 0 {
			writePaint(track.op, track.from)
			active = true
			continue
		}
		if progress >= 1 {
			writePaint(track.op, track.to)
			continue
		}
		if track.discrete {
			if easeProgress(track.easing, progress) < halfScalar {
				writePaint(track.op, track.from)
			} else {
				writePaint(track.op, track.to)
			}
			active = true
			continue
		}
		writePaint(track.op, blendPaint(track.from, track.to, easeProgress(track.easing, progress)))
		active = true
	}
	return active
}

func bindLocalTransitionTo(
	d *ownframe.Display, box ownframe.Box, property string,
	indices []int, initial, final []paint, start time.Time,
) (*localTransition, error) {
	ops := resolveOps(d, indices)
	var rec layout.Transition
	found := false
	for _, item := range box.Transitions {
		if item.Property == property || item.Property == "all" {
			rec = item
			found = true
			break
		}
	}
	if !found || len(ops) == 0 || len(initial) != len(ops) {
		return &localTransition{start: start}, nil
	}
	discrete := property == "position" || property == "visibility" || property == "display"
	tracks := make([]transitionTrack, len(ops))
	for i, op := range ops {
		to := readPaint(op)
		if final != nil && i < len(final) {
			to = final[i]
		}
		from, dest := adjustPaint(initial[i], initial[i], to, readPaint(op)), adjustPaint(to, initial[i], to, readPaint(op))
		dur := time.Duration(rec.DurationMS * float64(time.Millisecond))
		del := time.Duration(rec.DelayMS * float64(time.Millisecond))
		if discrete && rec.Behavior != "allow-discrete" {
			dur, del = 0, 0
		}
		tracks[i] = transitionTrack{
			op: op, from: from, to: dest,
			duration: dur, delay: del,
			easing:   rec.TimingFunction,
			discrete: discrete && rec.Behavior == "allow-discrete",
		}
	}
	return &localTransition{tracks: tracks, start: start}, nil
}

func adjustPaint(captured, initial, final, current paint) paint {
	field := func(c, lo, hi, cur float64) float64 {
		if lo > hi {
			lo, hi = hi, lo
		}
		if cur < lo-0.5 || cur > hi+0.5 {
			return cur
		}
		return c
	}
	return paint{
		X:         field(captured.X, initial.X, final.X, current.X),
		Y:         field(captured.Y, initial.Y, final.Y, current.Y),
		W:         field(captured.W, initial.W, final.W, current.W),
		H:         field(captured.H, initial.H, final.H, current.H),
		R:         field(captured.R, initial.R, final.R, current.R),
		G:         field(captured.G, initial.G, final.G, current.G),
		B:         field(captured.B, initial.B, final.B, current.B),
		Alpha:     field(captured.Alpha, initial.Alpha, final.Alpha, current.Alpha),
		Opacity:   field(captured.Opacity, initial.Opacity, final.Opacity, current.Opacity),
		Matrix:    captured.Matrix,
		HasMatrix: captured.HasMatrix,
	}
}

type localClock struct {
	Start  time.Time
	Offset time.Duration
}

type localKeyframe struct {
	offset     float64
	tx, ty     float64
	hasTx      bool
	opacity    float64
	hasOpacity bool
	r, g, b    float64
	hasColor   bool
}

type localPlayer struct {
	ops       []*ownframe.DisplayOp
	initial   []paint
	frames    []localKeyframe
	duration  time.Duration
	delay     time.Duration
	direction string
	easing    string
	scale     float64
	start     time.Time
	started   bool
}

func (p *localPlayer) Clock() (localClock, bool) {
	if p == nil || !p.started {
		return localClock{}, false
	}
	return localClock{Start: p.start}, true
}

func (p *localPlayer) Step(now time.Time) bool {
	if p == nil || !p.started || len(p.frames) == 0 || p.duration <= 0 {
		return false
	}
	elapsed := now.Sub(p.start) - p.delay
	active := true
	progress := float64(elapsed) / float64(p.duration)
	if progress < 0 {
		progress = 0
	} else if progress >= 1 {
		progress = 1
		active = false
	}
	if p.direction == "reverse" || p.direction == "alternate-reverse" {
		progress = 1 - progress
	}
	eased := easeProgress(p.easing, progress)
	sample := sampleLocalKeyframes(p.frames, eased, p.initial[0])
	for i, op := range p.ops {
		out := p.initial[i]
		if sample.hasTx {
			out.HasMatrix = true
			out.Matrix = layout.Matrix2D{A: 1, D: 1, E: sample.tx * p.scale, F: sample.ty * p.scale}
		}
		if sample.hasOpacity {
			out.Opacity = sample.opacity
		}
		if sample.hasColor {
			out.R, out.G, out.B = sample.r, sample.g, sample.b
		}
		writePaint(op, out)
	}
	return active
}

func sampleLocalKeyframes(frames []localKeyframe, t float64, base paint) localKeyframe {
	if len(frames) == 1 {
		return frames[0]
	}
	prev := frames[0]
	next := frames[len(frames)-1]
	for i := 1; i < len(frames); i++ {
		if t <= frames[i].offset {
			prev = frames[i-1]
			next = frames[i]
			break
		}
	}
	span := next.offset - prev.offset
	localT := 0.0
	if span > 0 {
		localT = (t - prev.offset) / span
	}
	lerp := func(a, b float64) float64 { return a + (b-a)*localT }
	out := localKeyframe{
		hasTx:      prev.hasTx || next.hasTx,
		tx:         lerp(prev.tx, next.tx),
		ty:         lerp(prev.ty, next.ty),
		hasOpacity: prev.hasOpacity || next.hasOpacity,
		opacity:    lerp(prev.opacity, next.opacity),
		hasColor:   prev.hasColor || next.hasColor,
	}
	pr, pg, pb := base.R, base.G, base.B
	if prev.hasColor {
		pr, pg, pb = prev.r, prev.g, prev.b
	}
	nr, ng, nb := base.R, base.G, base.B
	if next.hasColor {
		nr, ng, nb = next.r, next.g, next.b
	}
	out.r, out.g, out.b = lerp(pr, nr), lerp(pg, ng), lerp(pb, nb)
	return out
}

func bindLocalAnimation(d *ownframe.Display, box ownframe.Box, indices []int, initial []paint, start time.Time) (*localPlayer, error) {
	ops := resolveOps(d, indices)
	if d == nil || len(box.Animations) == 0 || len(ops) == 0 || len(initial) != len(ops) {
		return &localPlayer{start: start}, nil
	}
	anim := box.Animations[0]
	scale := d.PointsPerPixel
	if scale <= 0 {
		scale = 1
	}
	var frames []localKeyframe
	for _, rule := range d.Keyframes {
		if rule.Name == anim.Name {
			for _, kf := range rule.Keyframes {
				frames = append(frames, parseLocalKeyframe(kf, initial[0]))
			}
			break
		}
	}
	return &localPlayer{
		ops:       ops,
		initial:   initial,
		frames:    frames,
		duration:  time.Duration(anim.DurationMS * float64(time.Millisecond)),
		delay:     time.Duration(anim.DelayStartMS * float64(time.Millisecond)),
		direction: anim.Direction,
		easing:    anim.TimingFunction,
		scale:     scale,
		start:     start,
	}, nil
}

func parseLocalKeyframe(kf layout.Keyframe, base paint) localKeyframe {
	out := localKeyframe{
		offset:  kf.Offset,
		opacity: base.Opacity,
		r:       base.R, g: base.G, b: base.B,
	}
	for _, decl := range kf.Declarations {
		val := strings.TrimSpace(decl.Value)
		switch decl.Property {
		case "transform":
			tx, ty := parseTranslate(val)
			out.tx, out.ty, out.hasTx = tx, ty, true
		case "opacity":
			if v, err := strconv.ParseFloat(val, 64); err == nil {
				out.opacity, out.hasOpacity = v, true
			}
		case "background-color", "background":
			if r, g, b, ok := parseHexColor(val); ok {
				out.r, out.g, out.b, out.hasColor = r, g, b, true
			}
		}
	}
	return out
}

func parseTranslate(s string) (float64, float64) {
	i := strings.IndexByte(s, '(')
	j := strings.LastIndexByte(s, ')')
	if i < 0 || j <= i {
		return 0, 0
	}
	name := strings.ToLower(strings.TrimSpace(s[:i]))
	parts := strings.Split(s[i+1:j], ",")
	parsePx := func(raw string) float64 {
		raw = strings.TrimSuffix(strings.TrimSpace(raw), "px")
		v, _ := strconv.ParseFloat(raw, 64)
		return v
	}
	switch name {
	case "translatex":
		return parsePx(parts[0]), 0
	case "translatey":
		return 0, parsePx(parts[0])
	case "translate":
		tx := parsePx(parts[0])
		var ty float64
		if len(parts) > 1 {
			ty = parsePx(parts[1])
		}
		return tx, ty
	}
	return 0, 0
}

func parseHexColor(s string) (float64, float64, float64, bool) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "#")
	if len(s) != hexRGBStringLen {
		return 0, 0, 0, false
	}
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return 0, 0, 0, false
	}
	r := float64((v>>hexChannelShiftR)&hexChannelByteMask) / rgbChannelMax
	g := float64((v>>hexChannelShiftG)&hexChannelByteMask) / rgbChannelMax
	b := float64(v&hexChannelByteMask) / rgbChannelMax
	return math.Round(r*colorRoundScale) / colorRoundScale,
		math.Round(g*colorRoundScale) / colorRoundScale,
		math.Round(b*colorRoundScale) / colorRoundScale,
		true
}
