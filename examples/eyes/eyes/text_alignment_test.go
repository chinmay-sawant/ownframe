package eyes

import (
	"bytes"
	"math"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/chinmay-sawant/ownframe"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

func TestMoodNameStaysCenteredWhenItsWidthChanges(t *testing.T) {
	a, c := newT(t)
	tk(t, a, c, time.Second/60)
	a.selectMood(poseIndex("worried"), c.at)
	tk(t, a, c, time.Second/60)
	assertTextCentered(t, a, "t-mood", a.pm.mood)
	a.selectMood(poseIndex("frustrated"), c.at)
	tk(t, a, c, time.Second/60)
	assertTextCentered(t, a, "t-mood", a.pm.mood)
}

func TestDevButtonTextStaysCenteredWhenItsWidthChanges(t *testing.T) {
	a, c := newT(t)
	tk(t, a, c, time.Second/60)
	if err := a.onTap(t.Context(), ownframe.Box{ID: "dev-toggle"}); err != nil {
		t.Fatal(err)
	}
	tk(t, a, c, time.Second/60)
	assertTextCentered(t, a, "dev-toggle", a.pm.dev)
}

func assertTextCentered(t *testing.T, a *App, id string, op *ownframe.DisplayOp) {
	t.Helper()
	var center float64
	for _, box := range a.page.Boxes() {
		if box.ID == id {
			x, _, w, _ := ownframe.BoxUnits(a.page.Display(), box)
			center = x + w/2
			break
		}
	}
	source, err := text.NewGoTextFaceSource(bytes.NewReader(op.Font.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	face := &text.GoTextFace{Source: source, Size: op.Size * 4 / 3}
	width, _ := text.Measure(op.Text, face, 0)
	width = width*3/4 + op.LetterSpacing*float64(utf8.RuneCountInString(op.Text))
	if got := op.X + width/2; math.Abs(got-center) > 1 {
		t.Fatalf("%s text center = %.1f, want %.1f", op.Text, got, center)
	}
}
