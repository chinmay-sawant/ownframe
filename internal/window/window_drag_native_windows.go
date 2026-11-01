package window

import (
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	dragUser32     = windows.NewLazySystemDLL("user32.dll")
	dragForeground = dragUser32.NewProc("GetForegroundWindow")
	dragProcess    = dragUser32.NewProc("GetWindowThreadProcessId")
	dragCursor     = dragUser32.NewProc("GetCursorPos")
	dragRect       = dragUser32.NewProc("GetWindowRect")
	dragPosition   = dragUser32.NewProc("SetWindowPos")
)

type nativeWindowDrag struct {
	hwnd             uintptr
	offsetX, offsetY int32
}

type dragPoint struct{ x, y int32 }
type dragRectangle struct{ left, top, right, bottom int32 }

func (d *nativeWindowDrag) begin() {
	*d = nativeWindowDrag{}
	hwnd, _, _ := dragForeground.Call()
	var pid uint32
	dragProcess.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	if pid != uint32(os.Getpid()) {
		return
	}
	var p dragPoint
	var r dragRectangle
	ok, _, _ := dragCursor.Call(uintptr(unsafe.Pointer(&p)))
	if ok == 0 {
		return
	}
	ok, _, _ = dragRect.Call(hwnd, uintptr(unsafe.Pointer(&r)))
	if ok == 0 {
		return
	}
	d.hwnd = hwnd
	d.offsetX, d.offsetY = p.x-r.left, p.y-r.top
}

// Both Win32 calls use desktop pixels. Window-relative cursor samples change
// when our own move lands, and cannot safely drive a cross-monitor drag.
func (d *nativeWindowDrag) move() (bool, error) {
	if d.hwnd == 0 {
		return false, nil
	}
	var p dragPoint
	ok, _, err := dragCursor.Call(uintptr(unsafe.Pointer(&p)))
	if ok == 0 {
		return true, fmt.Errorf("read drag cursor: %w", err)
	}
	const flags = 0x0001 | 0x0004 | 0x0010 | 0x4000 // NOSIZE | NOZORDER | NOACTIVATE | ASYNCWINDOWPOS
	ok, _, err = dragPosition.Call(d.hwnd, 0, uintptr(p.x-d.offsetX), uintptr(p.y-d.offsetY), 0, 0, flags)
	if ok == 0 {
		return true, fmt.Errorf("move dragged window: %w", err)
	}
	return true, nil
}

func (d *nativeWindowDrag) end() { *d = nativeWindowDrag{} }

func (*nativeWindowDrag) managed() bool { return false }

func setNativeWindowRect(x, y, w, h int) bool { return false }

func nativeWindowPosition() (int, int, bool) { return 0, 0, false }
