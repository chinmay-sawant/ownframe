package previewwindow

import (
	"context"
	"os"
	"path/filepath"

	"github.com/chinmay-sawant/ownframe"
)

// Config describes one local OwnFrame test-data window.
type Config struct {
	Title, FixtureDir string
	Data              any
	Width, Height     int
	// RetainState keeps the layout state an element patch needs. Set it on a
	// page that patches elements in the parsed tree.
	RetainState bool
}

// New builds and lays out a window from an example fixture directory.
func New(ctx context.Context, config Config) (*ownframe.Page, error) {
	html, err := os.ReadFile(filepath.Join(config.FixtureDir, "index.html"))
	if err != nil {
		return nil, err
	}
	css, err := os.ReadFile(filepath.Join(config.FixtureDir, "styles.css"))
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	page, err := ownframe.New(ownframe.Config{Title: config.Title, HTML: string(html),
		Theme: string(css), Width: config.Width, Height: config.Height,
		RetainState: config.RetainState})
	if err != nil {
		return nil, err
	}
	page.SetData(config.Data)
	if err = page.Redraw(ctx); err != nil {
		return nil, err
	}
	return page, nil
}

// Redraw applies updated template data to a page.
func Redraw(ctx context.Context, page *ownframe.Page, data any) error {
	page.SetData(data)
	return page.Redraw(ctx)
}
