package main

import (
	"context"
	"fmt"
	"regexp"

	"github.com/chinmay-sawant/ownframe"
)

// sideWheelStep is how many CSS pixels one wheel notch moves the tab list.
const sideWheelStep = 40

// sideWheel scrolls the fixed tab list when the pointer is over it. The window
// only pans the page on a wheel event, so the pinned list, which has more tabs
// than a short window shows, moves its own content through a relative top
// offset and draws a thumb. The offset is read back from the laid-out boxes, so a page swap
// resets it with no state kept here.
func sideWheel(page *ownframe.Page) func(context.Context, float64, float64, float64, float64) (bool, error) {
	return func(ctx context.Context, x, y, _, dy float64) (bool, error) {
		var side, in *ownframe.Box

		boxes := page.Boxes()
		for i := range boxes {
			switch boxes[i].ID {
			case "side":
				side = &boxes[i]
			case "side-in":
				in = &boxes[i]
			}
		}

		if side == nil || in == nil {
			return false, nil
		}

		_, scrollY := page.Scroll().ScrollOffset()
		vy := y - float64(scrollY)

		if x < side.X || x > side.X+side.W || vy < side.Y || vy > side.Y+side.H {
			return false, nil
		}

		limit := in.H - side.H
		if limit <= 0 {
			return true, nil
		}

		off := side.Y - in.Y
		off = min(max(off-dy*sideWheelStep, 0), limit)

		thumb := side.H * side.H / in.H
		thumbTop := off / limit * (side.H - thumb)

		inStyle := fmt.Sprintf("position:relative;top:-%dpx", int(off))
		barStyle := fmt.Sprintf("top:%dpx;height:%dpx", int(thumbTop), int(thumb))

		err := page.PatchElements(ctx, []ownframe.ElementPatch{
			{ID: "side-in", Edits: []ownframe.ElementEdit{{Attribute: "style", Value: inStyle}}},
			{ID: "side-bar", Edits: []ownframe.ElementEdit{{Attribute: "style", Value: barStyle}}},
		})
		if err == nil {
			return true, nil
		}

		// The engine refuses a subtree patch on a page with anchored boxes, so
		// rewrite the two styles in the source and reload it in place.
		src := sideInStyle.ReplaceAllString(page.HTML(), `${1}`+inStyle+`"`)
		src = sideBarStyle.ReplaceAllString(src, `${1}`+barStyle+`"`)

		return true, page.Load(ctx, src, ownframe.LoadOptions{ReplaceCurrent: true})
	}
}

var (
	sideInStyle  = regexp.MustCompile(`(id="side-in" style=")[^"]*"`)
	sideBarStyle = regexp.MustCompile(`(id="side-bar" class="sidebar" style=")[^"]*"`)
)
