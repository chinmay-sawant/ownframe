package main

import (
	"context"
	"strings"
	"testing"
)

func TestCatalogStartRunsEachPreviewKind(t *testing.T) {
	for _, name := range []string{"-webkit-animation", "transition-property", "animation-composition"} {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			p, d, v := fixturePage(t)
			var selected catalogRow
			for _, row := range v.Catalog {
				if row.Name == name {
					selected = row
					break
				}
			}
			if selected.Name == "" {
				t.Fatalf("missing catalog row %q", name)
			}
			button, ok := findBox(p.Boxes(), selected.Button)
			if !ok {
				t.Fatalf("missing button %q", selected.Button)
			}
			if err := p.Testing().Click(ctx, button.X+button.W/2, button.Y+button.H/2); err != nil {
				t.Fatal(err)
			}
			if !v.Started[selected.Button] || v.Selected != name {
				t.Fatalf("start state not recorded: %+v", v)
			}
			if selected.SampleKey == "" {
				if len(d.samples) != 0 || !strings.Contains(v.Readout, selected.Value) {
					t.Fatalf("metadata selection started motion or lost value: %+v", v)
				}
				return
			}
			if err := d.tick(ctx); err != nil {
				t.Fatal(err)
			}
			if d.samples[selected.SampleKey] == nil {
				t.Fatalf("button %q did not start sample %q", selected.Button, selected.SampleKey)
			}
		})
	}
}
