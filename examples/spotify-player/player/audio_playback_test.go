package player

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/music"
)

func TestBundledPlaybackStartsWithoutNetwork(t *testing.T) {
	app, err := NewAt("http://127.0.0.1:1")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(app.Close)
	voice := &fakeVoice{dur: 118 * time.Second}
	app.audio.Open = func(data []byte) (music.Voice, error) {
		if len(data) < 2 || data[0] != 0xff || data[1]&0xe0 != 0xe0 {
			t.Fatal("playback did not receive the bundled MP3")
		}
		return voice, nil
	}
	redraw(t, app)
	click(t, app, "play")
	waitAudio(t, app)
	if !voice.playing || app.audio.Credit().Title != "2 AM Debug Loop" {
		t.Fatalf("voice playing = %v, credit = %+v", voice.playing, app.audio.Credit())
	}
}
