package main

import (
	"math"
	"testing"
)

func TestCSSPreviewTimingFunctions(t *testing.T) {
	cases := []struct {
		value    string
		at, want float64
	}{
		{"linear", .5, .5}, {"ease-in-out", .5, .5}, {"steps(5, end)", .5, .4},
		{"linear(0, 0.35 30%, 0.8 65%, 1)", .5, 0.6071428571428572},
	}
	for _, tc := range cases {
		if got := easing(tc.value, tc.at); math.Abs(got-tc.want) > 1e-6 {
			t.Errorf("easing(%q, %v) = %v, want %v", tc.value, tc.at, got, tc.want)
		}
	}
	if got := easing("cubic-bezier(0.2, 0.7, 0.2, 1)", .5); got <= .5 || got >= 1 {
		t.Fatalf("cubic-bezier sample = %v", got)
	}
}
