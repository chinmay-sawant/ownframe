package player

import "testing"

func TestNewDrawsReplayable(t *testing.T) {
	app := mustApp(t)
	redraw(t, app)
}

func TestSampleDataRenderPath(t *testing.T) {
	app := mustApp(t)

	v := app.View()
	if v.Status != sampleStatus {
		t.Fatalf("Status = %q", v.Status)
	}

	for i, track := range v.Tracks {
		want := bundledTracks()[i].Cover
		if track.Cover != want {
			t.Fatalf("track %d cover = %q", i, track.Cover)
		}
	}

	if v.Now.Cover != bundledTracks()[0].Cover || v.Now.Title != v.Tracks[0].Title {
		t.Fatalf("Now = %+v", v.Now)
	}

	redraw(t, app)
}

func TestRedrawAtMinSize(t *testing.T) {
	app := mustApp(t)
	app.Page().SetSize(MinWidth, MinHeight)
	redraw(t, app)
}

func TestDefaultViewShape(t *testing.T) {
	v := DefaultView()

	if len(v.Tracks) != 30 || len(v.PageTracks) != trackPageSize || v.TrackPages != 6 || len(v.Picks) != 6 || len(v.Shelf) != 6 || len(v.Playlists) != 5 {
		t.Fatalf("shape = %d/%d/%d/%d", len(v.Tracks), len(v.Picks), len(v.Shelf), len(v.Playlists))
	}

	if v.Now != v.Tracks[0] {
		t.Fatalf("Now = %+v want %+v", v.Now, v.Tracks[0])
	}

	if v.Status != sampleStatus || v.ShelfTitle != "Recently played" {
		t.Fatalf("Status = %q ShelfTitle = %q", v.Status, v.ShelfTitle)
	}

	if v.Progress != 32 || v.Elapsed != "1:08" || v.Remaining != "2:24" {
		t.Fatalf("progress = %d %q %q", v.Progress, v.Elapsed, v.Remaining)
	}
}
