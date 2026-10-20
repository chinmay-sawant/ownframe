package transition

import "testing"

func TestEasingFunctions(t *testing.T) {
	for _, test := range []struct {
		name string
		at   float64
		want float64
	}{
		{name: "linear", at: .5, want: .5},
		{name: "steps(4, end)", at: .5, want: .5},
		{name: "linear(0, 1)", at: .25, want: .25},
	} {
		if got := easing(test.name, test.at); got != test.want {
			t.Errorf("easing(%q, %g) = %g, want %g", test.name, test.at, got, test.want)
		}
	}
	if got := easing("ease", .5); got <= 0 || got >= 1 {
		t.Fatalf("ease midpoint = %g, want an interior value", got)
	}
}
