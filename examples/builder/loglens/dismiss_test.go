package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestHeaderSurfacesDismissOnOutsideClick(t *testing.T) {
	tests := []struct {
		name  string
		id    string
		open  func(*app) bool
		setup func(*app)
	}{
		{name: "navigation", id: "app-main", setup: func(a *app) { a.navOpen = "product-platform" }, open: func(a *app) bool { return a.navOpen != "" }},
		{name: "menubar", id: "app-main", setup: func(a *app) { a.menubarOpen = "menu-file" }, open: func(a *app) bool { return a.menubarOpen != "" }},
		{name: "notifications", id: "app-main", setup: func(a *app) { a.notifOpen = true }, open: func(a *app) bool { return a.notifOpen }},
		{name: "account", id: "app-main", setup: func(a *app) { a.userMenuOpen = true }, open: func(a *app) bool { return a.userMenuOpen }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := &app{}
			test.setup(a)
			if err := a.handleClick(context.Background(), ownframe.Box{ID: test.id}); err != nil {
				t.Fatal(err)
			}
			if test.open(a) {
				t.Fatal("surface remained open after clicking outside")
			}
		})
	}
}

func TestHeaderSurfacesStayOpenWhenClickingTheirContent(t *testing.T) {
	tests := []struct {
		name  string
		id    string
		setup func(*app)
		open  func(*app) bool
	}{
		{name: "navigation", id: "product-platform-content", setup: func(a *app) { a.navOpen = "product-platform" }, open: func(a *app) bool { return a.navOpen != "" }},
		{name: "menubar", id: "menu-file-content", setup: func(a *app) { a.menubarOpen = "menu-file" }, open: func(a *app) bool { return a.menubarOpen != "" }},
		{name: "notifications", id: "notifications-popover-content", setup: func(a *app) { a.notifOpen = true }, open: func(a *app) bool { return a.notifOpen }},
		{name: "account", id: "user-menu-menu", setup: func(a *app) { a.userMenuOpen = true }, open: func(a *app) bool { return a.userMenuOpen }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := &app{}
			test.setup(a)
			if err := a.handleClick(context.Background(), ownframe.Box{ID: test.id}); err != nil {
				t.Fatal(err)
			}
			if !test.open(a) {
				t.Fatal("surface closed after clicking inside its content")
			}
		})
	}
}

func TestOpenTopMenuDismissesOnOutsidePointerClick(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	a.navOpen = "product-platform"
	if err := a.refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, ok := findDismissBox(a.page); !ok {
		t.Fatal("open top menu did not mount its outside-click layer")
	}
	if err := a.page.Click(context.Background(), 900, 500); err != nil {
		t.Fatal(err)
	}
	if a.navOpen != "" {
		t.Fatal("navigation menu remained open after clicking outside")
	}
}

func TestResourceEntriesCanBeSelected(t *testing.T) {
	tests := []struct {
		label string
		toast string
	}{
		{label: "Documentation", toast: "Documentation"},
		{label: "Status page", toast: "Status page"},
	}
	for _, test := range tests {
		t.Run(test.label, func(t *testing.T) {
			a, err := newApp()
			if err != nil {
				t.Fatal(err)
			}
			a.navOpen = "product-resources"
			if err := a.refresh(context.Background()); err != nil {
				t.Fatal(err)
			}
			entry := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool { return box.Text == test.label })
			if err := a.page.Click(context.Background(), entry.X+entry.W/2, entry.Y+entry.H/2); err != nil {
				t.Fatal(err)
			}
			if a.navOpen != "" || !a.toastOpen || a.toastTitle != test.toast {
				t.Fatalf("resource selection state: menu=%q toastOpen=%t toastTitle=%q", a.navOpen, a.toastOpen, a.toastTitle)
			}
		})
	}
}

func findDismissBox(page *ownframe.Page) (ownframe.Box, bool) {
	for _, box := range page.Boxes() {
		if box.ID == "top-menu-dismiss" {
			return box, true
		}
	}
	return ownframe.Box{}, false
}
