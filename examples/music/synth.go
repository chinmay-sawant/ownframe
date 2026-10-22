package music

import (
	"hash/fnv"
	"math"
)

// Demo tune constants: a short generated WAV so playback works offline.
const (
	demoRate    = 48000
	demoSeconds = 16
	demoStep    = 0.32 // seconds per arpeggio step

	semitones  = 12   // pitch classes per octave
	pitchBase  = 2    // frequency ratio of one octave
	tuneBaseHz = 220  // A3, the arpeggio root
	tuneAmp    = 0.35 // peak amplitude keeps headroom
	twoPi      = 2 * math.Pi
)

// DemoTune returns a Clip holding a generated tune. The query picks the key,
// so different tracks sound different.
func DemoTune(query string) Clip {
	return Clip{
		Title:   "Demo tone",
		Creator: "ownframe",
		License: "generated",
		Data:    wavTune(seed(query)),
	}
}

// seed folds a query into a small deterministic key offset.
func seed(query string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(query))

	return int(h.Sum32() % semitones)
}

// wavTune renders an arpeggio over a four-chord loop to 16-bit stereo PCM.
func wavTune(root int) []byte {
	frames := demoRate * demoSeconds
	pcm := make([]int16, frames*wavChannels)
	step := int(demoStep * demoRate)
	chords := [4][3]int{{0, 4, 7}, {-4, 0, 3}, {-1, 2, 7}, {-3, 2, 5}}

	for i := 0; i < frames; i++ {
		chord := chords[i/step%4]
		note := chord[i/step%3]
		t := float64(i%demoRate) / demoRate
		env := math.Exp(-3 * t)
		freq := tuneBaseHz * math.Pow(pitchBase, float64(root+note)/semitones)
		v := math.Sin(twoPi*freq*float64(i)/demoRate) * env * tuneAmp
		s := int16(v * math.MaxInt16)
		pcm[i*2] = s
		pcm[i*2+1] = s
	}

	return wavBytes(pcm)
}
