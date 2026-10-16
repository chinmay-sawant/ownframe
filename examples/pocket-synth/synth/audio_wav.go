package synth

import "encoding/binary"

const (
	sampleRate   = 48000
	noteDuration = 0.35
)

// encodeMonoWAV wraps 16-bit mono PCM into a canonical RIFF/WAVE container.
func encodeMonoWAV(pcm []int16, rate int) []byte {
	dataLen := len(pcm) * 2
	out := make([]byte, 44+dataLen)

	copy(out[0:4], "RIFF")
	binary.LittleEndian.PutUint32(out[4:8], uint32(36+dataLen))
	copy(out[8:16], "WAVEfmt ")
	binary.LittleEndian.PutUint32(out[16:20], 16)
	binary.LittleEndian.PutUint16(out[20:22], 1) // PCM format
	binary.LittleEndian.PutUint16(out[22:24], 1) // mono
	binary.LittleEndian.PutUint32(out[24:28], uint32(rate))
	binary.LittleEndian.PutUint32(out[28:32], uint32(rate*2)) // byte rate
	binary.LittleEndian.PutUint16(out[32:34], 2)              // block align
	binary.LittleEndian.PutUint16(out[34:36], 16)             // bits per sample
	copy(out[36:40], "data")
	binary.LittleEndian.PutUint32(out[40:44], uint32(dataLen))

	for i, s := range pcm {
		binary.LittleEndian.PutUint16(out[44+i*2:], uint16(s))
	}
	return out
}
