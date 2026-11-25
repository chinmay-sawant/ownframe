//go:build windows

package window

import (
	"unsafe"
)

var (
	hzMonitorFromRect     = fitUser32.NewProc("MonitorFromRect")
	hzEnumDisplaySettings = fitUser32.NewProc("EnumDisplaySettingsW")
)

const (
	cchDeviceName              = 32
	cchFormName                = 32
	enumCurrentSettings uint32 = 0xFFFFFFFF
)

// winMonitorInfoExW mirrors Win32 MONITORINFOEXW (104 bytes).
type winMonitorInfoExW struct {
	size   uint32
	_      fitWinRect // rcMonitor
	_      fitWinRect // rcWork
	_      uint32     // dwFlags
	device [cchDeviceName]uint16
}

// winDevModeW mirrors Win32 DEVMODEW (220 bytes) with exact field offsets for
// dmSize (68) and dmDisplayFrequency (184).
type winDevModeW struct {
	_                [cchDeviceName]uint16 // 0..64: dmDeviceName
	_                uint16                // 64..66: dmSpecVersion
	_                uint16                // 66..68: dmDriverVersion
	size             uint16                // 68..70: dmSize
	_                uint16                // 70..72: dmDriverExtra
	_                uint32                // 72..76: dmFields
	_                [4]uint32             // 76..92: dmPosition/dmDisplayOrientation/dmDisplayFixedOutput
	_                [5]int16              // 92..102: dmColor..dmCollate
	_                [cchFormName]uint16   // 102..166: dmFormName
	_                uint16                // 166..168: dmLogPixels
	_                uint32                // 168..172: dmBitsPerPel
	_                uint32                // 172..176: dmPelsWidth
	_                uint32                // 176..180: dmPelsHeight
	_                uint32                // 180..184: dmDisplayFlags
	displayFrequency uint32                // 184..188: dmDisplayFrequency
	_                [8]uint32             // 188..220: dmICMMethod..dmPanningHeight
}

// Compile-time layout assertions for Win32 MONITORINFOEXW and DEVMODEW.
var (
	_ [104 - unsafe.Sizeof(winMonitorInfoExW{})]byte
	_ [unsafe.Sizeof(winMonitorInfoExW{}) - 104]byte
	_ [220 - unsafe.Sizeof(winDevModeW{})]byte
	_ [unsafe.Sizeof(winDevModeW{}) - 220]byte
	_ [68 - unsafe.Offsetof(winDevModeW{}.size)]byte
	_ [unsafe.Offsetof(winDevModeW{}.size) - 68]byte
	_ [184 - unsafe.Offsetof(winDevModeW{}.displayFrequency)]byte
	_ [unsafe.Offsetof(winDevModeW{}.displayFrequency) - 184]byte
)

// detectMonitorHz queries the refresh rate (Hz) of the Windows monitor
// currently hosting this process's window. It prefers MonitorFromWindow on the
// live GLFW HWND so mixed-DPI multi-monitor setups match exact physical
// monitor bounds, and falls back to MonitorFromRect at the window center when
// the HWND is not yet available.
func detectMonitorHz(posX, posY, winW, winH int) int {
	var monitor uintptr
	if hwnd := fitWindow(); hwnd != 0 {
		monitor, _, _ = fitMonitorFrom.Call(hwnd, fitMonitorNearest)
	}
	if monitor == 0 {
		cx := int32(posX + winW/2)
		cy := int32(posY + winH/2)
		rc := fitWinRect{left: cx, top: cy, right: cx + 1, bottom: cy + 1}
		monitor, _, _ = hzMonitorFromRect.Call(uintptr(unsafe.Pointer(&rc)), fitMonitorNearest)
	}
	if monitor == 0 {
		return defaultMonitorHz
	}

	var info winMonitorInfoExW
	info.size = uint32(unsafe.Sizeof(info))
	if ok, _, _ := fitMonitorInfo.Call(monitor, uintptr(unsafe.Pointer(&info))); ok == 0 {
		return defaultMonitorHz
	}

	var dm winDevModeW
	dm.size = uint16(unsafe.Sizeof(dm))
	ok, _, _ := hzEnumDisplaySettings.Call(
		uintptr(unsafe.Pointer(&info.device[0])),
		uintptr(enumCurrentSettings),
		uintptr(unsafe.Pointer(&dm)),
	)
	if ok == 0 {
		return defaultMonitorHz
	}
	return sanitizeMonitorHz(int(dm.displayFrequency))
}
