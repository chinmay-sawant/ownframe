package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// loglensLayoutBox finds one box in the current frame by id, text, or tag.
func loglensLayoutBox(t *testing.T, p *ownframe.Page, match func(ownframe.Box) bool) ownframe.Box {
	t.Helper()
	for _, box := range p.Boxes() {
		if match(box) {
			return box
		}
	}
	t.Fatal("expected box is missing from the frame")
	return ownframe.Box{}
}

// TestSettingsTabsKeepHiddenValues moves through the settings sections and
// saves, proving controls on an inactive section still reach Page.Form.
func TestSettingsTabsKeepHiddenValues(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	a.settingsOpen = true
	if err := a.refresh(ctx); err != nil {
		t.Fatal(err)
	}
	a.setFormValue("settings-email", "oncall@example.com")
	a.setFormValue("settings-notes", "handover")
	for index := range settingsOTPLength {
		a.setFormValue("settings-otp-"+string(rune('0'+index)), string(rune('1'+index)))
	}
	for _, tab := range []string{"noise", "verification"} {
		a.settingsTab = tab
		if err := a.refresh(ctx); err != nil {
			t.Fatal(err)
		}
	}
	save := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool {
		return box.Tag == "button" && box.Text == "Save"
	})
	if err := a.page.Click(ctx, save.X+save.W/2, save.Y+save.H/2); err != nil {
		t.Fatal(err)
	}
	if a.settingsOpen {
		t.Fatal("Save left the settings sheet open")
	}
	if a.email != "oncall@example.com" || a.notes != "handover" || a.otp != "123456" {
		t.Fatalf("hidden values lost: email=%q notes=%q otp=%q", a.email, a.notes, a.otp)
	}
}

// TestDialogFitsMinimumWindow checks the create-alert controls stay above the
// footer at the 1100x720 minimum instead of scrolling under it.
func TestDialogFitsMinimumWindow(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	a.page.SetSize(minWidth, minHeight)
	a.dialogOpen = true
	if err := a.refresh(ctx); err != nil {
		t.Fatal(err)
	}
	footer := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool {
		return box.Tag == "footer" && box.Text == "CancelSave"
	})
	for _, id := range []string{"alert-enabled", "alert-page"} {
		box := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool { return box.ID == id })
		if box.Y+box.H > footer.Y {
			t.Fatalf("%s ends at %.0f, under the footer at %.0f", id, box.Y+box.H, footer.Y)
		}
	}
}

// TestDrawerTabsFitMinimumWindow checks both drawer tabs end above the footer
// at the 1100x720 minimum.
func TestDrawerTabsFitMinimumWindow(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	a.page.SetSize(minWidth, minHeight)
	a.drawerOpen = true
	a.drawerIndex = 0
	for _, tab := range []string{"payload", "trace"} {
		a.drawerTab = tab
		if err := a.refresh(ctx); err != nil {
			t.Fatal(err)
		}
		footer := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool {
			return box.Tag == "footer" && box.Text == "Close"
		})
		note := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool {
			return box.Text == "Retained for 30 days. Tokens and passwords are redacted in exports."
		})
		if note.Y+note.H > footer.Y {
			t.Fatalf("%s content ends at %.0f, under the footer at %.0f", tab, note.Y+note.H, footer.Y)
		}
	}
}

func loglensSameBox(a, b ownframe.Box) bool {
	return a.X == b.X && a.Y == b.Y && a.W == b.W && a.H == b.H
}

// TestTopbarChromeStaysClear keeps the brand, nav, and menubar from
// overlapping and proves an open nav panel does not move them.
func TestTopbarChromeStaysClear(t *testing.T) {
	for _, width := range []int{minWidth, windowWidth} {
		a, err := newApp()
		if err != nil {
			t.Fatal(err)
		}
		ctx := context.Background()
		a.page.SetSize(width, minHeight)
		if err := a.refresh(ctx); err != nil {
			t.Fatal(err)
		}
		brand := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool { return box.Text == "Loglens" })
		nav := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool { return box.ID == "top-nav" })
		menubar := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool { return box.ID == "top-menubar" })
		if brand.X+brand.W > nav.X {
			t.Fatalf("%d: brand ends at %.0f under the nav at %.0f", width, brand.X+brand.W, nav.X)
		}
		if nav.X+nav.W > menubar.X {
			t.Fatalf("%d: nav ends at %.0f under the menubar at %.0f", width, nav.X+nav.W, menubar.X)
		}
		closeBox := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool {
			return box.Tag == "button" && box.Text == "×"
		})
		if closeBox.X+closeBox.W > float64(width) {
			t.Fatalf("%d: window controls end at %.0f", width, closeBox.X+closeBox.W)
		}
		a.navOpen = "product-platform"
		if err := a.refresh(ctx); err != nil {
			t.Fatal(err)
		}
		movedNav := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool { return box.ID == "top-nav" })
		movedMenubar := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool { return box.ID == "top-menubar" })
		if !loglensSameBox(nav, movedNav) || !loglensSameBox(menubar, movedMenubar) {
			t.Fatalf("%d: open panel moved the chrome: nav %+v -> %+v menubar %+v -> %+v", width, nav, movedNav, menubar, movedMenubar)
		}
		panel := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool { return box.ID == "product-platform-content" })
		if panel.Y < nav.Y+nav.H {
			t.Fatalf("%d: nav panel starts at %.0f inside the trigger row ending at %.0f", width, panel.Y, nav.Y+nav.H)
		}
	}
}

// TestSettingsTabsFitMinimumWindow checks each settings section keeps its
// fields above the Save footer at the 1100x720 minimum.
func TestSettingsTabsFitMinimumWindow(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	a.page.SetSize(minWidth, minHeight)
	a.settingsOpen = true
	checks := map[string][]string{
		"routing":      {"settings-email", "settings-channel", "settings-webhook"},
		"noise":        {"settings-threshold"},
		"verification": {"settings-otp-5", "settings-notes"},
	}
	for _, tab := range []string{"routing", "noise", "verification"} {
		a.settingsTab = tab
		if err := a.refresh(ctx); err != nil {
			t.Fatal(err)
		}
		footer := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool {
			return box.Tag == "button" && box.Text == "Save"
		})
		for _, id := range checks[tab] {
			box := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool { return box.ID == id })
			if box.Y+box.H > footer.Y {
				t.Fatalf("%s: %s ends at %.0f under Save at %.0f", tab, id, box.Y+box.H, footer.Y)
			}
		}
	}
}
