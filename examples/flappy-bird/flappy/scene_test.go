package flappy

import (
	"context"
	"strings"
	"testing"
)

// sceneTextIDs are the HUD ids the paint step writes text into.
var sceneTextIDs = []string{
	"t-score", "t-title", "t-hint", "board",
	"t-over", "t-oscore", "t-obest", "t-again",
}

// sceneIDs is every id the scene owns: the bird, the pipe slots, the
// clouds, the dashes, and the HUD.
func sceneIDs() []string {
	ids := append([]string{}, birdIDs[:]...)
	ids = append(ids, cloudIDs[:]...)
	ids = append(ids, stripeIDs[:]...)

	for slot := range pipeSlots {
		for _, part := range pipeParts {
			ids = append(ids, pipeID(slot, part))
		}
	}

	return append(ids, sceneTextIDs...)
}

// TestSceneIDs pins one occurrence of every id the paint step binds. A
// duplicate would make bind pick the wrong box; a missing one would leave a
// fill unmovable.
func TestSceneIDs(t *testing.T) {
	const want = birdCount + pipeSlots*pipePartCount + cloudMax + stripeMax + 8

	ids := sceneIDs()
	if len(ids) != want {
		t.Fatalf("sceneIDs has %d ids, want %d", len(ids), want)
	}

	html := buildHTML()

	for _, id := range ids {
		if got := strings.Count(html, `id="`+id+`"`); got != 1 {
			t.Errorf("id %q appears %d times, want 1", id, got)
		}
	}
}

// TestSceneStaysReplayable keeps the page on the display-list path and free
// of controls.
func TestSceneStaysReplayable(t *testing.T) {
	html := buildHTML()

	for _, bad := range []string{"<button", "<input", "data-action"} {
		if strings.Contains(html, bad) {
			t.Errorf("buildHTML contains %q", bad)
		}
	}
}

// TestSceneReplaysWithoutABitmap renders the page and demands the vector
// display list, so the window never falls back to the raster picture.
func TestSceneReplaysWithoutABitmap(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatal(err)
	}

	if err := app.page.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	if app.page.Display() == nil {
		t.Fatal("the scene fell back to the bitmap path")
	}
}
