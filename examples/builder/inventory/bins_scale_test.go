package main

import (
	"context"
	"strings"
	"testing"
)

func TestBinScaleCoversRange(t *testing.T) {
	var order []string
	seen := map[string]bool{}
	for v := 0; v <= 100; v++ {
		bg, fg := binColors(v)
		if bg == "" || fg == "" {
			t.Fatalf("binColors(%d) = (%q, %q), want non-empty bg/fg", v, bg, fg)
		}
		if !seen[bg] {
			seen[bg] = true
			order = append(order, bg)
		}
	}
	if len(order) != 5 {
		t.Fatalf("distinct bin bgs = %d (%v), want 5", len(order), order)
	}
	if len(binStops) != 5 {
		t.Fatalf("len(binStops) = %d, want 5", len(binStops))
	}
	bg100, _ := binColors(100)
	if bg100 != "#b91c1c" {
		t.Errorf("binColors(100) bg = %q, want %q (red stop)", bg100, "#b91c1c")
	}
	bg0, _ := binColors(0)
	if bg0 != "#1d5c3f" {
		t.Errorf("binColors(0) bg = %q, want %q (green stop)", bg0, "#1d5c3f")
	}
}

func TestBinStatusLabels(t *testing.T) {
	cases := []struct {
		value int
		want  string
	}{
		{97, "Overfull"},
		{90, "Full"},
		{75, "High"},
		{60, "Healthy"},
		{20, "Room"},
	}
	for _, c := range cases {
		if got := binStatus(c.value); got != c.want {
			t.Errorf("binStatus(%d) = %q, want %q", c.value, got, c.want)
		}
	}
}

func TestBinCellTitleData(t *testing.T) {
	title := binCellTitle(2, "L3", 92)
	for _, want := range []string{"C", "L3", "92"} {
		if !strings.Contains(title, want) {
			t.Errorf("binCellTitle(2, %q, 92) = %q, want it to contain %q", "L3", title, want)
		}
	}
	status := binStatus(92)
	if status == "" {
		t.Fatal("binStatus(92) is empty")
	}
	if !strings.Contains(title, status) {
		t.Errorf("binCellTitle(2, %q, 92) = %q, want it to contain status %q", "L3", title, status)
	}
}

func TestHottestColdestSane(t *testing.T) {
	hot, hotAvg := hottestAisle()
	cold, coldAvg := coldestAisle()
	if hotAvg < coldAvg {
		t.Errorf("hottest avg %d < coldest avg %d", hotAvg, coldAvg)
	}
	for _, name := range []string{hot, cold} {
		if len(name) != 1 || name[0] < 'A' || name[0] > 'J' {
			t.Errorf("aisle name = %q, want single letter A-J", name)
		}
	}
}

func TestBinCellsHaveNoTitleTooltip(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	if err := a.page.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	html := a.page.HTML()
	for _, cell := range []string{"bin-G-L2", "bin-E-L5", "bin-H-L4"} {
		idx := strings.Index(html, `id="`+cell+`"`)
		if idx < 0 {
			t.Fatalf("missing cell %q", cell)
		}
		probe := html[idx:]
		if end := strings.Index(probe, ">"); end >= 0 {
			probe = probe[:end]
		}
		if strings.Contains(probe, "title=") {
			t.Fatalf("cell %q still carries a title tooltip: %q", cell, probe)
		}
	}
}
