package remote

import (
	"sync/atomic"
	"time"
)

type repeatState struct {
	id, host, mode string
	schedule       repeatScheduler
	epoch          atomic.Uint64
	pending        atomic.Bool
}

var remoteRepeatPolicy = repeatPolicy{Delay: repeatInitDelay, Rates: []repeatRate{
	{After: 0, Interval: repeatFirstTick},
	{After: repeatSecondAfter, Interval: repeatSecondTick},
	{After: repeatThirdAfter, Interval: repeatThirdTick},
}}

func repeatable(id string) bool {
	switch id {
	case "volup", "voldn", "chup", "chdn", "up", "down", "left", "right":
		return true
	}
	return false
}

func (a *App) armRepeat(id string) {
	a.stopRepeat()
	if !repeatable(id) {
		return
	}
	r := &a.repeat
	r.id, r.host, r.mode = id, a.hostValue(), a.view.Mode
	r.schedule.Start(remoteRepeatPolicy, time.Now())
}

func (a *App) stopRepeat() {
	a.repeat.id = ""
	a.repeat.schedule.Stop()
	a.repeat.epoch.Add(1)
}

func (a *App) repeatTick() bool {
	r := &a.repeat
	if r.id == "" {
		return false
	}
	if (a.page.PressedID() != r.id && a.nativeHeld != r.id) || a.view.Mode != r.mode || a.hostValue() != r.host {
		a.stopRepeat()
		return false
	}
	if !r.schedule.Tick(time.Now()) {
		return false
	}
	if r.mode == "Bluetooth" {
		a.sendBT(r.id)
		return true
	}
	if !r.pending.CompareAndSwap(false, true) {
		return false
	}
	host, epoch := r.host, r.epoch.Load()
	spec, _ := actionOf(r.id)
	if !a.later(func() {
		defer r.pending.Store(false)
		if r.epoch.Load() == epoch {
			a.doControl(host, spec, true)
		}
	}) {
		r.pending.Store(false)
	}
	return false
}
