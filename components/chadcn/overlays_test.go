package chadcn_test

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// overlayMarkup renders an overlay tree and reports a builder error as a test
// failure.
func overlayMarkup(t *testing.T, root *ui.Element) string {
	t.Helper()
	if root == nil {
		t.Fatal("overlay rendered nil")
	}
	markup, err := ui.Render(root)
	if err != nil {
		t.Fatal(err)
	}
	return markup
}

// overlayWant checks that every expected fragment appears in the markup.
func overlayWant(t *testing.T, markup string, fragments ...string) {
	t.Helper()
	for _, fragment := range fragments {
		if !strings.Contains(markup, fragment) {
			t.Errorf("markup missing %q:\n%s", fragment, markup)
		}
	}
}

// overlayDeny checks that none of the fragments appears in the markup.
func overlayDeny(t *testing.T, markup string, fragments ...string) {
	t.Helper()
	for _, fragment := range fragments {
		if strings.Contains(markup, fragment) {
			t.Errorf("markup unexpectedly contains %q:\n%s", fragment, markup)
		}
	}
}

func TestOverlayDialogOpen(t *testing.T) {
	t.Parallel()
	root := chadcn.Dialog(chadcn.DialogOptions{
		ID:           "profile",
		Open:         true,
		Title:        "Edit profile",
		Description:  "Make changes to your profile here.",
		CloseAction:  "profile:close",
		Trigger:      chadcn.Button("Edit profile", chadcn.ButtonOptions{}),
		Footer:       []*ui.Element{chadcn.Button("Save changes", chadcn.ButtonOptions{Variant: chadcn.ButtonFilled})},
		OnOpenChange: func(context.Context, bool) error { return nil },
	}, ui.Text("Body content"))
	markup := overlayMarkup(t, root)
	overlayWant(t, markup,
		`class="of-dialog"`,
		`class="of-dialog-backdrop"`,
		`class="of-dialog-panel"`,
		`class="of-dialog-header"`,
		`class="of-dialog-heading"`,
		`class="of-dialog-title"`,
		`class="of-dialog-description"`,
		`class="of-dialog-body"`,
		`class="of-dialog-footer"`,
		`class="of-dialog-close"`,
		`id="profile"`,
		`id="profile-panel"`,
		`id="profile-panel-title"`,
		`id="profile-panel-description"`,
		`role="dialog"`,
		`aria-modal="true"`,
		`aria-labelledby="profile-panel-title"`,
		`aria-describedby="profile-panel-description"`,
		`aria-haspopup="dialog"`,
		`aria-expanded="true"`,
		`aria-controls="profile-panel"`,
		`data-action="profile:close"`,
		`aria-label="Close"`,
		"Edit profile",
		"Make changes to your profile here.",
		"Body content",
		"Save changes",
	)
}

func TestOverlayDialogClosed(t *testing.T) {
	t.Parallel()
	root := chadcn.Dialog(chadcn.DialogOptions{
		ID:           "profile",
		Title:        "Edit profile",
		Trigger:      chadcn.Button("Open profile", chadcn.ButtonOptions{}),
		OnOpenChange: func(context.Context, bool) error { return nil },
	}, ui.Text("Body content"))
	markup := overlayMarkup(t, root)
	overlayWant(t, markup, "Open profile", `aria-expanded="false"`, "data-action=")
	overlayDeny(t, markup, "of-dialog-backdrop", "of-dialog-body", `aria-modal="true"`)

	if closed := chadcn.Dialog(chadcn.DialogOptions{Title: "Edit profile"}); closed != nil {
		t.Fatal("closed dialog without a trigger rendered a surface")
	}
}

func TestOverlayDialogLegacyCallShape(t *testing.T) {
	t.Parallel()
	root := chadcn.Dialog(chadcn.DialogOptions{
		BackdropID: "legacy-backdrop", PanelID: "legacy-panel", Label: "Legacy dialog",
		CloseAction: "legacy:close", Width: ui.Px(320),
	}, ui.Text("Legacy body"))
	markup := overlayMarkup(t, root)
	overlayWant(t, markup,
		`id="legacy-backdrop"`,
		`id="legacy-panel"`,
		`class="of-dialog-backdrop"`,
		`class="of-dialog-panel"`,
		`role="dialog"`,
		`aria-label="Legacy dialog"`,
		`data-action="legacy:close"`,
		"Legacy body",
	)
	// The legacy call shape supplies its own header, so no generated title or
	// close control is added.
	overlayDeny(t, markup, `class="of-dialog-title"`, `class="of-dialog-close"`)
}

func TestOverlayAlertDialog(t *testing.T) {
	t.Parallel()
	root := chadcn.AlertDialog(chadcn.AlertDialogOptions{
		ID:          "delete",
		Open:        true,
		Title:       "Are you absolutely sure?",
		Description: "This action cannot be undone.",
		Trigger:     chadcn.Button("Show dialog", chadcn.ButtonOptions{}),
		OnConfirm:   func(context.Context) error { return nil },
		OnCancel:    func(context.Context) error { return nil },
	})
	markup := overlayMarkup(t, root)
	overlayWant(t, markup,
		`class="of-alert-dialog"`,
		`class="of-alert-dialog-backdrop"`,
		`class="of-alert-dialog-panel"`,
		`class="of-alert-dialog-header"`,
		`class="of-alert-dialog-title"`,
		`class="of-alert-dialog-description"`,
		`class="of-alert-dialog-footer"`,
		`class="of-alert-dialog-cancel"`,
		`class="of-alert-dialog-confirm"`,
		`id="delete"`,
		`id="delete-panel"`,
		`role="alertdialog"`,
		`aria-modal="true"`,
		`aria-labelledby="delete-panel-title"`,
		`aria-describedby="delete-panel-description"`,
		"Are you absolutely sure?",
		"This action cannot be undone.",
		"Cancel",
		"Continue",
	)
	overlayDeny(t, markup, `class="of-alert-dialog-close"`)

	custom := chadcn.AlertDialog(chadcn.AlertDialogOptions{
		ID: "remove", Open: true, Title: "Remove file?", Destructive: true,
		ConfirmLabel: "Delete", CancelLabel: "Keep",
	})
	customMarkup := overlayMarkup(t, custom)
	overlayWant(t, customMarkup, "Delete", "Keep")

	if closed := chadcn.AlertDialog(chadcn.AlertDialogOptions{Title: "Remove file?"}); closed != nil {
		t.Fatal("closed alert dialog without a trigger rendered a surface")
	}
}

func TestOverlaySheetSides(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name  string
		side  chadcn.SheetSide
		class string
	}{
		{"default", 0, "of-sheet--right"},
		{"left", chadcn.SheetLeft, "of-sheet--left"},
		{"top", chadcn.SheetTop, "of-sheet--top"},
		{"bottom", chadcn.SheetBottom, "of-sheet--bottom"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			root := chadcn.Sheet(chadcn.SheetOptions{
				ID:          "filters",
				Open:        true,
				Side:        tt.side,
				Title:       "Filters",
				Description: "Narrow the results.",
				Trigger:     chadcn.Button("Open filters", chadcn.ButtonOptions{}),
				Footer:      []*ui.Element{chadcn.Button("Apply", chadcn.ButtonOptions{Variant: chadcn.ButtonFilled})},
				OnOpenChange: func(context.Context, bool) error {
					return nil
				},
			}, ui.Text("Filter body"))
			markup := overlayMarkup(t, root)
			overlayWant(t, markup,
				`class="of-sheet-panel `+tt.class+`"`,
				`class="of-sheet-backdrop"`,
				`class="of-sheet-header"`,
				`class="of-sheet-title"`,
				`class="of-sheet-description"`,
				`class="of-sheet-body"`,
				`class="of-sheet-footer"`,
				`class="of-sheet-footer-slot"`,
				`class="of-sheet-close"`,
				`id="filters"`,
				`id="filters-panel"`,
				`aria-labelledby="filters-panel-title"`,
				`aria-describedby="filters-panel-description"`,
				`aria-expanded="true"`,
				"Filters",
				"Narrow the results.",
				"Filter body",
				"Apply",
			)
		})
	}
}

func TestOverlaySheetClosed(t *testing.T) {
	t.Parallel()
	root := chadcn.Sheet(chadcn.SheetOptions{
		ID:      "filters",
		Title:   "Filters",
		Trigger: chadcn.Button("Open filters", chadcn.ButtonOptions{}),
	}, ui.Text("Filter body"))
	markup := overlayMarkup(t, root)
	overlayWant(t, markup, "Open filters", `aria-expanded="false"`)
	overlayDeny(t, markup, "of-sheet-backdrop", "Filter body")
}

func TestOverlayDrawer(t *testing.T) {
	t.Parallel()
	root := chadcn.Drawer(chadcn.DrawerOptions{
		ID:          "delivery",
		Open:        true,
		Title:       "Pick a delivery time",
		Description: "We prepare your order as soon as possible.",
		Trigger:     chadcn.Button("Open drawer", chadcn.ButtonOptions{}),
		Footer: []*ui.Element{
			chadcn.Button("Confirm delivery time", chadcn.ButtonOptions{Variant: chadcn.ButtonFilled}),
			chadcn.Button("Cancel", chadcn.ButtonOptions{Variant: chadcn.ButtonOutlined}),
		},
		OnOpenChange: func(context.Context, bool) error { return nil },
	}, ui.Text("Delivery slots"))
	markup := overlayMarkup(t, root)
	overlayWant(t, markup,
		`class="of-drawer"`,
		`class="of-drawer-backdrop"`,
		`class="of-drawer-panel"`,
		`class="of-drawer-grab"`,
		`class="of-drawer-handle"`,
		`class="of-drawer-header"`,
		`class="of-drawer-title"`,
		`class="of-drawer-description"`,
		`class="of-drawer-body"`,
		`class="of-drawer-footer"`,
		`class="of-drawer-footer-slot"`,
		`id="delivery"`,
		`id="delivery-panel"`,
		`role="dialog"`,
		`aria-modal="true"`,
		`aria-labelledby="delivery-panel-title"`,
		`aria-describedby="delivery-panel-description"`,
		"Pick a delivery time",
		"We prepare your order as soon as possible.",
		"Delivery slots",
		"Confirm delivery time",
		"Cancel",
	)

	closed := chadcn.Drawer(chadcn.DrawerOptions{
		ID:      "delivery",
		Title:   "Pick a delivery time",
		Trigger: chadcn.Button("Open drawer", chadcn.ButtonOptions{}),
	})
	closedMarkup := overlayMarkup(t, closed)
	overlayWant(t, closedMarkup, "Open drawer", `aria-expanded="false"`)
	overlayDeny(t, closedMarkup, "of-drawer-backdrop", "Delivery slots")
}
