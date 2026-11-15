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

// Tick advances audio and animation once per frame. Per-tick screen motion
// stays in the retained ops the tick patches in place (seek fill, time
// labels, credit); the tick never touches template-bound view fields and
// never forces damage, so a quiet frame stays clean and can idle.
func (a *App) Tick(ctx context.Context) error {
	a.audio.Update()
	a.publishClip()
	a.publishPlayback()
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

// publishClip hands the current encoded clip to the page. The web preview
// plays whatever a page published. The same bytes are not published twice.
func (a *App) publishClip() {
	clip := a.audio.Credit()
	if len(clip.Data) == 0 {
		return
	}

	key := clip.Title + "\x00" + clip.Creator + "\x00" + clip.URL
	if key == a.clipKey {
		return
	}

	a.clipKey = key
	a.page.SetAudio(clip.Data, clipType(clip.Data))
}

// publishPlayback tells the preview whether the clip is running and how
// loud it is. Volume on the page is 0 to 1. The view stores percent.
func (a *App) publishPlayback() {
	a.page.SetAudioPlayback(a.view.Playing, float64(a.view.Volume)/percentScale)
}

func clipType(data []byte) string {
	if len(data) >= 12 && string(data[:4]) == "RIFF" && string(data[8:12]) == "WAVE" {
		return "audio/wav"
	}

	return "audio/mpeg"
}
