package motion

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
		if got := Ease(test.name, test.at); got != test.want {
			t.Errorf("Ease(%q, %g) = %g, want %g", test.name, test.at, got, test.want)
		}
	}
	if got := Ease("ease", .5); got <= 0 || got >= 1 {
		t.Fatalf("ease midpoint = %g, want an interior value", got)
	}
}
