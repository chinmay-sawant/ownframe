package animation

import (
	"math"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
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
	if table.mask != 1<<fieldX|1<<fieldY {
		t.Fatalf("mask = %b, want only the move fields", table.mask)
	}
	field := table.fields[fieldX]
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
	for other := fieldOpacity; other < fieldCount; other++ {
		if table.fields[other].present {
			t.Fatalf("field %d has stops but no keyframe carries it", other)
		}
	}

	sparse := buildEffectTable(Effect{Frames: []Keyframe{{Offset: 0.5, X: 30, Move: true}}})
	stops := sparse.fields[fieldX].stops
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
	field := &table.fields[fieldX]
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
	p, err := New([]Target{{Op: op, From: cssmotion.Read(op)}}, []Keyframe{
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
	field := &table.fields[fieldX]
	for _, at := range []float64{0.9, 0.6, 0.3, 0.1, 0.45} {
		got, _ := field.sample(at)
		fresh := buildEffectTable(e).fields[fieldX]
		want, _ := fresh.sample(at)
		if math.Abs(got-want) > 1e-9 {
			t.Fatalf("at %v x = %v, want %v", at, got, want)
		}
	}

	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	p, err := New([]Target{{Op: op, From: cssmotion.Read(op)}}, []Keyframe{
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
	base := cssmotion.Paint{X: 12, Y: 7, W: 40, H: 20,
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
	base := cssmotion.Read(op)
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
	if !paintClose(cssmotion.Read(op), want) {
		t.Fatalf("paint = %+v, want %+v", cssmotion.Read(op), want)
	}
}

// TestStepKeepsParsedTransformOnCompose checks that an effect carrying a
// parsed transform still reaches the compose oracle. The six-field table only
// covers the scalar fields; transform sampling stays in phase 02 until it has
// its own precomputed stops.
func TestStepKeepsParsedTransformOnCompose(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 10, Y: 5, Alpha: 1}
	base := cssmotion.Read(op)
	target := Target{Op: op, From: base}
	frames := []Keyframe{
		{Offset: 0, Transform: TransformValue{Decomposed: IdentityTransform()}},
		{Offset: 1, Transform: TransformValue{Decomposed: Transform{
			Tx: 100, Ty: 20, Sx: 1, Sy: 1, Set: true,
		}}},
	}
	timing := Timing{Duration: time.Second, Iterations: 1}
	p, err := New([]Target{target}, frames, timing)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(44, 0)
	p.Start(start)
	p.Step(start.Add(500 * time.Millisecond))
	want := compose(&target, base, Effect{Frames: frames, Timing: timing}, 0.5)
	got := cssmotion.Read(op)
	if got.X != want.X || got.Y != want.Y {
		t.Fatalf("transform position = %v,%v, want %v,%v", got.X, got.Y, want.X, want.Y)
	}
	if got.Matrix != want.Matrix || got.HasMatrix != want.HasMatrix {
		t.Fatalf("transform matrix = %+v (set %v), want %+v (set %v)",
			got.Matrix, got.HasMatrix, want.Matrix, want.HasMatrix)
	}
}

func paintClose(got, want cssmotion.Paint) bool {
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
