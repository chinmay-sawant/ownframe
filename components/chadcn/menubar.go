package chadcn

import (
	"context"

	"github.com/chinmay-sawant/ownframe/ui"
)

func menubarMenuSkin() menuSkin {
	return menuSkin{
		surface:   "of-menubar-menu",
		item:      "of-menubar-menu-item",
		danger:    "of-menubar-menu-item--danger",
		shortcut:  "of-menubar-menu-shortcut",
		label:     "of-menubar-menu-label",
		separator: "of-menubar-menu-separator",
		chevron:   "of-menubar-menu-chevron",
		row:       "of-menubar-menu-row",
		submenu:   "of-menubar-menu-submenu",
	}
}

// MenubarMenu is one top-level trigger and its items. ID identifies the menu
// in MenubarOptions.Open and derives the trigger and menu part IDs.
type MenubarMenu struct {
	ID    string
	Label string
	Items []MenuItem
}

// MenubarOptions describes a controlled menubar. Open holds the ID of the one
// open menu, or the empty string when every menu is closed. Menus render left
// to right. OnOpenChange receives the next open ID, or the empty string when
// the open trigger is clicked again. OnSelect receives the activated item
// label.
type MenubarOptions struct {
	ID           string
	Label        string
	Open         string
	Menus        []MenubarMenu
	OnOpenChange func(ctx context.Context, id string) error
	OnSelect     func(ctx context.Context, label string) error
}

// Menubar renders a horizontal bar of menu triggers and, when Open matches a
// menu ID, that menu below its trigger. The bar carries role="menubar" and
// each trigger role="menuitem" with aria-haspopup and aria-expanded. Menus
// need an ID to open. The host owns Escape handling and click-away.
func Menubar(options MenubarOptions) *ui.Element {
	root := ui.Div().Class("of-menubar")
	if options.ID != "" {
		root.ID(options.ID)
	}
	bar := ui.Div().Class("of-menubar-bar").Attr("role", "menubar")
	if options.Label != "" {
		bar.Attr("aria-label", options.Label)
	}
	for _, menu := range options.Menus {
		bar.Children(menubarSlot(menu, options))
	}
	return root.Children(bar)
}

// menubarSlot builds one trigger and, when it is the open menu, its surface.
func menubarSlot(menu MenubarMenu, options MenubarOptions) *ui.Element {
	open := menu.ID != "" && menu.ID == options.Open
	menuID := partID(menu.ID, "-menu")
	trigger := ui.Button(menu.Label).Attr("type", "button").Attr("role", "menuitem").
		Class(classes("of-menubar-trigger", "of-focus", "of-transition")).
		Attr("aria-haspopup", "menu").Attr("aria-expanded", ariaBool(open))
	if id := partID(menu.ID, "-trigger"); id != "" {
		trigger.ID(id)
	}
	if open && menuID != "" {
		trigger.Attr("aria-controls", menuID)
	}
	if options.OnOpenChange != nil {
		next := ""
		if !open {
			next = menu.ID
		}
		trigger.OnClick(func(ctx context.Context) error { return options.OnOpenChange(ctx, next) })
	}
	slot := ui.Div().Class("of-menubar-slot").Children(trigger)
	if open {
		skin := menubarMenuSkin()
		slot.Children(menuPane(&skin, menuID, menu.Label, menu.Items, nil, options.OnSelect))
	}
	return slot
}
