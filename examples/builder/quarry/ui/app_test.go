package ui

import "testing"

func TestNewAppMounts(t *testing.T) {
	for _, sz := range [][2]int{{1440, 960}, {1280, 800}} {
		r := newRig(t, sz[0], sz[1], "")
		r.png("smoke")
		if len(r.a.Page().Boxes()) == 0 {
			t.Fatal("no boxes")
		}
	}
}
