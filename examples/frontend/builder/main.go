package main

import (
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"log"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

const (
	windowWidth  = 640
	windowHeight = 640
)

//go:embed template.html
var templateHTML string

//go:embed style.css
var stylesheet string

type view struct {
	// Body holds only ui.Render output from the fixed chadcn
	// constructors in renderBody, so it is safe to skip escaping.
	Body    template.HTML
	Pressed bool
	Notify  bool
	Clicks  int
	Status  string
}

func renderBody(state *view) (template.HTML, error) {
	count := "Run action"
	if state.Clicks > 0 {
		count = fmt.Sprintf("Ran %d time", state.Clicks)
		if state.Clicks != 1 {
			count += "s"
		}
	}
	card := chadcn.Card(
		chadcn.CardHeader(
			chadcn.CardTitle("Builder basics"),
			chadcn.CardDescription("One card, host-owned state, dark theme."),
		),
		chadcn.CardContent(
			chadcn.Button(count, chadcn.ButtonOptions{Variant: chadcn.ButtonFilled}).ID("press").Attr("data-action", "press"),
			chadcn.Toggle("Bold", state.Pressed).ID("toggle").Attr("data-action", "toggle"),
			chadcn.Switch("Notifications", state.Notify).ID("notify").Attr("data-action", "notify"),
			chadcn.Muted(state.Status),
		),
		chadcn.CardFooter(
			chadcn.Muted("State lives in Go; every click re-renders this card."),
		),
	)
	out, err := ui.Render(card)
	if err != nil {
		return "", err
	}
	return template.HTML(out), nil
}

func refresh(page *ownframe.Page, state *view) error {
	state.Status = fmt.Sprintf("Button ran %d times. Toggle %v. Switch %v.", state.Clicks, state.Pressed, state.Notify)
	body, err := renderBody(state)
	if err != nil {
		return err
	}
	state.Body = body
	page.SetData(state)
	return nil
}

func newPage() (*ownframe.Page, *view, error) {
	page, err := ownframe.New(ownframe.Config{
		Title:  "Builder basics",
		Width:  windowWidth,
		Height: windowHeight,
		HTML:   templateHTML,
		Theme:  chadcn.Dark + chadcn.Stylesheet + stylesheet,
	})
	if err != nil {
		return nil, nil, err
	}
	state := &view{}
	if err := refresh(page, state); err != nil {
		return nil, nil, err
	}
	page.Handle(ownframe.Handlers{Click: func(_ context.Context, box ownframe.Box) error {
		switch box.Action {
		case "press":
			state.Clicks++
		case "toggle":
			state.Pressed = !state.Pressed
		case "notify":
			state.Notify = !state.Notify
		default:
			return nil
		}
		return refresh(page, state)
	}})
	return page, state, nil
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
