package music

import (
	"encoding/binary"
	"math"
)

// Canonical 16-bit stereo RIFF/WAVE header layout.
const (
	wavHeaderLen   = 44
	wavRIFFBase    = 36 // header length minus the 8 RIFF bytes
	wavFmtLen      = 16
	wavFormatPCM   = 1
	wavChannels    = 2
	wavFrameBytes  = 4 // channels times 16 bits over 8
	wavBits        = 16
	wavSampleBytes = 2
)

// maxWAVSamples caps wavBytes input so the RIFF sizes fit uint32.
const maxWAVSamples = (math.MaxUint32 - wavHeaderLen) / wavSampleBytes

// wavBytes wraps 16-bit stereo PCM in a canonical RIFF/WAVE header.
func wavBytes(pcm []int16) []byte {
	if len(pcm) > maxWAVSamples {
		return nil
	}
	data := make([]byte, wavHeaderLen+len(pcm)*wavSampleBytes)
	copy(data, "RIFF")
	binary.LittleEndian.PutUint32(data[4:], uint32(wavRIFFBase+len(pcm)*wavSampleBytes)) //nolint:gosec // reason: len(pcm) capped at maxWAVSamples above, sum fits uint32.
	copy(data[8:], "WAVEfmt ")
	binary.LittleEndian.PutUint32(data[16:], wavFmtLen)
	binary.LittleEndian.PutUint16(data[20:], wavFormatPCM)
	binary.LittleEndian.PutUint16(data[22:], wavChannels)
	binary.LittleEndian.PutUint32(data[24:], demoRate)
	binary.LittleEndian.PutUint32(data[28:], demoRate*wavFrameBytes)
	binary.LittleEndian.PutUint16(data[32:], wavFrameBytes)
	binary.LittleEndian.PutUint16(data[34:], wavBits)
	copy(data[36:], "data")
	binary.LittleEndian.PutUint32(data[40:], uint32(len(pcm)*wavSampleBytes)) //nolint:gosec // reason: len(pcm) capped at maxWAVSamples above, product fits uint32.

	for i, s := range pcm {
		binary.LittleEndian.PutUint16(data[44+i*2:], uint16(s)) //nolint:gosec // reason: int16 to uint16 keeps the PCM two's-complement bits; negative samples must wrap.
	}

	return data
}
