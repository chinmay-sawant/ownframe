package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func BenchmarkPointerCycleAll43(b *testing.B) {
	benchCycle(b, true)
}

func BenchmarkPointerCycleAll43NoBatch(b *testing.B) {
	benchCycle(b, false)
}

func benchCycle(b *testing.B, useBatch bool) {
	b.Helper()

	ctx := context.Background()
	v := newView()
	p, err := openCatalog(ctx, catalogOpen{Data: v, Width: 960, Height: 1000})
	if err != nil {
		b.Fatal(err)
	}
	v.Catalog = catalogRows(nil)
	v.Rows = v.Catalog
	if err := redrawCatalog(ctx, p, v); err != nil {
		b.Fatal(err)
	}
	v.Catalog = catalogRows(p)
	v.Rows = v.Catalog
	v.Readout = describe(p)
	if err := redrawCatalog(ctx, p, v); err != nil {
		b.Fatal(err)
	}
	d := newDemo(p, v)
	p.Handle(ownframe.Handlers{Click: d.click})

	// The setup is not a click, so the timer starts after it.
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		v.Started = map[string]bool{}
		v.Restart = map[string]bool{}
		v.Restarting = map[string]bool{}
		v.Selected = ""
		v.applyRows()
		if err := p.Redraw(ctx); err != nil {
			b.Fatal(err)
		}
		b.StartTimer()
		for _, row := range v.Catalog {
			box, ok := findBox(p.Boxes(), row.Button)
			if !ok {
				b.Fatalf("missing %s", row.Button)
			}
			x, y := box.X+box.W/2, box.Y+box.H/2
			if err := p.Edit().Hover(ctx, x, y); err != nil {
				b.Fatal(err)
			}
			run := func() error {
				if err := p.Edit().Press(ctx, x, y); err != nil {
					return err
				}
				if err := p.Testing().SelectAt(ctx, x, y); err != nil {
					return err
				}
				if err := p.Click(ctx, x, y); err != nil {
					return err
				}
				return p.Edit().Release(ctx)
			}
			if useBatch {
				err = redrawBatch(ctx, p, run)
			} else {
				err = run()
			}
			if err != nil {
				b.Fatal(err)
			}
		}
	}
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*43)/1e6, "ms/click")
}
