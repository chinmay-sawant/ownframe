package player

import "testing"

func TestTrackPaginationAndGlobalSelection(t *testing.T) {
	app, _ := newFakeApp(t)
	redraw(t, app)
	if v := app.View(); v.TrackPage != 0 || !v.HasNext || v.HasPrev || len(v.PageTracks) != 5 {
		t.Fatalf("first page = %d/%d, prev=%v next=%v", len(v.PageTracks), v.TrackPages, v.HasPrev, v.HasNext)
	}
	first, _ := findBox(app.Boxes(), "row-0")
	second, _ := findBox(app.Boxes(), "row-1")
	if gap := second.Y - (first.Y + first.H); gap < 6 {
		t.Fatalf("popular row gap = %.1f px", gap)
	}

	click(t, app, "tracks-next")
	redraw(t, app)
	prev, ok := findBox(app.Boxes(), "tracks-prev")
	if !ok {
		t.Fatal("Previous button is missing")
	}
	next, ok := findBox(app.Boxes(), "tracks-next")
	if !ok {
		t.Fatal("Next button is missing")
	}
	if gap := next.X - (prev.X + prev.W); gap < 16 {
		t.Fatalf("pagination button gap = %.1f px", gap)
	}
	mainRight := 0.0
	for _, box := range app.Boxes() {
		if box.Tag == "main" {
			mainRight = box.X + box.W
			break
		}
	}
	if mainRight == 0 || next.X+next.W > mainRight-16 {
		t.Fatalf("Next button right edge %.1f exceeds main panel inset %.1f", next.X+next.W, mainRight-16)
	}
	if next.X < DefaultWidth*0.75 {
		t.Fatalf("Next button is not far enough right: x=%.1f", next.X)
	}
	if next.Y+next.H > 840 {
		t.Fatalf("Next button extends below the main panel: bottom edge %.1f", next.Y+next.H)
	}
	if v := app.View(); v.TrackPage != 1 || v.PageTracks[0].Index != 5 || !v.HasPrev || !v.HasNext {
		t.Fatalf("second page = %+v", v)
	}

	click(t, app, "tracks-prev")
	redraw(t, app)
	if v := app.View(); v.TrackPage != 0 || !v.HasNext || v.HasPrev {
		t.Fatalf("previous page = %+v", v)
	}

	paginateTracks(&app.view, app.view.TrackPages-1)
	app.page.SetData(app.view)
	redraw(t, app)
	click(t, app, "row-28")
	if v := app.View(); v.Now.Index != 28 || v.TrackPage != 5 {
		t.Fatalf("selected track = %d on page %d", v.Now.Index, v.TrackPage)
	}
}
