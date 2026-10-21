package main

import "math"

// nearColor reports whether a paint carries the given 8 bit colour.
func nearColor(paint paint, r, g, b uint8) bool {
	want := [3]float64{float64(r) / 255, float64(g) / 255, float64(b) / 255}
	got := [3]float64{paint.R, paint.G, paint.B}
	for i := range want {
		if math.Abs(want[i]-got[i]) > 0.01 {
			return false
		}
	}
	return true
}
