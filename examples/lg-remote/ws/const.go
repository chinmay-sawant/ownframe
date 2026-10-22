package ws

import "time"

// WebSocket frame constants (RFC 6455 opcodes and length markers).
// Timeouts bound each network wait so a silent peer cannot park a
// goroutine forever. Limits bound handshake and message memory.
const (
	opcodeCont  = 0x0
	opcodeText  = 0x1
	opcodeClose = 0x8
	opcodePing  = 0x9
	opcodePong  = 0xA
	extLen16    = 126
	extLen64    = 127
	maxShortLen = 65536
	payloadMask = 0x7f
	maskBit     = 0x80
	finBit      = 0x80
	opcodeMask  = 0x0f
	// Control frames carry at most 125 bytes of payload per RFC 6455.
	maxCtrlLen   = 125
	writeTimeout = 2 * time.Second
	readTimeout  = 10 * time.Second
	dialTimeout  = 5 * time.Second
	shakeTimeout = 5 * time.Second
	maxFrameLen  = 1 << 24
	maxMsgLen    = 1 << 22
	maxHeadLines = 64
	maxHeadBytes = 16 << 10
	byteShift    = 8
)
