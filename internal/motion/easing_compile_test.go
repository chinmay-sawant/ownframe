package motion

import "testing"

// TestCompileEaseMatchesEaseAndAllocsZero pins PERF-MOT-01: compiled samplers
// agree with Ease and sample without allocating.
func TestCompileEaseMatchesEaseAndAllocsZero(t *testing.T) {
	for _, value := range []string{
		"ease", "ease-in-out", "step-end", "steps(4, jump-both)",
		"cubic-bezier(0.1, 0.7, 1, 0.1)", "linear(0, 0.25 40%, 1)", "bogus(1)", "",
	} {
		sample := compileEase(value)
		for _, p := range []float64{-1, 0, 0.1, 0.5, 0.99, 1, 2} {
			if got, want := sample(p), compileEase(value)(p); got != want || Ease(value, p) != want {
				t.Fatalf("%q at %v: compiled %v, Ease %v", value, p, got, want)
			}
		}
		Ease(value, 0.5)
		allocs := testing.AllocsPerRun(100, func() { sample(0.37); Ease(value, 0.37) })
		if allocs != 0 {
			t.Fatalf("%q sampling allocs/op = %v, want 0", value, allocs)
		}
	}
}
