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

func TestOverlayDismissDialogInsideClickDoesNotClose(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	closes := 0
	page := overlayDismissPage(t, chadcn.Dialog(chadcn.DialogOptions{
		Open:  true,
		Title: "Edit profile",
		OnOpenChange: func(_ context.Context, open bool) error {
			if !open {
				closes++
			}
			return nil
		},
	}, ui.Text("Notes")))
	overlayDismissClickID(t, ctx, page, "of-dialog-panel-title")
	overlayDismissClickText(t, ctx, page, "Notes")
	if closes != 0 {
		t.Fatalf("inside click dismissed the dialog %d times", closes)
	}
	overlayDismissClickID(t, ctx, page, "of-dialog-panel-close")
	if closes != 1 {
		t.Fatalf("close control dismissals=%d, want 1", closes)
	}
}

func TestOverlayDismissDialogBackdropClickCloses(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	closes := 0
	page := overlayDismissPage(t, chadcn.Dialog(chadcn.DialogOptions{
		Open:        true,
		Title:       "Edit profile",
		CloseAction: "",
		OnOpenChange: func(_ context.Context, open bool) error {
			if !open {
				closes++
			}
			return nil
		},
	}, ui.Text("Notes")))
	overlayDismissClickOutside(t, ctx, page, "of-dialog-backdrop", "of-dialog-panel")
	if closes != 1 {
		t.Fatalf("backdrop dismissals=%d, want 1", closes)
	}
}

func TestOverlayDismissClosedDialogIsNotMounted(t *testing.T) {
	t.Parallel()
	root := chadcn.Dialog(chadcn.DialogOptions{
		Title:        "Edit profile",
		Trigger:      chadcn.Button("Open", chadcn.ButtonOptions{}),
		OnOpenChange: func(context.Context, bool) error { return nil },
	})
	markup := overlayMarkup(t, root)
	overlayDeny(t, markup, "of-dialog-backdrop", "of-dialog-panel", `role="dialog"`)
	if closed := chadcn.Dialog(chadcn.DialogOptions{Open: false, Title: "Edit profile"}); closed != nil {
		t.Fatal("closed dialog without a trigger mounted a surface")
	}
}

func TestOverlayDismissDialogWithoutIDKeepsDistinctParts(t *testing.T) {
	t.Parallel()
	markup := overlayMarkup(t, chadcn.Dialog(chadcn.DialogOptions{
		Open:        true,
		Title:       "Edit profile",
		CloseAction: "dialog:close",
	}, ui.Text("Notes")))
	overlayWant(t, markup,
		`id="of-dialog-panel"`,
		`id="of-dialog-backdrop"`,
		`id="of-dialog-panel-title"`,
		`id="of-dialog-panel-close"`,
		`data-action="dialog:close"`,
	)
	if strings.Count(markup, `id="of-dialog-panel"`) != 1 {
		t.Fatalf("panel id is not distinct:\n%s", markup)
	}
}

func TestOverlayDismissAlertDialogOmitsDeadButtons(t *testing.T) {
	t.Parallel()
	markup := overlayMarkup(t, chadcn.AlertDialog(chadcn.AlertDialogOptions{
		Open:  true,
		Title: "Remove file?",
	}))
	overlayDeny(t, markup, "of-alert-dialog-cancel", "of-alert-dialog-confirm", ">Cancel<", ">Continue<")

	wired := overlayMarkup(t, chadcn.AlertDialog(chadcn.AlertDialogOptions{
		Open:         true,
		OnOpenChange: func(context.Context, bool) error { return nil },
	}))
	overlayWant(t, wired,
		`aria-label="Alert dialog"`,
		`class="of-alert-dialog-cancel"`,
		`class="of-alert-dialog-confirm"`,
		">Cancel<",
		">Continue<",
		"data-action=",
	)
	overlayDeny(t, wired, "aria-labelledby")
}

func TestOverlayDismissAlertDialogNilActionCloses(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	closes := 0
	page := overlayDismissPage(t, chadcn.AlertDialog(chadcn.AlertDialogOptions{
		ID:    "remove",
		Open:  true,
		Title: "Remove file?",
		OnOpenChange: func(_ context.Context, open bool) error {
			if !open {
				closes++
			}
			return nil
		},
	}))
	overlayDismissClickText(t, ctx, page, "Cancel")
	overlayDismissClickText(t, ctx, page, "Continue")
	if closes != 2 {
		t.Fatalf("nil actions closed the dialog %d times, want 2", closes)
	}
}

func TestOverlayDismissDrawerHandleClosesAndBodyDoesNot(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	closes := 0
	page := overlayDismissPage(t, chadcn.Drawer(chadcn.DrawerOptions{
		Open:  true,
		Title: "Delivery",
		OnOpenChange: func(_ context.Context, open bool) error {
			if !open {
				closes++
			}
			return nil
		},
	}, ui.Text("Slots")))
	overlayDismissClickText(t, ctx, page, "Slots")
	if closes != 0 {
		t.Fatalf("body click dismissed the drawer %d times", closes)
	}
	overlayDismissClickID(t, ctx, page, "of-drawer-panel-handle")
	if closes != 1 {
		t.Fatalf("handle dismissals=%d, want 1", closes)
	}
	markup := overlayMarkup(t, chadcn.Drawer(chadcn.DrawerOptions{Open: true, Title: "Delivery"}, ui.Text("Slots")))
	overlayWant(t, markup, `id="of-drawer-panel"`, `id="of-drawer-backdrop"`)
	overlayDeny(t, markup, `id="of-drawer-panel-handle"`)
}

func TestOverlayDismissSheetBodyDoesNotClose(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	closes := 0
	page := overlayDismissPage(t, chadcn.Sheet(chadcn.SheetOptions{
		Open:  true,
		Title: "Filters",
		OnOpenChange: func(_ context.Context, open bool) error {
			if !open {
				closes++
			}
			return nil
		},
	}, ui.Text("Query")))
	overlayDismissClickText(t, ctx, page, "Query")
	if closes != 0 {
		t.Fatalf("body click dismissed the sheet %d times", closes)
	}
	overlayDismissClickID(t, ctx, page, "of-sheet-panel-close")
	if closes != 1 {
		t.Fatalf("close control dismissals=%d, want 1", closes)
	}
	markup := overlayMarkup(t, chadcn.Sheet(chadcn.SheetOptions{Open: true, Title: "Filters"}, ui.Text("Query")))
	overlayWant(t, markup, `id="of-sheet-panel"`, `id="of-sheet-backdrop"`)
	overlayDeny(t, markup, "of-sheet-close")
}

func TestOverlayDismissToastOmitsDeadClose(t *testing.T) {
	t.Parallel()
	markup := overlayMarkup(t, chadcn.Toast(chadcn.ToastOptions{Open: true, Title: "Saved"}))
	overlayDeny(t, markup, "of-toast-close", "data-action", `aria-label="Close"`)
	withClose := overlayMarkup(t, chadcn.Toast(chadcn.ToastOptions{
		Open:        true,
		Title:       "Saved",
		CloseAction: "toast:close",
	}))
	overlayWant(t, withClose, `class="of-toast-close"`, `data-action="toast:close"`)
}

func TestOverlayDismissAlertOmitsEmptyCopy(t *testing.T) {
	t.Parallel()
	if alert := chadcn.Alert(chadcn.AlertOptions{}); alert != nil {
		t.Fatal("empty alert mounted a live region")
	}
	if alert := chadcn.Alert(chadcn.AlertOptions{Icon: ui.Text("!"), Action: ui.Text("Retry")}); alert != nil {
		t.Fatal("alert with no copy mounted a live region")
	}
	markup := overlayMarkup(t, chadcn.Alert(chadcn.AlertOptions{Description: "Saved."}))
	overlayWant(t, markup, `role="alert"`, "Saved.")
	overlayDeny(t, markup, "of-alert-title")
}

func TestOverlayDismissCSSHoverShowsFromTrigger(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	for _, selector := range []string{
		".of-tooltip--css-hover:has(:hover) > .of-tooltip-anchor > .of-tooltip-content",
		".of-hover-card--css-hover:has(:hover) > .of-hover-card-anchor > .of-hover-card-content",
	} {
		if !strings.Contains(chadcn.Stylesheet, selector+" {") {
			t.Fatalf("stylesheet is missing %s", selector)
		}
	}

	tooltip := overlayDismissPage(t, chadcn.Tooltip(chadcn.TooltipOptions{
		ID:       "live-tip",
		CSSHover: true,
		Content:  "Updated just now",
	}, ui.Button("Live")))
	overlayDismissHidden(t, tooltip, "live-tip-content")
	overlayDismissHoverID(t, ctx, tooltip, "live-tip-trigger")
	overlayDismissShown(t, tooltip, "live-tip-content")

	inner := overlayDismissPage(t, chadcn.Tooltip(chadcn.TooltipOptions{
		ID:       "library-tip",
		CSSHover: true,
		Content:  "Add to library",
	}, ui.Button("Hover").ID("library")))
	overlayDismissHoverID(t, ctx, inner, "library")
	overlayDismissShown(t, inner, "library-tip-content")

	card := overlayDismissPage(t, chadcn.HoverCard(chadcn.HoverCardOptions{
		ID:       "invite-card",
		CSSHover: true,
	}, ui.Button("Invite"), ui.Text("Invite a teammate")))
	overlayDismissHidden(t, card, "invite-card-content")
	overlayDismissHoverID(t, ctx, card, "invite-card-trigger")
	overlayDismissShown(t, card, "invite-card-content")

	top := overlayDismissPage(t, chadcn.HoverCard(chadcn.HoverCardOptions{
		ID:       "top-card",
		CSSHover: true,
		Side:     chadcn.PopupTop,
	}, ui.Button("Above").ID("above-trigger"), ui.Text("Above the trigger")))
	overlayDismissHoverID(t, ctx, top, "above-trigger")
	overlayDismissShown(t, top, "top-card-content")
}

func overlayDismissPage(t *testing.T, root *ui.Element) *ownframe.Page {
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

func overlayDismissClickID(t *testing.T, ctx context.Context, page *ownframe.Page, id string) {
	t.Helper()
	if err := pagetest.Click(ctx, page, id); err != nil {
		t.Fatal(err)
	}
}

func overlayDismissClickText(t *testing.T, ctx context.Context, page *ownframe.Page, text string) {
	t.Helper()
	for _, box := range page.Boxes() {
		if box.Text != text || box.W <= 0 || box.H <= 0 {
			continue
		}
		if err := page.Click(ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
			t.Fatal(err)
		}
		return
	}
	t.Fatalf("missing text %q", text)
}

func overlayDismissClickOutside(t *testing.T, ctx context.Context, page *ownframe.Page, backdropID, panelID string) {
	t.Helper()
	backdrop, ok := pagetest.BoxByID(page, backdropID)
	if !ok {
		t.Fatalf("missing backdrop %s", backdropID)
	}
	panel, ok := pagetest.BoxByID(page, panelID)
	if !ok {
		t.Fatalf("missing panel %s", panelID)
	}
	const inset = 4
	points := [][2]float64{
		{backdrop.X + inset, backdrop.Y + inset},
		{backdrop.X + backdrop.W - inset, backdrop.Y + inset},
		{backdrop.X + inset, backdrop.Y + backdrop.H - inset},
		{backdrop.X + backdrop.W - inset, backdrop.Y + backdrop.H - inset},
	}
	for _, point := range points {
		if point[0] >= panel.X && point[0] <= panel.X+panel.W && point[1] >= panel.Y && point[1] <= panel.Y+panel.H {
			continue
		}
		if err := page.Click(ctx, point[0], point[1]); err != nil {
			t.Fatal(err)
		}
		return
	}
	t.Fatal("backdrop has no point outside the panel")
}

func overlayDismissHoverID(t *testing.T, ctx context.Context, page *ownframe.Page, id string) {
	t.Helper()
	box, ok := pagetest.BoxByID(page, id)
	if !ok {
		t.Fatalf("missing hover target %s", id)
	}
	if err := page.Edit().Hover(ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
		t.Fatal(err)
	}
}

func overlayDismissHidden(t *testing.T, page *ownframe.Page, id string) {
	t.Helper()
	box, ok := pagetest.BoxByID(page, id)
	if ok && box.W > 0 && box.H > 0 {
		t.Fatalf("%s is visible before hover: %+v", id, box)
	}
}

func overlayDismissShown(t *testing.T, page *ownframe.Page, id string) {
	t.Helper()
	box, ok := pagetest.BoxByID(page, id)
	if !ok || box.W <= 0 || box.H <= 0 {
		t.Fatalf("%s did not open on hover: ok=%v box=%+v", id, ok, box)
	}
}
