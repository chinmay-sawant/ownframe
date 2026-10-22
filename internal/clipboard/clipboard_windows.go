//go:build windows

package clipboard

import (
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	cfUnicode = 13
	gmemMove  = 0x0002
)

var (
	user32   = windows.NewLazySystemDLL("user32.dll")
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")

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
	rtlMove = kernel32.NewProc("RtlMoveMemory")
)

// putMem copies Go bytes to a locked block without a uintptr to
// Pointer conversion, which go vet flags as unsafeptr.
func putMem(dst uintptr, src []byte) {
	if len(src) == 0 {
		return
	}
	rtlMove.Call(dst, uintptr(unsafe.Pointer(&src[0])), uintptr(len(src)))
	runtime.KeepAlive(src)
}

// getMem copies n bytes from a locked block into dst without a
// uintptr to Pointer conversion, which go vet flags as unsafeptr.
func getMem(src, n uintptr, dst []uint16) {
	if len(dst) == 0 || n == 0 {
		return
	}
	rtlMove.Call(uintptr(unsafe.Pointer(&dst[0])), src, n)
	runtime.KeepAlive(dst)
}

func writeOS(text string) {
	opened, _, _ := openClip.Call(0)
	if opened == 0 {
		return
	}

	defer closeClip.Call()
	emptyClip.Call()

	buf := utf16Clipboard(text)
	if len(buf) == 0 {
		return
	}
	h, _, _ := gAlloc.Call(gmemMove, uintptr(len(buf)))
	if h == 0 {
		return
	}

	ptr, _, _ := gLock.Call(h)
	if ptr == 0 {
		gFree.Call(h)

		return
	}

	putMem(ptr, buf)
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
