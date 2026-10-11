// Package demo shows the semantic tree, accessibility actions, and desktop IME.
package demo

import (
	_ "embed"
	"fmt"

	"github.com/chinmay-sawant/ownframe"
)

//go:embed page.html
var source string

const DefaultWidth, DefaultHeight = 980, 720

type View struct {
	Name, Secret, Status, Tree string
	Rows                       []string
}

type App struct {
	page *ownframe.Page
	view View
}

func New() (*App, error) {
	p, err := ownframe.New(ownframe.Config{
		Title: "Accessibility", HTML: source,
		Width: DefaultWidth, Height: DefaultHeight,
		Scrollbar: ownframe.ScrollbarVisible,
	})
	if err != nil {
		return nil, err
	}
	a := &App{page: p, view: View{Name: "Ada", Secret: "private", Status: "Type in the fields, then refresh the tree."}}
	for i := 1; i <= 12; i++ {
		a.view.Rows = append(a.view.Rows, fmt.Sprintf("Scrollable example row %d", i))
	}
	p.Handle(ownframe.Handlers{Click: a.onClick, Change: a.onChange})
	p.SetData(a.view)
	return a, nil
}

func (a *App) Page() *ownframe.Page { return a.page }
