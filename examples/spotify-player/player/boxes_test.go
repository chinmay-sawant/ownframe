package player

import "testing"

func TestBoxesExist(t *testing.T) {
	app := mustApp(t)
	redraw(t, app)

	ids := []string{
		"play", "seek", "volume", "row-0", "row-4",
		"card-0", "pick-0", "nav-home", "list-0", "seek-0", "volume-9", "avatar",
	}

	for _, id := range ids {
		if _, ok := findBox(app.Boxes(), id); !ok {
			t.Fatalf("no box %s", id)
		}
	}
}

func TestPlayToggles(t *testing.T) {
	app := mustApp(t)
	redraw(t, app)

	before := app.View().Playing
	click(t, app, "play")

	if app.View().Playing == before {
		t.Fatal("Playing did not toggle")
	}

	redraw(t, app)
}

func TestRowSelectAndNext(t *testing.T) {
	app := mustApp(t)
	redraw(t, app)

	click(t, app, "row-2")

	v := app.View()
	if !v.Tracks[2].Active || v.Now.Title != v.Tracks[2].Title {
		t.Fatalf("row-2 = %+v", v.Now)
	}

	if v.Progress != 0 || v.Elapsed != "0:00" {
		t.Fatalf("progress = %d %q", v.Progress, v.Elapsed)
	}

	click(t, app, "next")

	if v = app.View(); !v.Tracks[3].Active || v.Now.Title != v.Tracks[3].Title {
		t.Fatalf("next = %+v", v.Now)
	}

	paginateTracks(&app.view, app.view.TrackPages-1)
	app.page.SetData(app.view)
	redraw(t, app)
	click(t, app, "row-29")
	click(t, app, "next")

	if !app.View().Tracks[0].Active {
		t.Fatal("next did not wrap")
	}
}

func TestSeekAndVolumeExtremes(t *testing.T) {
	app := mustApp(t)
	redraw(t, app)

	click(t, app, "seek-0")

	v := app.View()
	if v.Progress != 10 || v.Elapsed == "1:08" || v.Remaining == "2:24" {
		t.Fatalf("seek-0 = %d %q %q", v.Progress, v.Elapsed, v.Remaining)
	}

	click(t, app, "seek-9")

	if v = app.View(); v.Progress != 100 || v.Elapsed != "1:58" || v.Remaining != "0:00" {
		t.Fatalf("seek-9 = %d %q %q", v.Progress, v.Elapsed, v.Remaining)
	}

	click(t, app, "volume-0")

	if v = app.View(); v.Volume != 10 {
		t.Fatalf("volume-0 = %d", v.Volume)
	}

	click(t, app, "volume-9")

	if v = app.View(); v.Volume != 100 {
		t.Fatalf("volume-9 = %d", v.Volume)
	}

	redraw(t, app)
}
