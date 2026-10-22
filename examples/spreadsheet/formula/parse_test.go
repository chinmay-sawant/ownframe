package formula

import (
	"errors"
	"testing"
)

func TestParseRefWord(t *testing.T) {
	cases := []struct {
		word string
		row  int
		col  int
		ok   bool
	}{
		{"A1", 1, 1, true},
		{"B2", 2, 2, true},
		{"$C$3", 3, 3, true},
		{"z26", 26, 26, true},
		{"AA1", 1, 27, true},
		{"ZZZ1", 1, 18278, true},
		{"A0", 0, 1, true},
		{"A", 0, 0, false},
		{"1A", 0, 0, false},
		{"A1B", 0, 0, false},
		{"ABCD1", 0, 0, false},
		{"$1", 0, 0, false},
		{"", 0, 0, false},
	}

	for _, c := range cases {
		row, col, ok := ParseRefWord(c.word)
		if row != c.row || col != c.col || ok != c.ok {
			t.Errorf("ParseRefWord(%q) = %d,%d,%v want %d,%d,%v",
				c.word, row, col, ok, c.row, c.col, c.ok)
		}
	}
}

func TestParseErrors(t *testing.T) {
	lim := DefaultLimits()

	cases := []struct {
		src  string
		code string
	}{
		{"1+", ErrSyntax},
		{"(1+2", ErrSyntax},
		{"1 2", ErrSyntax},
		{"SUM()", ErrSyntax},
		{"SUM(1,)", ErrSyntax},
		{"A", ErrName},
		{"FOOBAR1", ErrName},
		{"A0", ErrRef},
		{"Z99999999", ErrRef},
		{`"open`, ErrSyntax},
		{"1.2.3", ErrSyntax},
	}

	for _, c := range cases {
		_, err := Parse(c.src, lim)
		if err == nil {
			t.Errorf("Parse(%q) succeeded, want %s", c.src, c.code)
			continue
		}

		var pe *ParseError
		if !errors.As(err, &pe) || pe.Code != c.code {
			t.Errorf("Parse(%q) = %v, want code %s", c.src, err, c.code)
		}
	}
}

func TestFormatNumber(t *testing.T) {
	cases := []struct {
		n    float64
		want string
	}{
		{0, "0"},
		{-5, "-5"},
		{0.5, "0.5"},
		{1000000, "1000000"},
		{-1000000, "-1000000"},
	}

	for _, c := range cases {
		if got := FormatNumber(c.n); got != c.want {
			t.Errorf("FormatNumber(%v) = %q want %q", c.n, got, c.want)
		}
	}
}
