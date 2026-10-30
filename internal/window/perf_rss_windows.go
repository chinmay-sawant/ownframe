//go:build windows

package window

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// System32 only (LOAD_LIBRARY_SEARCH_SYSTEM32): a planted DLL in the app
// dir or on PATH cannot win the load.
var (
	modPsapiWindows                = windows.NewLazySystemDLL("psapi.dll")
	procGetProcessMemoryInfoWindow = modPsapiWindows.NewProc("GetProcessMemoryInfo")
)

// windowsProcessMemoryCounters matches PROCESS_MEMORY_COUNTERS. Only the
// working set field is read; the rest keeps the struct size correct for
// the API call.
type windowsProcessMemoryCounters struct {
	cb                         uint32
	pageFaultCount             uint32
	peakWorkingSetSize         uintptr
	workingSetSize             uintptr
	quotaPeakPagedPoolUsage    uintptr
	quotaPagedPoolUsage        uintptr
	quotaPeakNonPagedPoolUsage uintptr
	quotaNonPagedPoolUsage     uintptr
	pagefileUsage              uintptr
	peakPagefileUsage          uintptr
}

// perfRSS reads the current working set via GetProcessMemoryInfo on the
// pseudo-handle, so no OpenProcess or CloseHandle runs per sample. The
// sampler already gates this to about twice a second. It returns 0 when
// the call fails; the caller renders that as unknown, since genuine
// resident memory is never zero.
func perfRSS() uint64 {
	var counters windowsProcessMemoryCounters
	counters.cb = uint32(unsafe.Sizeof(counters))
	r, _, _ := procGetProcessMemoryInfoWindow.Call(
		uintptr(windows.CurrentProcess()),
		uintptr(unsafe.Pointer(&counters)),
		uintptr(counters.cb),
	)
	if r == 0 {
		return 0
	}
	return uint64(counters.workingSetSize)
}
