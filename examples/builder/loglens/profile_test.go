package main

import (
	"context"
	"math"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestProfileAvatarAndMenuAlignment(t *testing.T) {
	for _, width := range []int{minWidth, windowWidth} {
		a, err := newApp()
		if err != nil {
			t.Fatal(err)
		}
		ctx := context.Background()
		a.page.SetSize(width, minHeight)
		if err = a.page.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
		box := func(id string) ownframe.Box {
			return loglensLayoutBox(t, a.page, func(b ownframe.Box) bool { return b.ID == id })
		}
		trigger := box("profile-trigger")
		avatar := loglensLayoutBox(t, a.page, func(b ownframe.Box) bool {
			return b.Tag == "span" && b.W == 32 && b.H == 32 && b.X >= trigger.X && b.X+b.W <= trigger.X+trigger.W && b.Y >= trigger.Y && b.Y+b.H <= trigger.Y+trigger.H
		})
		if math.Abs(avatar.X+avatar.W/2-trigger.X-trigger.W/2) > 1 || math.Abs(avatar.Y+avatar.H/2-trigger.Y-trigger.H/2) > 1 {
			t.Fatalf("%d: avatar not centered: avatar=%+v trigger=%+v", width, avatar, trigger)
		}
		if err = a.page.Click(ctx, trigger.X+trigger.W/2, trigger.Y+trigger.H/2); err != nil {
			t.Fatal(err)
		}
		if !a.userMenuOpen {
			t.Fatal("profile menu did not open")
		}
		menu := box("user-menu-menu")
		if menu.Y < trigger.Y+trigger.H || math.Abs(menu.X+menu.W-trigger.X-trigger.W) > 1 || menu.X < 0 || menu.X+menu.W > float64(width) {
			t.Fatalf("%d: profile panel does not align beneath trigger: menu=%+v trigger=%+v", width, menu, trigger)
		}
		if after := box("profile-trigger"); !loglensSameBox(trigger, after) {
			t.Fatalf("profile moved on open: before=%+v after=%+v", trigger, after)
		}
	}
}
