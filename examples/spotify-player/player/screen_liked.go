package player

import (
	"context"
	"strings"
)

// LikedData is the data behind the Liked Songs screen: the saved count and
// the visible rows of the playlist.
type LikedData struct {
	Count  int
	Tracks []Track
}

// defaultLiked returns the offline Liked Songs screen with eight rows.
func defaultLiked() LikedData {
	tracks := sampleTracks()
	tracks = tracks[:8]

	return LikedData{Count: 214, Tracks: tracks}
}

// clickLiked applies one action on the Liked Songs screen: the header play
// and heart buttons, and a row that takes over the now bar.
func (a *App) clickLiked(ctx context.Context, action string) error {
	switch {
	case action == "liked-play":
		a.view.Playing = !a.view.Playing

		if a.view.Playing {
			if a.audio.Duration() > 0 {
				a.audio.Play()
			} else {
				a.playNow(ctx)
			}
		} else {
			a.audio.Pause()
		}
	case action == "liked-heart":
		a.view.Liked = !a.view.Liked
	case strings.HasPrefix(action, "liked-row-"):
		i := slot(action, "liked-row-")
		if i < 0 || i >= len(a.view.LikedSongs.Tracks) {
			break
		}

		for j := range a.view.LikedSongs.Tracks {
			a.view.LikedSongs.Tracks[j].Active = j == i
		}

		a.view.Now = a.view.LikedSongs.Tracks[i]
		a.resetProgress()
		a.playNow(ctx)
	}

	return nil
}
