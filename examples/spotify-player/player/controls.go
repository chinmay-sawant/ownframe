package player

import (
	"context"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

// onClick applies one rendered box action.
func (a *App) onClick(ctx context.Context, box ownframe.Box) error {
	action := box.Action

	switch {
	case action == "play":
		a.togglePlayback(ctx)
	case action == "next":
		a.step(1)
		a.playNow(ctx)
	case action == "prev":
		a.step(-1)
		a.playNow(ctx)
	case action == "back":
		a.moveNav(-1)
	case action == "forward":
		a.moveNav(1)
	case strings.HasPrefix(action, "row-"):
		a.selectRow(slot(action, "row-"))
		a.playNow(ctx)
	case strings.HasPrefix(action, "pick-"):
		a.selectPick(slot(action, "pick-"))
		a.playNow(ctx)
	case strings.HasPrefix(action, "card-"):
		a.selectShelf(slot(action, "card-"))
		a.playNow(ctx)
	case strings.HasPrefix(action, "list-"):
		a.selectList(slot(action, "list-"))
	case action == "queue":
		a.navigate("queue")
	case strings.HasPrefix(action, "nav-"):
		a.navigate(strings.TrimPrefix(action, "nav-"))
	case action == "heart":
		a.view.Liked = !a.view.Liked
	case action == "shuffle":
		a.view.Shuffle = !a.view.Shuffle
	case action == "repeat":
		a.view.Repeat = !a.view.Repeat
	case action == "mute":
		a.toggleMute()
		a.audio.SetVolume(float64(a.view.Volume) / percentScale)
	case strings.HasPrefix(action, "seek-"):
		a.setProgress(slot(action, "seek-")*seekStep + seekBase)
		a.audio.SeekFraction(float64(a.view.Progress) / percentScale)
	case strings.HasPrefix(action, "volume-"):
		a.view.Volume = slot(action, "volume-")*volumeStep + volumeBase
		a.audio.SetVolume(float64(a.view.Volume) / percentScale)
	case action == "search":
		return a.searchNow(ctx)
	case action == "tracks-prev":
		paginateTracks(&a.view, a.view.TrackPage-1)
	case action == "tracks-next":
		paginateTracks(&a.view, a.view.TrackPage+1)
	default:
		handled, err := a.screenClick(ctx, action)
		if err != nil {
			return err
		}

		if !handled {
			return nil
		}
	}

	a.page.SetData(&a.view)
	a.logEvent("click", action)

	return nil
}
