// Command hello-builder is the builder-pattern Hello World.
// The HTML template version lives in examples/frontend/hello.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/ui"
)

const (
	windowWidth  = 640
	windowHeight = 480
)

type state struct {
	Message string
}

func newPage() (*ownframe.Page, *state, error) {
	current := &state{Message: "Hello, world!"}
	var view *ui.View
	var onGreet func(context.Context) error
	onGreet = func(ctx context.Context) error {
		current.Message = "Hello from Go!"
		if err := view.Update(ctx, helloTree(current, onGreet)); err != nil {
			return fmt.Errorf("update hello page: %w", err)
		}
		return nil
	}
	mounted, err := ui.Mount(ownframe.Config{
		Title:  "Hello World",
		Width:  windowWidth,
		Height: windowHeight,
	}, helloTree(current, onGreet))
	if err != nil {
		return nil, nil, fmt.Errorf("mount hello page: %w", err)
	}
	view = mounted
	return mounted.Page, current, nil
}

func helloTree(current *state, onGreet func(context.Context) error) *ui.Element {
	return ui.Div().
		BoxSizing(ui.BoxSizingBorderBox).
		Display(ui.DisplayFlex).
		JustifyContent(ui.JustifyContentCenter).
		AlignItems(ui.AlignItemsFlexStart).
		MinHeight(ui.Vh(100)).
		Padding(ui.Edges(ui.Px(48), ui.Px(24))).
		FontFamily(ui.FontFamilySansSerif).
		BackgroundColor(ui.Hex("#f1f4fa")).
		Color(ui.Hex("#17233c")).
		Children(ui.Div().
			BoxSizing(ui.BoxSizingBorderBox).
			MaxWidth(ui.Px(440)).
			Padding(ui.Px(32)).
			BackgroundColor(ui.Hex("#ffffff")).
			Border(ui.SolidBorder(ui.Px(1), ui.Hex("#dce3ef"))).
			BorderRadius(ui.Px(16)).
			Children(
				ui.P().Text("YOUR FIRST OWNFRAME APP").
					Color(ui.Hex("#4969bf")).
					FontSize(ui.Px(12)).
					FontWeight(ui.FontWeightBold),
				ui.Tag("h1").Text(current.Message).
					Margin(ui.Edges(ui.Px(16), ui.Px(0), ui.Px(8))).
					FontSize(ui.Px(32)),
				ui.P().Text("The page tree is Go. The handler updates the message.").
					LineHeight(ui.Number(1.5)),
				ui.Button("Say hello").
					ID("greet").
					OnClick(onGreet).
					BoxSizing(ui.BoxSizingBorderBox).
					Display(ui.DisplayFlex).
					AlignItems(ui.AlignItemsCenter).
					JustifyContent(ui.JustifyContentCenter).
					Height(ui.Px(44)).
					MarginTop(ui.Px(16)).
					Padding(ui.Edges(ui.Px(0), ui.Px(20))).
					Border(ui.SolidBorder(ui.Px(0), ui.Transparent())).
					BorderRadius(ui.Px(8)).
					BackgroundColor(ui.Hex("#4265c9")).
					Color(ui.Hex("#ffffff")).
					FontSize(ui.Px(14)).
					FontWeight(ui.FontWeightBold),
			))
}

func main() {
	page, _, err := newPage()
	if err != nil {
		log.Fatal(err)
	}
	if err := ownframe.Run(context.Background(), page); err != nil {
		log.Fatal(err)
	}
}
