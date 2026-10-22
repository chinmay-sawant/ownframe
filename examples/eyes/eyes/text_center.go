package eyes

import (
	"bytes"
	"unicode/utf8"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/frame"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

const pointsPerPixel = 4.0 / 3

func textCenterX(d *ownframe.Display, boxes []ownframe.Box, id string) float64 {
	for _, box := range boxes {
		if box.ID == id {
			x, _, w, _ := frame.BoxUnits(d, box)
			return x + w/2
		}
	}
	return 0
}

func setCenteredText(op *ownframe.DisplayOp, value string, center float64) {
	if op == nil || op.Text == value {
		return
	}
	op.Text = value
	if op.Font == nil {
		return
	}
	source, err := text.NewGoTextFaceSource(bytes.NewReader(op.Font.Bytes()))
	if err != nil {
		return
	}
	face := &text.GoTextFace{Source: source, Size: op.Size * pointsPerPixel}
	width, _ := text.Measure(value, face, 0)
	width = width/pointsPerPixel + op.LetterSpacing*float64(utf8.RuneCountInString(value))
	op.X, op.W = center-width/halfDivisor, width
}
