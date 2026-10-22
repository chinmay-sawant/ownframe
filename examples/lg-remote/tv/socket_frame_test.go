package tv

import (
	"bufio"
	"encoding/binary"
	"io"
)

func readSocketFrame(r *bufio.ReadWriter) ([]byte, error) {
	h := make([]byte, 2)
	if _, err := io.ReadFull(r, h); err != nil {
		return nil, err
	}
	if h[0]&15 == 8 {
		return nil, io.EOF
	}
	n := int(h[1] & 127)
	masked := h[1]&128 != 0
	if n == 126 {
		if _, err := io.ReadFull(r, h); err != nil {
			return nil, err
		}
		n = int(binary.BigEndian.Uint16(h))
	}
	mask := make([]byte, 4)
	if masked {
		if _, err := io.ReadFull(r, mask); err != nil {
			return nil, err
		}
	}
	b := make([]byte, n)
	if _, err := io.ReadFull(r, b); err != nil {
		return nil, err
	}
	for i := range b {
		b[i] ^= mask[i%4]
	}
	return b, nil
}

func writeSocketFrame(w io.Writer, b []byte) {
	h := []byte{0x81, byte(len(b))} //nolint:gosec // reason: test frames are short fixtures, length fits in one byte.
	if len(b) >= 126 {
		h = []byte{0x81, 126, byte(len(b) >> 8), byte(len(b))} //nolint:gosec // reason: test frames are short fixtures, extended length bytes hold the value.
	}
	_, _ = w.Write(append(h, b...))
}
