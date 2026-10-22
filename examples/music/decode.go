package music

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

// sniffLen is the longest magic decode checks, the RIFF prefix.
const sniffLen = 4

// decode sniffs data and returns 16-bit stereo PCM at sampleRate with the
// decoded length in bytes.
func decode(data []byte, sampleRate int) (io.ReadSeeker, int64, error) {
	if len(data) < sniffLen {
		return nil, 0, errors.New("music: audio too short")
	}

	switch {
	case bytes.HasPrefix(data, []byte("RIFF")):
		stream, err := wav.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
		if err != nil {
			return nil, 0, err
		}

		return stream, stream.Length(), nil
	case bytes.HasPrefix(data, []byte("ID3")) || isMP3Frame(data):
		stream, err := mp3.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
		if err != nil {
			return nil, 0, err
		}

		return stream, stream.Length(), nil
	default:
		return nil, 0, fmt.Errorf("music: unknown audio format")
	}
}

// isMP3Frame reports an MPEG audio frame sync at the head of data.
func isMP3Frame(data []byte) bool {
	return len(data) >= 2 && data[0] == 0xff && data[1]&0xe0 == 0xe0
}
