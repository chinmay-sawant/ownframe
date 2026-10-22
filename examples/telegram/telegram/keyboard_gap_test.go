package telegram_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestPhoneKeyboardRotationRestoresThreadBackground(t *testing.T) {
	ctx := context.Background()
	app := phoneApp(t)
	click(t, ctx, app, "", "open-anna")
	app.Page().Window().SetDevTools(false)
	game := ownframe.NewGame(ctx, app.Page())
	for _, state := range [][3]int{{420, 934, 16}, {420, 934, 400}, {885, 420, 200}, {885, 420, 16}, {420, 934, 16}} {
		app.SetInsets(28, state[2])
		game.Layout(state[0], state[1])
		for i := 0; i < 3; i++ {
			if err := game.Update(); err != nil {
				t.Fatal(err)
			}
		}
		thread := boxByID(t, app, "thread")
		composer := boxByID(t, app, "composebar")
		_, offset := app.Page().Scroll().ScrollOffset()
		composerTop := float64(offset+state[1]-state[2]) - composer.H
		t.Logf("state=%v offset=%d threadEnd=%.1f composerTop=%.1f", state, offset, thread.Y+thread.H, composerTop)
		if thread.Y+thread.H+1 < composerTop {
			t.Fatalf("keyboard gap: thread ends at %.1f, composer begins at %.1f", thread.Y+thread.H, composerTop)
		}
	}
}
