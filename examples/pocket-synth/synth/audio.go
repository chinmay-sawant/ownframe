package synth

import (
	"math"

	"github.com/chinmay-sawant/ownframe/examples/music"
)

// noteOffsets maps chromatic note names to semitone offsets from C.
var noteOffsets = map[string]int{
	"C": 0, "C#": 1, "CS": 1, "DB": 1,
	"D": 2, "D#": 3, "DS": 3, "EB": 3,
	"E": 4,
	"F": 5, "F#": 6, "FS": 6, "GB": 6,
	"G": 7, "G#": 8, "GS": 8, "AB": 8,
	"A": 9, "A#": 10, "AS": 10, "BB": 10,
	"B": 11,
}

// noteFrequency returns the frequency in Hz for a note and octave.
func noteFrequency(note string, octave int) float64 {
	offset, ok := noteOffsets[note]
	if !ok {
		offset = 0
	}
	semitones := 12*(octave-4) + (offset - 9)
	return 440.0 * math.Pow(2.0, float64(semitones)/12.0)
}

// AudioHost manages note synthesis and playback through music.Voice.
type AudioHost struct {
	Open        func(data []byte) (music.Voice, error)
	LastNote    string
	LastWave    string
	PlayCount   int
	activeVoice music.Voice
}

// NewAudioHost returns an initialized AudioHost.
func NewAudioHost() *AudioHost {
	return &AudioHost{
		Open: func(data []byte) (music.Voice, error) {
			return music.NewSound(data)
		},
	}
}

// PlayNote synthesizes and plays a note. In headless environments without
// an audio device (e.g. tests or -web mode), it falls back gracefully.
func (h *AudioHost) PlayNote(note string, octave int, wave string) {
	h.LastNote = note
	h.LastWave = wave
	h.PlayCount++

	freq := noteFrequency(note, octave)
	pcm := synthPCM(freq, wave, sampleRate, noteDuration)
	wavBytes := encodeMonoWAV(pcm, sampleRate)

	if h.Open == nil {
		return
	}
	voice, err := h.Open(wavBytes)
	if err != nil {
		return
	}
	if h.activeVoice != nil {
		_ = h.activeVoice.Close()
	}
	h.activeVoice = voice
	voice.Play()
}

// Close releases any active audio voice.
func (h *AudioHost) Close() {
	if h.activeVoice != nil {
		_ = h.activeVoice.Close()
		h.activeVoice = nil
	}
}
