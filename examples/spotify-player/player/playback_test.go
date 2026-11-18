package player

import (
	"testing"
	"time"
)

func TestPlayClickStartsVoice(t *testing.T) {
	app, voice := newFakeApp(t)
	redraw(t, app)

	click(t, app, "play")
	waitAudio(t, app)

	if !voice.playing {
		t.Fatal("voice did not start")
	}

	if !app.View().Playing {
		t.Fatal("Playing intent lost")
	}
}

func TestSeekClickMovesVoice(t *testing.T) {
	app, voice := newFakeApp(t)
	redraw(t, app)

	click(t, app, "play")
	waitAudio(t, app)

	click(t, app, "seek-60")

	if want := time.Duration(0.6 * float64(voice.dur)); voice.pos != want {
		t.Fatalf("pos = %v want %v", voice.pos, want)
	}
}

func TestVolumeClicksDriveVoice(t *testing.T) {
	app, voice := newFakeApp(t)
	redraw(t, app)

	click(t, app, "play")
	waitAudio(t, app)

	click(t, app, "volume-2")

	if voice.volume != 0.3 {
		t.Fatalf("volume = %v", voice.volume)
	}

	click(t, app, "mute")

	if voice.volume != 0 {
		t.Fatalf("muted volume = %v", voice.volume)
	}
}
