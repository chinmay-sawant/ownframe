// Package demo shows the semantic tree, accessibility actions, and desktop IME.
package demo

import (
	_ "embed"
	"errors"
	"fmt"

	"github.com/chinmay-sawant/ownframe"
)

//go:embed page.html
var source string

const DefaultWidth, DefaultHeight = 980, 720

type View struct {
	Name, Secret, ActionValue, Status, Tree string
	ActivationCount                         int
	Rows                                    []string
}

type App struct {
	page *ownframe.Page
	view View
}

func New() (*App, error) {
	p, err := ownframe.New(ownframe.Config{DevTools: true,
		Title: "Accessibility", HTML: source,
		Width: DefaultWidth, Height: DefaultHeight,
		Scrollbar: ownframe.ScrollbarVisible,
	})
	if err != nil {
		return nil, err
	}
	a := &App{page: p, view: View{
		Name: "Ada", Secret: "private", ActionValue: "Updated through the tree action",
		Tree: "Click Refresh tree after the first render to load the snapshot.",
	}}
	if _, err := a.accessible().AccessibilityTree(); errors.Is(err, ownframe.ErrAccessibilityUnavailable) {
		a.view.Status = "Pre-render availability check passed; render the page to create its first tree."
	} else if err != nil {
		return nil, err
	} else {
		a.view.Status = "Pre-render availability check returned a tree."
	}
	for i := 1; i <= 8; i++ {
		a.view.Rows = append(a.view.Rows, fmt.Sprintf("Scrollable example row %d", i))
	}
	p.Handle(ownframe.Handlers{Click: a.onClick, Change: a.onChange})
	p.SetData(a.view)
	return a, nil
}

func (a *App) Page() *ownframe.Page { return a.page }

func (a *App) accessible() ownframe.Accessible { return a.page }
