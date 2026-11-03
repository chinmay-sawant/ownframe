package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/ui"
)

func BenchmarkLoglensSidebarHover(b *testing.B) {
	for _, retain := range []bool{false, true} {
		name := "full-layout"
		if retain {
			name = "retained-layout"
		}
		b.Run(name, func(b *testing.B) {
			a, err := newApp()
			if err != nil {
				b.Fatal(err)
			}
			view, err := ui.Mount(ownframe.Config{
				Width: windowWidth, Height: windowHeight,
				Theme: a.themeSheet(), RetainState: retain,
			}, a.screen())
			if err != nil {
				b.Fatal(err)
			}
			a.view, a.page = view, view.Page
			view.Handle(ownframe.Handlers{Hover: a.handleHover})
			ctx := context.Background()
			if err := view.Page.Redraw(ctx); err != nil {
				b.Fatal(err)
			}
			buttons := []ownframe.Box{}
			for _, box := range view.Page.Boxes() {
				if box.Action == "nav:overview" || box.Action == "nav:logs" {
					buttons = append(buttons, box)
				}
			}
			if len(buttons) < 2 {
				b.Fatal("dashboard has fewer than two hover targets")
			}
			buttons = append(buttons, ownframe.Box{X: 400, Y: 20})
			buttons = append(buttons, ownframe.Box{X: 500, Y: 400})
			index := 0
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				box := buttons[index%len(buttons)]
				if err := view.Page.Edit().Hover(ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
					b.Fatal(err)
				}
				index++
			}
		})
	}
}

func BenchmarkLoglensPointerHover(b *testing.B) {
	a, err := newApp()
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()
	if err = a.page.Redraw(ctx); err != nil {
		b.Fatal(err)
	}
	var targets []ownframe.Box
	for _, box := range a.page.Boxes() {
		if box.Action == "nav:overview" || box.Action == "nav:logs" || box.ID == "menu-file-trigger" {
			targets = append(targets, box)
		}
	}
	if len(targets) != 3 {
		b.Fatalf("hover targets=%d", len(targets))
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		box := targets[i%len(targets)]
		if err = a.page.Edit().Hover(ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
			b.Fatal(err)
		}
	}
}
