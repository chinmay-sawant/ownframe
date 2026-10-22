package window

// devHeaderRight places the header buttons from the right edge: the close
// key, the F12 hint, and the picker toggle. It returns the close rect and
// the picker origin so the draw and the hits agree.
func devHeaderRight(r devRect) (cls devRect, px float64) {
	wX := devTextWidth("X")
	xX := r.X + r.W - devPanelPad - wX
	cls = devRect{X: xX - devClosePad, Y: r.Y, W: wX + devClosePad + devPanelPad, H: devHeaderH}
	fx := xX - devHeaderGap - devTextWidth("F12")

	return cls, fx - devHeaderGap - devTextWidth("Picker")
}
