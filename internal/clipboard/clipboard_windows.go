//go:build windows

package clipboard

import (
	"context"
	"runtime"
	"time"
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

const (
	openAttempts = 4
	openBackoff  = 5 * time.Millisecond
)

// openClipboard opens the clipboard for the calling OS thread, retrying a
// few times because another process may hold it briefly. The caller must
// have called runtime.LockOSThread: Win32 binds clipboard ownership to the
// thread, so CloseClipboard has to run on the same one.
func openClipboard() bool {
	for i := range openAttempts {
		if opened, _, _ := openClip.Call(0); opened != 0 {
			return true
		}
		if i < openAttempts-1 {
			time.Sleep(openBackoff << i)
		}
	}

	return false
}

func writeOS(_ context.Context, text string) bool {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if !openClipboard() {
		return false
	}

	defer closeClip.Call()
	emptyClip.Call()

	buf := utf16Clipboard(text)
	if len(buf) == 0 {
		return false
	}
	h, _, _ := gAlloc.Call(gmemMove, uintptr(len(buf)))
	if h == 0 {
		return false
	}

	ptr, _, _ := gLock.Call(h)
	if ptr == 0 {
		gFree.Call(h)

		return false
	}

	putMem(ptr, buf)
	gUnlock.Call(h)

	set, _, _ := setClip.Call(cfUnicode, h)
	if set == 0 {
		gFree.Call(h)
	}
	return set != 0
}

// clearOS empties the desktop clipboard.
func clearOS() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if !openClipboard() {
		return
	}

	defer closeClip.Call()
	emptyClip.Call()
}

func diagnoseOS() Status { return Status{Backend: "windows", Available: true} }
