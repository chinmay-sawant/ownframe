package main

import "testing"

func TestProfileCardPosAnchorsToRow(t *testing.T) {
	a := &app{profileY: 120, profileAnchored: true}
	left, top := a.profileCardPos()
	if left != 888 {
		t.Fatalf("left = %d, want 888", left)
	}
	if top != 120 {
		t.Fatalf("top = %d, want row 120", top)
	}
}

func TestProfileCardPosClampsBottom(t *testing.T) {
	a := &app{profileY: 850, profileAnchored: true}
	_, top := a.profileCardPos()
	if want := 900 - 480 - 12; top != want {
		t.Fatalf("top = %d, want clamped %d", top, want)
	}
}

func TestMsgMenuPosFlipsNearBottom(t *testing.T) {
	a := &app{msgX: 1000, msgY: 850, msgAnchored: true}
	_, y := a.msgMenuPos()
	if y >= 850 {
		t.Fatalf("y = %d, want flipped above 850", y)
	}
}

func TestMsgMenuPosFallsBackWithoutAnchor(t *testing.T) {
	a := &app{}
	x, y := a.msgMenuPos()
	if x != overlayMsgMenuFallbackX || y != overlayMsgMenuFallbackY {
		t.Fatalf("fallback = (%d,%d), want (%d,%d)", x, y, overlayMsgMenuFallbackX, overlayMsgMenuFallbackY)
	}
}
