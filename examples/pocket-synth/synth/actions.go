package synth

import (
	"context"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

// onClick handles mouse clicks on controls, keys, and buttons.
func (a *App) onClick(ctx context.Context, box ownframe.Box) error {
	act := box.Action
	if act == "" {
		act = box.ID
	}
	if act == "" {
		return nil
	}

	switch {
	case act == "preset" || act == "preset-btn":
		a.cyclePreset()
	case act == "oct-down":
		a.shiftOctave(-1)
	case act == "oct-up":
		a.shiftOctave(1)
	case act == "play-pause" || act == "play":
		a.togglePlayback()
	case act == "stop":
		a.stopPlayback()
	case strings.HasPrefix(act, "note-") || strings.HasPrefix(act, "key-"):
		name := strings.TrimPrefix(strings.TrimPrefix(act, "note-"), "key-")
		a.triggerByName(name)
	default:
		return nil
	}

	a.page.SetData(&a.view)
	return nil
}

// cyclePreset switches to the next available synthesizer waveform.
func (a *App) cyclePreset() {
	switch a.view.Preset {
	case "square":
		a.view.Preset = "sawtooth"
		a.view.PresetName = "ACID SAW"
	case "sawtooth":
		a.view.Preset = "triangle"
		a.view.PresetName = "WARM TRIANGLE"
	case "triangle":
		a.view.Preset = "sine"
		a.view.PresetName = "PURE SINE"
	default:
		a.view.Preset = "square"
		a.view.PresetName = "8-BIT SQUARE"
	}
}

// shiftOctave adjusts the current octave up or down within [2, 6].
func (a *App) shiftOctave(delta int) {
	newOct := a.view.Octave + delta
	if newOct >= 2 && newOct <= 6 {
		a.view.Octave = newOct
	}
}

// togglePlayback starts or pauses the arpeggio sequence loop.
func (a *App) togglePlayback() {
	a.view.Playing = !a.view.Playing
	if a.view.Playing {
		a.view.Status = "PLAYING"
	} else {
		a.view.Status = "PAUSED"
		a.view.BeatLED = false
	}
}

// stopPlayback stops the sequence and resets active note.
func (a *App) stopPlayback() {
	a.view.Playing = false
	a.view.Status = "READY"
	a.view.BeatLED = false
	a.view.ActiveNote = "--"
}
