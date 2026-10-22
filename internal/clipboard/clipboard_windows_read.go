//go:build windows

package clipboard

import "unicode/utf16"

func readOS() (string, bool) {
	opened, _, _ := openClip.Call(0)
	if opened == 0 {
		return "", false
	}

	defer closeClip.Call()

	h, _, _ := getClip.Call(cfUnicode)
	if h == 0 {
		return "", false
	}

	ptr, _, _ := gLock.Call(h)
	if ptr == 0 {
		return "", false
	}

	n, _, _ := gSize.Call(h)
	if n < 2 {
		gUnlock.Call(h)

		return "", false
	}

	tmp := make([]uint16, int(n/2))
	if len(tmp) == 0 {
		gUnlock.Call(h)

		return "", false
	}

	getMem(ptr, uintptr(len(tmp))*utf16Unit, tmp)
	gUnlock.Call(h)

	end := 0

	for end < len(tmp) && tmp[end] != 0 {
		end++
	}

	return string(utf16.Decode(tmp[:end])), true
}
