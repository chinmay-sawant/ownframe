// Package csspreview runs an OwnFrame window for Blinkless CSS metadata.
package csspreview

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/previewwindow"
)

// Run opens the native CSS property catalog window.
func Run() error {
	ctx := context.Background()
	v := &view{Started: map[string]bool{}}
	p, err := previewwindow.New(ctx, previewwindow.Config{Title: "OwnFrame CSS property catalog",
		FixtureDir: "testdata/css-preview", Data: v, Width: 960, Height: 860})
	if err != nil {
		return err
	}
	v.Catalog = catalogRows(nil)
	if err := previewwindow.Redraw(ctx, p, v); err != nil {
		return err
	}
	v.Catalog = catalogRows(p)
	v.Readout = describe(p)
	if err := previewwindow.Redraw(ctx, p, v); err != nil {
		return err
	}
	d := newDemo(p, v)
	p.Handle(ownframe.Handlers{Click: d.click})
	return ownframe.Run(ctx, p)
}
