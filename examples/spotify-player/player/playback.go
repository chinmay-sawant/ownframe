package player

import (
	"context"
	"time"
)

// Close releases the audio device. main defers it after a successful New.
func (a *App) Close() {
	if a.audio != nil {
		a.audio.Close()
	}
	a.closeLog()
}

// Tick advances audio and animation once per frame.
func (a *App) Tick(ctx context.Context) error {
	a.page.Testing().UseFrameDirty()
	a.audio.Update()
	a.autoStart(ctx)

	if a.runtime != nil {
		wasSlow := false
		if a.runtime.Loading != nil {
			wasSlow = a.runtime.Loading.Slow
		}
		if a.audio.Loading() {
			a.runtime.StartLoading("audio")
			if a.runtime.SlowLoading(time.Now(), slowThreshold) && !wasSlow {
				a.logEvent("runtime", "audio-slow-loading")
			}
		} else if a.runtime.Loading != nil && a.runtime.Loading.Visible {
			a.runtime.ResolveLoading()
			a.logEvent("runtime", "audio-loaded")
		}
	}

	if !a.audio.Loading() {
		if err := a.advance(ctx); err != nil {
			return err
		}
	}

	a.animate()

	if a.runtime != nil && a.runtime.SampleWatchdog(time.Now(), time.Second, watchdogLimit) {
		a.logEvent("runtime", "watchdog-feedback-loop")
	}

	return nil
}
