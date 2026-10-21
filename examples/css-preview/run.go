// Package main runs the OwnFrame CSS property catalog example.
package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/previewwindow"
)

func run(perf bool, width, height int) error {
	ctx := context.Background()
	v := &view{Started: map[string]bool{}}
	p, err := previewwindow.New(ctx, previewwindow.Config{Title: "OwnFrame CSS property catalog",
		FixtureDir: catalogFixtureDir(), Data: v, Width: width, Height: height,
		RetainState: true})
	if err != nil {
		return err
	}
	if perf {
		p.SetPerf(true)
		p.SetDevTools(true)
	}
	v.Catalog = catalogRows(nil)
	v.Rows = v.Catalog
	if err := previewwindow.Redraw(ctx, p, v); err != nil {
		return err
	}
	v.Catalog = catalogRows(p)
	v.Rows = v.Catalog
	v.Readout = describe(p)
	if err := previewwindow.Redraw(ctx, p, v); err != nil {
		return err
	}
	metrics, ok := measureCatalog(p)
	v.Readout = "Select a Start button.\nIts value and action appear here."
	d := newDemo(p, v)
	p.Handle(ownframe.Handlers{Click: d.click, ClickPaint: d.clickPaint})
	p.SetAfterRedraw(func() { d.afterRedraw(context.Background()) })
	if ok {
		if err := installCatalogWindow(ctx, p, v, d, metrics); err != nil {
			return err
		}
	}
	d.gen = p.Generation()
	startWalk(p, d, v, metrics)
	return ownframe.Run(ctx, p)
}

func catalogFixtureDir() string {
	candidates := []string{
		filepath.Join("examples", "css-preview", "testdata", "css-preview"),
		filepath.Join("css-preview", "testdata", "css-preview"),
		filepath.Join("testdata", "css-preview"),
	}
	for _, dir := range candidates {
		if _, err := os.Stat(filepath.Join(dir, "index.html")); err == nil {
			return dir
		}
	}
	return filepath.Join("testdata", "css-preview")
}
