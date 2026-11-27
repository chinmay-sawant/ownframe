package ui

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// focusRing walks Tab n times from the current focus and returns the ids.
func (r *rig) focusRing(n int, back bool) []string {
	r.t.Helper()
	var out []string
	for i := 0; i < n; i++ {
		if back {
			r.must(r.a.page.Form().FocusPrev(r.ctx))
		} else {
			r.must(r.a.page.Form().FocusNext(r.ctx))
		}
		out = append(out, r.a.page.Form().FocusID())
	}
	return out
}

func (r *rig) boxIndex(id string) int {
	for i, b := range r.a.Page().Boxes() {
		if b.ID == id {
			return i
		}
	}
	return -1
}

// inside reports whether box id lies within the box named host.
func (r *rig) inside(id, host string) bool {
	h, b := r.box(host), r.box(id)
	return b.X >= h.X-1 && b.Y >= h.Y-1 && b.X+b.W <= h.X+h.W+1 && b.Y+b.H <= h.Y+h.H+1
}

func TestTabStaysInsideEveryDialog(t *testing.T) {
	open := map[string]func(r *rig){
		"conn":     func(r *rig) { r.tap("rail-connections") },
		"settings": func(r *rig) { r.tap("rail-settings") },
		"history":  func(r *rig) { r.tap("rail-history") },
		"confirm": func(r *rig) {
			r.must(r.a.askConfirm(r.ctx, &confirmState{Title: "Sure?", Body: "b", Label: "Do it", OK: func(context.Context) error { return nil }}))
		},
	}
	for name, fn := range open {
		r := newRig(t, 1440, 960, "")
		fn(r)
		if r.a.modal != name {
			t.Fatalf("%s: modal = %q", name, r.a.modal)
		}
		first := r.a.page.Form().FocusID()
		if first == "" || !r.inside(first, "modal") {
			t.Fatalf("%s: opening focus %q is not inside the dialog", name, first)
		}
		fwd := r.focusRing(40, false)
		for _, id := range fwd {
			if id == "" || !r.inside(id, "modal") {
				t.Fatalf("%s: Tab escaped the dialog to %q", name, id)
			}
		}
		if !slices.Contains(fwd, first) {
			t.Errorf("%s: Tab never wrapped back to %s: %v", name, first, fwd)
		}
		for _, id := range r.focusRing(40, true) {
			if id == "" || !r.inside(id, "modal") {
				t.Fatalf("%s: Shift+Tab escaped the dialog to %q", name, id)
			}
		}
	}
}

func TestTabOrderIsDocumentOrderInDialog(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.tap("rail-connections")
	r.must(r.a.page.Form().Focus(r.ctx, ""))
	ids := r.focusRing(6, false)
	last := -1
	for _, id := range ids {
		i := r.boxIndex(id)
		if i < last {
			t.Fatalf("Tab order %v is not document order", ids)
		}
		last = i
	}
	for _, want := range []string{"cd-engine-sqlite", "cd-name"} {
		if !slices.Contains(ids, want) {
			t.Errorf("%s missing from the first Tab stops: %v", want, ids)
		}
	}
	// Shift+Tab reverses it.
	back := r.focusRing(2, true)
	if back[0] != ids[4] || back[1] != ids[3] {
		t.Errorf("Shift+Tab = %v, want %v then %v", back, ids[4], ids[3])
	}
}

func TestPaletteTabKeepsFocusInSearch(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.chord("k", false)
	r.must(r.a.Page().Redraw(r.ctx))
	for _, id := range append(r.focusRing(4, false), r.focusRing(4, true)...) {
		if id != "pal-input" {
			t.Fatalf("Tab left the palette search: %q", id)
		}
	}
}

func TestToolbarTabOrderFollowsLayout(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.must(r.a.page.Form().Focus(r.ctx, ""))
	ids := r.focusRing(30, false)
	var bar []string
	for _, id := range ids {
		if strings.HasPrefix(id, "ed-") && r.has(id) && r.inside(id, "editor-toolbar") {
			bar = append(bar, id)
		}
	}
	if len(bar) < 3 {
		t.Fatalf("toolbar buttons not reachable by Tab: %v", ids)
	}
	for i := 1; i < len(bar); i++ {
		a, b := r.box(bar[i-1]), r.box(bar[i])
		if b.Y < a.Y-1 || (b.Y <= a.Y+1 && b.X < a.X) {
			t.Errorf("Tab goes %s then %s, against reading order", bar[i-1], bar[i])
		}
	}
}

func TestEscapeClosesTopmostOverlayAndReturnsFocus(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	// Dialog opened from a focused rail button: Escape gives focus back.
	r.must(r.a.page.Form().Focus(r.ctx, "rail-connections"))
	r.must(r.a.openConnManager(r.ctx))
	if r.a.modal != "conn" || r.a.page.Form().FocusID() == "rail-connections" {
		t.Fatal("dialog should take focus on open")
	}
	r.must(r.a.page.Edit().KeyDown(r.ctx, "escape"))
	if r.a.modal != "" || r.a.page.Form().FocusID() != "rail-connections" {
		t.Fatalf("after Escape modal=%q focus=%q, want the opener", r.a.modal, r.a.page.Form().FocusID())
	}
	// Tab is free again: it no longer stays in the dialog's controls.
	ring := r.focusRing(3, false)
	if slices.Contains(ring, "cd-name") {
		t.Fatalf("focus scope leaked after close: %v", ring)
	}
	// Palette: Escape closes it and returns to where Ctrl+K was pressed.
	r.must(r.a.page.Form().Focus(r.ctx, "rail-settings"))
	r.chord("k", false)
	r.must(r.a.Page().Redraw(r.ctx))
	r.must(r.a.page.Edit().KeyDown(r.ctx, "escape"))
	if r.a.modal != "" || r.a.page.Form().FocusID() != "rail-settings" {
		t.Fatalf("palette close: modal=%q focus=%q", r.a.modal, r.a.page.Form().FocusID())
	}
	// The topmost overlay goes first: a title-bar menu over the page closes
	// before anything beneath it.
	r.a.menuOpen = "conn"
	r.must(r.a.redraw(r.ctx))
	r.must(r.a.page.Edit().KeyDown(r.ctx, "escape"))
	if r.a.menuOpen != "" {
		t.Fatal("Escape should close the open menu")
	}
}

func TestEscapeFromTextFieldClosesDialogOnce(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.must(r.a.page.Form().Focus(r.ctx, "rail-settings"))
	r.tap("rail-connections")
	r.must(r.a.page.Form().Focus(r.ctx, "cd-name"))
	r.must(r.a.page.Edit().KeyDown(r.ctx, "escape"))
	if r.a.modal != "" {
		t.Fatal("Escape in a text field must close the dialog")
	}
}

func TestEnterSubmitsPrimaryAction(t *testing.T) {
	// Connection manager: Enter in a field connects.
	r := newRig(t, 1440, 960, "")
	other := r.a.cfg.FilePath
	r.tap("rail-connections")
	r.setField("cd-name", "Via Enter")
	r.setField("cd-path", other)
	r.must(r.a.page.Edit().Submit(r.ctx))
	if r.a.modal != "" || r.a.cfg.Name != "Via Enter" {
		t.Fatalf("Enter did not connect: modal=%q cfg=%q", r.a.modal, r.a.cfg.Name)
	}

	// Invalid draft: Enter reports errors and keeps the dialog.
	r.tap("rail-connections")
	r.setField("cd-name", "")
	r.must(r.a.page.Edit().Submit(r.ctx))
	if r.a.modal != "conn" || !r.has("cd-name-err") {
		t.Fatal("Enter on an invalid draft should show errors and stay open")
	}
	r.tap("conn-close")

	// Settings: Enter in the threshold field applies it and closes.
	r.tap("rail-settings")
	r.setField("set-slow", "321")
	r.must(r.a.page.Edit().Submit(r.ctx))
	if r.a.modal != "" || r.a.slow.Threshold().Milliseconds() != 321 {
		t.Fatalf("settings Enter: modal=%q slow=%v", r.a.modal, r.a.slow.Threshold())
	}

	// Confirm: Enter on the text field confirms; Enter on Cancel cancels.
	done := 0
	ask := func(typed string) {
		r.must(r.a.askConfirm(r.ctx, &confirmState{Title: "Drop?", Body: "b", Label: "Drop", Typed: typed,
			OK: func(context.Context) error { done++; return nil }}))
	}
	ask("orders")
	if r.a.page.Form().FocusID() != "confirm-typed" {
		t.Fatalf("typed confirm should focus its field, got %q", r.a.page.Form().FocusID())
	}
	r.must(r.a.page.Form().Focus(r.ctx, "confirm-typed"))
	r.a.page.Form().SetFormValue("confirm-typed", "wrong")
	r.must(r.a.page.Edit().Submit(r.ctx))
	if done != 0 || r.a.modal != "confirm" || !r.has("confirm-error") {
		t.Fatal("Enter with the wrong text must not confirm")
	}
	r.a.page.Form().SetFormValue("confirm-typed", "orders")
	r.must(r.a.page.Edit().Submit(r.ctx))
	if done != 1 || r.a.modal != "" {
		t.Fatalf("Enter with the right text: done=%d modal=%q", done, r.a.modal)
	}
	ask("")
	if r.a.page.Form().FocusID() != "confirm-cancel" {
		t.Fatalf("plain confirm should start on Cancel, got %q", r.a.page.Form().FocusID())
	}
	r.must(r.a.page.Edit().Submit(r.ctx))
	if done != 1 || r.a.modal != "" {
		t.Fatalf("Enter on Cancel must cancel: done=%d modal=%q", done, r.a.modal)
	}
}

// ownLayerOps serialises the paint ops that touch box id, so a focus change
// shows up as a different string.
func (r *rig) opsAround(id string) string {
	b := r.box(id)
	d := r.a.Page().Probe().Display()
	k := d.PixelPerPoint
	if k <= 0 {
		k = 1
	}
	var sb strings.Builder
	for _, op := range d.Ops {
		x, y, w, h := op.X*k, op.Y*k, op.W*k, op.H*k
		if x+w >= b.X-2 && x <= b.X+b.W+2 && y+h >= b.Y-2 && y <= b.Y+b.H+2 && w < b.W*3 && op.Kind != ownframe.DisplayOpText {
			fmt.Fprintf(&sb, "%v %.0f %.0f %.0f %.0f %.2f,%.2f,%.2f,%v|", op.Kind, x, y, w, h, op.R, op.G, op.B, op.Alpha)
		}
	}
	return sb.String()
}

func TestKeyboardFocusIsVisible(t *testing.T) {
	for _, dark := range []bool{true, false} {
		r := newRig(t, 1440, 960, "")
		if !dark {
			r.must(r.a.ToggleTheme(r.ctx))
		}
		r.tap("rail-connections")
		for _, id := range []string{"conn-save", "cd-engine-postgresql", "cd-name"} {
			r.must(r.a.page.Form().Focus(r.ctx, ""))
			plain := r.opsAround(id)
			r.must(r.a.page.Form().Focus(r.ctx, id))
			if r.opsAround(id) == plain {
				t.Errorf("dark=%v: focusing %s changes nothing on screen", dark, id)
			}
		}
		r.tap("conn-close")
		r.tap("rail-settings")
		r.must(r.a.page.Form().Focus(r.ctx, ""))
		plain := r.opsAround("set-confirm")
		r.must(r.a.page.Form().Focus(r.ctx, "set-confirm"))
		if r.opsAround("set-confirm") == plain {
			t.Errorf("dark=%v: switch focus is invisible", dark)
		}
	}
}
