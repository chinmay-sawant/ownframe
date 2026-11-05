package chadcn_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

func TestSearchOpenHitsFieldBackdropAndPanel(t *testing.T) {
	var events []string
	root := chadcn.Search(chadcn.SearchOptions{
		ID:    "site",
		Open:  true,
		Query: "but",
		Items: []chadcn.SearchItem{
			{Value: "buttons", Label: "Buttons"},
			{Value: "badges", Label: "Badges"},
		},
		OnOpenChange: func(_ context.Context, open bool) error {
			events = append(events, "open:"+boolWord(open))
			return nil
		},
		OnSelect: func(_ context.Context, value string) error {
			events = append(events, "select:"+value)
			return nil
		},
	})
	view, err := ui.Mount(ownframe.Config{
		Width:  400,
		Height: 300,
		Theme:  chadcn.Stylesheet,
	}, ui.Div().Height(ui.Px(300)).Children(
		ui.Div().MarginTop(ui.Px(80)).MarginLeft(ui.Px(40)).Width(ui.Px(240)).Children(root),
	))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := view.Page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	p := view.Page
	var hit string
	view.Handle(ownframe.Handlers{Click: func(_ context.Context, b ownframe.Box) error {
		hit = b.ID
		return nil
	}})
	field, ok := findBox(p, "site-input")
	if !ok {
		t.Fatalf("no field\n%s", searchHitDump(p))
	}
	rootBox, ok := findBox(p, "site")
	if !ok {
		t.Fatalf("no search root\n%s", searchHitDump(p))
	}
	if field.W < rootBox.W-1 {
		t.Fatalf("field width %.0f does not cover the root %.0f\n%s", field.W, rootBox.W, searchHitDump(p))
	}
	back, ok := findBox(p, "site-backdrop")
	if !ok || back.H < 100 || back.W < 100 {
		t.Fatalf("backdrop box = %+v, want a viewport layer\n%s", back, searchHitDump(p))
	}

	for _, pt := range []struct {
		name string
		x, y float64
	}{
		{"left", field.X + 4, field.Y + field.H/2},
		{"middle", field.X + field.W/2, field.Y + field.H/2},
		{"right", field.X + field.W - 2, field.Y + field.H/2},
	} {
		hit = ""
		before := len(events)
		if err := p.Click(ctx, pt.x, pt.y); err != nil {
			t.Fatal(err)
		}
		if hit != "site-input" {
			t.Fatalf("%s of the field hit %q, want site-input\n%s", pt.name, hit, searchHitDump(p))
		}
		if len(events) != before {
			t.Fatalf("%s of the field dispatched %v", pt.name, events[before:])
		}
	}

	hit = ""
	if err := p.Click(ctx, 10, 10); err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0] != "open:false" {
		t.Fatalf("outside click events=%v, want backdrop close\n%s", events, searchHitDump(p))
	}
	gapY := field.Y + field.H + 4
	if panel, ok := findBox(p, "site-panel"); ok && gapY >= panel.Y {
		t.Fatalf("gap point %.0f is inside the panel at %.0f", gapY, panel.Y)
	}
	if err := p.Click(ctx, field.X+8, gapY); err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 || events[1] != "open:false" {
		t.Fatalf("gap between field and panel events=%v, want backdrop close", events)
	}

	option, ok := findBox(p, "site-option-0")
	if !ok {
		t.Fatalf("no option\n%s", searchHitDump(p))
	}
	if err := p.Click(ctx, option.X+option.W/2, option.Y+option.H/2); err != nil {
		t.Fatal(err)
	}
	if len(events) != 3 || events[2] != "select:buttons" {
		t.Fatalf("panel click events=%v, want select:buttons\n%s", events, searchHitDump(p))
	}
}

func TestComboboxRebuildShowsSameString(t *testing.T) {
	items := []chadcn.ComboboxItem{
		{Value: "next", Label: "Next.js"},
		{Value: "svelte", Label: "SvelteKit"},
	}
	closed := chadcn.Combobox(chadcn.ComboboxOptions{
		ID: "pick", Query: "sv", Value: "next", Items: items,
	})
	closedHTML, err := ui.Render(closed)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(closedHTML, `value="Next.js"`) || strings.Contains(closedHTML, `value="sv"`) {
		t.Fatalf("closed field is not the host value:\n%s", closedHTML)
	}
	if strings.Contains(closedHTML, `role="listbox"`) {
		t.Fatalf("closed combobox rendered a list:\n%s", closedHTML)
	}

	open := chadcn.Combobox(chadcn.ComboboxOptions{
		ID: "pick", Open: true, Query: "sv", Value: "next", Items: items,
	})
	openHTML, err := ui.Render(open)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(openHTML, `value="sv"`) || strings.Contains(openHTML, `value="Next.js"`) {
		t.Fatalf("open field is not the filter string:\n%s", openHTML)
	}
	if !strings.Contains(openHTML, "SvelteKit") || strings.Contains(openHTML, "Next.js") {
		t.Fatalf("open list is not filtered by the field string:\n%s", openHTML)
	}

	ctx := context.Background()
	view, err := ui.Mount(ownframe.Config{Width: 480, Height: 320, Theme: chadcn.Stylesheet}, closed)
	if err != nil {
		t.Fatal(err)
	}
	if err := view.Page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if got := view.Page.Form().FormValue("pick-input"); got != "Next.js" {
		t.Fatalf("closed field %q, want the host value", got)
	}
	if err := view.Update(ctx, open); err != nil {
		t.Fatal(err)
	}
	if err := view.Page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if got := view.Page.Form().FormValue("pick-input"); got != "sv" {
		t.Fatalf("open field after one rebuild %q, want the query", got)
	}
	rebuilt := view.Page.HTML()
	if !strings.Contains(rebuilt, "SvelteKit") || strings.Contains(rebuilt, "Next.js") {
		t.Fatalf("rebuilt list is not filtered by the field string:\n%s", rebuilt)
	}
	if err := view.Update(ctx, closed); err != nil {
		t.Fatal(err)
	}
	if err := view.Page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if got := view.Page.Form().FormValue("pick-input"); got != "Next.js" {
		t.Fatalf("closed field after one rebuild %q, want the host value", got)
	}
}

func boolWord(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func searchHitDump(p *ownframe.Page) string {
	var lines []string
	for _, b := range p.Boxes() {
		if b.ID == "" && b.W < 2 && b.H < 2 {
			continue
		}
		lines = append(lines, fmt.Sprintf("id=%q tag=%s x=%.0f y=%.0f w=%.0f h=%.0f", b.ID, b.Tag, b.X, b.Y, b.W, b.H))
	}
	return strings.Join(lines, "\n")
}

func findBox(p *ownframe.Page, id string) (ownframe.Box, bool) {
	var found ownframe.Box
	ok := false
	for _, b := range p.Boxes() {
		if b.ID == id {
			found = b
			ok = true
		}
	}
	return found, ok
}
