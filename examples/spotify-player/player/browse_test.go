package player

import (
	"fmt"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe"
)

func TestBrowseScreenNavAndIDs(t *testing.T) {
	app := mustApp(t)
	redraw(t, app)

	click(t, app, "nav-browse")

	if app.View().Nav != "browse" {
		t.Fatalf("Nav = %q", app.View().Nav)
	}

	redraw(t, app)

	ids := []string{"browse-hero-0", "browse-hero-2", "browse-tile-0", "browse-tile-11"}

	for _, id := range ids {
		if _, ok := findBox(app.Boxes(), id); !ok {
			t.Fatalf("no box %s", id)
		}
	}
}

func TestBrowseScreenTileSelect(t *testing.T) {
	app := mustApp(t)
	redraw(t, app)
	click(t, app, "nav-browse")
	click(t, app, "browse-tile-4")

	v := app.View()
	if !v.Browse.Tiles[4].Active {
		t.Fatal("tile 4 not active")
	}

	for i, tile := range v.Browse.Tiles {
		if i != 4 && tile.Active {
			t.Fatalf("tile %d is active too", i)
		}
	}
}

func TestBrowseScreenHeroSelect(t *testing.T) {
	app := mustApp(t)
	redraw(t, app)
	click(t, app, "nav-browse")
	click(t, app, "browse-hero-1")

	v := app.View()
	if !v.Browse.Hero[1].Active {
		t.Fatal("hero 1 not active")
	}

	for i, hero := range v.Browse.Hero {
		if i != 1 && hero.Active {
			t.Fatalf("hero %d is active too", i)
		}
	}
}

func TestBrowseHoverKeepsGenreArtworkAndBackground(t *testing.T) {
	app := mustApp(t)
	redraw(t, app)
	click(t, app, "nav-browse")
	redraw(t, app)
	page := app.Page()
	ctx := t.Context()

	hover := func(id string) {
		t.Helper()
		box, ok := findBox(app.Boxes(), id)
		if !ok {
			t.Fatalf("missing hover target %s", id)
		}
		if err := page.Edit().Hover(ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
			t.Fatal(err)
		}
		if err := page.Testing().Tick(ctx); err != nil {
			t.Fatal(err)
		}
	}

	tones := [][3]float64{}
	check := func(step string) {
		t.Helper()
		display := page.Probe().Display()
		if display == nil {
			t.Fatalf("%s: missing display list", step)
		}
		boxes := app.Boxes()
		for i, genre := range app.View().Browse.Tiles {
			id := fmt.Sprintf("browse-tile-%d", i)
			tile, ok := findBox(boxes, id)
			if !ok {
				t.Fatalf("%s: %s has no tile", step, genre.Title)
			}
			x, y, w, h := ownframe.BoxUnits(display, tile)
			if i == len(tones) {
				var found bool
				for _, op := range display.Ops {
					if op.Kind == layout.DisplayOpFillRect && op.X == x && op.Y == y && op.W == w && op.H == h {
						tones = append(tones, [3]float64{op.R, op.G, op.B})
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("%s: %s has no tone to track", step, genre.Title)
				}
			}
			if ownframe.Fill(display, tile, tones[i]) == nil {
				t.Fatalf("%s: %s lost its background", step, genre.Title)
			}

			art, ok := findBox(boxes, fmt.Sprintf("browse-tile-art-%d", i))
			if !ok {
				t.Fatalf("%s: %s has no artwork box", step, genre.Title)
			}
			ax, ay, aw, ah := ownframe.BoxUnits(display, art)
			cx, cy := ax+aw/2, ay+ah/2
			var hasImage bool
			for _, op := range display.Ops {
				if op.Kind == layout.DisplayOpImage && cx >= op.X && cx < op.X+op.W && cy >= op.Y && cy < op.Y+op.H {
					data, _, _ := op.ImageBytes()
					hasImage = len(data) > 0
					if hasImage {
						break
					}
				}
			}
			if !hasImage {
				t.Fatalf("%s: %s lost its artwork", step, genre.Title)
			}
		}
	}

	check("after opening Browse")
	for _, id := range []string{"browse-hero-2", "browse-tile-11"} {
		hover(id)
		click(t, app, id)
	}
	for range 3 {
		for _, id := range []string{
			"browse-tile-8", "browse-tile-9", "browse-tile-10",
			"browse-tile-art-11", "browse-tile-art-8", "browse-hero-2", "nav-browse",
		} {
			hover(id)
			check(id)
		}
	}
}
