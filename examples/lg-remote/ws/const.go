package ws

import "time"

// WebSocket frame constants (RFC 6455 opcodes and length markers).
const (
	opcodeCont   = 0x0
	opcodeText   = 0x1
	opcodeClose  = 0x8
	opcodePing   = 0x9
	opcodePong   = 0xA
	extLen16     = 126
	extLen64     = 127
	maxShortLen  = 65536
	payloadMask  = 0x7f
	maskBit      = 0x80
	finBit       = 0x80
	opcodeMask   = 0x0f
	writeTimeout = 2 * time.Second
	dialTimeout  = 5 * time.Second
	maxFrameLen  = 1 << 24
	byteShift    = 8
)
