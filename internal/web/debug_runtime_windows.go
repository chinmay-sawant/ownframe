//go:build windows

package web

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// System32 only (LOAD_LIBRARY_SEARCH_SYSTEM32): a planted DLL in the app
// dir or on PATH cannot win the load.
var (
	modPsapiWeb                 = windows.NewLazySystemDLL("psapi.dll")
	procGetProcessMemoryInfoWeb = modPsapiWeb.NewProc("GetProcessMemoryInfo")
)

// webProcessMemoryCounters matches PROCESS_MEMORY_COUNTERS. Only the
// working set field is read; the rest keeps the struct size correct for
// the API call.
type webProcessMemoryCounters struct {
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

func runtimeRSS() (uint64, bool) {
	var counters webProcessMemoryCounters
	counters.cb = uint32(unsafe.Sizeof(counters))
	r, _, _ := procGetProcessMemoryInfoWeb.Call(
		uintptr(windows.CurrentProcess()),
		uintptr(unsafe.Pointer(&counters)),
		uintptr(counters.cb),
	)
	if r == 0 {
		return 0, false
	}
	if counters.workingSetSize == 0 {
		return 0, false
	}
	return uint64(counters.workingSetSize), true
}
