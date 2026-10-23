package main

import (
	"context"
	"testing"
)

func BenchmarkClickRedraw(b *testing.B) {
	ctx := context.Background()
	v := &view{Started: map[string]bool{}}
	p, err := openCatalog(ctx, catalogOpen{Data: v, Width: 960, Height: 860})
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
	row := v.Catalog[14]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.Started[row.Button] = i%2 == 0
		v.Selected = row.Name
		if err := redrawCatalog(ctx, p, v); err != nil {
			b.Fatal(err)
		}
	}
}
