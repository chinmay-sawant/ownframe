package synth

import (
	"math"
)

// pulseWaveform excites the waveform display when a note is struck.
func (a *App) pulseWaveform(note string, octave int) {
	for i := range a.view.Bars {
		phase := float64(i) / float64(scopeBarCount) * 2 * math.Pi
		var h float64
		switch a.view.Preset {
		case "square":
			if i%4 < 2 {
				h = 85
			} else {
				h = 25
			}
		case "sawtooth":
			h = 20 + float64(i%8)*10
		case "triangle":
			if i < 8 {
				h = 20 + float64(i)*8
			} else {
				h = 20 + float64(15-i)*8
			}
		default: // sine
			h = 50 + 40*math.Sin(phase*2)
		}
		a.view.Bars[i] = int(h)
	}
}

// animateWaveform continuously steps the oscilloscope bars each tick.
func (a *App) animateWaveform() {
	phase := float64(a.tickCount) * 0.15
	for i := range a.view.Bars {
		var h float64
		if a.view.Playing || a.view.BeatLED {
			x := float64(i)/float64(scopeBarCount)*2*math.Pi + phase
			switch a.view.Preset {
			case "square":
				if math.Sin(x) >= 0 {
					h = 75 + 15*math.Sin(phase*1.5)
				} else {
					h = 25 + 10*math.Sin(phase*1.5)
				}
			case "sawtooth":
				mod := math.Mod(x, 2*math.Pi) / (2 * math.Pi)
				h = 20 + 70*mod
			case "triangle":
				mod := math.Mod(x, 2*math.Pi) / math.Pi
				if mod > 1.0 {
					mod = 2.0 - mod
				}
				h = 20 + 70*mod
			default: // sine
				h = 50 + 40*math.Sin(x*2)
			}
		} else {
			x := float64(i)/float64(scopeBarCount)*2*math.Pi + phase*0.5
			h = 14 + 6*math.Sin(x)
		}
		a.view.Bars[i] = int(h)
	}
}
