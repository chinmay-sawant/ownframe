package chadcn_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// TestProbePatchDialogs proves open->open content edits on the four overlay
// components take the ui.View.Update patch path (parses+0) instead of falling
// back to a full Load. Closed->open transitions mount the backdrop+panel, so
// they stay FULL by design and are not covered here.
func TestProbePatchDialogs(t *testing.T) {
	t.Parallel()

	t.Run("dialog body edit", func(t *testing.T) {
		t.Parallel()
		build := func(body string) *ui.Element {
			return ui.Div().ID("probe-root").Children(chadcn.Dialog(chadcn.DialogOptions{
				ID: "pp-dialog", Open: true,
				Title: "Edit profile", Description: "Make changes here.",
				CloseAction: "pp-dialog:close",
				Trigger:     chadcn.Button("Edit profile", chadcn.ButtonOptions{}).ID("pp-dialog-trigger"),
				Footer:      []*ui.Element{chadcn.Button("Save", chadcn.ButtonOptions{Variant: chadcn.ButtonFilled}).ID("pp-dialog-save")},
			}, ui.P().ID("pp-dialog-body").Text(body)))
		}
		v, ctx := mountProbeView(t, build("body v1"))
		expectProbePatch(t, v, ctx, build("body v2"))
	})

	t.Run("dialog description edit", func(t *testing.T) {
		t.Parallel()
		build := func(description string) *ui.Element {
			return ui.Div().ID("probe-root").Children(chadcn.Dialog(chadcn.DialogOptions{
				ID: "pp-dialog", Open: true,
				Title: "Edit profile", Description: description,
				CloseAction: "pp-dialog:close",
				Trigger:     chadcn.Button("Edit profile", chadcn.ButtonOptions{}).ID("pp-dialog-trigger"),
				Footer:      []*ui.Element{chadcn.Button("Save", chadcn.ButtonOptions{Variant: chadcn.ButtonFilled}).ID("pp-dialog-save")},
			}, ui.P().ID("pp-dialog-body").Text("stable body")))
		}
		v, ctx := mountProbeView(t, build("description v1"))
		expectProbePatch(t, v, ctx, build("description v2"))
	})

	t.Run("sheet body edit", func(t *testing.T) {
		t.Parallel()
		onOpenChange := func(context.Context, bool) error { return nil }
		build := func(body string) *ui.Element {
			return ui.Div().ID("probe-root").Children(chadcn.Sheet(chadcn.SheetOptions{
				ID: "pp-sheet", Open: true,
				Title: "Filters", Description: "Narrow the results.",
				Trigger:      chadcn.Button("Open filters", chadcn.ButtonOptions{}).ID("pp-sheet-trigger"),
				Footer:       []*ui.Element{chadcn.Button("Apply", chadcn.ButtonOptions{Variant: chadcn.ButtonFilled}).ID("pp-sheet-apply")},
				OnOpenChange: onOpenChange,
			}, ui.P().ID("pp-sheet-body").Text(body)))
		}
		v, ctx := mountProbeView(t, build("body v1"))
		expectProbePatch(t, v, ctx, build("body v2"))
	})

	t.Run("drawer body edit", func(t *testing.T) {
		t.Parallel()
		onOpenChange := func(context.Context, bool) error { return nil }
		build := func(body string) *ui.Element {
			return ui.Div().ID("probe-root").Children(chadcn.Drawer(chadcn.DrawerOptions{
				ID: "pp-drawer", Open: true,
				Title: "Delivery", Description: "Pick a time.",
				Trigger:      chadcn.Button("Open drawer", chadcn.ButtonOptions{}).ID("pp-drawer-trigger"),
				Footer:       []*ui.Element{chadcn.Button("Confirm", chadcn.ButtonOptions{Variant: chadcn.ButtonFilled}).ID("pp-drawer-confirm")},
				OnOpenChange: onOpenChange,
			}, ui.P().ID("pp-drawer-body").Text(body)))
		}
		v, ctx := mountProbeView(t, build("body v1"))
		expectProbePatch(t, v, ctx, build("body v2"))
	})

	t.Run("alert dialog description edit", func(t *testing.T) {
		t.Parallel()
		onConfirm := func(context.Context) error { return nil }
		onCancel := func(context.Context) error { return nil }
		build := func(description string) *ui.Element {
			return ui.Div().ID("probe-root").Children(chadcn.AlertDialog(chadcn.AlertDialogOptions{
				ID: "pp-alert", Open: true,
				Title: "Remove file?", Description: description,
				Trigger:   chadcn.Button("Show dialog", chadcn.ButtonOptions{}).ID("pp-alert-trigger"),
				OnConfirm: onConfirm,
				OnCancel:  onCancel,
			}))
		}
		v, ctx := mountProbeView(t, build("description v1"))
		expectProbePatch(t, v, ctx, build("description v2"))
	})

	t.Run("dialog identical rebuild noop", func(t *testing.T) {
		t.Parallel()
		build := func() *ui.Element {
			return ui.Div().ID("probe-root").Children(chadcn.Dialog(chadcn.DialogOptions{
				ID: "pp-dialog", Open: true,
				Title: "Edit profile", Description: "Make changes here.",
				CloseAction: "pp-dialog:close",
				Trigger:     chadcn.Button("Edit profile", chadcn.ButtonOptions{}).ID("pp-dialog-trigger"),
				Footer:      []*ui.Element{chadcn.Button("Save", chadcn.ButtonOptions{Variant: chadcn.ButtonFilled}).ID("pp-dialog-save")},
			}, ui.P().ID("pp-dialog-body").Text("stable body")))
		}
		v, ctx := mountProbeView(t, build())
		expectProbeNoop(t, v, ctx, build())
	})
}
