//go:build windows

#include "textflag.h"

// IRawElementProviderFragmentRoot::ElementProviderFromPoint(this, x, y, out)
// arrives with this in R0, out in R1, and the doubles x and y in F0 and F1:
// the integer and floating-point argument registers count separately. The Go
// callback reads integer registers only, so out moves to R3, and the bits of
// x and y to R1 and R2, which makes the callback's argument list
// (this, x bits, y bits, out). The thunk uses only registers a call may
// change, and the load of the callback's address uses R9 for that reason.
TEXT ·hitThunk(SB), NOSPLIT|NOFRAME, $0-0
	MOVD  R1, R3
	FMOVD F0, R1
	FMOVD F1, R2
	MOVD  $·fromPointCallback(SB), R9
	MOVD  (R9), R9
	B     (R9)

// func hitThunkAddr() uintptr
TEXT ·hitThunkAddr(SB), NOSPLIT, $0-8
	MOVD $·hitThunk(SB), R0
	MOVD R0, ret+0(FP)
	RET
