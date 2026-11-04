// Command mygo opens the compare thousand-row table in MyGo native UI.
//
// Best-practice form: the rows live in a virtualized ui.List that builds
// only the rows in view, with the choice held in ListState.
package main

import (
	"fmt"
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

const (
	listWidth    = 800
	listHeight   = 600
	listRows     = 1000
	listSubtitle = "mygo · Go view → GPU, no HTML"
	listBody     = "Lorem ipsum dolor sit amet."
	// columnsWidth keeps rows and header at one width so the table aligns
	// and never scrolls sideways.
	columnsWidth = 752
)

// listApp is the app state the window shows: the list place and the choice.
type listApp struct {
	list     ui.ListState
	selected int
}

// view builds the interface from the state, on the main thread, whenever the
// window needs a frame.
func (s *listApp) view(c *ui.Context) {
	s.list.Selected = &s.selected
	s.list.Label = func(i int) string { return "Item " + fmt.Sprint(i+1) }
	t := c.Theme()
	white := ui.Hex("#ffffff")
	ui.Column(c).Fill().Children(func() {
		ui.Column(c).Gap(4).Padding(12).Children(func() {
			ui.Text(c, "Compare Table · 1000 rows").FontSize(18).Bold()
			ui.Text(c, listSubtitle).FontSize(12).TextColor(t.TextMuted)
			ui.Text(c, s.selectedText()).FontSize(14)
		})
		ui.Row(c).Gap(12).Padding(8).Background(ui.Hex("#1c1915")).Children(func() {
			ui.Text(c, "Num").FontSize(13).Bold().Width(64).TextColor(white)
			ui.Text(c, "Title").FontSize(13).Bold().Width(200).TextColor(white)
			ui.Text(c, "Description").FontSize(13).Bold().TextColor(white)
		})
		ui.List(c, &s.list, listRows, func(i int) {
			row := ui.Row(c).Gap(12).Padding(8)
			if i != s.selected {
				if i%2 == 1 {
					row.Background(ui.Hex("#f3f4f6"))
				} else {
					row.Background(ui.Hex("#ffffff"))
				}
			}
			row.Children(func() {
				ui.Text(c, "#"+fmt.Sprint(i+1)).FontSize(13).Width(64)
				ui.Text(c, "Item "+fmt.Sprint(i+1)).FontSize(14).Bold().Width(200)
				ui.Text(c, listBody).FontSize(13)
			})
		}).Grow(1).RowWidth(columnsWidth)
	})
}

func (s *listApp) selectedText() string {
	if s.selected < 0 || s.selected >= listRows {
		return "Selected: none"
	}
	return "Selected: Item " + fmt.Sprint(s.selected+1)
}

func main() {
	s := &listApp{selected: -1}
	mygo.App.WhenReady(func() {
		mygo.NewWindow(mygo.WindowOptions{
			Title:   "Compare Table",
			Width:   listWidth,
			Height:  listHeight,
			Content: ui.View(s.view),
		})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
