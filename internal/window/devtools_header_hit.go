package window

// devHeaderRight places the header buttons from the right edge: the F12
// close hint, the X key, and the picker toggle. The close rect starts at
// the F12 label and runs to the right edge, so the hint, the gap after it,
// and the X key all close the dock. The picker hit runs from the picker
// origin to the close rect, so the gap between them toggles the picker and
// no dead space sits between the two hits. It returns the close rect and
// the picker origin so the draw and the hits agree.
func devHeaderRight(r devRect) (cls devRect, px float64) {
	wX := devTextWidth("X")
	wF12 := devTextWidth("F12")
	wPicker := devTextWidth("Picker")
	xX := r.X + r.W - devPanelPad - wX
	fx := xX - devHeaderGap - wF12
	px = fx - devHeaderGap - wPicker
	cls = devRect{X: fx, Y: r.Y, W: r.X + r.W - fx, H: devHeaderH}

	return cls, px
}
