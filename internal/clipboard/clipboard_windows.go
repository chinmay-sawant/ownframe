//go:build windows

package clipboard

import (
	"syscall"
	"unicode/utf16"
	"unsafe"
)

const (
	cfUnicode = 13
	gmemMove  = 0x0002
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	openClip  = user32.NewProc("OpenClipboard")
	emptyClip = user32.NewProc("EmptyClipboard")
	setClip   = user32.NewProc("SetClipboardData")
	getClip   = user32.NewProc("GetClipboardData")
	closeClip = user32.NewProc("CloseClipboard")

	gAlloc  = kernel32.NewProc("GlobalAlloc")
	gLock   = kernel32.NewProc("GlobalLock")
	gUnlock = kernel32.NewProc("GlobalUnlock")
	gFree   = kernel32.NewProc("GlobalFree")
	gSize   = kernel32.NewProc("GlobalSize")
)

func writeOS(text string) {
	opened, _, _ := openClip.Call(0)
	if opened == 0 {
		return
	}

	defer closeClip.Call()
	emptyClip.Call()

	buf := utf16Clipboard(text)
	h, _, _ := gAlloc.Call(gmemMove, uintptr(len(buf)))
	if h == 0 {
		return
	}

	ptr, _, _ := gLock.Call(h)
	if ptr == 0 {
		gFree.Call(h)

		return
	}

	dst := unsafe.Slice((*byte)(unsafe.Pointer(ptr)), len(buf))
	copy(dst, buf)
	gUnlock.Call(h)

	set, _, _ := setClip.Call(cfUnicode, h)
	if set == 0 {
		gFree.Call(h)
	}
}

// clearOS empties the desktop clipboard.
func clearOS() {
	opened, _, _ := openClip.Call(0)
	if opened == 0 {
		return
	}

	defer closeClip.Call()
	emptyClip.Call()
}

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

	raw := unsafe.Slice((*uint16)(unsafe.Pointer(ptr)), int(n/2))
	end := 0

	for end < len(raw) && raw[end] != 0 {
		end++
	}

	u := append([]uint16(nil), raw[:end]...)
	gUnlock.Call(h)

	return string(utf16.Decode(u)), true
}
