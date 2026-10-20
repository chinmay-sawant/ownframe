package main

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/previewwindow"
)

func BenchmarkPointerCycleAll43(b *testing.B) {
	benchCycle(b, true)
}

func BenchmarkPointerCycleAll43NoBatch(b *testing.B) {
	benchCycle(b, false)
}

func benchCycle(b *testing.B, batch bool) {
	ctx := context.Background()
	v := &view{Started: map[string]bool{}}
	p, err := previewwindow.New(ctx, previewwindow.Config{
		FixtureDir: filepath.Join("testdata", "css-preview"),
		Data:       v, Width: 960, Height: 1000})
	if err != nil {
		b.Fatal(err)
	}
	v.Catalog = catalogRows(nil)
	if err := previewwindow.Redraw(ctx, p, v); err != nil {
		b.Fatal(err)
	}
	v.Catalog = catalogRows(p)
	v.Readout = describe(p)
	if err := previewwindow.Redraw(ctx, p, v); err != nil {
		b.Fatal(err)
	}
	d := newDemo(p, v)
	p.Handle(ownframe.Handlers{Click: d.click})

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		v.Started = map[string]bool{}
		v.Selected = ""
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
			if err := p.Hover(ctx, x, y); err != nil {
				b.Fatal(err)
			}
			run := func() error {
				if err := p.Press(ctx, x, y); err != nil {
					return err
				}
				if err := p.SelectAt(ctx, x, y); err != nil {
					return err
				}
				if err := p.Click(ctx, x, y); err != nil {
					return err
				}
				return p.Release(ctx)
			}
			if batch {
				err = host.Batch(ctx, p, run)
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
