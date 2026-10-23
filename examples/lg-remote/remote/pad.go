package remote

import (
	"context"
	"github.com/chinmay-sawant/ownframe"
)

func (a *App) padStart(_ context.Context, box ownframe.Box) (bool, error) {
	claim := a.view.Panel == "pad" && (box.ID == "sensitivity" || (box.ID == "pad" && a.view.Mode == "Wi-Fi"))
	if claim {
		a.dragTarget = box.ID
	}
	return claim, nil
}

func (a *App) movePad(ctx context.Context, dx, dy float64) error {
	if a.dragTarget == "sensitivity" {
		a.moveSensitivity(dx)
		return a.page.Redraw(ctx)
	}
	if a.view.Panel != "pad" {
		return nil
	}

	if a.view.Mode == "Bluetooth" {
		a.view.Status = "The pad needs Wi-Fi"
		a.setData()
		return a.page.Redraw(ctx)
	}

	if dx == 0 && dy == 0 {
		return nil
	}

	host := a.page.Form().FormValue("host")
	a.queueMotion(host, dx, dy)

	return nil
}

func (a *App) padPress(id string) bool {
	spec := ""
	switch id {
	case "pad", "clicker":
		spec = "click:"
	case "wheelup":
		spec = "scroll:-3"
	case "wheeldown":
		spec = "scroll:3"
	default:
		return false
	}

	if a.view.Mode == "Bluetooth" {
		a.view.Status = "The pad needs Wi-Fi"
		return true
	}

	host := a.host
	a.later(func() { a.doExec(host, spec) })

	return true
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}

	if v > hi {
		return hi
	}

	return v
}
