package masking

import (
	"reflect"
	"testing"
)

func TestParseFontFeatures(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want []FontFeature
	}{
		{"", nil},
		{`"smcp" 1`, []FontFeature{{"smcp", 1}}},
		{`"liga" 0, "clig" 0, "calt" 0`, []FontFeature{{"liga", 0}, {"clig", 0}, {"calt", 0}}},
		{`"tnum"`, []FontFeature{{"tnum", 1}}},
		{`'kern' off, "onum" on, "ss01" 3`, []FontFeature{{"kern", 0}, {"onum", 1}, {"ss01", 3}}},
		{`"toolong" 1, "ok  " x, "zero" 1`, []FontFeature{{"zero", 1}}},
		{`,,  , "a b c" 1 2`, nil},
	}
	for _, c := range cases {
		if got := ParseFontFeatures(c.in); !reflect.DeepEqual(got, c.want) {
			t.Errorf("ParseFontFeatures(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestAutospaceGapBetween(t *testing.T) {
	t.Parallel()
	const gap = 2.0
	cases := []struct {
		prev, next rune
		want       float64
	}{
		{'語', 'a', gap},
		{'a', '語', gap},
		{'か', 'Z', gap},
		{'a', 'b', 0},
		{'語', '日', 0},
		{'語', '1', 0},
		{'語', ' ', 0},
		{0, 'a', 0},
		{'語', 0, 0},
		{'語', softHyphen, 0},
		{softHyphen, 'a', 0},
	}
	for _, c := range cases {
		if got := AutospaceGapBetween(c.prev, c.next, gap); got != c.want {
			t.Errorf("AutospaceGapBetween(%q, %q) = %v, want %v", c.prev, c.next, got, c.want)
		}
	}
	if got := AutospaceGapBetween('語', 'a', 0); got != 0 {
		t.Errorf("a zero gap must stay zero, got %v", got)
	}
}
