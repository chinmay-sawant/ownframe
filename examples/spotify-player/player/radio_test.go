package player

import "testing"

func TestRadioScreenNavAndBoxes(t *testing.T) {
	app := mustApp(t)
	redraw(t, app)

	click(t, app, "nav-radio")

	if v := app.View(); v.Nav != "radio" {
		t.Fatalf("Nav = %q", v.Nav)
	}

	redraw(t, app)

	ids := []string{"radio-feat", "radio-play", "radio-st-0", "radio-st-5"}
	for _, id := range ids {
		if _, ok := findBox(app.Boxes(), id); !ok {
			t.Fatalf("no box %s", id)
		}
	}
}

func TestRadioScreenStationClick(t *testing.T) {
	app := mustApp(t)
	redraw(t, app)
	click(t, app, "nav-radio")
	redraw(t, app)

	want := app.View().Radio.Stations[1]

	click(t, app, "radio-st-1")

	v := app.View()
	if v.Now.Title != want.Title || v.Now.Cover != want.Cover {
		t.Fatalf("Now = %+v want %+v", v.Now, want)
	}

	for i, st := range v.Radio.Stations {
		if st.Active != (i == 1) {
			t.Fatalf("station %d Active = %v", i, st.Active)
		}
	}

	redraw(t, app)
}

func TestRadioScreenPlayStartsAudio(t *testing.T) {
	app, voice := newFakeApp(t)
	redraw(t, app)
	click(t, app, "nav-radio")
	redraw(t, app)

	click(t, app, "radio-play")

	if !app.View().Playing {
		t.Fatal("Playing = false after radio-play")
	}

	waitAudio(t, app)

	if !voice.playing {
		t.Fatal("fake voice is not playing")
	}

	click(t, app, "radio-play")

	if app.View().Playing || voice.playing {
		t.Fatal("radio-play did not pause")
	}

	redraw(t, app)
}
