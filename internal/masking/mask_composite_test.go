package masking

import "testing"

func TestCombineMaskOperators(t *testing.T) {
	cases := []struct {
		mode string
		want uint8
	}{
		{"add", 160}, {"subtract", 96}, {"intersect", 32}, {"exclude", 128},
	}
	for _, tc := range cases {
		if got := combineMask(128, 64, tc.mode); got != tc.want {
			t.Errorf("%s = %d, want %d", tc.mode, got, tc.want)
		}
	}
}
