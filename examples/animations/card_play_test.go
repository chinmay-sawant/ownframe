package main

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

func TestEveryCardCanPlayAlone(t *testing.T) {
	app, err := newApp(false, false)
	if err != nil {
		t.Fatal(err)
	}
	p := app.Page()
	for _, selected := range showcaseTargetIDs() {
		t.Run(selected, func(t *testing.T) {
			clickCard(t, p, "play-"+selected)
			if app.view.Solo != selected || app.view.IsPlaying || app.view.Reset {
				t.Fatalf("unexpected playback state: %+v", app.view)
			}
			for _, id := range showcaseTargetIDs() {
				want := "paused"
				if id == selected {
					want = "running"
				}
				for _, animation := range boxByID(t, p, id).Animations {
					if animation.PlayState != want {
						t.Fatalf("%s play state = %q, want %q", id, animation.PlayState, want)
					}
				}
			}
			button := boxByID(t, p, "play-"+selected)
			card := boxByID(t, p, "card-"+selected)
			if math.Abs(button.X+button.W-(card.X+card.W-15)) > 2 ||
				math.Abs(button.Y+button.H-(card.Y+card.H-15)) > 2 {
				t.Fatalf("button is not at card's bottom right: button=%+v card=%+v", button, card)
			}
		})
	}
	clickCard(t, p, togglePlayID)
	if !app.view.IsPlaying || app.view.Solo != "" {
		t.Fatal("Play all did not leave solo mode")
	}
	for _, id := range showcaseTargetIDs() {
		for _, animation := range boxByID(t, p, id).Animations {
			if animation.PlayState != "running" {
				t.Fatalf("%s stayed paused after Play all", id)
			}
		}
	}
}

func TestCardReplayRestartsAndKeepsOthersFrozen(t *testing.T) {
	app, err := newApp(false, false)
	if err != nil {
		t.Fatal(err)
	}
	p := app.Page()
	ctx := context.Background()
	clickCard(t, p, "play-show-translate")
	time.Sleep(40 * time.Millisecond)
	if err := p.Testing().Tick(ctx); err != nil {
		t.Fatal(err)
	}
	moved := translationX(t, p, "show-translate")
	frozen := translationX(t, p, "show-linear")
	clickCard(t, p, "play-show-translate")
	if err := p.Testing().Tick(ctx); err != nil {
		t.Fatal(err)
	}
	if got := translationX(t, p, "show-translate"); got >= moved {
		t.Fatalf("Replay did not restart: got %g, previous %g", got, moved)
	}
	if got := translationX(t, p, "show-linear"); math.Abs(got-frozen) > cardPaintTolerance {
		t.Fatalf("replaying moved another card: got %g, frozen %g", got, frozen)
	}
	clickCard(t, p, "play-show-delay-positive")
	if err := p.Testing().Tick(ctx); err != nil {
		t.Fatal(err)
	}
	if got := requireFill(t, p, "show-delay-positive").Opacity(); math.Abs(got-0.3) > cardPaintTolerance {
		t.Fatalf("solo delay did not restart at the faded pose: %g", got)
	}
}

func TestCompletedCardCanReplay(t *testing.T) {
	app, err := newApp(false, false)
	if err != nil {
		t.Fatal(err)
	}
	p := app.Page()
	ctx := context.Background()
	if err := p.SetTheme(`#show-iteration-once { animation-duration: 100ms !important; }`); err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	clickCard(t, p, "play-show-iteration-once")
	time.Sleep(120 * time.Millisecond)
	if err := p.Testing().Tick(ctx); err != nil {
		t.Fatal(err)
	}
	completed := translationX(t, p, "show-iteration-once")
	const endpoint = 130 * 72.0 / 96.0
	if math.Abs(completed-endpoint) > cardPaintTolerance {
		t.Fatalf("one-shot did not finish: %g", completed)
	}
	clickCard(t, p, "play-show-iteration-once")
	if err := p.Testing().Tick(ctx); err != nil {
		t.Fatal(err)
	}
	if got := translationX(t, p, "show-iteration-once"); got >= completed {
		t.Fatalf("completed card did not replay: %g", got)
	}
}

// Match the host's batched press/click dispatch and the following frame tick.
func clickCard(t *testing.T, p *ownframe.Page, id string) {
	t.Helper()
	p.Probe().BeginRedrawBatch()
	clickID(t, p, id)
	if err := p.Probe().EndRedrawBatch(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := p.Testing().Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestDelayAndFillCardsReplayThroughPointerBatch(t *testing.T) {
	for _, id := range []string{
		"show-delay-positive", "show-fill-none", "show-fill-forwards", "show-fill-both",
	} {
		t.Run(id, func(t *testing.T) {
			app, err := newApp(false, false)
			if err != nil {
				t.Fatal(err)
			}
			p := app.Page()
			ctx := context.Background()
			// Keep the original keyframes and fill modes, shortening only the wait.
			theme := "#" + id + " { animation-duration: 2s !important; animation-delay: 40ms !important; }"
			if err := p.SetTheme(theme); err != nil {
				t.Fatal(err)
			}
			if err := p.Redraw(ctx); err != nil {
				t.Fatal(err)
			}
			for pass := 0; pass < 2; pass++ {
				clickCard(t, p, "play-"+id)
				if err := p.Testing().Tick(ctx); err != nil {
					t.Fatal(err)
				}
				if got := translationX(t, p, id); math.Abs(got) > cardPaintTolerance {
					t.Fatalf("play %d did not restart its delay: translation=%g", pass, got)
				}
				time.Sleep(120 * time.Millisecond)
				if err := p.Testing().Tick(ctx); err != nil {
					t.Fatal(err)
				}
				if got := translationX(t, p, id); got <= 0 {
					t.Fatalf("play %d changed its label but did not animate: translation=%g", pass, got)
				}
			}
			if p.Probe().Motion().LastError != "" {
				t.Fatal(p.Probe().Motion().LastError)
			}
		})
	}
}
