package ws

import (
	"encoding/binary"
	"io"
)

func frameLen(r io.Reader, n int) (int, error) {
	switch n {
	case extLen16:
		var b [2]byte
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return 0, err
		}

		return int(binary.BigEndian.Uint16(b[:])), nil
	case extLen64:
		var b [8]byte
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return 0, err
		}

		v := binary.BigEndian.Uint64(b[:])
		if v > uint64(maxFrameLen) {
			return 0, ErrTooLarge
		}

		return int(v), nil
	default:
		return n, nil
	}
}
