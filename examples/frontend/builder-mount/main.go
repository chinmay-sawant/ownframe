// Command builder-mount is the builder-pattern Builder basics.
// The template splice version lives in examples/frontend/builder.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

const (
	windowWidth  = 640
	windowHeight = 640
)

type view struct {
	Pressed bool
	Notify  bool
	Clicks  int
	Status  string
}

func statusText(state *view) string {
	return fmt.Sprintf("Button ran %d times. Toggle %v. Switch %v.", state.Clicks, state.Pressed, state.Notify)
}

func buttonLabel(clicks int) string {
	label := "Run action"
	if clicks > 0 {
		label = fmt.Sprintf("Ran %d time", clicks)
		if clicks != 1 {
			label += "s"
		}
	}
	return label
}

func newPage() (*ownframe.Page, *view, error) {
	state := &view{}
	state.Status = statusText(state)
	var mounted *ui.View
	var refresh func(context.Context) error
	onPress := func(ctx context.Context) error {
		state.Clicks++
		return refresh(ctx)
	}
	onToggle := func(ctx context.Context) error {
		state.Pressed = !state.Pressed
		return refresh(ctx)
	}
	onNotify := func(ctx context.Context) error {
		state.Notify = !state.Notify
		return refresh(ctx)
	}
	refresh = func(ctx context.Context) error {
		state.Status = statusText(state)
		if err := mounted.Update(ctx, basicsCard(state, onPress, onToggle, onNotify)); err != nil {
			return fmt.Errorf("update builder page: %w", err)
		}
		return nil
	}
	page, err := ui.Mount(ownframe.Config{
		Title:  "Builder basics",
		Width:  windowWidth,
		Height: windowHeight,
		Theme:  chadcn.Dark + chadcn.Stylesheet,
	}, basicsCard(state, onPress, onToggle, onNotify))
	if err != nil {
		return nil, nil, fmt.Errorf("mount builder page: %w", err)
	}
	mounted = page
	return page.Page, state, nil
}

func basicsCard(state *view, onPress, onToggle, onNotify func(context.Context) error) *ui.Element {
	return chadcn.Card(
		chadcn.CardHeader(
			chadcn.CardTitle("Builder basics"),
			chadcn.CardDescription("One card, host-owned state, dark theme."),
		),
		chadcn.CardContent(
			chadcn.Button(buttonLabel(state.Clicks), chadcn.ButtonOptions{Variant: chadcn.ButtonFilled}).
				ID("press").OnClick(onPress),
			chadcn.Toggle("Bold", state.Pressed).ID("toggle").OnClick(onToggle),
			chadcn.Switch("Notifications", state.Notify).ID("notify").OnClick(onNotify),
			chadcn.Muted(state.Status),
		),
		chadcn.CardFooter(
			chadcn.Muted("State lives in Go; every click re-renders this card."),
		),
	)
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
