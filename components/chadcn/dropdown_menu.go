package chadcn

import (
	"context"

	"github.com/chinmay-sawant/ownframe/ui"
)

// MenuItem is one row in a menu surface. Separator draws a rule and ignores the
// other fields. Children mark a submenu affordance and render as a nested
// surface shown on hover. Action becomes data-action when the item has no
// select callback, so a host can route it through Handlers.Click.
//
// Select precedence is item OnSelect, then the menu-level OnSelect with the
// item label, then Action.
type MenuItem struct {
	Label     string
	Shortcut  string
	Disabled  bool
	Danger    bool
	Separator bool
	Action    string
	OnSelect  func(ctx context.Context) error
	Children  []MenuItem
}

// MenuGroup is a labeled run of menu items. A menu renders its flat Items
// first, then each group with a separator between sections.
type MenuGroup struct {
	Label string
	Items []MenuItem
}

// menuSkin names the classes of one menu surface. Each menu component keeps
// its own skin so one stylesheet file owns one component.
type menuSkin struct {
	surface   string
	item      string
	danger    string
	shortcut  string
	label     string
	separator string
	chevron   string
	row       string
	submenu   string
}

func dropdownMenuSkin() menuSkin {
	return menuSkin{
		surface:   "of-dropdown-menu-content",
		item:      "of-dropdown-menu-item",
		danger:    "of-dropdown-menu-item--danger",
		shortcut:  "of-dropdown-menu-shortcut",
		label:     "of-dropdown-menu-label",
		separator: "of-dropdown-menu-separator",
		chevron:   "of-dropdown-menu-chevron",
		row:       "of-dropdown-menu-row",
		submenu:   "of-dropdown-menu-submenu",
	}
}

// DropdownMenuOptions describes a controlled dropdown menu. Open comes from the
// host. ID names the root and derives the menu ID. Trigger is rendered in both
// states. Items are flat rows; Groups add labeled sections after them.
// OnOpenChange receives the next state when the trigger is clicked. OnSelect
// receives the label of the activated item.
type DropdownMenuOptions struct {
	// AlignEnd aligns the panel and trigger right edges.
	AlignEnd     bool
	ID           string
	Label        string
	Open         bool
	Trigger      *ui.Element
	Items        []MenuItem
	Groups       []MenuGroup
	OnOpenChange func(ctx context.Context, open bool) error
	OnSelect     func(ctx context.Context, label string) error
}

// DropdownMenu renders a host-supplied trigger and, when Open is true, an
// anchored menu. The trigger receives aria-haspopup, aria-expanded, and, while
// open, aria-controls. When OnOpenChange is set the trigger click is wired to
// it, so pass a trigger without its own action. The host owns Escape handling,
// click-away, and focus scope.
func DropdownMenu(options DropdownMenuOptions) *ui.Element {
	if options.Trigger == nil && !options.Open {
		return nil
	}
	menuID := partID(options.ID, "-menu")
	trigger := options.Trigger
	if trigger != nil {
		trigger.Attr("aria-haspopup", "menu").Attr("aria-expanded", ariaBool(options.Open))
		if options.Open && menuID != "" {
			trigger.Attr("aria-controls", menuID)
		}
		if options.OnOpenChange != nil {
			next := !options.Open
			trigger.OnClick(func(ctx context.Context) error { return options.OnOpenChange(ctx, next) })
		}
	}
	root := ui.Div().Class("of-dropdown-menu")
	if options.AlignEnd {
		root.Class("of-dropdown-menu of-dropdown-menu--align-end")
	}
	if options.ID != "" {
		root.ID(options.ID)
	}
	root.Children(trigger)
	if options.Open {
		skin := dropdownMenuSkin()
		root.Children(menuPane(&skin, menuID, options.Label, options.Items, options.Groups, options.OnSelect))
	}
	return root
}

// menuPane builds one menu surface with its items and groups.
func menuPane(skin *menuSkin, id, label string, items []MenuItem, groups []MenuGroup, onSelect func(context.Context, string) error) *ui.Element {
	pane := ui.Div().Class(skin.surface).Attr("role", "menu")
	if id != "" {
		pane.ID(id)
	}
	if label != "" {
		pane.Attr("aria-label", label)
	}
	appendMenuSections(pane, skin, items, groups, onSelect)
	return pane
}

// appendMenuSections writes the flat items then the labeled groups, with a
// separator between every written section.
func appendMenuSections(pane *ui.Element, skin *menuSkin, items []MenuItem, groups []MenuGroup, onSelect func(context.Context, string) error) {
	wrote := appendMenuItems(pane, skin, items, onSelect)
	for _, group := range groups {
		if wrote {
			pane.Children(menuSeparator(skin))
		}
		if group.Label != "" {
			pane.Children(ui.Div().Class(skin.label).Text(group.Label))
			wrote = true
		}
		if appendMenuItems(pane, skin, group.Items, onSelect) {
			wrote = true
		}
	}
}

// appendMenuItems writes rows in order and reports whether it wrote anything.
func appendMenuItems(pane *ui.Element, skin *menuSkin, items []MenuItem, onSelect func(context.Context, string) error) bool {
	wrote := false
	for _, item := range items {
		if item.Separator {
			pane.Children(menuSeparator(skin))
			wrote = true
			continue
		}
		pane.Children(menuItemNode(skin, item, onSelect))
		wrote = true
	}
	return wrote
}

// menuItemNode builds a row, wrapping it in a positioned slot when the item
// owns a submenu.
func menuItemNode(skin *menuSkin, item MenuItem, onSelect func(context.Context, string) error) *ui.Element {
	button := menuItemButton(skin, item, onSelect)
	if len(item.Children) == 0 {
		return button
	}
	row := ui.Div().Class(skin.row).Children(button)
	submenu := ui.Div().Class(classes(skin.surface, skin.submenu)).Attr("role", "menu")
	appendMenuItems(submenu, skin, item.Children, onSelect)
	return row.Children(submenu)
}

// menuItemButton builds the interactive row. A disabled row keeps its markup
// but carries disabled and aria-disabled so it cannot dispatch.
func menuItemButton(skin *menuSkin, item MenuItem, onSelect func(context.Context, string) error) *ui.Element {
	class := skin.item
	if item.Danger {
		class = classes(class, skin.danger)
	}
	button := ui.Button("").Attr("type", "button").Attr("role", "menuitem").
		Class(classes(class, "of-focus", "of-transition"))
	if item.Disabled {
		setDisabled(button, true)
		button.Attr("aria-disabled", "true")
	} else {
		wireMenuSelect(button, item, onSelect)
	}
	if len(item.Children) > 0 {
		button.Attr("aria-haspopup", "menu")
	}
	button.Children(ui.Span().Text(item.Label))
	if item.Shortcut != "" {
		button.Children(ui.Span().Class(skin.shortcut).Text(item.Shortcut))
	}
	if len(item.Children) > 0 {
		button.Children(ui.Span().Class(skin.chevron).Attr("aria-hidden", "true").Text("›"))
	}
	return button
}

func wireMenuSelect(button *ui.Element, item MenuItem, onSelect func(context.Context, string) error) {
	switch {
	case item.OnSelect != nil:
		button.OnClick(item.OnSelect)
	case onSelect != nil:
		button.OnClick(func(ctx context.Context) error { return onSelect(ctx, item.Label) })
	case item.Action != "":
		button.Attr("data-action", item.Action)
	}
}

func menuSeparator(skin *menuSkin) *ui.Element {
	return ui.Div().Class(skin.separator).Attr("role", "separator")
}
