package perfcheck

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestCheckPass(t *testing.T) {
	budgets := Budgets{
		Template: StageBudget{
			PerBox: 1,
			PerOp:  0.25,
		},
		Style: StageBudget{
			PerBox: 2,
			PerOp:  0.5,
		},
		Layout: StageBudget{
			PerBox: 3,
			PerOp:  0.75,
		},
		Paint: StageBudget{
			PerBox: 3,
			PerOp:  0.75,
		},
		Other: StageBudget{
			PerBox: 1,
			PerOp:  0.25,
		},
		AllocPerBox: 512,
	}
	stats := Stats{
		Template: 2 * time.Microsecond,
		Style:    4 * time.Microsecond,
		Layout:   6 * time.Microsecond,
		Paint:    6 * time.Microsecond,
		Other:    2 * time.Microsecond,
		Alloc:    1024,
	}
	boxes, ops := 10, 20

	if got := Violations(budgets, stats, boxes, ops); len(got) != 0 {
		t.Fatalf("Violations = %v, want none", got)
	}

	report := ReportWith(budgets, stats, boxes, ops)
	for _, want := range []string{
		"status=ok",
		"stage",
		"actual",
		"budget",
		"0.20 µs",
		"1.00 µs",
		"102 B",
		"512 B",
	} {
		if !strings.Contains(report, want) {
			t.Errorf("report missing %q:\n%s", want, report)
		}
	}

	reporter := &fakeReporter{}
	check(reporter, budgets, stats, boxes, ops)
	if len(reporter.errors) != 0 {
		t.Fatalf("check reported %v, want none", reporter.errors)
	}

	Check(t, stats, boxes, ops)
}

func TestCheckFail(t *testing.T) {
	budgets := Budgets{
		Layout: StageBudget{
			PerBox: 5,
			PerOp:  0.5,
		},
		Paint: StageBudget{
			PerBox: 5,
			PerOp:  0.5,
		},
		AllocPerBox: 100,
	}
	stats := Stats{
		Layout: 100 * time.Microsecond,
		Paint:  20 * time.Microsecond,
		Alloc:  2048,
	}
	boxes, ops := 10, 20

	violations := Violations(budgets, stats, boxes, ops)
	if len(violations) != 4 {
		t.Fatalf("Violations = %v, want 4", violations)
	}
	want := []Violation{
		{
			Stage:  "layout",
			Unit:   "box",
			Actual: 10,
			Budget: 5,
		},
		{
			Stage:  "layout",
			Unit:   "op",
			Actual: 5,
			Budget: 0.5,
		},
		{
			Stage:  "paint",
			Unit:   "op",
			Actual: 1,
			Budget: 0.5,
		},
		{
			Stage:  "alloc",
			Unit:   "box",
			Actual: 204.8,
			Budget: 100,
		},
	}
	for i, violation := range violations {
		if violation != want[i] {
			t.Errorf("Violations[%d] = %+v, want %+v", i, violation, want[i])
		}
	}

	report := ReportWith(budgets, stats, boxes, ops)
	for _, wantText := range []string{
		"status=FAIL (4 over budget)",
		"10.00 µs",
		"5.00 µs",
		"0.50 µs",
		"205 B",
		"100 B",
		"over by",
	} {
		if !strings.Contains(report, wantText) {
			t.Errorf("report missing %q:\n%s", wantText, report)
		}
	}

	reporter := &fakeReporter{}
	check(reporter, budgets, stats, boxes, ops)
	if len(reporter.errors) != 1 {
		t.Fatalf("check reported %d errors, want 1", len(reporter.errors))
	}
	if !strings.Contains(reporter.errors[0], "status=FAIL") {
		t.Fatalf("check error = %q, want the budget table", reporter.errors[0])
	}
	if reporter.helpers == 0 {
		t.Fatal("check did not mark itself as a helper")
	}
}

func TestZeroCounts(t *testing.T) {
	budgets := Budgets{
		Layout: StageBudget{
			PerBox: 5,
			PerOp:  5,
		},
	}
	stats := Stats{Layout: 100 * time.Microsecond}

	t.Run("both zero", func(t *testing.T) {
		if got := Violations(DefaultBudgets(), stats, 0, 0); len(got) != 0 {
			t.Fatalf("Violations = %v, want none", got)
		}
		report := ReportWith(DefaultBudgets(), stats, 0, 0)
		if strings.Contains(report, "NaN") || strings.Contains(report, "Inf") {
			t.Fatalf("report contains a non-finite value:\n%s", report)
		}
		for _, want := range []string{"skipped (no boxes)", "skipped (no ops)"} {
			if !strings.Contains(report, want) {
				t.Errorf("report missing %q:\n%s", want, report)
			}
		}
	})

	t.Run("no boxes", func(t *testing.T) {
		violations := Violations(budgets, stats, 0, 10)
		if len(violations) != 1 {
			t.Fatalf("Violations = %v, want one per-op violation", violations)
		}
		if violations[0].Unit != "op" {
			t.Fatalf("violation unit = %q, want op", violations[0].Unit)
		}
	})

	t.Run("no ops", func(t *testing.T) {
		violations := Violations(budgets, stats, 10, 0)
		if len(violations) != 1 {
			t.Fatalf("Violations = %v, want one per-box violation", violations)
		}
		if violations[0].Unit != "box" {
			t.Fatalf("violation unit = %q, want box", violations[0].Unit)
		}
	})
}

type fakeReporter struct {
	helpers int
	errors  []string
}

func (f *fakeReporter) Helper() {
	f.helpers++
}

func (f *fakeReporter) Errorf(format string, args ...any) {
	f.errors = append(f.errors, fmt.Sprintf(format, args...))
}
