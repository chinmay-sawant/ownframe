package player

import (
	"context"
	"testing"
	"time"
)

func TestNavigationCommitsCurrentPlaybackPosition(t *testing.T) {
	for _, action := range []string{"nav-library", "list-1"} {
		t.Run(action, func(t *testing.T) {
			app, voice := newFakeApp(t)
			redraw(t, app)

			click(t, app, "play")
			waitAudio(t, app)

			voice.pos = 30 * time.Second
			if err := app.Tick(context.Background()); err != nil {
				t.Fatalf("Tick: %v", err)
			}
			if app.View().Progress != 0 {
				t.Fatalf("progress before navigation = %d, want the uncommitted template value 0", app.View().Progress)
			}

			click(t, app, action)

			if got := app.View().Progress; got != 15 {
				t.Fatalf("progress after navigation = %d, want 15 for 30s of a 200s track", got)
			}
			if got := app.View().Elapsed; got != "0:30" {
				t.Fatalf("elapsed after navigation = %q, want 0:30", got)
			}
			if voice.pos != 30*time.Second {
				t.Fatalf("navigation moved audio to %v, want it to remain at 30s", voice.pos)
			}
		})
	}
}
