package replay

import (
	"testing"
)

func TestGlyphAt(t *testing.T) {
	cases := []struct {
		input string
		want  []string
	}{
		{
			input: "Hello, world!",
			want:  []string{"H", "e", "l", "l", "o", ",", " ", "w", "o", "r", "l", "d", "!"},
		},
		{
			input: "Привет мир", // Cyrillic multi-byte
			want:  []string{"П", "р", "и", "в", "е", "т", " ", "м", "и", "р"},
		},
		{
			input: "こんにちは", // Japanese 3-byte
			want:  []string{"こ", "ん", "に", "ち", "は"},
		},
		{
			input: "A🚀B", // 4-byte emoji
			want:  []string{"A", "🚀", "B"},
		},
	}

	for _, tc := range cases {
		var got []string
		for i := 0; i < len(tc.input); {
			g, size := glyphAt(tc.input, i)
			got = append(got, g)
			i += size
		}
		if len(got) != len(tc.want) {
			t.Fatalf("for %q: got %d glyphs, want %d", tc.input, len(got), len(tc.want))
		}
		for idx := range got {
			if got[idx] != tc.want[idx] {
				t.Errorf("for %q at %d: got %q, want %q", tc.input, idx, got[idx], tc.want[idx])
			}
		}
	}
}

func TestGlyphAt_ASCIIZeroAllocs(t *testing.T) {
	str := "The quick brown fox jumps over the lazy dog"
	allocs := testing.AllocsPerRun(100, func() {
		for i := 0; i < len(str); {
			g, size := glyphAt(str, i)
			_ = g
			i += size
		}
	})
	if allocs > 0 {
		t.Fatalf("expected 0 allocs for glyphAt, got %f", allocs)
	}
}
