//go:build windows && (amd64 || arm64)

package uia

import (
	"math"
	"syscall"
)

// ElementProviderFromPoint(this, x, y, out) takes two doubles. The Windows x64
// and arm64 conventions pass them in floating-point registers, which a Go
// syscall callback never reads. hit_windows_amd64.s and hit_windows_arm64.s
// hold a thunk that copies the bits of x and y into the integer registers of
// the same argument positions, and then jumps to fromPointCallback, so the Go
// side receives (this, bits(x), bits(y), out).

// fromPointCallback is the address of fromPointEntry. The thunks read it.
var fromPointCallback = syscall.NewCallback(fromPointEntry)

// hitThunkAddr returns the address of the assembly thunk for the vtable.
func hitThunkAddr() uintptr

func fromPointEntry(this, xBits, yBits, out uintptr) (ret uintptr) {
	defer func() {
		if recover() != nil {
			ret = hr(eFail)
		}
	}()

	o, _, ok := resolve(this)
	if !ok {
		return hr(eElementNoAvl)
	}

	return o.fromPointAt(math.Float64frombits(uint64(xBits)), math.Float64frombits(uint64(yBits)), out)
}
