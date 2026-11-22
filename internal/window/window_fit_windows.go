//go:build windows

package window

import (
	"os"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	fitUser32        = windows.NewLazySystemDLL("user32.dll")
	fitEnumWindows   = fitUser32.NewProc("EnumWindows")
	fitWindowPID     = fitUser32.NewProc("GetWindowThreadProcessId")
	fitClassName     = fitUser32.NewProc("GetClassNameW")
	fitIsVisible     = fitUser32.NewProc("IsWindowVisible")
	fitIsWindow      = fitUser32.NewProc("IsWindow")
	fitIsZoomed      = fitUser32.NewProc("IsZoomed")
	fitIsIconic      = fitUser32.NewProc("IsIconic")
	fitGetWindowRect = fitUser32.NewProc("GetWindowRect")
	fitMonitorFrom   = fitUser32.NewProc("MonitorFromWindow")
	fitMonitorInfo   = fitUser32.NewProc("GetMonitorInfoW")
	fitSetWindowPos  = fitUser32.NewProc("SetWindowPos")
)

const (
	fitGLFWClass      = "GLFW30"
	fitMonitorNearest = 2 // MONITOR_DEFAULTTONEAREST
	fitSwpNoZOrder    = 0x0004
	fitSwpNoActivate  = 0x0010
	// fitSwpAsync queues the move on the window's own thread. A synchronous
	// SetWindowPos from inside Update fires the window's resize handlers
	// re-entrantly while Ebiten is mid-update, as the drag path found out.
	fitSwpAsync         = 0x4000
	fitMonitorInfoBytes = 40
)

type fitWinRect struct{ left, top, right, bottom int32 }

type fitMonitorInfoW struct {
	size    uint32
	monitor fitWinRect
	work    fitWinRect
	flags   uint32
}

var fitHandle struct {
	sync.Mutex
	hwnd uintptr
}

var (
	fitEnumMu     sync.Mutex
	fitEnumResult uintptr
	fitEnumProc   = syscall.NewCallback(func(hwnd, _ uintptr) uintptr {
		var pid uint32

		fitWindowPID.Call(hwnd, uintptr(unsafe.Pointer(&pid)))

		if pid != uint32(os.Getpid()) {
			return 1
		}

		if visible, _, _ := fitIsVisible.Call(hwnd); visible == 0 {
			return 1
		}

		var name [16]uint16

		n, _, _ := fitClassName.Call(hwnd, uintptr(unsafe.Pointer(&name[0])), uintptr(len(name)))
		if windows.UTF16ToString(name[:n]) != fitGLFWClass {
			return 1
		}

		fitEnumResult = hwnd

		return 0
	})
)

// fitWindow returns this process's visible Ebiten (GLFW) window, found once
// and looked up again if the handle goes stale.
func fitWindow() uintptr {
	fitHandle.Lock()
	defer fitHandle.Unlock()

	if fitHandle.hwnd != 0 {
		if alive, _, _ := fitIsWindow.Call(fitHandle.hwnd); alive != 0 {
			return fitHandle.hwnd
		}
	}

	fitEnumMu.Lock()
	fitEnumResult = 0
	fitEnumWindows.Call(fitEnumProc, 0)
	hwnd := fitEnumResult
	fitEnumMu.Unlock()

	fitHandle.hwnd = hwnd

	return hwnd
}

func toFitRect(r fitWinRect) fitRect {
	return fitRect{x: int(r.left), y: int(r.top), w: int(r.right - r.left), h: int(r.bottom - r.top)}
}

// nativeFitProbe reads the window rectangle and the work area of the monitor
// the window is mostly on, in desktop pixels. Ebiten makes the process
// per-monitor DPI aware, so both come back in the same physical units.
func nativeFitProbe() (fitProbe, bool) {
	hwnd := fitWindow()
	if hwnd == 0 {
		return fitProbe{}, false
	}

	var win fitWinRect
	if ok, _, _ := fitGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&win))); ok == 0 {
		return fitProbe{}, false
	}

	monitor, _, _ := fitMonitorFrom.Call(hwnd, fitMonitorNearest)
	if monitor == 0 {
		return fitProbe{}, false
	}

	info := fitMonitorInfoW{size: fitMonitorInfoBytes}
	if ok, _, _ := fitMonitorInfo.Call(monitor, uintptr(unsafe.Pointer(&info))); ok == 0 {
		return fitProbe{}, false
	}

	zoomed, _, _ := fitIsZoomed.Call(hwnd)
	iconic, _, _ := fitIsIconic.Call(hwnd)

	covers := win.left <= info.monitor.left && win.top <= info.monitor.top &&
		win.right >= info.monitor.right && win.bottom >= info.monitor.bottom

	return fitProbe{
		monitor: monitor,
		window:  toFitRect(win),
		work:    toFitRect(info.work),
		skip:    zoomed != 0 || iconic != 0 || covers,
	}, true
}

// nativeFitApply moves and resizes the window without changing its z-order or
// taking the focus.
func nativeFitApply(r fitRect) bool {
	hwnd := fitWindow()
	if hwnd == 0 {
		return false
	}

	ok, _, _ := fitSetWindowPos.Call(hwnd, 0,
		uintptr(int32(r.x)), uintptr(int32(r.y)), uintptr(int32(r.w)), uintptr(int32(r.h)),
		fitSwpNoZOrder|fitSwpNoActivate|fitSwpAsync)

	return ok != 0
}
