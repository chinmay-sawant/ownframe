package main

import "testing"

// TestComboMultiplierFloor pins the multiplier rule: no ratio from a sum
// inside timer noise, and none from a combination that did not cost time.
func TestComboMultiplierFloor(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		name       string
		delta, sum int64
		want       float64
	}{
		{"above floor", 300_000, 100_000, 3},
		{"sum at floor", 300_000, comboMultiplierMinNS, 0},
		{"sum below floor", 300_000, 1_000, 0},
		{"zero delta", 0, 100_000, 0},
		{"negative delta", -20_000, 100_000, 0},
	} {
		if got := comboMultiplier(c.delta, c.sum); got != c.want {
			t.Fatalf("%s: comboMultiplier(%d, %d) = %v, want %v", c.name, c.delta, c.sum, got, c.want)
		}
	}

	entry := comboEntry{Scale64DeltaNS: 160_000, HoverDeltaNS: 900_000, RasterDeltaNS: -5_000}
	applyComboMultipliers(&entry, comboSums{scale: 100_000, hover: 10_000, raster: 200_000})
	if entry.Scale64Multiplier != 1.6 || entry.HoverMultiplier != 0 || entry.RasterMultiplier != 0 {
		t.Fatalf("multipliers = %v/%v/%v, want 1.6/0/0", entry.Scale64Multiplier, entry.HoverMultiplier, entry.RasterMultiplier)
	}
	if !entry.SuperLinear || entry.MaxMultiplier != 1.6 {
		t.Fatalf("super_linear = %t max = %v, want true 1.6", entry.SuperLinear, entry.MaxMultiplier)
	}

	noisy := comboEntry{HoverDeltaNS: 900_000}
	applyComboMultipliers(&noisy, comboSums{hover: 2_000})
	if noisy.SuperLinear {
		t.Fatal("a sub-floor denominator set super_linear")
	}
}

// TestComboSingleCacheMeasuresOncePerSampling checks that each single is
// measured once per sample count and GC mode, that a refined pair gets
// singles at its own sampling, and that failed singles drop out of the sum.
func TestComboSingleCacheMeasuresOncePerSampling(t *testing.T) {
	t.Parallel()
	calls := map[singleKey]int{}
	cache := newComboSingleCache(func(spec comboSpec, iters int, highPrecision bool) comboEntry {
		calls[singleKey{spec.BaseStyle, iters, highPrecision}]++
		if spec.Kind != "single" || spec.BaseStyle != spec.HoverStyle {
			t.Fatalf("single spec = %+v", spec)
		}
		if spec.BaseStyle == "bad:x;" {
			return comboEntry{Err: "boom", Scale64DeltaNS: 1 << 40}
		}
		return comboEntry{Scale64DeltaNS: 100_000, HoverDeltaNS: -50_000, RasterDeltaNS: int64(iters) * 1_000}
	})
	pair := comboSpec{Singles: []string{"a:1;", "b:2;"}}
	stack := comboSpec{Singles: []string{"a:1;", "bad:x;"}}

	if s := cache.sums(pair, 1, false); s != (comboSums{scale: 200_000, hover: 0, raster: 2_000}) {
		t.Fatalf("sweep sums = %+v", s)
	}
	_ = cache.sums(pair, 1, false)
	if s := cache.sums(stack, 1, false); s.scale != 100_000 {
		t.Fatalf("failed single counted: %+v", s)
	}
	if s := cache.sums(pair, 3, true); s.raster != 6_000 {
		t.Fatalf("refined sums = %+v, want singles at 3 iters", s)
	}
	for key, n := range calls {
		if n != 1 {
			t.Fatalf("single %+v measured %d times", key, n)
		}
	}
	if len(calls) != 5 {
		t.Fatalf("measured %d singles, want 5 (a,b,bad at sweep; a,b refined)", len(calls))
	}
}
