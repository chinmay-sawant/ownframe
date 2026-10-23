package grid

import (
	"context"
	"fmt"

	"github.com/chinmay-sawant/ownframe"
)

// New parses the embedded grid template and registers its handlers.
func New() (*App, error) {
	page, err := ownframe.New(ownframe.Config{
		Title:  "Grid",
		HTML:   gridHTML,
		Width:  DefaultWidth,
		Height: DefaultHeight,
	})
	if err != nil {
		return nil, err
	}

	tiles := []Tile{}
	labels := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L"}
	spans := []string{"span-4", "span-4", "span-4", "span-6", "span-6", "span-3", "span-3", "span-3", "span-3", "span-12", "span-8", "span-4"}
	for i, l := range labels {
		span := ""
		if i < len(spans) {
			span = spans[i]
		}
		tiles = append(tiles, Tile{
			ID:    fmt.Sprintf("t%d", i),
			Label: l,
			Class: span,
		})
	}

	app := &App{page: page, view: View{
		Title: "CSS grid in ownframe",
		Tiles: tiles,
	}}
	page.SetData(&app.view)
	page.Handle(ownframe.Handlers{
		Click: app.onClick,
	})

	return app, nil
}

func (a *App) onClick(_ context.Context, box ownframe.Box) error {
	if box.Action != "toggle" {
		return nil
	}
	for i := range a.view.Tiles {
		if a.view.Tiles[i].ID == box.ID {
			a.view.Tiles[i].On = !a.view.Tiles[i].On
		}
	}
	a.page.SetData(&a.view)
	return nil
}
