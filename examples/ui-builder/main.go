// Command ui-builder demonstrates the experimental fluent UI API.
//
// Run as a desktop app or with -web for the existing OwnFrame web preview.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/ui"
)

func main() {
	web := flag.Bool("web", false, "serve the preview in a browser")
	addr := flag.String("addr", "127.0.0.1:8131", "address for -web")
	flag.Parse()

	page, err := ui.NewPage(ownframe.Config{
		Title: "OwnFrame Builder Experiment",
		Width: 920, Height: 640,
	}, screen())
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	if err := page.Redraw(ctx); err != nil {
		log.Fatal(err)
	}

	if *web {
		if err := ownframe.Serve(ctx, page, *addr); err != nil {
			log.Fatal(err)
		}
		return
	}
	if err := ownframe.Run(ctx, page); err != nil {
		log.Fatal(err)
	}
}

func screen() *ui.Element {
	return ui.Div().Display("flex").FlexDirection("column").
		Gap("16px").Padding("20px").BackgroundColor("#171b23").
		Color("#e4e9f0").FontSize("15px").
		Children(
			// Example 1: a CSS-styled card.
			ui.Div().BackgroundColor("#252b36").
				Padding("16px").BorderRadius("10px").
				Children(
					ui.Text("OwnFrame: declarative UI").FontWeight("bold").FontSize("22px"),
					ui.P().Text("GPUI-inspired Go builder; Blinkless still owns CSS layout."),
				),
			// Example 2: flex toolbar, native OwnFrame action handling.
			ui.Div().Display("flex").FlexDirection("row").
				AlignItems("center").Gap("12px").
				Children(
					ui.Button("Save").Padding("10px").OnClick(func(context.Context) error {
						log.Print("Save clicked through OwnFrame")
						return nil
					}),
					ui.Button("Cancel").Padding("10px").OnClick(func(context.Context) error {
						log.Print("Cancel clicked through OwnFrame")
						return nil
					}),
				),
			// Example 3: a file-explorer sidebar composed from list nodes.
			ui.Div().Width("280px").BackgroundColor("#202630").
				Padding("12px").BorderRadius("8px").
				Children(
					ui.Text("Explorer").FontWeight("bold"),
					ui.Ul().Children(
						ui.Li().Text("main.go"),
						ui.Li().Text("go.mod"),
						ui.Li().Text("README.md"),
					),
				),
		)
}
