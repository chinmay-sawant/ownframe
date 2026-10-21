//go:build linux && !android

package clipboard

import "math"

func pad4(n int) int { return (n + xMask) &^ xMask }

func put16(b []byte, v uint16) {
	b[0] = byte(v)
	b[1] = byte(v >> xShift8)
}

func put32(b []byte, v uint32) {
	b[0] = byte(v)
	b[1] = byte(v >> xShift8)
	b[2] = byte(v >> xShift16)
	b[3] = byte(v >> xShift24)
}

func get16(b []byte) uint16 {
	return uint16(b[0]) | uint16(b[1])<<xShift8
}

func get32(b []byte) uint32 {
	return uint32(b[0]) | uint32(b[1])<<xShift8 |
		uint32(b[2])<<xShift16 | uint32(b[3])<<xShift24
}

func setupReq(name, data []byte) []byte {
	if len(name) > math.MaxUint16 {
		name = name[:math.MaxUint16]
	}

	if len(data) > math.MaxUint16 {
		data = data[:math.MaxUint16]
	}

	n := pad4(len(name))
	buf := make([]byte, 12+n+pad4(len(data)))
	buf[0] = 'l'
	put16(buf[2:], xProtoMajor)
	put16(buf[6:], uint16(len(name))) //nolint:gosec // reason: name clamped to MaxUint16 above.
	put16(buf[8:], uint16(len(data))) //nolint:gosec // reason: data clamped to MaxUint16 above.
	copy(buf[12:], name)
	copy(buf[12+n:], data)

	return buf
}

func setupIDs(buf []byte) (uint32, uint32, int, bool) {
	if len(buf) < 40 || buf[0] != 1 {
		return 0, 0, 0, false
	}

	max := int(get16(buf[26:])) * xWord

	return get32(buf[12:]), get32(buf[16:]), max, true
}

func rootAt(buf []byte, screen int) (uint32, bool) {
	if screen < 0 || len(buf) < 40 || buf[0] != 1 {
		return 0, false
	}

	off := (xHdr40 + int(get16(buf[24:])) + xMask) &^ xMask
	off += int(buf[29]) * xHdr8

	for s := 0; ; s++ {
		if off+40 > len(buf) {
			return 0, false
		}

		if s == screen {
			return get32(buf[off:]), true
		}

		nd := int(buf[off+39])
		off += 40

		for d := 0; d < nd; d++ {
			if off+8 > len(buf) {
				return 0, false
			}

			nv := int(get16(buf[off+2:]))
			off += xHdr8 + nv*xHdr24
		}
	}
}
