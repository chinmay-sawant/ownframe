package animationpreview

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/previewwindow"
)

// Run opens the standalone animation fixture window.
func Run() error {
	ctx := context.Background()
	data := &view{Started: map[string]bool{}}
	page, err := previewwindow.New(ctx, previewwindow.Config{Title: "OwnFrame CSS animations",
		FixtureDir: "testdata/animation", Data: data, Width: previewWidth, Height: previewHeight})
	if err != nil {
		return err
	}
	data.Readout = describe(page)
	if err := previewwindow.Redraw(ctx, page, data); err != nil {
		return err
	}
	demo := newDemo(page, data)
	page.Handle(ownframe.Handlers{Click: demo.click})
	return ownframe.Run(ctx, page)
}

type view struct {
	Started map[string]bool
	Readout string
}
