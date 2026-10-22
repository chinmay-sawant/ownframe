package player

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoadFromTestServer(t *testing.T) {
	api := &testAPI{}
	srv := httptest.NewServer(api)
	t.Cleanup(srv.Close)

	app, err := NewAt(srv.URL)
	if err != nil {
		t.Fatalf("NewAt: %v", err)
	}

	if err := app.Load(testCtx(), "canned"); err != nil {
		t.Fatalf("Load: %v", err)
	}

	v := app.View()
	if len(v.Tracks) != 6 || v.Tracks[0].Title != "Song 0" || v.Tracks[5].Title != "Song 5" {
		t.Fatalf("tracks = %+v", v.Tracks)
	}

	if v.Tracks[0].Artist != "Artist 0" || v.Tracks[0].Album != "Album 0" || v.Tracks[0].Year != "2021" {
		t.Fatalf("track fields = %+v", v.Tracks[0])
	}

	if v.Tracks[0].Length != "3:20" || v.Tracks[0].Cover != "track-0" {
		t.Fatalf("track length/cover = %+v", v.Tracks[0])
	}

	if len(v.Picks) != 6 || v.Picks[0].Title != "Record 0" {
		t.Fatalf("picks = %+v", v.Picks)
	}

	if len(v.Shelf) != 6 || v.Shelf[5].Title != "Record 11" {
		t.Fatalf("shelf = %+v", v.Shelf)
	}

	if len(v.Playlists) != 5 || v.Playlists[0].Cover != "card-0" {
		t.Fatalf("playlists = %+v", v.Playlists)
	}

	if !strings.Contains(v.Status, "Live") || v.ShelfTitle != "Top results for canned" {
		t.Fatalf("status = %q shelf = %q", v.Status, v.ShelfTitle)
	}

	if v.Now.Title != v.Tracks[0].Title {
		t.Fatalf("now = %+v", v.Now)
	}

	for _, frag := range []string{"entity=song", "entity=album", "300x300bb.png", "300x300bb.jpg"} {
		if !api.saw(frag) {
			t.Fatalf("missing %s in %v", frag, api.paths)
		}
	}

	redraw(t, app)
}

func TestLoadOfflineKeepsSample(t *testing.T) {
	app, err := NewAt("http://127.0.0.1:1")
	if err != nil {
		t.Fatalf("NewAt: %v", err)
	}

	if err := app.Load(context.Background(), "canned"); err == nil {
		t.Fatal("Load: want an error")
	}

	if app.View().Status != sampleStatus {
		t.Fatalf("Status = %q", app.View().Status)
	}

	if len(app.View().Tracks) != 30 {
		t.Fatalf("tracks = %d", len(app.View().Tracks))
	}

	redraw(t, app)
}
