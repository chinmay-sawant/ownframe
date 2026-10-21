//go:build linux && !android

package clipboard

import "math"

func internReq(name string) []byte {
	if len(name) > math.MaxUint16 {
		name = name[:math.MaxUint16]
	}

	size := pad4(xHdr8 + len(name))
	buf := make([]byte, size)
	buf[0] = xOpIntern
	put16(buf[2:], uint16(size/xWord)) //nolint:gosec // reason: name clamped so size/4 fits uint16.
	put16(buf[4:], uint16(len(name)))  //nolint:gosec // reason: name clamped to MaxUint16 above.
	copy(buf[8:], name)

	return buf
}

func winReq(id, parent uint32) []byte {
	buf := make([]byte, xEvSize)
	buf[0] = xOpCreate
	put16(buf[2:], xWordsCreate)
	put32(buf[4:], id)
	put32(buf[8:], parent)
	put16(buf[16:], 1)
	put16(buf[18:], 1)
	put16(buf[22:], 1)

	return buf
}

func ownReq(owner, sel uint32) []byte {
	buf := make([]byte, xSizeSetOwner)
	buf[0] = xOpSetOwner
	put16(buf[2:], xWordsSetOwner)
	put32(buf[4:], owner)
	put32(buf[8:], sel)

	return buf
}

func convReq(win, sel, target, prop uint32) []byte {
	buf := make([]byte, xSizeConvert)
	buf[0] = xOpConvert
	put16(buf[2:], xWordsConvert)
	put32(buf[4:], win)
	put32(buf[8:], sel)
	put32(buf[12:], target)
	put32(buf[16:], prop)

	return buf
}

func ownerReq(sel uint32) []byte {
	buf := make([]byte, xSizeGetOwner)
	buf[0] = xOpGetOwner
	put16(buf[2:], xWordsGetOwner)
	put32(buf[4:], sel)

	return buf
}

func getReq(win, prop uint32) []byte {
	buf := make([]byte, xSizeConvert)
	buf[0] = xOpGetProp
	buf[1] = 1
	put16(buf[2:], xWordsConvert)
	put32(buf[4:], win)
	put32(buf[8:], prop)
	put32(buf[20:], xMaxWords)

	return buf
}
