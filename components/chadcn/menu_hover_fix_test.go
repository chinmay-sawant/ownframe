package chadcn_test

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/pagetest"
	"github.com/chinmay-sawant/ownframe/ui"
)

// menuHoverItems is the shared row set: a leaf, an enabled submenu, and a
// disabled submenu. Separators are omitted so the generated indexes stay put.
func menuHoverItems() []chadcn.MenuItem {
	return []chadcn.MenuItem{
		{Label: "Profile"},
		{Label: "Invite", Children: []chadcn.MenuItem{{Label: "Email"}}},
		{Label: "Locked", Disabled: true, Children: []chadcn.MenuItem{{Label: "Hidden"}}},
	}
}

type menuHoverSurface struct {
	name   string
	prefix string
	row    string
	item   string
	sub    string
	root   func(func(context.Context, string) error) *ui.Element
}

func menuHoverSurfaces() []menuHoverSurface {
	items := menuHoverItems()
	return []menuHoverSurface{
		{
			name:   "dropdown",
			prefix: "actions-menu",
			row:    "of-dropdown-menu-row",
			item:   "of-dropdown-menu-item",
			sub:    "of-dropdown-menu-submenu",
			root: func(onSelect func(context.Context, string) error) *ui.Element {
				return chadcn.DropdownMenu(chadcn.DropdownMenuOptions{
					ID:       "actions",
					Label:    "Account",
					Open:     true,
					Trigger:  ui.Button("Open"),
					Items:    items,
					OnSelect: onSelect,
				})
			},
		},
		{
			name:   "context",
			prefix: "canvas-menu",
			row:    "of-context-menu-row",
			item:   "of-context-menu-item",
			sub:    "of-context-menu-submenu",
			root: func(onSelect func(context.Context, string) error) *ui.Element {
				return chadcn.ContextMenu(chadcn.ContextMenuOptions{
					ID:       "canvas-menu",
					Label:    "Canvas",
					Open:     true,
					X:        24,
					Y:        40,
					Items:    items,
					OnSelect: onSelect,
				})
			},
		},
		{
			name:   "menubar",
			prefix: "file-menu",
			row:    "of-menubar-menu-row",
			item:   "of-menubar-menu-item",
			sub:    "of-menubar-menu-submenu",
			root: func(onSelect func(context.Context, string) error) *ui.Element {
				return chadcn.Menubar(chadcn.MenubarOptions{
					ID:    "app-menus",
					Label: "Application",
					Open:  "file",
					Menus: []chadcn.MenubarMenu{{
						ID:    "file",
						Label: "File",
						Items: items,
					}},
					OnSelect: onSelect,
				})
			},
		},
	}
}

// TestMenuHoverRulesOpenFromTheRow pins the selectors that show a submenu from
// the row id and keep a disabled row shut. Each menu skin has the same shape.
func TestMenuHoverRulesOpenFromTheRow(t *testing.T) {
	t.Parallel()
	for _, surface := range menuHoverSurfaces() {
		for _, selector := range []string{
			"." + surface.row + ":hover ." + surface.sub,
			"." + surface.row + ":has(:hover) ." + surface.sub,
			"." + surface.row + ":has(> ." + surface.item + "[disabled]) ." + surface.sub,
		} {
			if !strings.Contains(chadcn.Stylesheet, selector+" {") &&
				!strings.Contains(chadcn.Stylesheet, selector+",") {
				t.Errorf("%s stylesheet missing %s", surface.name, selector)
			}
		}
	}
}

// TestMenuHoverRowOwnsEnabledID checks the id the :hover rule names. An
// enabled submenu puts it on the row and leaves the button unnamed. A
// disabled submenu keeps it on the button.
func TestMenuHoverRowOwnsEnabledID(t *testing.T) {
	t.Parallel()
	for _, surface := range menuHoverSurfaces() {
		t.Run(surface.name, func(t *testing.T) {
			t.Parallel()
			markup := renderMenuMarkup(t, surface.root(nil))
			rowID := surface.prefix + "-item-1"
			lockedID := surface.prefix + "-item-2"
			row := compositeTag(t, markup, `id="`+rowID+`"`)
			if !strings.HasPrefix(row, "<div ") || !strings.Contains(row, `class="`+surface.row+`"`) {
				t.Fatalf("enabled submenu id is not on the row: %s", row)
			}
			trigger := menuHoverRowButton(t, markup, rowID)
			if strings.Contains(trigger, "id=") {
				t.Fatalf("enabled submenu button took an id: %s", trigger)
			}
			if strings.Contains(markup, `class="`+surface.row+`" id="`+lockedID+`"`) {
				t.Fatalf("disabled submenu id moved onto the row:\n%s", markup)
			}
			locked := compositeTag(t, markup, `id="`+lockedID+`"`)
			if !strings.HasPrefix(locked, "<button ") || !strings.Contains(locked, `disabled="disabled"`) {
				t.Fatalf("disabled submenu id is not on the button: %s", locked)
			}
			sub := compositeTag(t, markup, `id="`+rowID+`-sub"`)
			if !strings.Contains(sub, surface.sub) {
				t.Fatalf("submenu id missing its class: %s", sub)
			}
		})
	}
}

// TestMenuHoverSubmenuFollowsRow proves an enabled row opens its submenu, a
// disabled row does not, and a click on an enabled item still runs OnSelect.
func TestMenuHoverSubmenuFollowsRow(t *testing.T) {
	t.Parallel()
	for _, surface := range menuHoverSurfaces() {
		t.Run(surface.name, func(t *testing.T) {
			t.Parallel()
			menuHoverSubmenuFollowsRow(t, surface)
		})
	}
}

func menuHoverSubmenuFollowsRow(t *testing.T, surface menuHoverSurface) {
	t.Helper()
	ctx := context.Background()
	var selected []string
	page := menuHoverPage(t, surface.root(func(_ context.Context, label string) error {
		selected = append(selected, label)
		return nil
	}))
	rowID := surface.prefix + "-item-1"
	subID := rowID + "-sub"
	lockedID := surface.prefix + "-item-2"
	lockedSubID := lockedID + "-sub"

	menuHoverHidden(t, page, subID)
	menuHoverHidden(t, page, lockedSubID)

	menuHoverAt(t, ctx, page, rowID)
	menuHoverShown(t, page, subID)
	menuHoverHidden(t, page, lockedSubID)

	menuHoverClick(t, ctx, page, surface.prefix+"-item-0")
	menuHoverClick(t, ctx, page, rowID)
	menuHoverClick(t, ctx, page, rowID+"-item-0")
	if strings.Join(selected, ",") != "Profile,Invite,Email" {
		t.Fatalf("selected=%v want [Profile Invite Email]", selected)
	}

	menuHoverAt(t, ctx, page, lockedID)
	menuHoverHidden(t, page, lockedSubID)
	before := len(selected)
	menuHoverClick(t, ctx, page, lockedID)
	if len(selected) != before {
		t.Fatalf("disabled row selected %v", selected[before:])
	}
}

// TestMenuHoverSelectWithoutHostID keeps OnSelect on a submenu row when the
// host gave the menu no id. The row still receives the action, so a click on
// the unnamed button resolves to it.
func TestMenuHoverSelectWithoutHostID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var selected []string
	page := menuHoverPage(t, chadcn.DropdownMenu(chadcn.DropdownMenuOptions{
		Open:    true,
		Trigger: ui.Button("Open"),
		Items:   menuHoverItems(),
		OnSelect: func(_ context.Context, label string) error {
			selected = append(selected, label)
			return nil
		},
	}))
	menuHoverClickText(t, ctx, page, "Profile")
	menuHoverClickText(t, ctx, page, "Invite")
	if strings.Join(selected, ",") != "Profile,Invite" {
		t.Fatalf("selected=%v want [Profile Invite]", selected)
	}
}

// TestNavigationMenuHoverTargets checks that a row carries the id named by
// the :hover rule that paints it. A static entry is the list item. A button
// entry puts the id on the button, which is the rule that highlights it.
func TestNavigationMenuHoverTargets(t *testing.T) {
	t.Parallel()
	for _, selector := range []string{
		".of-navigation-menu-entry:hover {",
		".of-navigation-menu-entry-button:hover,",
	} {
		if !strings.Contains(chadcn.Stylesheet, selector) {
			t.Fatalf("stylesheet missing %s", selector)
		}
	}

	markup := renderMenuMarkup(t, chadcn.NavigationMenu(chadcn.NavigationMenuOptions{
		ID:    "primary",
		Label: "Product",
		Open:  "platform",
		Items: []chadcn.NavigationItem{
			{
				ID:    "platform",
				Label: "Platform",
				Children: []chadcn.NavigationItem{
					{ID: "logs", Label: "Logs", OnSelect: func(context.Context) error { return nil }},
					{Label: "Traces", Action: "nav:traces"},
					{Label: "Incident room"},
				},
			},
		},
	}))

	logs := compositeTag(t, markup, `id="logs"`)
	if !strings.HasPrefix(logs, "<button ") || !strings.Contains(logs, `class="of-navigation-menu-entry-button"`) {
		t.Fatalf("selectable entry id is not on the button: %s", logs)
	}
	traces := compositeTag(t, markup, `id="platform-entry-1"`)
	if !strings.HasPrefix(traces, "<button ") || !strings.Contains(traces, `class="of-navigation-menu-entry-button"`) {
		t.Fatalf("action entry id is not on the button: %s", traces)
	}
	incident := compositeTag(t, markup, `id="platform-entry-2"`)
	if !strings.HasPrefix(incident, "<li ") || !strings.Contains(incident, `class="of-navigation-menu-entry"`) {
		t.Fatalf("static entry id is not on the list item: %s", incident)
	}
	if strings.Contains(incident, "entry-button") {
		t.Fatalf("static entry id is on the button rule: %s", incident)
	}
	interactive := compositeTag(t, markup, `class="of-navigation-menu-entry of-navigation-menu-entry--interactive"`)
	if strings.Contains(interactive, "id=") {
		t.Fatalf("interactive row kept the id the button rule needs: %s", interactive)
	}
}

func menuHoverPage(t *testing.T, root *ui.Element) *ownframe.Page {
	t.Helper()
	page, err := ui.NewPage(ownframe.Config{
		Width:  800,
		Height: 600,
		Theme:  chadcn.Stylesheet,
	}, root)
	if err != nil {
		t.Fatal(err)
	}
	if err := page.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	return page
}

func menuHoverAt(t *testing.T, ctx context.Context, page *ownframe.Page, id string) {
	t.Helper()
	box, ok := pagetest.BoxByID(page, id)
	if !ok || box.W <= 0 || box.H <= 0 {
		t.Fatalf("missing hover target %s: ok=%v box=%+v", id, ok, box)
	}
	if err := page.Edit().Hover(ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
		t.Fatal(err)
	}
}

func menuHoverClick(t *testing.T, ctx context.Context, page *ownframe.Page, id string) {
	t.Helper()
	if err := pagetest.Click(ctx, page, id); err != nil {
		t.Fatal(err)
	}
}

func menuHoverClickText(t *testing.T, ctx context.Context, page *ownframe.Page, label string) {
	t.Helper()
	for _, box := range page.Boxes() {
		if box.Tag != "button" || box.W <= 0 || box.H <= 0 {
			continue
		}
		if box.Text != label && !strings.HasPrefix(box.Text, label) {
			continue
		}
		if strings.Contains(box.Text, "Email") || strings.Contains(box.Text, "Hidden") {
			continue
		}
		if err := page.Click(ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
			t.Fatal(err)
		}
		return
	}
	t.Fatalf("missing button %q", label)
}

func menuHoverShown(t *testing.T, page *ownframe.Page, id string) {
	t.Helper()
	box, ok := pagetest.BoxByID(page, id)
	if !ok || box.W <= 0 || box.H <= 0 {
		t.Fatalf("%s did not open: ok=%v box=%+v", id, ok, box)
	}
}

func menuHoverHidden(t *testing.T, page *ownframe.Page, id string) {
	t.Helper()
	box, ok := pagetest.BoxByID(page, id)
	if ok && box.W > 0 && box.H > 0 {
		t.Fatalf("%s is visible: %+v", id, box)
	}
}

// menuHoverRowButton is the item button inside the row that owns rowID. The
// menubar trigger also carries aria-haspopup, so the search starts at the row.
func menuHoverRowButton(t *testing.T, markup, rowID string) string {
	t.Helper()
	marker := `id="` + rowID + `"`
	start := strings.Index(markup, marker)
	if start < 0 {
		t.Fatalf("missing %s in %s", rowID, markup)
	}
	rest := markup[start:]
	buttonAt := strings.Index(rest, "<button ")
	if buttonAt < 0 {
		t.Fatalf("row %s has no button in %s", rowID, markup)
	}
	end := strings.Index(rest[buttonAt:], ">")
	if end < 0 {
		t.Fatalf("row %s button is malformed", rowID)
	}
	return rest[buttonAt : buttonAt+end+1]
}
