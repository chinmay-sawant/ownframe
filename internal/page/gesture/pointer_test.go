package gesture

import "testing"

func TestPointerPressHoverAndSwipe(t *testing.T) {
	t.Parallel()

	if !SameHit("a", "a") || SameHit("a", "b") {
		t.Fatal("hit identity")
	}

	if !RunPress(true, false, true) || RunPress(true, true, true) || !RunPress(true, true, false) {
		t.Fatal("press handler gate")
	}

	if RunPress(false, true, false) {
		t.Fatal("missing handler ran")
	}

	if !CanPaintHover(true, true) || CanPaintHover(true, false) || CanPaintHover(false, true) {
		t.Fatal("hover paint gate")
	}

	if !SkipHoverRedraw(true, true) || !SkipHoverRedraw(false, false) || SkipHoverRedraw(false, true) {
		t.Fatal("hover redraw")
	}

	if !ClaimsWord(textField("hi"), true) {
		t.Fatal("text hold")
	}

	disabled := textField("hi")
	disabled.Disabled = true
	if ClaimsWord(disabled, true) || ClaimsWord(textField("hi"), false) {
		t.Fatal("word claim")
	}

	dx, dy, ok := Swipe(false, 3, -4)
	if ok || dx != 0 || dy != 0 {
		t.Fatalf("no handler %v %v %v", dx, dy, ok)
	}

	dx, dy, ok = Swipe(true, 3, -4)
	if !ok || dx != 3 || dy != -4 {
		t.Fatalf("swipe %v %v %v", dx, dy, ok)
	}

	var flag uint32
	if TakeCancel(&flag) {
		t.Fatal("empty cancel")
	}

	QueueCancel(&flag)
	if !TakeCancel(&flag) || TakeCancel(&flag) {
		t.Fatal("cancel was not one-shot")
	}

	if !TouchZoom(true, false) || !TouchZoom(false, true) || TouchZoom(false, false) {
		t.Fatal("touch zoom")
	}
}
