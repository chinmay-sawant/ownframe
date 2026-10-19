package csspreview

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func fixturePage(t *testing.T, kind string) (*ownframe.Page, *demo, *view) {
	t.Helper()
	fixture := kind
	if kind == "all" {
		fixture = "css-preview"
	}
	dir := filepath.Join("..", "..", "testdata", fixture)
	html, err := os.ReadFile(filepath.Join(dir, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	css, err := os.ReadFile(filepath.Join(dir, "styles.css"))
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	p, err := ownframe.New(ownframe.Config{HTML: string(html), Theme: string(css), Width: 960, Height: 1000})
	if err != nil {
		t.Fatal(err)
	}
	v := &view{Started: map[string]bool{}}
	p.SetData(v)
	if err = p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	if kind == "all" {
		v.Catalog = catalogRows(nil)
		p.SetData(v)
		if err = p.Redraw(context.Background()); err != nil {
			t.Fatal(err)
		}
		v.Catalog = catalogRows(p)
	}
	v.Readout = describe(p, kind)
	p.SetData(v)
	if err = p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	d := newDemo(p, kind, v)
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
