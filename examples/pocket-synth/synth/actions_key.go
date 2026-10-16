package synth

import (
	"context"
	"fmt"
	"strings"
)

// onKeyDown routes computer keyboard presses to synth notes and controls.
func (a *App) onKeyDown(ctx context.Context, key string) error {
	k := strings.ToLower(key)
	switch k {
	case "a":
		a.triggerByName("c")
	case "w":
		a.triggerByName("cs")
	case "s":
		a.triggerByName("d")
	case "e":
		a.triggerByName("ds")
	case "d":
		a.triggerByName("e")
	case "f":
		a.triggerByName("f")
	case "t":
		a.triggerByName("fs")
	case "g":
		a.triggerByName("g")
	case "y":
		a.triggerByName("gs")
	case "h":
		a.triggerByName("a")
	case "u":
		a.triggerByName("as")
	case "j":
		a.triggerByName("b")
	case "k":
		a.triggerByName("c2")
	case "z":
		a.shiftOctave(-1)
	case "x":
		a.shiftOctave(1)
	case "tab":
		a.cyclePreset()
	case "space", "p":
		a.togglePlayback()
	default:
		return nil
	}

	a.page.SetData(&a.view)
	return nil
}

// triggerByName parses a note suffix and triggers the corresponding tone.
func (a *App) triggerByName(name string) {
	norm := strings.ToLower(name)
	switch norm {
	case "c":
		a.triggerNote("C", 0)
	case "cs", "c#":
		a.triggerNote("C#", 0)
	case "d":
		a.triggerNote("D", 0)
	case "ds", "d#":
		a.triggerNote("D#", 0)
	case "e":
		a.triggerNote("E", 0)
	case "f":
		a.triggerNote("F", 0)
	case "fs", "f#":
		a.triggerNote("F#", 0)
	case "g":
		a.triggerNote("G", 0)
	case "gs", "g#":
		a.triggerNote("G#", 0)
	case "a":
		a.triggerNote("A", 0)
	case "as", "a#":
		a.triggerNote("A#", 0)
	case "b":
		a.triggerNote("B", 0)
	case "c2":
		a.triggerNote("C", 1)
	}
}

// triggerNote plays the note, updates the display and lights the beat LED.
func (a *App) triggerNote(note string, octOffset int) {
	oct := a.view.Octave + octOffset
	a.view.ActiveNote = fmt.Sprintf("%s%d", note, oct)
	a.view.BeatLED = true
	a.audio.PlayNote(note, oct, a.view.Preset)
	a.pulseWaveform(note, oct)
}
