package main

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/previewwindow"
)

func BenchmarkClickRedraw(b *testing.B) {
	ctx := context.Background()
	v := &view{Started: map[string]bool{}}
	p, err := previewwindow.New(ctx, previewwindow.Config{
		FixtureDir: filepath.Join("testdata", "css-preview"),
		Data:       v, Width: 960, Height: 860})
	if err != nil {
		b.Fatal(err)
	}
	v.Catalog = catalogRows(nil)
	v.Rows = v.Catalog
	if err := previewwindow.Redraw(ctx, p, v); err != nil {
		b.Fatal(err)
	}
	v.Catalog = catalogRows(p)
	v.Rows = v.Catalog
	v.Readout = describe(p)
	if err := previewwindow.Redraw(ctx, p, v); err != nil {
		b.Fatal(err)
	}
	row := v.Catalog[14]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.Started[row.Button] = i%2 == 0
		v.Selected = row.Name
		if err := previewwindow.Redraw(ctx, p, v); err != nil {
			b.Fatal(err)
		}
	}
}
