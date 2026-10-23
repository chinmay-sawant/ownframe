package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/chinmay-sawant/ownframe"
)

type catalogOpen struct {
	Title         string
	Data          any
	Width, Height int
	RetainState   bool
}

func openCatalog(ctx context.Context, cfg catalogOpen) (*ownframe.Page, error) {
	dir := catalogFixtureDir()
	html, err := os.ReadFile(filepath.Join(dir, "index.html")) //nolint:gosec // reason: fixed repo-relative testdata path.
	if err != nil {
		return nil, err
	}
	css, err := os.ReadFile(filepath.Join(dir, "styles.css")) //nolint:gosec // reason: fixed repo-relative testdata path.
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	p, err := ownframe.New(ownframe.Config{
		Title: cfg.Title, HTML: string(html), Theme: string(css),
		Width: cfg.Width, Height: cfg.Height, RetainState: cfg.RetainState,
	})
	if err != nil {
		return nil, err
	}
	return p, redrawCatalog(ctx, p, cfg.Data)
}

func redrawCatalog(ctx context.Context, p *ownframe.Page, data any) error {
	p.SetData(data)
	return p.Redraw(ctx)
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
