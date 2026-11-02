package chadcn_test

import (
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

func renderMenuMarkup(t *testing.T, root *ui.Element) string {
	t.Helper()
	markup, err := ui.Render(root)
	if err != nil {
		t.Fatalf("render menu: %v", err)
	}
	return markup
}

func requireMenuMarkup(t *testing.T, markup string, fragments ...string) {
	t.Helper()
	for _, fragment := range fragments {
		if !strings.Contains(markup, fragment) {
			t.Errorf("menu markup missing %q:\n%s", fragment, markup)
		}
	}
}

func requireMenuEmpty(t *testing.T, markup string, fragments ...string) {
	t.Helper()
	for _, fragment := range fragments {
		if strings.Contains(markup, fragment) {
			t.Errorf("menu markup unexpectedly contains %q:\n%s", fragment, markup)
		}
	}
}

func TestDropdownMenuClosedRendersTrigger(t *testing.T) {
	t.Parallel()
	markup := renderMenuMarkup(t, chadcn.DropdownMenu(chadcn.DropdownMenuOptions{
		ID:      "actions",
		Trigger: ui.Button("Open").ID("actions-trigger").Class("host-trigger"),
		Items:   []chadcn.MenuItem{{Label: "Profile", Action: "account:profile"}},
	}))
	requireMenuMarkup(t, markup,
		`class="of-dropdown-menu"`,
		`class="host-trigger"`,
		`id="actions-trigger"`,
		`aria-haspopup="menu"`,
		`aria-expanded="false"`,
		`Open`,
	)
	requireMenuEmpty(t, markup,
		"of-dropdown-menu-content",
		`role="menu"`,
		"Profile",
		`aria-controls`,
	)
}

func TestDropdownMenuOpenRendersRows(t *testing.T) {
	t.Parallel()
	markup := renderMenuMarkup(t, chadcn.DropdownMenu(chadcn.DropdownMenuOptions{
		ID:      "actions",
		Label:   "Account",
		Open:    true,
		Trigger: ui.Button("Open").ID("actions-trigger"),
		Items: []chadcn.MenuItem{
			{Label: "Profile", Shortcut: "⇧⌘P", Action: "account:profile"},
			{Separator: true},
			{Label: "Invite users", Children: []chadcn.MenuItem{{Label: "Email"}}},
			{Label: "Delete account", Danger: true, Disabled: true},
		},
		Groups: []chadcn.MenuGroup{{
			Label: "Team",
			Items: []chadcn.MenuItem{{Label: "Settings", Shortcut: "⌘S"}},
		}},
	}))
	requireMenuMarkup(t, markup,
		`class="of-dropdown-menu-content"`,
		`id="actions-menu"`,
		`role="menu"`,
		`aria-label="Account"`,
		`aria-expanded="true"`,
		`aria-controls="actions-menu"`,
		`data-action="account:profile"`,
		`class="of-dropdown-menu-item of-focus of-transition"`,
		`role="menuitem"`,
		`class="of-dropdown-menu-shortcut"`,
		`⇧⌘P`,
		`class="of-dropdown-menu-separator"`,
		`role="separator"`,
		`class="of-dropdown-menu-label"`,
		`Team`,
		`class="of-dropdown-menu-item of-dropdown-menu-item--danger of-focus of-transition"`,
		`aria-disabled="true"`,
		`disabled="disabled"`,
		`aria-haspopup="menu"`,
		`of-dropdown-menu-submenu`,
		`Email`,
		`⌘S`,
	)
}

func TestDropdownMenuOpenWithoutTrigger(t *testing.T) {
	t.Parallel()
	markup := renderMenuMarkup(t, chadcn.DropdownMenu(chadcn.DropdownMenuOptions{
		ID:    "floating",
		Open:  true,
		Items: []chadcn.MenuItem{{Label: "Reload", Action: "view:reload"}},
	}))
	requireMenuMarkup(t, markup,
		`class="of-dropdown-menu"`,
		`id="floating-menu"`,
		`role="menu"`,
		`Reload`,
	)
}

func TestContextMenuClosedIsNil(t *testing.T) {
	t.Parallel()
	if markup := renderMenuMarkup(t, chadcn.ContextMenu(chadcn.ContextMenuOptions{
		Items: []chadcn.MenuItem{{Label: "Back"}},
	})); markup != "" {
		t.Fatalf("closed context menu rendered %q", markup)
	}
}

func TestContextMenuOpenUsesHostCoordinates(t *testing.T) {
	t.Parallel()
	markup := renderMenuMarkup(t, chadcn.ContextMenu(chadcn.ContextMenuOptions{
		ID:    "canvas-menu",
		Label: "Canvas actions",
		Open:  true,
		X:     24,
		Y:     40,
		Items: []chadcn.MenuItem{
			{Label: "Back", Shortcut: "⌘["},
			{Label: "Forward", Disabled: true},
		},
		Groups: []chadcn.MenuGroup{{
			Label: "People",
			Items: []chadcn.MenuItem{{Label: "Pedro Duarte"}},
		}},
	}))
	requireMenuMarkup(t, markup,
		`class="of-context-menu"`,
		`id="canvas-menu"`,
		`role="menu"`,
		`aria-label="Canvas actions"`,
		`position:absolute`,
		`left:24px`,
		`top:40px`,
		`class="of-context-menu-item of-focus of-transition"`,
		`class="of-context-menu-shortcut"`,
		`⌘[`,
		`class="of-context-menu-label"`,
		`People`,
		`class="of-context-menu-separator"`,
		`disabled="disabled"`,
	)
}

func TestMenubarRendersOneOpenMenu(t *testing.T) {
	t.Parallel()
	markup := renderMenuMarkup(t, chadcn.Menubar(chadcn.MenubarOptions{
		ID:    "app-menus",
		Label: "Application",
		Open:  "file",
		Menus: []chadcn.MenubarMenu{
			{ID: "file", Label: "File", Items: []chadcn.MenuItem{{Label: "New Tab", Shortcut: "⌘T"}}},
			{ID: "edit", Label: "Edit", Items: []chadcn.MenuItem{{Label: "Undo"}}},
		},
	}))
	requireMenuMarkup(t, markup,
		`class="of-menubar"`,
		`id="app-menus"`,
		`class="of-menubar-bar"`,
		`role="menubar"`,
		`aria-label="Application"`,
		`class="of-menubar-slot"`,
		`class="of-menubar-trigger of-focus of-transition"`,
		`id="file-trigger"`,
		`id="edit-trigger"`,
		`aria-expanded="true"`,
		`aria-expanded="false"`,
		`aria-controls="file-menu"`,
		`class="of-menubar-menu"`,
		`id="file-menu"`,
		`role="menu"`,
		`class="of-menubar-menu-item of-focus of-transition"`,
		`New Tab`,
		`⌘T`,
	)
	requireMenuEmpty(t, markup,
		`id="edit-menu"`,
		`aria-controls="edit-menu"`,
		"Undo",
	)
}

func TestMenubarWithoutOpenMenu(t *testing.T) {
	t.Parallel()
	markup := renderMenuMarkup(t, chadcn.Menubar(chadcn.MenubarOptions{
		Open: "missing",
		Menus: []chadcn.MenubarMenu{
			{ID: "file", Label: "File", Items: []chadcn.MenuItem{{Label: "New Tab"}}},
		},
	}))
	requireMenuMarkup(t, markup,
		`class="of-menubar-bar"`,
		`role="menubar"`,
		`id="file-trigger"`,
		`aria-expanded="false"`,
	)
	requireMenuEmpty(t, markup,
		`aria-expanded="true"`,
		`class="of-menubar-menu"`,
		`id="file-menu"`,
		"New Tab",
	)
}
