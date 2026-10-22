//go:build linux && !android

package clipboard

import "math"

// sliceCap clamps counts to a bound valid on every arch: on 386 and
// arm int is 32 bits, so math.MaxUint32 is not a slice bound (GEN-25).
const sliceCap = math.MaxInt32

func prop32(win, prop, typ uint32, vals []uint32) []byte {
	if uint64(len(vals)) > uint64(math.MaxUint32) {
		vals = vals[:sliceCap] //nolint:gosec // reason: sliceCap is MaxInt32, always a valid bound.
	}

	size := xHdr24 + xWord*len(vals)
	buf := make([]byte, size)
	buf[0] = xOpChange
	put16(buf[2:], uint16(size/xWord)) //nolint:gosec // reason: vals tiny, size fits uint16.
	put32(buf[4:], win)
	put32(buf[8:], prop)
	put32(buf[12:], typ)
	buf[16] = xFmt32
	put32(buf[20:], uint32(len(vals))) //nolint:gosec // reason: vals clamped to MaxUint32 above.

	for i, v := range vals {
		put32(buf[xHdr24+xWord*i:], v)
	}

	return buf
}

func prop8(win, prop, typ uint32, data []byte) []byte {
	if uint64(len(data)) > uint64(math.MaxUint32) {
		data = data[:sliceCap]
	}

	size := pad4(xHdr24 + len(data))
	buf := make([]byte, size)
	buf[0] = xOpChange
	put16(buf[2:], uint16(size/xWord)) //nolint:gosec // reason: clipboard bounded by server max, fits uint16.
	put32(buf[4:], win)
	put32(buf[8:], prop)
	put32(buf[12:], typ)
	buf[16] = xFmt8
	put32(buf[20:], uint32(len(data))) //nolint:gosec // reason: data clamped to MaxUint32 above.
	copy(buf[24:], data)

	return buf
}

func selEvent(when, who, sel, target, prop uint32) []byte {
	ev := make([]byte, xEvSize)
	ev[0] = xSelNotify
	put32(ev[4:], when)
	put32(ev[8:], who)
	put32(ev[12:], sel)
	put32(ev[16:], target)
	put32(ev[20:], prop)

	return ev
}

func sendEv(dst uint32, ev []byte) []byte {
	buf := make([]byte, xSizeSend)
	buf[0] = xOpSend
	put16(buf[2:], xWordsSend)
	put32(buf[4:], dst)
	copy(buf[12:], ev[:xEvSize])

	return buf
}
