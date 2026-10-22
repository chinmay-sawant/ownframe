package remote

import (
	"context"
	"strings"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

func (a *App) onPress(ctx context.Context, box ownframe.Box, x, y float64) (bool, error) {
	if box.ID == "sensitivity" {
		a.sensitivityAt(x)
		a.flash(box.ID)
		a.setData()
		return true, nil
	}
	if box.ID == "pad" || box.ID == "host" || box.ID == "" {
		return false, nil
	}
	_, wifi := actionOf(box.ID)
	_, bt := btCommand(box.ID)
	if !wifi && !bt && box.ID != "wake" {
		return false, nil
	}
	a.armRepeat(box.ID)
	return true, a.onClick(ctx, box)
}

func (a *App) flash(id string) {
	a.view.PressedID = id
	a.pressedUntil = time.Now().Add(pressFlashDelay)
}

func (a *App) releaseFlash() bool {
	if a.nativeHeld != "" || a.view.PressedID == "" || time.Now().Before(a.pressedUntil) {
		return false
	}
	a.view.PressedID = ""
	return true
}

func (a *App) pressedClass(id string) string {
	if id == a.view.PressedID {
		return " pressed"
	}
	return ""
}

func (a *App) hostValue() string { return strings.TrimSpace(a.page.FormValue("host")) }
