package game

import (
	"crypto/rand"
	"encoding/hex"
)

// newID returns a fresh stable game ID from the system random source.
// It panics when that source fails instead of falling back to time.
func newID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}

	return hex.EncodeToString(b[:])
}
