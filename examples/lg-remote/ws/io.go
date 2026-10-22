package ws

import (
	"crypto/rand"
	"encoding/binary"
	"io"
)

func writeFrame(w io.Writer, op byte, p []byte, mask bool) error {
	hdr := []byte{finBit | op}
	n := len(p)

	switch {
	case n < extLen16:
		hdr = append(hdr, byte(n))
	case n < maxShortLen:
		hdr = append(hdr, extLen16, byte(n>>byteShift), byte(n)) //nolint:gosec // reason: n is range-checked below 65536, so the two length bytes hold the value.
	default:
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], uint64(n))
		hdr = append(hdr, extLen64)
		hdr = append(hdr, b[:]...)
	}

	var m [4]byte
	if mask {
		hdr[1] |= maskBit
		if _, err := rand.Read(m[:]); err != nil {
			return err
		}
		hdr = append(hdr, m[:]...)
	}

	if _, err := w.Write(hdr); err != nil {
		return err
	}

	if !mask {
		_, err := w.Write(p)
		return err
	}

	out := make([]byte, len(p))
	for i := range p {
		out[i] = p[i] ^ m[i%4]
	}

	_, err := w.Write(out)

	return err
}

func readFrame(r io.Reader) (bool, byte, []byte, error) {
	var h [2]byte
	if _, err := io.ReadFull(r, h[:]); err != nil {
		return false, 0, nil, err
	}

	n, err := frameLen(r, int(h[1]&payloadMask))
	if err != nil {
		return false, 0, nil, err
	}

	var mask [4]byte
	masked := h[1]&maskBit != 0
	if masked {
		if _, err = io.ReadFull(r, mask[:]); err != nil {
			return false, 0, nil, err
		}
	}

	p := make([]byte, n)
	if _, err = io.ReadFull(r, p); err != nil {
		return false, 0, nil, err
	}

	if masked {
		for i := range p {
			p[i] ^= mask[i%4]
		}
	}

	return h[0]&finBit != 0, h[0] & opcodeMask, p, nil
}
