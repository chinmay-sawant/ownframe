package synth

const (
	// DefaultWidth and DefaultHeight are the initial window dimensions.
	DefaultWidth  = 540
	DefaultHeight = 680

	// MinWidth and MinHeight are the minimum allowed window dimensions.
	MinWidth  = 360
	MinHeight = 480

	// Number of oscilloscope waveform display bars.
	scopeBarCount = 16
)

// View holds the current display state of the synthesizer.
type View struct {
	ActiveNote string
	Octave     int
	Preset     string
	PresetName string
	BPM        int
	Playing    bool
	BeatLED    bool
	Status     string
	Bars       []int
}

// initialView returns the default synthesizer state.
func initialView() View {
	bars := make([]int, scopeBarCount)
	for i := range bars {
		bars[i] = 12
	}
	return View{
		ActiveNote: "--",
		Octave:     4,
		Preset:     "square",
		PresetName: "8-BIT SQUARE",
		BPM:        120,
		Playing:    false,
		BeatLED:    false,
		Status:     "READY",
		Bars:       bars,
	}
}
