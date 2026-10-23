package remote

import (
	"context"

	"github.com/chinmay-sawant/ownframe/examples/lg-remote/bridge"
)

func (a *App) onTick(ctx context.Context) error {
	if err := a.accessibilityActions(ctx); err != nil {
		return err
	}
	defer a.publishAccessibility()
	changed := a.load() || a.touchDirty
	a.touchDirty = false
	changed = a.repeatTick() || changed
	changed = a.releaseFlash() || changed
	if size := bridge.FontSize(); size != a.view.FontSize {
		a.view.FontSize = size
		changed = true
	}
	if a.pollWake() {
		changed = true
	}
	if a.drain() {
		changed = true
	}
	if radio, rev := bridge.BluetoothStatus(); a.view.Mode == "Bluetooth" && radio != "" && rev != a.btSeen {
		a.btSeen = rev
		a.view.Status = radio
		changed = true
	}

	if !changed {
		return nil
	}

	a.setData()

	return a.page.Redraw(ctx)
}

func (a *App) load() bool {
	if a.loaded {
		return false
	}

	a.loaded = true
	link := a.use()
	if host := link.SavedHost(); host != "" {
		a.page.Form().SetFormValue("host", host)
	}

	dirty := false
	if model := link.SavedModel(); model != "" {
		a.view.Title = model
		dirty = true
	}

	if !a.wantSearch {
		return dirty
	}

	a.view.Status = "Searching this Wi-Fi"
	a.later(a.doFind)

	return true
}

func (a *App) drain() bool {
	changed := false

	for {
		select {
		case u := <-a.notes:
			if u.powerSeq != 0 && u.powerSeq != a.powerIntent.seq {
				continue
			}
			if u.status != "" {
				a.view.Status = u.status
			}
			if u.title != "" {
				a.view.Title = u.title
			}
			if u.host != "" {
				a.page.Form().SetFormValue("host", u.host)
			}
			if u.setPower && (u.powerSeq != 0 || !a.powerIntent.pending) {
				a.view.PowerOn = u.powerOn
				if u.powerSeq != 0 {
					a.powerIntent.pending = false
				}
				if u.powerOn {
					a.wake = wakeState{}
				}
			}
			if u.wakeHost != "" && a.async {
				a.armWake(u.wakeHost)
			}
			changed = true
		default:
			return changed
		}
	}
}
