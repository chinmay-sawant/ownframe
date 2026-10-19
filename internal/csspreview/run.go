// Package csspreview runs an OwnFrame window for Blinkless CSS metadata.
package csspreview

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/chinmay-sawant/ownframe"
)

// Run opens the native CSS preview window for one fixture.
func Run(kind string) error {
	if kind != "all" && kind != "animation" && kind != "transition" {
		return fmt.Errorf("usage: go run ./cmd/css-preview [all|animation|transition]")
	}
	dir := filepath.Join("testdata", kind)
	if kind == "all" {
		dir = filepath.Join("testdata", "css-preview")
	}
	html, err := os.ReadFile(filepath.Join(dir, "index.html"))
	if err != nil {
		return err
	}
	css, err := os.ReadFile(filepath.Join(dir, "styles.css"))
	if err != nil && (kind != "animation" || !os.IsNotExist(err)) {
		return err
	}
	p, err := ownframe.New(ownframe.Config{
		Title: "OwnFrame CSS " + kind, HTML: string(html), Theme: string(css),
		Width: 960, Height: 860,
	})
	if err != nil {
		return err
	}
	ctx := context.Background()
	v := &view{Started: map[string]bool{}}
	p.SetData(v)
	if err := p.Redraw(ctx); err != nil {
		return err
	}
	if kind == "all" {
		v.Catalog = catalogRows(nil)
		p.SetData(v)
		if err := p.Redraw(ctx); err != nil {
			return err
		}
		v.Catalog = catalogRows(p)
	}
	v.Readout = describe(p, kind)
	p.SetData(v)
	if err := p.Redraw(ctx); err != nil {
		return err
	}
	d := newDemo(p, kind, v)
	p.Handle(ownframe.Handlers{Click: d.click})
	return ownframe.Run(ctx, p)
}
