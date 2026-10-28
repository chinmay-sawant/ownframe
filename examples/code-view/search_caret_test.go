package main

import (
	"context"
	"math"
	"testing"
)

func TestSearchCaretAlignsWithPlaceholder(t *testing.T) {
	for _, tc := range []struct{ id, placeholder string }{
		{"palette", "Search files by name"},
		{"workspace-search", "Search workspace contents"},
	} {
		t.Run(tc.id, func(t *testing.T) {
			app := newTestApp(t)
			app.openFile("backend/probe.go")
			app.scrollTo(1000)
			var err error
			if tc.id == "palette" {
				err = app.openPalette(context.Background())
			} else {
				err = app.openWorkspaceSearch(context.Background())
			}
			if err != nil {
				t.Fatal(err)
			}
			d := app.Page().Probe().Display()
			caret := d.Ops[len(d.Ops)-1]
			found := false
			for _, op := range d.Ops {
				if op.Text != tc.placeholder {
					continue
				}
				found = true
				if math.Abs(caret.X-op.X) > 1 || caret.Y < op.Y-op.Size-1 || caret.Y+caret.H > op.Y+1 || !caret.Fixed {
					t.Fatalf("caret %+v does not align with placeholder %+v", caret, op)
				}
			}
			if !found {
				t.Fatal("placeholder not rendered")
			}
		})
	}
}
