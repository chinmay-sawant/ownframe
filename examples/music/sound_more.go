package music

import "time"

// Seek moves to an absolute position inside the stream.
func (s *Sound) Seek(at time.Duration) error {
	if at < 0 {
		at = 0
	}

	if at > s.dur {
		at = s.dur
	}

	return s.player.SetPosition(at)
}

// SetVolume sets the stream volume from 0 to 1.
func (s *Sound) SetVolume(volume float64) { s.player.SetVolume(clampVolume(volume)) }

// Close releases the stream. The player stops reading the source, so the
// decoded bytes can be dropped without a gap on the next Play.
func (s *Sound) Close() error {
	s.player.PauseAndStopReading()
	return nil
}

// clampVolume keeps a volume inside the range Ebiten accepts.
func clampVolume(volume float64) float64 {
	if volume < 0 {
		return 0
	}

	if volume > 1 {
		return 1
	}

	return volume
}
