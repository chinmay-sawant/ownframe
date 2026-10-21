//go:build linux && !android

package clipboard

const (
	xShift8  = 8
	xShift16 = 16
	xShift24 = 24

	xProtoMajor = 11

	xEvSize    = 32
	xSelClear  = 29
	xSelReq    = 30
	xSelNotify = 31

	xMaxWords = 1 << 20

	xAtomId = 4
	xStrId  = 31

	xOpIntern   = 16
	xOpCreate   = 1
	xOpSetOwner = 22
	xOpGetOwner = 23
	xOpConvert  = 24
	xOpGetProp  = 20
	xOpChange   = 18
	xOpSend     = 25

	xWordsCreate   = 8
	xWordsSetOwner = 4
	xWordsConvert  = 6
	xWordsSend     = 11
	xWordsGetOwner = 2

	xSizeSetOwner = 16
	xSizeConvert  = 24
	xSizeSend     = 44
	xSizeGetOwner = 8

	xFmt8  = 8
	xFmt32 = 32

	xWord = 4
	xMask = 3

	xHdr8  = 8
	xHdr24 = 24
	xHdr40 = 40

	xAuthHdr = 4
	xAuthLen = 2
)
