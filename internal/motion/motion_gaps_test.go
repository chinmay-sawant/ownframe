package motion

import (
	"math"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestBorderRadiusInPlaceAndZeroAllocs(t *testing.T) {
	t0 := time.Unix(100, 0)
	const ppp = 0.75
	d := &layout.Display{
		PointsPerPixel: ppp,
		Boxes: []layout.Box{{
			ID:           "card",
			X:            10,
			Y:            10,
			W:            140,
			H:            60,
			OperationIDs: []uint64{1, 2},
			Animations: []layout.Animation{{
				Name:              "morph",
				DurationMS:        1000,
				IterationInfinite: true,
				TimingFunction:    "linear",
				FillMode:          "both",
			}},
		}},
		Ops: []layout.DisplayOp{
			{ID: 1, Kind: layout.DisplayOpFillRect, X: 7.5, Y: 7.5, W: 105, H: 45, Radius: 3, RadiusY: 3, R: 0.1, G: 0.3, B: 0.8, Alpha: 1},
			{ID: 2, Kind: layout.DisplayOpStrokeRect, X: 7.5, Y: 7.5, W: 105, H: 45, Radius: 3, RadiusY: 3, Width: 1, R: 0, G: 0, B: 0, Alpha: 1},
		},
		OpOwners: []int{0, 0},
		Keyframes: []layout.KeyframesRule{{
			Name: "morph",
			Keyframes: []layout.Keyframe{
				{Offset: 0, Declarations: []layout.KeyframeDeclaration{{Property: "border-radius", Value: "4px"}}},
				{Offset: 1, Declarations: []layout.KeyframeDeclaration{{Property: "border-radius", Value: "24px"}}},
			},
		}},
	}

	r := New()
	r.now = func() time.Time { return t0 }
	if err := r.Bind(d); err != nil {
		t.Fatalf("Bind failed: %v", err)
	}
	if stats := r.Stats(); stats.LayoutBound != 0 {
		t.Fatalf("LayoutBound = %d, want 0 (uniform px border-radius should animate in place)", stats.LayoutBound)
	}
	if stats := r.Stats(); stats.Bound != 1 {
		t.Fatalf("Bound = %d, want 1", stats.Bound)
	}

	// Sample at 50%: midpoint of 4px and 24px is 14px -> 14 * 0.75 = 10.5pt.
	r.Step(d, t0.Add(500*time.Millisecond))
	wantRadius := 14.0 * ppp
	for i, op := range d.Ops {
		if math.Abs(op.Radius-wantRadius) > 1e-6 || math.Abs(op.RadiusY-wantRadius) > 1e-6 {
			t.Fatalf("op[%d] Radius=(%v,%v), want (%v,%v)", i, op.Radius, op.RadiusY, wantRadius, wantRadius)
		}
	}

	// Verify 0 allocs/op on steady-state Step.
	now := t0.Add(500 * time.Millisecond)
	allocs := testing.AllocsPerRun(20, func() {
		now = now.Add(16 * time.Millisecond)
		r.Step(d, now)
	})
	if allocs != 0 {
		t.Fatalf("Step allocs = %v, want 0", allocs)
	}
}

func TestBorderRadiusComplexStaysOnRelayout(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		prop string
		from string
		to   string
	}{
		{"multi-value", "border-radius", "4px 8px", "12px 24px"},
		{"elliptical", "border-radius", "4px / 8px", "12px / 24px"},
		{"percentage", "border-radius", "10%", "50%"},
		{"per-corner", "border-top-left-radius", "4px", "24px"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			d := &layout.Display{
				PointsPerPixel: 0.75,
				Boxes: []layout.Box{{
					ID:           "card",
					W:            100,
					H:            50,
					OperationIDs: []uint64{1},
					Animations: []layout.Animation{{
						Name:       "morph",
						DurationMS: 1000,
					}},
				}},
				Ops:      []layout.DisplayOp{{ID: 1, Kind: layout.DisplayOpFillRect, W: 75, H: 37.5, Alpha: 1}},
				OpOwners: []int{0},
				Keyframes: []layout.KeyframesRule{{
					Name: "morph",
					Keyframes: []layout.Keyframe{
						{Offset: 0, Declarations: []layout.KeyframeDeclaration{{Property: tc.prop, Value: tc.from}}},
						{Offset: 1, Declarations: []layout.KeyframeDeclaration{{Property: tc.prop, Value: tc.to}}},
					},
				}},
			}
			r := New()
			if err := r.Bind(d); err != nil {
				t.Fatalf("Bind failed: %v", err)
			}
			if got := r.Stats().LayoutBound; got != 1 {
				t.Fatalf("LayoutBound = %d, want 1 for %s", got, tc.name)
			}
		})
	}
}

func TestColorAlphaInPlaceAndZeroAllocs(t *testing.T) {
	t0 := time.Unix(100, 0)
	d := &layout.Display{
		PointsPerPixel: 1,
		Boxes: []layout.Box{{
			ID:           "badge",
			W:            120,
			H:            40,
			OperationIDs: []uint64{1},
			Animations: []layout.Animation{{
				Name:              "glow",
				DurationMS:        1000,
				IterationInfinite: true,
				TimingFunction:    "linear",
				FillMode:          "both",
			}},
		}},
		Ops:      []layout.DisplayOp{{ID: 1, Kind: layout.DisplayOpFillRect, W: 120, H: 40, R: 61.0 / 255, G: 220.0 / 255, B: 132.0 / 255, Alpha: 0.1}},
		OpOwners: []int{0},
		Keyframes: []layout.KeyframesRule{{
			Name: "glow",
			Keyframes: []layout.Keyframe{
				{Offset: 0, Declarations: []layout.KeyframeDeclaration{{Property: "background-color", Value: "rgba(61, 220, 132, 0.1)"}}},
				{Offset: 1, Declarations: []layout.KeyframeDeclaration{{Property: "background-color", Value: "rgba(61, 220, 132, 0.9)"}}},
			},
		}},
	}

	r := New()
	r.now = func() time.Time { return t0 }
	if err := r.Bind(d); err != nil {
		t.Fatalf("Bind failed: %v", err)
	}
	if gaps := r.Stats().Gaps; len(gaps) != 0 {
		t.Fatalf("Gaps = %v, want none", gaps)
	}

	r.Step(d, t0.Add(500*time.Millisecond))
	if math.Abs(d.Ops[0].Alpha-0.5) > 1e-6 {
		t.Fatalf("Alpha at 50%% = %v, want 0.5", d.Ops[0].Alpha)
	}

	now := t0.Add(500 * time.Millisecond)
	allocs := testing.AllocsPerRun(20, func() {
		now = now.Add(16 * time.Millisecond)
		r.Step(d, now)
	})
	if allocs != 0 {
		t.Fatalf("Step allocs = %v, want 0", allocs)
	}
}

func TestIndividualTransformProperties2DAndZeroAllocs(t *testing.T) {
	t0 := time.Unix(100, 0)
	const ppp = 0.75

	t.Run("translate", func(t *testing.T) {
		d := &layout.Display{
			PointsPerPixel: ppp,
			Boxes: []layout.Box{{
				ID:           "chip",
				W:            100,
				H:            40,
				OperationIDs: []uint64{1},
				Animations: []layout.Animation{{
					Name:              "slide",
					DurationMS:        1000,
					IterationInfinite: true,
					TimingFunction:    "linear",
					FillMode:          "both",
				}},
			}},
			Ops:      []layout.DisplayOp{{ID: 1, Kind: layout.DisplayOpFillRect, W: 75, H: 30, Alpha: 1}},
			OpOwners: []int{0},
			Keyframes: []layout.KeyframesRule{{
				Name: "slide",
				Keyframes: []layout.Keyframe{
					{Offset: 0, Declarations: []layout.KeyframeDeclaration{{Property: "translate", Value: "0px 0px"}}},
					{Offset: 1, Declarations: []layout.KeyframeDeclaration{{Property: "translate", Value: "80px 20px"}}},
				},
			}},
		}

		r := New()
		r.now = func() time.Time { return t0 }
		if err := r.Bind(d); err != nil {
			t.Fatalf("Bind failed: %v", err)
		}
		if stats := r.Stats(); len(stats.Gaps) != 0 || stats.NeedsRetainedState || stats.Bound != 1 {
			t.Fatalf("unexpected stats: %+v", stats)
		}

		r.Step(d, t0.Add(500*time.Millisecond))
		m := d.Ops[0].Transform()
		wantTx, wantTy := 40.0*ppp, 10.0*ppp
		if math.Abs(m.E-wantTx) > 1e-6 || math.Abs(m.F-wantTy) > 1e-6 {
			t.Fatalf("translate midpoint = (%v,%v), want (%v,%v)", m.E, m.F, wantTx, wantTy)
		}
		if entries := r.LayerEntries(); len(entries) != 1 || !entries[0].Has2DTransform {
			t.Fatalf("LayerEntries = %+v, want 1 promotable 2D transform entry", entries)
		}

		now := t0.Add(500 * time.Millisecond)
		allocs := testing.AllocsPerRun(20, func() {
			now = now.Add(16 * time.Millisecond)
			r.Step(d, now)
		})
		if allocs != 0 {
			t.Fatalf("Step allocs = %v, want 0", allocs)
		}
	})

	t.Run("rotate_and_scale_combined", func(t *testing.T) {
		d := &layout.Display{
			PointsPerPixel: 1,
			Boxes: []layout.Box{{
				ID:           "spin",
				W:            100,
				H:            100,
				OperationIDs: []uint64{1},
				Animations: []layout.Animation{{
					Name:           "combo",
					DurationMS:     1000,
					IterationCount: 1,
					TimingFunction: "linear",
					FillMode:       "both",
				}},
			}},
			Ops:      []layout.DisplayOp{{ID: 1, Kind: layout.DisplayOpFillRect, W: 100, H: 100, Alpha: 1}},
			OpOwners: []int{0},
			Keyframes: []layout.KeyframesRule{{
				Name: "combo",
				Keyframes: []layout.Keyframe{
					{Offset: 0, Declarations: []layout.KeyframeDeclaration{
						{Property: "translate", Value: "0px 0px 0px"},
						{Property: "rotate", Value: "0deg"},
						{Property: "scale", Value: "1 1 1"},
					}},
					{Offset: 1, Declarations: []layout.KeyframeDeclaration{
						{Property: "translate", Value: "40px 0px 0px"},
						{Property: "rotate", Value: "90deg"},
						{Property: "scale", Value: "2 2 1"},
					}},
				},
			}},
		}

		r := New()
		r.now = func() time.Time { return t0 }
		if err := r.Bind(d); err != nil {
			t.Fatalf("Bind failed: %v", err)
		}
		if stats := r.Stats(); stats.NeedsRetainedState || len(stats.Gaps) != 0 {
			t.Fatalf("planar translate/rotate/scale should stay 2D with no gaps: %+v", stats)
		}

		r.Step(d, t0.Add(500*time.Millisecond))
		dec := decomposeMatrix2D(d.Ops[0].Transform())
		if math.Abs(dec.Tx-20) > 1e-6 || math.Abs(dec.Ty) > 1e-6 {
			t.Fatalf("Tx,Ty = (%v,%v), want (20,0)", dec.Tx, dec.Ty)
		}
		if math.Abs(dec.Rotate-math.Pi/4) > 1e-6 {
			t.Fatalf("Rotate = %v, want %v", dec.Rotate, math.Pi/4)
		}
		if math.Abs(dec.Sx-1.5) > 1e-6 || math.Abs(dec.Sy-1.5) > 1e-6 {
			t.Fatalf("Scale = (%v,%v), want (1.5,1.5)", dec.Sx, dec.Sy)
		}
	})

	t.Run("3d_individual_triggers_restamp", func(t *testing.T) {
		for _, tc := range []struct {
			prop string
			val  string
		}{
			{"translate", "10px 20px 30px"},
			{"rotate", "1 0 0 45deg"},
			{"scale", "1 1 2"},
		} {
			d := &layout.Display{
				PointsPerPixel: 1,
				Boxes: []layout.Box{{
					ID:           "box3d",
					W:            100,
					H:            100,
					OperationIDs: []uint64{1},
					Animations:   []layout.Animation{{Name: "a", DurationMS: 1000}},
				}},
				Ops:      []layout.DisplayOp{{ID: 1, Kind: layout.DisplayOpFillRect, W: 100, H: 100, Alpha: 1}},
				OpOwners: []int{0},
				Keyframes: []layout.KeyframesRule{{
					Name: "a",
					Keyframes: []layout.Keyframe{
						{Offset: 0, Declarations: []layout.KeyframeDeclaration{{Property: tc.prop, Value: "none"}}},
						{Offset: 1, Declarations: []layout.KeyframeDeclaration{{Property: tc.prop, Value: tc.val}}},
					},
				}},
			}
			r := New()
			if err := r.Bind(d); err != nil {
				t.Fatalf("%s bind failed: %v", tc.prop, err)
			}
			if !r.Stats().NeedsRetainedState {
				t.Fatalf("%s=%q should require 3D restamp", tc.prop, tc.val)
			}
		}
	})
}
