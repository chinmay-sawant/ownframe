package pet

import (
	"context"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

func (a *App) onClick(_ context.Context, b ownframe.Box) error {
	switch b.ID {
	case "feed", "action-feed", "btn-feed":
		a.Feed()
	case "play", "action-play", "btn-play":
		a.Play()
	case "pet", "action-pet", "btn-pet":
		a.Pet()
	case "sleep", "action-sleep", "btn-sleep":
		a.Sleep()
	case "talk", "action-talk", "btn-talk":
		a.Talk()
	default:
		if isFaceBox(b.ID) {
			a.Pet()
		}
	}
	return nil
}

func (a *App) onKeyDown(_ context.Context, k string) error {
	switch k {
	case "1":
		a.Feed()
	case "2":
		a.Play()
	case "3":
		a.Pet()
	case "4":
		a.Sleep()
	case "5":
		a.Talk()
	case "space":
		a.Pet()
	}
	return nil
}

func (a *App) onDragStart(_ context.Context, b ownframe.Box) (bool, error) {
	if isFaceBox(b.ID) {
		a.draggingFace = true
		a.Pet()
		return true, nil
	}
	return false, nil
}

func (a *App) onDragMove(_ context.Context, _, _ float64) error {
	if a.draggingFace {
		now := a.now()
		if now.Sub(a.lastPetDrag) > 250*time.Millisecond {
			a.lastPetDrag = now
			a.Pet()
		}
	}
	return nil
}

func (a *App) onDragEnd(_ context.Context) error {
	a.draggingFace = false
	return nil
}

func isFaceBox(id string) bool {
	switch id {
	case "pet-face", "face", "pet-avatar", "face-body",
		"eye-left", "eye-right", "lid-left", "lid-right",
		"blush-left", "blush-right", "pet-mouth", "mouth",
		"ear-l", "ear-r":
		return true
	default:
		return false
	}
}
