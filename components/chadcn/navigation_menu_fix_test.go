package chadcn_test

import (
	"context"
	"math"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/pagetest"
	"github.com/chinmay-sawant/ownframe/ui"
)

// navigationFixTree builds a menu with one plain child, one selectable child,
// and one action child so the panel covers every row shape.
func navigationFixTree(open string) *ui.Element {
	return chadcn.NavigationMenu(chadcn.NavigationMenuOptions{
		ID:    "primary",
		Label: "Product",
		Open:  open,
		Items: []chadcn.NavigationItem{
			{
				ID:    "platform",
				Label: "Platform",
				Children: []chadcn.NavigationItem{
					{
						ID:          "logs",
						Label:       "Logs",
						Description: "Search the request stream",
						OnSelect: func(context.Context) error {
							return nil
						},
					},
					{
						ID:          "traces",
						Label:       "Traces",
						Description: "Follow one request across services",
						Action:      "nav:traces",
					},
					{
						ID:          "incident",
						Label:       "Incident room",
						Description: "Coordinate a live response",
					},
				},
			},
			{
				ID:    "resources",
				Label: "Resources",
				Children: []chadcn.NavigationItem{
					{Label: "Documentation"},
				},
			},
		},
	})
}

// navigationFixPage mounts a menu inside a fixed-width topbar, the shape a real
// application chrome uses, so a panel that took part in layout would still be
// caught by a changed bar width.
func navigationFixPage(t *testing.T, open string) *ui.View {
	t.Helper()
	view, err := ui.Mount(ownframe.Config{
		Width:  640,
		Height: 320,
		Theme:  chadcn.Stylesheet,
	}, ui.Div().ID("topbar").Width(ui.Px(600)).Children(navigationFixTree(open)))
	if err != nil {
		t.Fatal(err)
	}
	if err := view.Page.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	return view
}

func navigationFixBox(t *testing.T, p *ownframe.Page, id string) ownframe.Box {
	t.Helper()
	box, ok := pagetest.BoxByID(p, id)
	if !ok {
		t.Fatalf("missing box %q", id)
	}
	return box
}

func TestNavigationMenuFixOpenPanelIsAbsolute(t *testing.T) {
	t.Parallel()
	closed := navigationFixPage(t, "")
	open := navigationFixPage(t, "platform")

	closedBar := navigationFixBox(t, closed.Page, "topbar")
	openBar := navigationFixBox(t, open.Page, "topbar")
	if math.Abs(closedBar.W-openBar.W) > 0.5 {
		t.Fatalf("open panel changed the row width: closed=%v open=%v", closedBar.W, openBar.W)
	}

	closedMenu := navigationFixBox(t, closed.Page, "primary")
	openMenu := navigationFixBox(t, open.Page, "primary")
	if math.Abs(closedMenu.H-openMenu.H) > 0.5 {
		t.Fatalf("open panel changed the menu height: closed=%v open=%v", closedMenu.H, openMenu.H)
	}

	trigger := navigationFixBox(t, open.Page, "platform-trigger")
	panel := navigationFixBox(t, open.Page, "platform-content")
	if panel.Y < trigger.Y+trigger.H {
		t.Fatalf("panel sits inside the trigger row: trigger bottom=%v panel top=%v",
			trigger.Y+trigger.H, panel.Y)
	}
	if math.Abs(panel.Y-(trigger.Y+36)) > 0.5 {
		t.Fatalf("panel top=%v want a fixed 36px below the row at %v", panel.Y, trigger.Y+36)
	}
	if math.Abs(panel.X-openMenu.X) > 0.5 {
		t.Fatalf("panel left=%v want the row left=%v", panel.X, openMenu.X)
	}
	if panel.W <= 0 {
		t.Fatalf("panel width=%v want a laid out panel", panel.W)
	}
}

func TestNavigationMenuFixChildEntriesAreClickable(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var selected int
	var actions []string

	tree := chadcn.NavigationMenu(chadcn.NavigationMenuOptions{
		ID:    "primary",
		Label: "Product",
		Open:  "platform",
		Items: []chadcn.NavigationItem{
			{
				ID:    "platform",
				Label: "Platform",
				Children: []chadcn.NavigationItem{
					{
						Label: "Logs",
						OnSelect: func(context.Context) error {
							selected++
							return nil
						},
					},
					{Label: "Traces", Action: "nav:traces"},
					{Label: "Incident room"},
				},
			},
		},
	})
	view, err := ui.Mount(ownframe.Config{
		Width:  640,
		Height: 320,
		Theme:  chadcn.Stylesheet,
	}, tree)
	if err != nil {
		t.Fatal(err)
	}
	view.Handle(ownframe.Handlers{Click: func(_ context.Context, box ownframe.Box) error {
		actions = append(actions, box.Action)
		return nil
	}})
	if err := view.Page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	source := view.Page.HTML()
	for _, want := range []string{
		`class="of-navigation-menu-content"`,
		`class="of-navigation-menu-entry of-navigation-menu-entry--interactive"`,
		`class="of-navigation-menu-entry-button"`,
		`data-action="nav:traces"`,
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("panel markup missing %q: %s", want, source)
		}
	}
	if got := strings.Count(source, `class="of-navigation-menu-entry-button"`); got != 2 {
		t.Fatalf("clickable rows=%d want 2: %s", got, source)
	}

	var logs ownframe.Box
	found := false
	for _, box := range view.Page.Boxes() {
		if box.Tag == "button" && strings.Contains(box.Text, "Logs") {
			logs = box
			found = true
			break
		}
	}
	if !found {
		t.Fatal("missing Logs button box")
	}
	if err := view.Page.Click(ctx, logs.X+logs.W/2, logs.Y+logs.H/2); err != nil {
		t.Fatal(err)
	}
	if selected != 1 {
		t.Fatalf("select callback ran %d times, want 1", selected)
	}

	if err := pagetest.ClickAction(ctx, view.Page, "nav:traces"); err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 || actions[0] != "nav:traces" {
		t.Fatalf("actions=%v want [nav:traces]", actions)
	}

	for _, box := range view.Page.Boxes() {
		if box.Tag == "button" && strings.Contains(box.Text, "Incident room") {
			t.Fatal("static entry rendered as a button")
		}
	}
}
