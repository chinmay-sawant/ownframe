package chadcn_test

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// TestModalsFixDialogKeepsBodyAboveFooter checks the structured dialog keeps
// the body, then the footer, and both close paths on their own controls.
func TestModalsFixDialogKeepsBodyAboveFooter(t *testing.T) {
	t.Parallel()
	root := chadcn.Dialog(chadcn.DialogOptions{
		ID:          "edit-profile",
		Open:        true,
		Title:       "Edit profile",
		Description: "Update the account details.",
		CloseAction: "profile:close",
		Trigger:     chadcn.Button("Edit profile", chadcn.ButtonOptions{}),
		Footer:      []*ui.Element{chadcn.Button("Save changes", chadcn.ButtonOptions{Variant: chadcn.ButtonFilled})},
	}, ui.Text("Profile body"))
	markup := modalsFixHTML(t, root)
	modalsFixWant(t, markup,
		`class="of-dialog"`,
		`class="of-dialog-backdrop"`,
		`class="of-dialog-panel"`,
		`class="of-dialog-header"`,
		`class="of-dialog-body"`,
		`class="of-dialog-footer"`,
		`class="of-dialog-close"`,
		`id="edit-profile"`,
		`id="edit-profile-panel"`,
		`role="dialog"`,
		`aria-modal="true"`,
		`aria-labelledby="edit-profile-panel-title"`,
		`aria-describedby="edit-profile-panel-description"`,
		"Profile body",
		"Save changes",
	)
	if got := strings.Count(markup, `data-action="profile:close"`); got != 2 {
		t.Fatalf("close action count=%d, want the backdrop and the close control: %s", got, markup)
	}
	body := strings.Index(markup, `class="of-dialog-body"`)
	footer := strings.Index(markup, `<footer class="of-dialog-footer"`)
	if body < 0 || footer < 0 || footer < body {
		t.Fatalf("dialog body does not precede the footer:\n%s", markup)
	}
	// The modal band belongs to the stylesheet, so --of-modal-z can move it.
	if strings.Contains(markup, "z-index") {
		t.Fatalf("dialog carries an inline z-index that would beat the modal layer:\n%s", markup)
	}
}

// TestModalsFixDrawerKeepsBodyAboveStackedFooter checks the drawer structure and
// the stacked footer slots that hold the action buttons.
func TestModalsFixDrawerKeepsBodyAboveStackedFooter(t *testing.T) {
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
	markup := modalsFixHTML(t, root)
	modalsFixWant(t, markup,
		`class="of-drawer"`,
		`class="of-drawer-backdrop"`,
		`class="of-drawer-panel"`,
		`class="of-drawer-grab"`,
		`class="of-drawer-header"`,
		`class="of-drawer-body"`,
		`class="of-drawer-footer"`,
		`id="delivery"`,
		`id="delivery-panel"`,
		`id="delivery-backdrop"`,
		`role="dialog"`,
		`aria-modal="true"`,
		"Delivery slots",
		"Confirm delivery time",
		"Cancel",
	)
	if got := strings.Count(markup, `class="of-drawer-footer-slot"`); got != 2 {
		t.Fatalf("footer slot count=%d, want one per action: %s", got, markup)
	}
	body := strings.Index(markup, `class="of-drawer-body"`)
	footer := strings.Index(markup, `<footer class="of-drawer-footer"`)
	if body < 0 || footer < 0 || footer < body {
		t.Fatalf("drawer body does not precede the footer:\n%s", markup)
	}
	if strings.Contains(markup, "z-index") {
		t.Fatalf("drawer carries an inline z-index that would beat the modal layer:\n%s", markup)
	}
}

// TestModalsFixBackdropUsesModalBand checks the shared layer sits above the app
// topbar and sidebar and under the toast and command surfaces.
func TestModalsFixBackdropUsesModalBand(t *testing.T) {
	t.Parallel()
	// The dashboard shell stacks the sidebar at 30, the topbar at 40, the toast
	// at 60, and the command palette at 90. Both modal backdrops read one token.
	const band = "z-index: var(--of-modal-z, 55);"
	dialog := modalsFixRule(t, chadcn.Stylesheet, ".of-dialog-backdrop")
	drawer := modalsFixRule(t, chadcn.Stylesheet, ".of-drawer-backdrop")
	for name, rule := range map[string]string{"dialog": dialog, "drawer": drawer} {
		if !strings.Contains(rule, band) {
			t.Fatalf("%s backdrop is outside the modal band %q:\n%s", name, band, rule)
		}
	}
	for _, want := range []string{"position: fixed;", "display: flex;", "align-items: center;", "justify-content: center;"} {
		if !strings.Contains(dialog, want) {
			t.Fatalf("dialog backdrop lost its centering %q:\n%s", want, dialog)
		}
	}
}

// TestModalsFixDialogBodyScrollsUnderFixedFooter checks the body shrinks and
// scrolls while the footer keeps its height.
func TestModalsFixDialogBodyScrollsUnderFixedFooter(t *testing.T) {
	t.Parallel()
	body := modalsFixRule(t, chadcn.Stylesheet, ".of-dialog-body")
	for _, want := range []string{"min-height: 0;", "overflow-y: auto;"} {
		if !strings.Contains(body, want) {
			t.Fatalf("dialog body is not a shrinking scroll area, missing %q:\n%s", want, body)
		}
	}
	footer := modalsFixRule(t, chadcn.Stylesheet, ".of-dialog-footer")
	if !strings.Contains(footer, "flex-shrink: 0;") {
		t.Fatalf("dialog footer can shrink with long content:\n%s", footer)
	}
}

// TestModalsFixDrawerAnchorsToBottom checks the panel stays on the bottom edge
// with the same shrinking body and fixed footer as the dialog.
func TestModalsFixDrawerAnchorsToBottom(t *testing.T) {
	t.Parallel()
	panel := modalsFixRule(t, chadcn.Stylesheet, ".of-drawer-panel")
	for _, want := range []string{"position: absolute;", "right: 0;", "bottom: 0;", "left: 0;"} {
		if !strings.Contains(panel, want) {
			t.Fatalf("drawer panel is not anchored to the bottom edge, missing %q:\n%s", want, panel)
		}
	}
	body := modalsFixRule(t, chadcn.Stylesheet, ".of-drawer-body")
	for _, want := range []string{"min-height: 0;", "overflow-y: auto;"} {
		if !strings.Contains(body, want) {
			t.Fatalf("drawer body is not a shrinking scroll area, missing %q:\n%s", want, body)
		}
	}
	footer := modalsFixRule(t, chadcn.Stylesheet, ".of-drawer-footer")
	if !strings.Contains(footer, "flex-shrink: 0;") {
		t.Fatalf("drawer footer can shrink with long content:\n%s", footer)
	}
}

// modalsFixHTML renders a modal tree and reports a builder error as a failure.
func modalsFixHTML(t *testing.T, root *ui.Element) string {
	t.Helper()
	if root == nil {
		t.Fatal("modal rendered nil")
	}
	markup, err := ui.Render(root)
	if err != nil {
		t.Fatal(err)
	}
	return markup
}

// modalsFixWant checks that every expected fragment appears in the markup.
func modalsFixWant(t *testing.T, markup string, fragments ...string) {
	t.Helper()
	for _, fragment := range fragments {
		if !strings.Contains(markup, fragment) {
			t.Fatalf("markup missing %q:\n%s", fragment, markup)
		}
	}
}

// modalsFixRule returns one rule body from the stylesheet, without its braces.
func modalsFixRule(t *testing.T, sheet, selector string) string {
	t.Helper()
	start := strings.Index(sheet, selector+" {")
	if start < 0 {
		t.Fatalf("stylesheet is missing %s", selector)
	}
	end := strings.Index(sheet[start:], "}")
	if end < 0 {
		t.Fatalf("stylesheet rule %s is unterminated", selector)
	}
	return sheet[start : start+end]
}
