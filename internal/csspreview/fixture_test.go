package csspreview

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/previewwindow"
)

func fixturePage(t *testing.T) (*ownframe.Page, *demo, *view) {
	t.Helper()
	ctx := context.Background()
	v := &view{Started: map[string]bool{}}
	p, err := previewwindow.New(ctx, previewwindow.Config{FixtureDir: filepath.Join("..", "..", "testdata", "css-preview"),
		Data: v, Width: 960, Height: 1000})
	if err != nil {
		t.Fatal(err)
	}
	v.Catalog = catalogRows(nil)
	if err = previewwindow.Redraw(ctx, p, v); err != nil {
		t.Fatal(err)
	}
	v.Catalog = catalogRows(p)
	v.Readout = describe(p)
	if err = previewwindow.Redraw(ctx, p, v); err != nil {
		t.Fatal(err)
	}
	d := newDemo(p, v)
	p.Handle(ownframe.Handlers{Click: d.click})
	return p, d, v
}

func hasRecords(p *ownframe.Page, b ownframe.Box, target target, kind string) bool {
	if target.kind == "transition" {
		for _, r := range b.Transitions {
			if r.Property == target.property {
				return true
			}
		}
		return false
	}
	if len(b.Animations) == 0 {
		return false
	}
	for _, rule := range p.Display().Keyframes {
		if rule.Name == b.Animations[0].Name {
			return true
		}
	}
	return false
}

func samePaint(a, b []paint) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
