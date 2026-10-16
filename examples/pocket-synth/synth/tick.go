package synth

import (
	"context"
	"time"
)

// arpeggioSeq defines the 16-step pattern for sequencer playback.
var arpeggioSeq = []string{
	"C", "E", "G", "B", "C", "B", "G", "E",
	"D", "F", "A", "C", "D", "C", "A", "F",
}

// Tick advances the sequence, animates the oscilloscope and pulses the LED.
func (a *App) Tick(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	a.tickCount++
	now := time.Now()

	if a.view.Playing {
		if a.lastStep.IsZero() || now.Sub(a.lastStep) >= 125*time.Millisecond {
			note := arpeggioSeq[a.seqStep%len(arpeggioSeq)]
			octOff := 0
			if a.seqStep%8 >= 4 {
				octOff = 1
			}
			a.triggerNote(note, octOff)
			a.seqStep++
			a.lastStep = now
		}
	}

	if a.view.BeatLED && !a.view.Playing {
		if now.Sub(a.lastStep) > 150*time.Millisecond && a.tickCount%10 == 0 {
			a.view.BeatLED = false
		}
	} else if a.view.Playing && now.Sub(a.lastStep) > 60*time.Millisecond {
		a.view.BeatLED = false
	}

	a.animateWaveform()

	a.page.UseFrameDirty()
	a.page.SetData(&a.view)

	return nil
}
