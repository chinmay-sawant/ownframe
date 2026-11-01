package window

// devClosePad is the pointer slop around the X and Picker labels. The F12
// hint draws between them but takes no hit, so a press aimed at Picker or
// the hint cannot close the dock.
const devClosePad = 6.0

// devHeaderRight places the header buttons from the right edge: the Picker
// toggle, the F12 hint, and the X key. Only the X key closes the dock; the
// F12 label is a hint with dead space on both sides. The picker hit covers
// only its own label, so the gaps around F12 do nothing. It returns the
// close rect and the picker text origin so the draw and the hits agree.
func devHeaderRight(r devRect) (cls devRect, px float64) {
	wX := devTextWidth("X")
	wPicker := devTextWidth("Picker")
	xX := r.X + r.W - devPanelPad - wX
	fx := xX - devHeaderGap - devTextWidth("F12")
	px = fx - devHeaderGap - wPicker
	cls = devRect{X: xX - devClosePad, Y: r.Y, W: wX + devClosePad*2, H: devHeaderH}

	if cls.X < r.X {
		cls.W -= r.X - cls.X
		cls.X = r.X
	}

	if cls.X+cls.W > r.X+r.W {
		cls.W = r.X + r.W - cls.X
	}

	return cls, px
}

// devHeaderF12X is the draw origin of the F12 hint. It takes no hit.
func devHeaderF12X(r devRect) float64 {
	wX := devTextWidth("X")
	xX := r.X + r.W - devPanelPad - wX

	return xX - devHeaderGap - devTextWidth("F12")
}

// devPickerRect is the tight hit around the Picker label drawn at px.
func devPickerRect(px, panelY float64) devRect {
	return devRect{X: px - devClosePad, Y: panelY, W: devTextWidth("Picker") + devClosePad*2, H: devHeaderH}
}
