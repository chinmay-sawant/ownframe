//go:build windows

#include "textflag.h"

// IRawElementProviderFragmentRoot::ElementProviderFromPoint(this, x, y, out)
// arrives with this in CX, x in X1, y in X2, and out in R9. The Go callback
// reads integer registers only, so the doubles' bits move to DX and R8, the
// registers of the same argument positions, and the thunk goes on to it. CX
// and R9 are already where the callback expects them, and the return address
// is on the stack, so the jump is a tail call.
TEXT ·hitThunk(SB), NOSPLIT|NOFRAME, $0-0
	MOVQ X1, DX
	MOVQ X2, R8
	MOVQ ·fromPointCallback(SB), AX
	JMP  AX

// func hitThunkAddr() uintptr
TEXT ·hitThunkAddr(SB), NOSPLIT, $0-8
	LEAQ ·hitThunk(SB), AX
	MOVQ AX, ret+0(FP)
	RET
