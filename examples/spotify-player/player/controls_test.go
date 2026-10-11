package player

import "testing"

func TestCardAndPlaylistClicks(t *testing.T) {
	app := mustApp(t)
	redraw(t, app)

	click(t, app, "pick-1")

	v := app.View()
	if !v.Picks[1].Active || v.Now.Title != v.Picks[1].Title || v.Now.Cover != v.Picks[1].Cover {
		t.Fatalf("pick = %+v", v.Now)
	}

	click(t, app, "card-3")

	v = app.View()
	if !v.Shelf[3].Active || v.Picks[1].Active || v.Now.Title != v.Shelf[3].Title || v.Now.Cover != v.Shelf[3].Cover {
		t.Fatalf("card = %+v picks = %+v", v.Now, v.Picks[1])
	}

	click(t, app, "list-2")

	v = app.View()
	if v.Nav != "library" || !v.Playlists[2].Active {
		t.Fatalf("list = %q %+v", v.Nav, v.Playlists[2])
	}

	redraw(t, app)
}

func TestTogglesAndNav(t *testing.T) {
	app := mustApp(t)
	redraw(t, app)

	before := app.View()

	click(t, app, "heart")
	click(t, app, "shuffle")
	click(t, app, "repeat")

	v := app.View()
	if v.Liked == before.Liked || v.Shuffle == before.Shuffle || v.Repeat == before.Repeat {
		t.Fatalf("toggles = %+v", v)
	}

	click(t, app, "nav-search")

	if v = app.View(); v.Nav != "search" {
		t.Fatalf("Nav = %q", v.Nav)
	}

	click(t, app, "nav-home")

	if v = app.View(); v.Nav != "home" {
		t.Fatalf("Nav = %q", v.Nav)
	}

	redraw(t, app)
}

func TestMuteToggle(t *testing.T) {
	app := mustApp(t)
	redraw(t, app)

	click(t, app, "mute")

	if v := app.View(); v.Volume != 0 {
		t.Fatalf("muted volume = %d", v.Volume)
	}

	click(t, app, "mute")

	if v := app.View(); v.Volume != defaultVolume {
		t.Fatalf("restored volume = %d", v.Volume)
	}

	redraw(t, app)
}
