package replay

import "testing"

func TestParseBackdropFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     string
		want      backdropFilter
		wantValid bool
	}{
		{
			name:      "blur and saturate",
			value:     "blur(14px) saturate(1.35)",
			want:      backdropFilter{BlurPx: 14, Saturation: 1.35},
			wantValid: true,
		},
		{
			name:      "blur saturation and refraction",
			value:     "blur(28px) saturate(1.4) refract(5px)",
			want:      backdropFilter{BlurPx: 28, Saturation: 1.4, RefractionPx: 5},
			wantValid: true,
		},
		{
			name:      "refraction only",
			value:     "refract(4px)",
			want:      backdropFilter{Saturation: 1, RefractionPx: 4},
			wantValid: true,
		},
		{
			name:      "percentage saturation",
			value:     "saturate(135%)",
			want:      backdropFilter{Saturation: 1.35},
			wantValid: true,
		},
		{name: "none", value: "none"},
		{name: "svg URL remains unsupported", value: "url(#liquid-glass)"},
		{name: "unsupported function rejects the list", value: "blur(12px) grayscale(0.5)"},
		{name: "negative blur is invalid", value: "blur(-4px)"},
		{name: "pixel blur only", value: "blur(1rem)"},
		{name: "duplicate blur is unsupported", value: "blur(4px) blur(8px)"},
		{name: "duplicate saturation is unsupported", value: "saturate(1.2) saturate(1.4)"},
		{name: "duplicate refraction is unsupported", value: "refract(2px) refract(4px)"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, ok := parseBackdropFilter(test.value)
			if ok != test.wantValid || got != test.want {
				t.Fatalf("parseBackdropFilter(%q) = %+v, %v; want %+v, %v",
					test.value, got, ok, test.want, test.wantValid)
			}
		})
	}
}
