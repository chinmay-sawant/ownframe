package remote

import (
	"bytes"
	"context"
	"image/color"
	"image/png"
	"testing"
)

func buttonColour(t *testing.T, a *App, id string) color.Color {
	t.Helper()
	b := remoteBox(t, a, id)
	img, err := png.Decode(bytes.NewReader(a.page.PNG()))
	if err != nil {
		t.Fatal(err)
	}
	return img.At(int(b.X+12), int(b.Y+12))
}

func TestPhoneHighlightsOnlySelectedButton(t *testing.T) {
	for _, light := range []bool{false, true} {
		a := newTest(t, WithPhone(true))
		if light {
			a.toggleTheme()
			if err := a.page.Redraw(context.Background()); err != nil {
				t.Fatal(err)
			}
		}
		selected := buttonColour(t, a, "volup")
		other := buttonColour(t, a, "voldn")
		if err := a.page.Focus(context.Background(), "volup"); err != nil {
			t.Fatal(err)
		}
		if buttonColour(t, a, "volup") == selected {
			t.Fatal("selected button did not highlight")
		}
		if buttonColour(t, a, "voldn") != other {
			t.Fatal("selection changed a different button")
		}
	}
}
