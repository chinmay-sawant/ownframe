package motion

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"math"
	"testing"
	"time"
)

// TestBuildEffectTableThreeStops checks the flat, sorted table shape: three
// real stops stay three stops, and a sparse field gains the 0 and 1 sentinels.
func TestBuildEffectTableThreeStops(t *testing.T) {
	e := Effect{Frames: []Keyframe{
		{Offset: 0, X: 10, Move: true},
		{Offset: 0.5, X: 30, Move: true},
		{Offset: 1, X: 10, Move: true},
	}}
	table := buildEffectTable(e)
	if table.mask != 1<<animFieldX|1<<animFieldY {
		t.Fatalf("mask = %b, want only the move fields", table.mask)
	}
	field := table.fields[animFieldX]
	if !field.present || len(field.stops) != 3 || len(field.segments) != 2 {
		t.Fatalf("stops = %d segments = %d present = %v, want 3/2/true",
			len(field.stops), len(field.segments), field.present)
	}
	for i, want := range []float64{0, 0.5, 1} {
		if field.stops[i].at != want {
			t.Fatalf("stop %d at = %v, want %v", i, field.stops[i].at, want)
		}
	}
	if field.segments[0].span != 0.5 || field.segments[1].span != 0.5 {
		t.Fatalf("segment spans = %v, %v", field.segments[0].span, field.segments[1].span)
	}
	for other := animFieldOpacity; other < fieldCount; other++ {
		if table.fields[other].present {
			t.Fatalf("field %d has stops but no keyframe carries it", other)
		}
	}

	sparse := buildEffectTable(Effect{Frames: []Keyframe{{Offset: 0.5, X: 30, Move: true}}})
	stops := sparse.fields[animFieldX].stops
	if len(stops) != 3 || stops[0].at != 0 || stops[2].at != 1 {
		t.Fatalf("sparse stops = %+v, want sentinels at 0 and 1", stops)
	}
	if stops[0].fromBase != 0 || stops[2].fromBase != 0 {
		t.Fatalf("position sentinels fromBase = %v/%v, want 0", stops[0].fromBase, stops[2].fromBase)
	}
}

// TestEffectTableAppliesSegmentEasing checks that each precomputed segment
// carries the parsed easing and applies it to its own local progress.
func TestEffectTableAppliesSegmentEasing(t *testing.T) {
	e := Effect{
		Frames: []Keyframe{
			{Offset: 0, X: 0, Move: true},
			{Offset: 0.5, X: 10, Move: true},
			{Offset: 1, X: 30, Move: true},
		},
		Timing: Timing{Easing: func(v float64) float64 { return v * v }},
	}
	table := buildEffectTable(e)
	field := &table.fields[animFieldX]
	if field.segments[0].ease == nil || field.segments[1].ease == nil {
		t.Fatal("segment easing was not precomputed")
	}
	if got, _ := field.sample(0.25); math.Abs(got-2.5) > 1e-9 {
		t.Fatalf("first segment x = %v, want 2.5", got)
	}
	if got, _ := field.sample(0.75); math.Abs(got-15) > 1e-9 {
		t.Fatalf("second segment x = %v, want 15", got)
	}
}

// TestStepRepeatedAtKeepsValues checks that sampling the same time repeatedly
// does not drift or skip a segment.
func TestStepRepeatedAtKeepsValues(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	p, err := newAnimationPlayer([]Target{{Op: op, From: Read(op)}}, []Keyframe{
		{Offset: 0, X: 0, Move: true},
		{Offset: 1, X: 100, Move: true},
	}, Timing{Duration: time.Second, Iterations: 1})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(30, 0)
	p.Start(start)
	mid := start.Add(500 * time.Millisecond)
	for i := 0; i < 5; i++ {
		p.Step(mid)
		if op.X != 50 {
			t.Fatalf("step %d x = %v, want 50", i, op.X)
		}
	}
}

// TestEffectTableCursorHandlesReverseProgress checks that a decreasing
// progress resets the forward cursor instead of reading a later segment.
func TestEffectTableCursorHandlesReverseProgress(t *testing.T) {
	e := Effect{
		Frames: []Keyframe{
			{Offset: 0, X: 0, Move: true},
			{Offset: 0.5, X: 100, Move: true},
			{Offset: 1, X: 0, Move: true},
		},
		Timing: Timing{TimingFunction: "linear"},
	}
	table := buildEffectTable(e)
	field := &table.fields[animFieldX]
	for _, at := range []float64{0.9, 0.6, 0.3, 0.1, 0.45} {
		got, _ := field.sample(at)
		fresh := buildEffectTable(e).fields[animFieldX]
		want, _ := fresh.sample(at)
		if math.Abs(got-want) > 1e-9 {
			t.Fatalf("at %v x = %v, want %v", at, got, want)
		}
	}

	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	p, err := newAnimationPlayer([]Target{{Op: op, From: Read(op)}}, []Keyframe{
		{Offset: 0, X: 0, Move: true},
		{Offset: 1, X: 100, Move: true},
	}, Timing{Duration: time.Second, Iterations: 1, Direction: Reverse})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(40, 0)
	p.Start(start)
	p.Step(start.Add(100 * time.Millisecond))
	if math.Abs(op.X-90) > 1e-9 {
		t.Fatalf("reverse start x = %v, want 90", op.X)
	}
	p.Step(start.Add(900 * time.Millisecond))
	if math.Abs(op.X-10) > 1e-9 {
		t.Fatalf("reverse end x = %v, want 10", op.X)
	}
}

// TestStepDoesNotAllocate is the phase 07 gate: a live frame writes values
// without heap traffic.
func TestStepDoesNotAllocate(t *testing.T) {
	cases := []struct {
		name    string
		effects int
	}{
		{"one effect", 1},
		{"ten effects", 10},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			player, start := liveStepPlayer(t, tc.effects)
			frame := 0
			allocs := testing.AllocsPerRun(200, func() {
				frame++
				player.Step(start.Add(time.Duration(frame) * 16 * time.Millisecond))
			})
			if allocs != 0 {
				t.Fatalf("live step allocated %v times per run", allocs)
			}
		})
	}
}

// TestEffectTableMatchesCompose checks the table path against the bare sample
// loop for every composition mode, sparse and dense keyframes, and offsets
// outside 0..1. The table may reorder the arithmetic, so compare with a small
// tolerance rather than bit equality.
func TestEffectTableMatchesCompose(t *testing.T) {
	base := Paint{X: 12, Y: 7, W: 40, H: 20,
		R: 0.2, G: 0.4, B: 0.6, Alpha: 1, Opacity: 0.7}
	smooth := func(v float64) float64 { return v * v * (3 - 2*v) }
	cases := []struct {
		name   string
		frames []Keyframe
		timing Timing
	}{
		{
			name: "replace sparse",
			frames: []Keyframe{
				{Offset: 0.25, X: 10, Move: true},
				{Offset: 0.4, R: 0.9, ColorSet: true},
				{Offset: 0.75, Opacity: 0.3, OpacitySet: true},
			},
		},
		{
			name: "replace dense",
			frames: []Keyframe{
				{Offset: 0, X: 0, Opacity: 1, OpacitySet: true, Move: true},
				{Offset: 0.5, X: 30, Opacity: 0, OpacitySet: true,
					R: 1, G: 0.5, B: 0.25, ColorSet: true, Move: true},
				{Offset: 1, X: 5, Opacity: 0.5, OpacitySet: true, Move: true},
			},
		},
		{
			name: "add sparse",
			frames: []Keyframe{
				{Offset: 0.3, X: 4, Opacity: 0.1, OpacitySet: true, Move: true},
				{Offset: 0.8, Y: 3, Move: true},
			},
			timing: Timing{Composition: Add},
		},
		{
			name: "accumulate sparse",
			frames: []Keyframe{
				{Offset: 0.2, X: 6, R: 0.3, ColorSet: true, Move: true},
				{Offset: 0.6, X: 9, G: 0.2, ColorSet: true, Move: true},
			},
			timing: Timing{Composition: Accumulate},
		},
		{
			name: "css easing",
			frames: []Keyframe{
				{Offset: 0, X: 0, Move: true},
				{Offset: 0.5, X: 40, Move: true},
				{Offset: 1, X: 10, Move: true},
			},
			timing: Timing{TimingFunction: "ease-in-out"},
		},
		{
			name: "custom easing",
			frames: []Keyframe{
				{Offset: 0.1, Opacity: 0.2, OpacitySet: true},
				{Offset: 0.9, Opacity: 0.8, OpacitySet: true},
			},
			timing: Timing{Easing: smooth},
		},
		{
			name: "duplicate offsets",
			frames: []Keyframe{
				{Offset: 0, X: 10, Move: true},
				{Offset: 0.5, X: 20, Move: true},
				{Offset: 0.5, X: 80, Move: true},
				{Offset: 1, X: 20, Move: true},
			},
		},
		{
			name:   "single stop",
			frames: []Keyframe{{Offset: 0.5, X: 80, Move: true}},
		},
	}
	offsets := []float64{-0.2, 0, 0.1, 0.25, 0.4, 0.5, 0.6, 0.75, 0.9, 1, 1.2}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := Effect{Frames: tc.frames, Timing: tc.timing}
			table := buildEffectTable(e)
			for _, at := range offsets {
				want := compose(&Target{From: base}, base, e, at)
				var set valueSet
				table.sample(at, &set)
				got := base
				set.applyTo(&got, base, tc.timing.Composition)
				if !paintClose(got, want) {
					t.Fatalf("at %v paint = %+v, want %+v", at, got, want)
				}
			}
		})
	}
}

// TestStepMatchesComposeAcrossEffects checks the inverted loop's effect order:
// each effect is sampled once and composed onto every target in list order,
// including a sparse replacing opacity that must fall back to the base paint.
func TestStepMatchesComposeAcrossEffects(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 10, Alpha: 1}
	op.SetPaintOpacity(0.8)
	base := Read(op)
	timing := Timing{Duration: time.Second, Iterations: 1}
	effects := []Effect{
		{Frames: []Keyframe{
			{Offset: 0, X: 10, Move: true},
			{Offset: 1, X: 30, Move: true},
		}, Timing: timing},
		{Frames: []Keyframe{
			{Offset: 0.5, Opacity: 0.5, OpacitySet: true},
		}, Timing: timing},
	}
	p, err := NewComposite([]Target{{Op: op, From: base}}, effects)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(48, 0)
	p.Start(start)
	p.Step(start.Add(250 * time.Millisecond))
	want := base
	for _, effect := range effects {
		want = compose(&Target{Op: op, From: base}, want, effect, 0.25)
	}
	if !paintClose(Read(op), want) {
		t.Fatalf("paint = %+v, want %+v", Read(op), want)
	}
}

// TestStepKeepsParsedTransformOnCompose checks that an effect carrying a
// parsed transform still reaches the compose oracle. The six-field table only
// covers the scalar fields; transform sampling stays in phase 02 until it has
// its own precomputed stops.
func TestStepKeepsParsedTransformOnCompose(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 10, Y: 5, Alpha: 1}
	base := Read(op)
	target := Target{Op: op, From: base}
	frames := []Keyframe{
		{Offset: 0, Transform: TransformValue{Decomposed: IdentityTransform()}},
		{Offset: 1, Transform: TransformValue{Decomposed: Transform{
			Tx: 100, Ty: 20, Sx: 1, Sy: 1, Set: true,
		}}},
	}
	timing := Timing{Duration: time.Second, Iterations: 1}
	p, err := newAnimationPlayer([]Target{target}, frames, timing)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(44, 0)
	p.Start(start)
	p.Step(start.Add(500 * time.Millisecond))
	want := compose(&target, base, Effect{Frames: frames, Timing: timing}, 0.5)
	got := Read(op)
	if got.X != want.X || got.Y != want.Y {
		t.Fatalf("transform position = %v,%v, want %v,%v", got.X, got.Y, want.X, want.Y)
	}
	if got.Matrix != want.Matrix || got.HasMatrix != want.HasMatrix {
		t.Fatalf("transform matrix = %+v (set %v), want %+v (set %v)",
			got.Matrix, got.HasMatrix, want.Matrix, want.HasMatrix)
	}
}

func paintClose(got, want Paint) bool {
	const tolerance = 1e-9
	return math.Abs(got.X-want.X) <= tolerance &&
		math.Abs(got.Y-want.Y) <= tolerance &&
		math.Abs(got.W-want.W) <= tolerance &&
		math.Abs(got.H-want.H) <= tolerance &&
		math.Abs(got.R-want.R) <= tolerance &&
		math.Abs(got.G-want.G) <= tolerance &&
		math.Abs(got.B-want.B) <= tolerance &&
		math.Abs(got.Alpha-want.Alpha) <= tolerance &&
		math.Abs(got.Opacity-want.Opacity) <= tolerance
}

const stepTargets = 4

// liveStepPlayer builds a player whose effects stay live for every frame of a
// benchmark or allocation test, with stops on every sampled field.
func liveStepPlayer(tb testing.TB, effectCount int) (*Player, time.Time) {
	tb.Helper()
	targets := make([]Target, stepTargets)
	for i := range targets {
		op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect,
			X: float64(i * 10), Y: float64(i * 5), W: 40, H: 20, Alpha: 1}
		targets[i] = Target{Op: op, From: Read(op)}
	}
	effects := make([]Effect, effectCount)
	for i := range effects {
		effects[i] = Effect{
			Frames: liveFrames(i),
			Timing: Timing{Duration: time.Hour, Infinite: true},
		}
	}
	player, err := NewComposite(targets, effects)
	if err != nil {
		tb.Fatal(err)
	}
	start := time.Unix(0, 0)
	player.Start(start)
	player.Step(start)
	return player, start
}

// liveFrames returns keyframes with every sampled field present, so the
// measurement covers all six stops of the table.
func liveFrames(seed int) []Keyframe {
	base := float64(seed * stepTargets)
	return []Keyframe{
		{Offset: 0, X: base, Y: base, Opacity: 1, OpacitySet: true,
			R: 0.1, G: 0.2, B: 0.3, ColorSet: true, Move: true},
		{Offset: 0.33, X: base + 3, Y: base + 2, Opacity: 0.7, OpacitySet: true,
			R: 0.4, G: 0.5, B: 0.6, ColorSet: true, Move: true},
		{Offset: 0.66, X: base + 6, Y: base + 4, Opacity: 0.4, OpacitySet: true,
			R: 0.7, G: 0.8, B: 0.9, ColorSet: true, Move: true},
		{Offset: 1, X: base + 9, Y: base + 6, Opacity: 1, OpacitySet: true,
			R: 0.1, G: 0.2, B: 0.3, ColorSet: true, Move: true},
	}
}

func benchmarkStep(b *testing.B, effectCount int) {
	b.Helper()
	player, start := liveStepPlayer(b, effectCount)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		player.Step(start.Add(time.Duration(i) * 16 * time.Millisecond))
	}
}

func BenchmarkSampleOneEffect(b *testing.B) { benchmarkStep(b, 1) }

func BenchmarkSampleTenEffects(b *testing.B) { benchmarkStep(b, 10) }

func TestStepSamplesTransformOpacityAndColor(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 10, Y: 4,
		W: 20, H: 10, R: 0, B: 1, Alpha: 1}
	p, err := newAnimationPlayer([]Target{{Op: op, From: Read(op)}}, []Keyframe{
		{Offset: 1, X: 180, Move: true, Opacity: 1, OpacitySet: true, R: 0, B: 1, ColorSet: true},
		{Offset: .5, X: 90, Move: true, Opacity: .72, OpacitySet: true, R: 226.0 / 255, G: 122.0 / 255, ColorSet: true},
		{Offset: 0, Move: true, Opacity: 1, OpacitySet: true, R: 0, B: 1, ColorSet: true},
	}, Timing{Duration: 2 * time.Second, Iterations: 1})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(10, 0)
	p.Start(start)
	if !p.Step(start.Add(time.Second)) {
		t.Fatal("midpoint should remain active")
	}
	if op.X != 100 || math.Abs(op.Opacity()-.72) > .001 || math.Abs(op.R-226.0/255) > .001 {
		t.Fatalf("sample x=%v opacity=%v rgb=%v,%v,%v", op.X, op.Opacity(), op.R, op.G, op.B)
	}
}

func TestStepUsesEasingInsideKeyframeInterval(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	p, err := newAnimationPlayer([]Target{{Op: op, From: Read(op)}}, []Keyframe{
		{Offset: 0, X: 0, Move: true}, {Offset: .5, X: 90, Move: true},
	}, Timing{Duration: time.Second, Iterations: 1, Easing: func(v float64) float64 { return v * v }})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(20, 0)
	p.Start(start)
	p.Step(start.Add(250 * time.Millisecond))
	if math.Abs(op.X-22.5) > .001 {
		t.Fatalf("x = %v, want 22.5", op.X)
	}
}

// TestTargetMaskLimitsWhatAKeyframeWrites pins the field mask: a target only
// writes the fields its keyframes own, and a write never changes W, H, or the
// operation's alpha.
func TestTargetMaskLimitsWhatAKeyframeWrites(t *testing.T) {
	const (
		wantW      = 10
		wantH      = 10
		wantAlpha  = .4
		wantMatrix = 50
	)
	tests := []struct {
		name       string
		mask       Mask
		wantMatrix bool
		wantColor  bool
	}{
		{"matrix only", MaskMatrix, true, false},
		{"colour only", MaskRGB, false, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect,
				X: 1, Y: 2, W: wantW, H: wantH, R: .1, G: .2, B: .3, Alpha: wantAlpha}
			from := Read(op)
			from.Mask = test.mask

			effect := Effect{Frames: []Keyframe{
				{Offset: 0, Transform: TransformValue{Decomposed: Transform{Set: true}}},
				{Offset: 1, Transform: TransformValue{Decomposed: Transform{Tx: 100, Set: true}},
					R: 1, G: 1, B: 1, ColorSet: true},
			}, Timing: Timing{Duration: time.Second, Iterations: 1, Fill: FillBoth}}
			player, err := NewComposite([]Target{{Op: op, From: from}}, []Effect{effect})
			if err != nil {
				t.Fatalf("bind: %v", err)
			}
			start := time.Unix(0, 0)
			player.Start(start)
			player.Step(start.Add(500 * time.Millisecond))

			if op.XformSet != test.wantMatrix {
				t.Fatalf("matrix written = %v, want %v", op.XformSet, test.wantMatrix)
			}
			if test.wantMatrix && op.Transform().E != wantMatrix {
				t.Fatalf("translation = %v, want %v", op.Transform().E, wantMatrix)
			}
			if got := op.R != .1; got != test.wantColor {
				t.Fatalf("colour written = %v, want %v (R = %v)", got, test.wantColor, op.R)
			}
			if op.Alpha != wantAlpha {
				t.Fatalf("alpha = %v, want the untouched %v", op.Alpha, wantAlpha)
			}
			if op.W != wantW || op.H != wantH {
				t.Fatalf("size = %vx%v, want the untouched %vx%v", op.W, op.H, wantW, wantH)
			}
		})
	}
}

// TestMissingKeyframeValuesUseTheUnderlyingPaint checks that a property
// omitted from a keyframe is not sampled as zero. A color omitted at 100%
// must hold the underlying color, and a transform omitted there must leave
// the operation where it started.
func TestMissingKeyframeValuesUseTheUnderlyingPaint(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 40, Y: 12,
		R: .42, G: .62, B: .21, Alpha: 1}
	base := Read(op)
	p, err := newAnimationPlayer([]Target{{Op: op, From: base}}, []Keyframe{
		{Offset: 0, Move: true},
		{Offset: .5, Move: true, X: 120, R: 1, ColorSet: true},
		{Offset: 1},
	}, Timing{Duration: time.Second, Iterations: 1, Fill: FillBoth,
		TimingFunction: "linear"})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(70, 0)
	p.Start(start)
	p.Step(start.Add(3 * time.Second))
	if math.Abs(op.R-base.R) > .001 || math.Abs(op.G-base.G) > .001 || math.Abs(op.B-base.B) > .001 {
		t.Fatalf("final color = %v,%v,%v, want underlying %v,%v,%v", op.R, op.G, op.B, base.R, base.G, base.B)
	}
	if op.X != base.X || op.Y != base.Y {
		t.Fatalf("final position = %v,%v, want underlying %v,%v", op.X, op.Y, base.X, base.Y)
	}
}

// threeDValue builds a sampled 3D list from its functions.
func threeDValue(functions ...TransformFunction) TransformValue {
	return TransformValue{Decomposed: IdentityTransform(), Functions: functions, ThreeD: true}
}

func TestTransformCSSSerializesThreeDimensionalFunctions(t *testing.T) {
	value := threeDValue(
		TransformFunction{Name: RotateY, Args: []float64{math.Pi / 4}},
		TransformFunction{Name: TranslateZ, Args: []float64{15}},
	)

	// 15 layout points are 20 CSS pixels, the unit the engine parses.
	if got, want := TransformCSS(value), "rotateY(45deg) translateZ(20px)"; got != want {
		t.Fatalf("TransformCSS = %q, want %q", got, want)
	}
}

func TestTransformCSSSerializesEveryFunctionUnit(t *testing.T) {
	tests := []struct {
		name string
		fn   TransformFunction
		want string
	}{
		{"translate", TransformFunction{Name: Translate, Args: []float64{7.5, 15}}, "translate(10px, 20px)"},
		{"translateX", TransformFunction{Name: TranslateX, Args: []float64{7.5}}, "translateX(10px)"},
		{"translateY", TransformFunction{Name: TranslateY, Args: []float64{15}}, "translateY(20px)"},
		{"translate3d", TransformFunction{Name: Translate3D, Args: []float64{7.5, 15, 22.5}},
			"translate3d(10px, 20px, 30px)"},
		{"translateZ", TransformFunction{Name: TranslateZ, Args: []float64{15}}, "translateZ(20px)"},
		{"scale", TransformFunction{Name: Scale, Args: []float64{2, 3}}, "scale(2, 3)"},
		{"scaleX", TransformFunction{Name: ScaleX, Args: []float64{2}}, "scaleX(2)"},
		{"scaleY", TransformFunction{Name: ScaleY, Args: []float64{3}}, "scaleY(3)"},
		{"scale3d", TransformFunction{Name: Scale3D, Args: []float64{2, 3, 4}}, "scale3d(2, 3, 4)"},
		{"scaleZ", TransformFunction{Name: ScaleZ, Args: []float64{4}}, "scaleZ(4)"},
		{"rotate", TransformFunction{Name: Rotate, Args: []float64{math.Pi / 2}}, "rotate(90deg)"},
		{"rotateX", TransformFunction{Name: RotateX, Args: []float64{math.Pi / 4}}, "rotateX(45deg)"},
		{"rotateY", TransformFunction{Name: RotateY, Args: []float64{math.Pi / 8}}, "rotateY(22.5deg)"},
		{"rotateZ", TransformFunction{Name: RotateZ, Args: []float64{math.Pi}}, "rotateZ(180deg)"},
		{"rotate3d", TransformFunction{Name: Rotate3D, Args: []float64{0, 1, 0, math.Pi / 2}},
			"rotate3d(0, 1, 0, 90deg)"},
		{"skew", TransformFunction{Name: Skew, Args: []float64{math.Pi / 4, 0}}, "skew(45deg, 0deg)"},
		{"skewX", TransformFunction{Name: SkewX, Args: []float64{math.Pi / 4}}, "skewX(45deg)"},
		{"skewY", TransformFunction{Name: SkewY, Args: []float64{0}}, "skewY(0deg)"},
		{"matrix", TransformFunction{Name: Matrix, Args: []float64{1, 0, 0, 1, 7.5, 15}},
			"matrix(1, 0, 0, 1, 10px, 20px)"},
		{"matrix3d", TransformFunction{Name: Matrix3D,
			Args: []float64{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 7.5, 15, 22.5, 1}},
			"matrix3d(1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 7.5, 15, 22.5, 1)"},
		{"perspective", TransformFunction{Name: Perspective, Args: []float64{375}}, "perspective(500px)"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := TransformCSS(threeDValue(test.fn)); got != test.want {
				t.Fatalf("TransformCSS = %q, want %q", got, test.want)
			}
		})
	}
}

func TestTransformCSSWritesATwoDimensionalValueAsAMatrix(t *testing.T) {
	value := TransformValue{
		Decomposed: Transform{Tx: 7.5, Ty: 15, Sx: 2, Sy: 3, Set: true},
	}

	// T(7.5,15) * S(2,3) in points, with the translation written in pixels.
	if got, want := TransformCSS(value), "matrix(2, 0, 0, 3, 10px, 20px)"; got != want {
		t.Fatalf("TransformCSS = %q, want %q", got, want)
	}
}

func TestTransformCSSWritesAnUnsetValueAsTheIdentity(t *testing.T) {
	if got, want := TransformCSS(TransformValue{}), "matrix(1, 0, 0, 1, 0px, 0px)"; got != want {
		t.Fatalf("TransformCSS = %q, want %q", got, want)
	}
}

func TestRestampValuesCarryTheSampledThreeDimensionalList(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	p, err := newAnimationPlayer([]Target{{Op: op, From: Read(op)}},
		[]Keyframe{
			{Offset: 0, Transform: rotateValue(RotateY, 0, true)},
			{Offset: 1, Transform: rotateValue(RotateY, math.Pi/2, true)},
		},
		Timing{Duration: time.Second, Iterations: 1, TimingFunction: "linear"})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(70, 0)
	p.Start(start)
	p.Step(start.Add(500 * time.Millisecond))

	values := p.RestampValues()
	if len(values) != 1 {
		t.Fatalf("restamp values = %d, want one per target", len(values))
	}
	if values[0].Target != 0 || values[0].Transform != "rotateY(45deg)" {
		t.Fatalf("restamp value = %+v, want target 0 at %q", values[0], "rotateY(45deg)")
	}
}

func TestRestampValuesAreNilForATwoDimensionalSample(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	p, err := newAnimationPlayer([]Target{{Op: op, From: Read(op)}},
		[]Keyframe{
			{Offset: 0, Transform: translateValue(0, 0)},
			{Offset: 1, Transform: translateValue(20, 10)},
		},
		Timing{Duration: time.Second, Iterations: 1, TimingFunction: "linear"})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(71, 0)
	p.Start(start)
	p.Step(start.Add(500 * time.Millisecond))

	if values := p.RestampValues(); values != nil {
		t.Fatalf("a 2D sample reported restamp values %+v", values)
	}
}

func TestRestampValuesFollowTheLastEffectOnAnElement(t *testing.T) {
	// A 3D effect followed by a 2D effect on the same element is a 2D frame:
	// the last effect wins, so the restamp list is empty.
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	target := Target{Op: op, From: Read(op)}
	threeD := Effect{
		Frames: []Keyframe{
			{Offset: 0, Transform: rotateValue(RotateY, 0, true)},
			{Offset: 1, Transform: rotateValue(RotateY, math.Pi, true)},
		},
		Timing: Timing{Duration: time.Second, Iterations: 1, TimingFunction: "linear"},
	}
	twoD := Effect{
		Frames: []Keyframe{
			{Offset: 0, Transform: translateValue(0, 0)},
			{Offset: 1, Transform: translateValue(10, 0)},
		},
		Timing: Timing{Duration: time.Second, Iterations: 1, TimingFunction: "linear"},
	}
	p, err := NewComposite([]Target{target}, []Effect{threeD, twoD})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(72, 0)
	p.Start(start)
	p.Step(start.Add(500 * time.Millisecond))

	if values := p.RestampValues(); values != nil {
		t.Fatalf("a 2D frame kept restamp values %+v", values)
	}
	if p.RestampPending() {
		t.Fatal("a 2D frame kept the restamp flag set")
	}
}

func TestRestampValuesSerializeATwoDimensionalSampleOnAThreeDElement(t *testing.T) {
	// A 2D keyframe on an element that already declares a 3D transform still
	// needs the engine pass, and the sampled matrix is the value it stamps.
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	target := Target{Op: op, From: Read(op), HasTransform3D: true}
	p, err := newAnimationPlayer([]Target{target},
		[]Keyframe{
			{Offset: 0, Transform: translateValue(0, 0)},
			{Offset: 1, Transform: translateValue(7.5, 0)},
		},
		Timing{Duration: time.Second, Iterations: 1, TimingFunction: "linear"})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(73, 0)
	p.Start(start)
	p.Step(start.Add(500 * time.Millisecond))

	values := p.RestampValues()
	if len(values) != 1 {
		t.Fatalf("restamp values = %d, want 1", len(values))
	}
	// Half of 7.5 points is 5 CSS pixels; the rest of the matrix is identity.
	if got, want := values[0].Transform, "matrix(1, 0, 0, 1, 5px, 0px)"; got != want {
		t.Fatalf("restamp transform = %q, want %q", got, want)
	}
}

func TestWantsRestampReadsTheKeyframesBeforeAnySample(t *testing.T) {
	threeD := []Keyframe{
		{Offset: 0, Transform: rotateValue(RotateY, 0, true)},
		{Offset: 1, Transform: rotateValue(RotateY, math.Pi, true)},
	}
	twoD := []Keyframe{
		{Offset: 0, Transform: translateValue(0, 0)},
		{Offset: 1, Transform: translateValue(10, 0)},
	}
	timing := Timing{Duration: time.Second}

	for _, test := range []struct {
		name    string
		target  Target
		frames  []Keyframe
		expects bool
	}{
		{name: "3D keyframe", target: Target{Op: &layout.DisplayOp{}}, frames: threeD, expects: true},
		{name: "2D keyframe", target: Target{Op: &layout.DisplayOp{}}, frames: twoD, expects: false},
		{name: "3D element", target: Target{Op: &layout.DisplayOp{}, HasTransform3D: true},
			frames: twoD, expects: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			p, err := newAnimationPlayer([]Target{test.target}, test.frames, timing)
			if err != nil {
				t.Fatal(err)
			}
			if got := p.WantsRestamp(); got != test.expects {
				t.Fatalf("WantsRestamp = %v, want %v", got, test.expects)
			}
		})
	}
}
