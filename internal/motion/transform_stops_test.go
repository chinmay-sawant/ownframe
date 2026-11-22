package motion

import (
	"reflect"
	"testing"
)

// referenceSampleTransform is the materialized algorithm sampleTransform
// replaced: collect every transform stop, pad the identity endpoints, and
// walk the list. The equivalence test holds the in-place sampler to it.
func referenceSampleTransform(e Effect, at float64) (TransformValue, bool) {
	stops := make([]transformStop, 0, len(e.Frames)+extraStops)
	for _, frame := range e.Frames {
		if value, ok := frame.transformValue(); ok {
			stops = append(stops, transformStop{at: frame.Offset, value: value, ease: frame.frameEasing(nil)})
		}
	}
	if len(stops) == 0 {
		return TransformValue{}, false
	}
	fallback := TransformValue{Decomposed: IdentityTransform()}
	if stops[0].at > 0 {
		stops = append([]transformStop{{at: 0, value: fallback}}, stops...)
	}
	if stops[len(stops)-1].at < 1 {
		stops = append(stops, transformStop{at: 1, value: fallback})
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
		part := segmentEase(prev.ease, e.Timing, (at-prev.at)/span)
		value, err := interpolateTransform(prev.value, next.value, part)
		if err != nil {
			if part < discreteMidpoint {
				return prev.value, true
			}
			return next.value, true
		}
		return value, true
	}
	return stops[len(stops)-1].value, true
}

// TestSampleTransformMatchesMaterializedStops compares the in-place sampler
// with the materialized list across padded and unpadded endpoints, duplicate
// offsets, move-only frames, keyframe easing, and frames with no transform.
func TestSampleTransformMatchesMaterializedStops(t *testing.T) {
	cases := map[string][]Keyframe{
		"from-to":       {{Offset: 0, Transform: translateValue(0, 0)}, {Offset: 1, Transform: translateValue(12, 0)}},
		"padded both":   {{Offset: 0.25, Transform: translateValue(4, 0)}, {Offset: 0.75, Transform: translateValue(8, 0)}},
		"single middle": {{Offset: 0.5, Transform: translateValue(10, 0)}},
		"duplicate offset": {
			{Offset: 0, Transform: translateValue(0, 0)}, {Offset: 0.5, Transform: translateValue(4, 0)},
			{Offset: 0.5, Transform: translateValue(9, 0)}, {Offset: 1, Transform: translateValue(2, 0)},
		},
		"skip opacity frames": {
			{Offset: 0, Opacity: 0.2, OpacitySet: true}, {Offset: 0.3, Transform: translateValue(6, 0)},
			{Offset: 0.6, Opacity: 1, OpacitySet: true}, {Offset: 1, Transform: translateValue(1, 0)},
		},
		"move only": {{Offset: 0, Move: true, X: 1, Y: 2}, {Offset: 1, Move: true, X: 9, Y: -3}},
		"keyframe easing": {
			{Offset: 0, Transform: translateValue(0, 0), TimingFunction: "ease-in"},
			{Offset: 0.4, Transform: translateValue(5, 0), TimingFunction: "steps(2)"}, {Offset: 1, Transform: translateValue(10, 0)},
		},
		"no transform": {{Offset: 0, Opacity: 0, OpacitySet: true}, {Offset: 1, Opacity: 1, OpacitySet: true}},
	}
	for name, frames := range cases {
		effect := Effect{Frames: frames, Timing: Timing{TimingFunction: "ease-out"}}
		for i := -2; i <= 22; i++ {
			at := float64(i) / 20
			got, gotOK := effect.sampleTransform(at)
			want, wantOK := referenceSampleTransform(effect, at)
			if gotOK != wantOK || !reflect.DeepEqual(got, want) {
				t.Fatalf("%s at %.2f = %+v, %t; want %+v, %t", name, at, got, gotOK, want, wantOK)
			}
		}
	}
}

// BenchmarkSampleTransform measures one per-tick transform sample. The stop
// list is walked in place, so only the blended function list allocates.
func BenchmarkSampleTransform(b *testing.B) {
	effect := Effect{Frames: []Keyframe{
		{Offset: 0, Transform: translateValue(0, 0)}, {Offset: 1, Transform: translateValue(12, 0)},
	}, Timing: Timing{TimingFunction: "linear"}}
	b.ReportAllocs()
	for i := range b.N {
		if _, ok := effect.sampleTransform(float64(i%100) / 100); !ok {
			b.Fatal("no sample")
		}
	}
}
