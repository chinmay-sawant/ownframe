package player

import "testing"

func TestShelfAndSidebarSpacing(t *testing.T) {
	app := mustApp(t)
	redraw(t, app)
	boxes := app.Boxes()
	artBottom := -1.0
	for _, box := range boxes {
		if box.Tag == "div" && box.ParentIndex >= 0 && boxes[box.ParentIndex].ID == "card-0" {
			artBottom = box.Y + box.H
			break
		}
	}
	if artBottom < 0 {
		t.Fatal("Recently played artwork is missing")
	}
	titleY := -1.0
	for _, box := range boxes {
		if box.Tag == "span" && box.Text == app.View().Shelf[0].Title {
			titleY = box.Y
			break
		}
	}
	if gap := titleY - artBottom; gap < 10 {
		t.Fatalf("shelf title starts too close to cover: gap %.1f px", gap)
	}
	first, _ := findBox(boxes, "list-0")
	second, _ := findBox(boxes, "list-1")
	if gap := second.Y - (first.Y + first.H); gap < 6 {
		t.Fatalf("sidebar library row gap = %.1f px", gap)
	}
}

func TestPlayerFitsViewport(t *testing.T) {
	for _, size := range []struct {
		width, height int
	}{
		{1920, 1080},
		{DefaultWidth, DefaultHeight},
		{1100, 960},
		{2560, 1600},
		{MinWidth, MinHeight},
	} {
		app := mustApp(t)
		app.Page().SetSize(size.width, size.height)
		redraw(t, app)
		display := app.Page().Probe().Display()
		var footerBottom float64
		for _, box := range display.Boxes {
			if box.Tag == "footer" {
				footerBottom = box.Y + box.H
				break
			}
		}
		if display.Height > size.height || footerBottom > float64(size.height) || footerBottom < float64(size.height-12) {
			t.Errorf("%dx%d viewport: display height %d, footer bottom %.1f",
				size.width, size.height, display.Height, footerBottom)
		}
	}
}

func TestScreenGridsUseAvailableWidth(t *testing.T) {
	for _, screen := range []struct {
		nav   string
		edges []string
	}{
		{"home", []string{"pick-5", "card-5"}},
		{"search", []string{"search-tile-3", "search-tile-11"}},
		{"browse", []string{"browse-hero-2", "browse-tile-5", "browse-tile-11"}},
		{"radio", []string{"radio-st-2", "radio-st-5"}},
	} {
		t.Run(screen.nav, func(t *testing.T) {
			app := mustApp(t)
			redraw(t, app)
			if screen.nav != "home" {
				click(t, app, "nav-"+screen.nav)
			}
			for _, width := range []int{MinWidth, DefaultWidth, 1920, 2560, MinWidth} {
				app.Page().SetSize(width, 1080)
				redraw(t, app)
				boxes := app.Page().Probe().Display().Boxes
				avatar, ok := findBox(boxes, "avatar")
				if !ok {
					t.Fatal("avatar is missing")
				}
				want := avatar.X + avatar.W
				for _, id := range screen.edges {
					box, ok := findBox(boxes, id)
					right := box.X + box.W
					if !ok || right < want-1 || right > want+1 {
						t.Errorf("%dpx: %s ends at %.1f, want %.1f", width, id, right, want)
					}
				}
			}
		})
	}
}
