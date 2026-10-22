package window

// The devtools dock follows the Chrome DevTools dark theme: a dark grey dock,
// a lighter header and tab strip, a blue accent, and syntax colours for the
// JSON view. Each hex packs 0xRRGGBBAA.
const (
	devPanelBgHex  = 0x202124ff
	devHeaderBgHex = 0x292a2dff
	devBorderHex   = 0x3c4043ff
	devFgHex       = 0xe8eaedff
	devDimHex      = 0x9aa0a6ff
	devAccentHex   = 0x8ab4f8ff
	devErrInkHex   = 0xf28b82ff
	devKeyInkHex   = 0x9cdcfeff
	devStrInkHex   = 0xce9178ff
	devNumInkHex   = 0xb5cea8ff
	devBoolInkHex  = 0x569cd6ff
	devPunctInkHex = 0xd4d4d4ff
	devRowHoverHex = 0x35363aff
	devRowPickHex  = 0x3d4d6dff
	devFillInkHex  = 0x1a56dbff
	devStrokeHex   = 0xe5484dff
	devLineInkHex  = 0xd97706ff
	devTextInkHex  = 0x176b45ff
	devImageInkHex = 0x7c3aedff
	devGridInkHex  = 0x0891b2ff
)

var (
	devPanelBg   = devRGBA(devPanelBgHex)
	devHeaderBg  = devRGBA(devHeaderBgHex)
	devBorder    = devRGBA(devBorderHex)
	devFg        = devRGBA(devFgHex)
	devDim       = devRGBA(devDimHex)
	devAccent    = devRGBA(devAccentHex)
	devErrInk    = devRGBA(devErrInkHex)
	devKeyInk    = devRGBA(devKeyInkHex)
	devStrInk    = devRGBA(devStrInkHex)
	devNumInk    = devRGBA(devNumInkHex)
	devBoolInk   = devRGBA(devBoolInkHex)
	devPunctInk  = devRGBA(devPunctInkHex)
	devRowHover  = devRGBA(devRowHoverHex)
	devRowPick   = devRGBA(devRowPickHex)
	devFillInk   = devRGBA(devFillInkHex)
	devStrokeInk = devRGBA(devStrokeHex)
	devLineInk   = devRGBA(devLineInkHex)
	devTextInk   = devRGBA(devTextInkHex)
	devImageInk  = devRGBA(devImageInkHex)
	devGridInk   = devRGBA(devGridInkHex)
	devHoverInk  = devRGBA(devStrokeHex)
	devPinInk    = devRGBA(devFillInkHex)
)
