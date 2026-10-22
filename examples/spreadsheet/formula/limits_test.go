package formula

import (
	"errors"
	"testing"
)

func TestLimits(t *testing.T) {
	env := testEnv{
		{0, 0}: Number(1),
		{1, 0}: Number(2),
		{2, 0}: Number(3),
		{3, 0}: Number(4),
	}

	cases := []struct {
		name string
		src  string
		lim  Limits
		work *Work
		code string
	}{
		{
			name: "length",
			src:  "1+1+1+1+1+1",
			lim:  Limits{MaxLen: 5},
			code: ErrLimit,
		},
		{
			name: "depth",
			src:  "((1+2))",
			lim:  Limits{MaxDepth: 1},
			code: ErrLimit,
		},
		{
			name: "range cells",
			src:  "SUM(A1:B3)",
			lim:  Limits{MaxRangeCells: 4},
			code: ErrLimit,
		},
		{
			name: "ref row",
			src:  "A10",
			lim:  Limits{MaxRows: 5},
			code: ErrRef,
		},
		{
			name: "work",
			src:  "SUM(A1:A4)",
			work: NewWork(2),
			code: ErrLimit,
		},
	}

	for _, c := range cases {
		e, err := Parse(c.src, c.lim)
		var pe *ParseError
		if errors.As(err, &pe) && pe.Code == c.code {
			continue
		}
		if err != nil {
			t.Errorf("%s: Parse = %v want %s", c.name, err, c.code)

			continue
		}

		v := Eval(e, env, c.lim, c.work)
		if !v.IsError() || v.Code != c.code {
			t.Errorf("%s: Eval = %v want error %s", c.name, v, c.code)
		}
	}
}

func TestWorkBudget(t *testing.T) {
	work := NewWork(3)
	if !work.Spend(2) || work.Left != 1 {
		t.Fatalf("Spend(2) left %d want 1", work.Left)
	}

	if work.Spend(2) {
		t.Fatal("Spend(2) with one left succeeded")
	}

	if work.Left != 0 {
		t.Fatalf("exhausted work left %d want 0", work.Left)
	}

	if !work.Spend(0) {
		t.Fatal("Spend(0) failed on empty budget")
	}
}
