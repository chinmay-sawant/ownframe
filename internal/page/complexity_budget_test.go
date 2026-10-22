package page

import (
	"context"
	"strings"
	"testing"
	"time"
)

const (
	budgetInputs = 2000
	budgetDepth  = 200
	budgetMaxLag = 10 * time.Second
	budgetWide   = 500000
	budgetDeep   = 20000
	budgetGuard  = 1000
)

func adversarialWide() string {
	return "<style>" + strings.Repeat(".c,", budgetInputs) + ".last{width:10px}</style><div>" + strings.Repeat(`<input value="v">`, budgetInputs) + "</div>"
}

func adversarialDeep() string {
	return strings.Repeat("<div>", budgetDepth) + "x" + strings.Repeat("</div>", budgetDepth)
}

func redrawNew(html string) error {
	p, err := New(Config{HTML: html, Width: 800, Height: 600})
	if err != nil {
		return err
	}

	return p.Redraw(context.Background())
}

// TestComplexityBudget pins adversarial costs (GEN-27): wide selector
// lists with thousands of fields and near-cap nesting finish inside a
// hard deadline and allocation budget.
func TestComplexityBudget(t *testing.T) {
	for _, html := range []string{adversarialWide(), adversarialDeep()} {
		start := time.Now()
		if err := redrawNew(html); err != nil {
			t.Fatal(err)
		}
		if d := time.Since(start); d > budgetMaxLag {
			t.Fatalf("redraw took %s", d)
		}
	}
	if n := testing.AllocsPerRun(3, func() { _ = redrawNew(adversarialWide()) }); n > budgetWide {
		t.Fatalf("wide allocs = %v", n)
	}
	if n := testing.AllocsPerRun(3, func() { _ = redrawNew(adversarialDeep()) }); n > budgetDeep {
		t.Fatalf("deep allocs = %v", n)
	}
	if n := testing.AllocsPerRun(20, func() { _ = checkDOMLimits(adversarialWide()) }); n > budgetGuard {
		t.Fatalf("guard allocs = %v", n)
	}
}

// BenchmarkAdversarialWide tracks the wide layout path.
func BenchmarkAdversarialWide(b *testing.B) {
	html := adversarialWide()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = redrawNew(html)
	}
}

// BenchmarkDOMGuard tracks the depth/node guard.
func BenchmarkDOMGuard(b *testing.B) {
	html := adversarialDeep()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = checkDOMLimits(html)
	}
}
