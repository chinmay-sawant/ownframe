package main

import (
	"context"
	_ "embed"

	"github.com/chinmay-sawant/ownframe"
)

//go:embed template.html
var templateHTML string

//go:embed style.css
var stylesheet string

func newPage() (*ownframe.Page, error) {
	page, err := ownframe.New(ownframe.Config{
		Title:  "Launch board",
		Width:  560,
		Height: 560,
		HTML:   templateHTML,
		Theme:  stylesheet,
	})
	if err != nil {
		return nil, err
	}

	tasks := Tasks{Filter: "all", Items: []Task{
		{ID: "design", Title: "Review the new app screens", Category: "Design", Done: true},
		{ID: "build", Title: "Build the desktop prototype", Category: "Engineering"},
		{ID: "test", Title: "Test the Android layout", Category: "QA"},
		{ID: "ship", Title: "Share the first release", Category: "Release"},
	}}
	page.SetData(&tasks)
	page.Handle(ownframe.Handlers{
		Click: func(_ context.Context, box ownframe.Box) error {
			switch box.Action {
			case "all", "open", "done":
				tasks.Filter = box.Action
			case "theme":
				tasks.Light = !tasks.Light
			default:
				for i := range tasks.Items {
					if box.Action == tasks.Items[i].ID {
						tasks.Items[i].Done = !tasks.Items[i].Done
						break
					}
				}
			}
			page.SetData(&tasks)
			return nil
		},
	})
	return page, nil
}
