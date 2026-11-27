package main

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/sqlkit"
)

// BenchmarkGridScroll100kRows measures one virtualized scroll step (window
// recompute plus tree update) over a 100,000-row grid.
func BenchmarkGridScroll100kRows(b *testing.B) {
	ctx := context.Background()
	a := newApp(ctx, b, 1440, 960)
	if err := a.LoadSyntheticRows(ctx, 100_000); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		step := 26.0 * 7
		if i%200 >= 100 {
			step = -step
		}
		if err := a.ScrollGridBy(ctx, step); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSQLHighlightLargeScript tokenizes a 500-line script.
func BenchmarkSQLHighlightLargeScript(b *testing.B) {
	var sb strings.Builder
	for i := 0; i < 500; i++ {
		fmt.Fprintf(&sb, "SELECT o.id, COUNT(*) AS n, 'row %d' AS label FROM orders o JOIN users u ON u.id = o.user_id WHERE o.total_cents > %d -- c%d\n", i, i*10, i)
	}
	src := sb.String()
	b.SetBytes(int64(len(src)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if len(sqlkit.Tokenize(src, sqlkit.SQLite)) == 0 {
			b.Fatal("no tokens")
		}
	}
}

// BenchmarkMonitorTickDrain pushes one sample into a full 60-sample history
// and repaints the Monitor tab.
func BenchmarkMonitorTickDrain(b *testing.B) {
	ctx := context.Background()
	a := newApp(ctx, b, 1440, 960)
	if err := a.OpenMonitor(ctx); err != nil {
		b.Fatal(err)
	}
	for i := 0; i < 60; i++ {
		a.PushSyntheticSample(i)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.PushSyntheticSample(60 + i)
		if err := a.RedrawMonitor(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGridHorizontalScroll100kRows scrolls columns of a 100k-row result;
// only the visible window of rows and columns is built.
func BenchmarkGridHorizontalScroll100kRows(b *testing.B) {
	ctx := context.Background()
	a := newApp(ctx, b, 1440, 960)
	if err := a.LoadSyntheticRows(ctx, 100_000); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		step := 1
		if i%2 == 1 {
			step = -1
		}
		if err := a.ScrollGridColumns(ctx, step); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGridClientSort10kRows sorts a 10k-row result in memory.
func BenchmarkGridClientSort10kRows(b *testing.B) {
	ctx := context.Background()
	a := newApp(ctx, b, 1440, 960)
	if err := a.LoadSyntheticRows(ctx, 10_000); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := a.SortGridBy(ctx, 2+i%2); err != nil {
			b.Fatal(err)
		}
	}
}
