package window

import (
	"image"
	"testing"
)

func TestPaceChromeUnchanged(t *testing.T) {
	g := &paceGame{}
	bar := []image.Rectangle{image.Rect(1, 1, 9, 60)}

	if g.chromeChanged(nil) {
		t.Fatal("no chrome twice must read unchanged")
	}

	if !g.chromeChanged(bar) {
		t.Fatal("a new scrollbar must read changed")
	}

	if g.chromeChanged(bar) {
		t.Fatal("a still scrollbar must read unchanged")
	}

	moved := []image.Rectangle{image.Rect(1, 20, 9, 79)}
	if !g.chromeChanged(moved) {
		t.Fatal("a moved scrollbar must read changed")
	}

	if !g.chromeChanged(nil) {
		t.Fatal("a dropped scrollbar must read changed")
	}

	if g.chromeChanged(nil) {
		t.Fatal("no chrome again must read unchanged")
	}
}
