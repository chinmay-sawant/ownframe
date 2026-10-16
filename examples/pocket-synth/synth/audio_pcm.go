package synth

import (
	"math"
)

// synthPCM produces 16-bit mono PCM samples for a frequency and waveform.
func synthPCM(freq float64, wave string, rate int, dur float64) []int16 {
	totalFrames := int(float64(rate) * dur)
	pcm := make([]int16, totalFrames)
	phase := 0.0
	phaseInc := 2.0 * math.Pi * freq / float64(rate)
	attackFrames := int(0.008 * float64(rate))

	for i := 0; i < totalFrames; i++ {
		t := float64(i) / float64(rate)
		env := math.Exp(-5.0 * t)
		if i < attackFrames {
			env *= float64(i) / float64(attackFrames)
		}

		var val float64
		normPhase := math.Mod(phase, 2*math.Pi)
		switch wave {
		case "square":
			if normPhase < math.Pi {
				val = 0.32
			} else {
				val = -0.32
			}
		case "sawtooth":
			val = 0.32 * (1.0 - normPhase/math.Pi)
		case "triangle":
			norm := normPhase / math.Pi
			if norm > 1.0 {
				norm = 2.0 - norm
			}
			val = 0.35 * (2.0*norm - 1.0)
		default: // sine
			val = 0.35 * math.Sin(normPhase)
		}

		pcm[i] = int16(val * env * math.MaxInt16)
		phase += phaseInc
	}
	return pcm
}
