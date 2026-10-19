// Package main runs the OwnFrame CSS property catalog example.
package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/previewwindow"
)

func run() error {
	ctx := context.Background()
	v := &view{Started: map[string]bool{}}
	p, err := previewwindow.New(ctx, previewwindow.Config{Title: "OwnFrame CSS property catalog",
		FixtureDir: catalogFixtureDir(), Data: v, Width: 960, Height: 860})
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
