package board

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// onClick handles the header buttons: a livery, the night toggle, or
// pause. Click draws after the handler, so no Redraw here. Unknown ids
// pass through.
func (a *App) onClick(_ context.Context, box ownframe.Box) error {
	switch box.ID {
	case "livery-amber", "livery-flap", "livery-lcd", "livery-paper":
		name := box.ID[len("livery-"):]
		if err := a.SetLivery(name); err != nil {
			return err
		}

		a.view.Status = name
	case "btn-night":
		on := !a.night
		if err := a.SetNight(on); err != nil {
			return err
		}

		a.view.Status = "day"
		if on {
			a.view.Status = "night"
		}
	case "btn-pause":
		a.paused = !a.paused
		a.view.Paused = a.paused

		a.view.Status = "live"
		if a.paused {
			a.view.Status = "paused"
		}
	}

	return nil
}
