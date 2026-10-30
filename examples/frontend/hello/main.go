package main

import (
	"context"
	_ "embed"
	"log"

	"github.com/chinmay-sawant/ownframe"
)

const (
	windowWidth  = 640
	windowHeight = 480
)

//go:embed template.html
var templateHTML string

//go:embed style.css
var stylesheet string

func main() {
	page, err := ownframe.New(ownframe.Config{
		Title:  "Hello World",
		Width:  windowWidth,
		Height: windowHeight,
		HTML:   templateHTML,
		Theme:  stylesheet,
	})
	if err != nil {
		log.Fatal(err)
	}

	view := struct{ Message string }{Message: "Hello, world!"}
	page.SetData(&view)
	page.Handle(ownframe.Handlers{
		Click: func(_ context.Context, box ownframe.Box) error {
			if box.Action == "greet" {
				view.Message = "Hello from Go!"
				page.SetData(&view)
			}
			return nil
		},
	})

	if err := ownframe.Run(context.Background(), page); err != nil {
		log.Fatal(err)
	}
}
