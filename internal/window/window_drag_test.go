package window

import "testing"

func TestWindowDragNeedsUserPress(t *testing.T) {
	var d windowDrag
	handled, move, _, _, _ := d.step(200, 100, false, true)
	if handled || move {
		t.Fatal("hover moved window")
	}
	handled, move, _, _, _ = d.step(200, 100, true, false)
	if handled || move {
		t.Fatal("transparent pixels began drag")
	}
	d.step(200, 100, false, false)
	d.step(200, 100, true, true)
	_, move, _, dx, dy := d.step(230, 120, true, false)
	if !move || dx != 30 || dy != 20 {
		t.Fatal("drag delta incorrect")
	}
	_, move, _, _, _ = d.step(200, 100, true, true)
	if move {
		t.Fatal("settled grab point moved again")
	}
	_, _, click, _, _ := d.step(200, 100, false, true)
	if click || d.active {
		t.Fatal("drag released as click")
	}
	d.step(200, 100, true, true)
	_, _, click, _, _ = d.step(201, 100, false, true)
	if !click {
		t.Fatal("small gesture did not click")
	}
}

func TestRecoverMotionCancelsOwnWindowShift(t *testing.T) {
	t.Parallel()

	// Steady hand while the window slides under it: the cursor mirror
	// cancels the window shift and the hand reads zero.
	if dx, dy := recoverMotion(900, 100, 1000, 100, 100, 50, 0, 50); dx != 0 || dy != 0 {
		t.Fatalf("steady hand reads %d,%d, want 0,0", dx, dy)
	}

	// Real hand motion passes through untouched.
	if dx, dy := recoverMotion(1010, 120, 1000, 100, 100, 50, 100, 50); dx != 10 || dy != 20 {
		t.Fatalf("hand motion reads %d,%d, want 10,20", dx, dy)
	}
}

func TestWindowRebaseIgnoresMonitorOrigin(t *testing.T) {
	t.Parallel()

	// Landed request, still in flight, and a partial step are all ours.
	if windowRebased(-1283, 384, -1290, 382, -1290, 382) {
		t.Fatal("landed request looked like a new origin")
	}

	if windowRebased(-1283, 384, -1290, 382, -1283, 384) {
		t.Fatal("unapplied request looked like a new origin")
	}

	if windowRebased(-1283, 384, -1290, 382, -1286, 383) {
		t.Fatal("partial apply looked like a new origin")
	}

	// drag.log line 189: the window origin jumped to the other monitor
	// while the cursor stayed in the titlebar.
	if !windowRebased(-1290, 382, -1290, 382, 630, 62) {
		t.Fatal("monitor origin jump was treated as our request")
	}
}

func TestDragAppliesFastHandInFull(t *testing.T) {
	t.Parallel()

	// drag.log line 26: the cursor moved 190px in one sample. The window
	// has to take the same step or it trails the pointer.
	var d windowDrag
	d.tracked = true
	d.lrx, d.lry = 1640, 138
	d.lsx, d.lsy = 1640, 138
	d.lwx, d.lwy = -452, 244

	dx, dy, ok := d.prepareDrag(dragSample{
		rawX: 1450, rawY: 174,
		sx: 1450, sy: 174,
		wx: -452, wy: 244,
	})
	if !ok || dx != -190 || dy != 36 {
		t.Fatalf("fast hand = %d,%d ok=%v, want -190,36", dx, dy, ok)
	}
}

func TestDragCursorJumpDoesNotReverse(t *testing.T) {
	t.Parallel()

	// drag.log lines 39-40. The raw cursor jumps by about one monitor
	// width while the window stays where we put it. The first sample is
	// refused and the smoother rewound. The second sample, still in the
	// new place, adopts that coordinate without moving the window.
	var d windowDrag
	d.tracked = true
	d.fx, d.fy = 1187, 219
	d.lrx, d.lry = 1187, 219
	d.lsx, d.lsy = 1187, 219
	d.lwx, d.lwy = -1096, 354
	d.reqx, d.reqy = -1096, 354
	d.haveReq = true

	_, _, ok := d.prepareDrag(dragSample{
		rawX: 1187, rawY: 219,
		sx: 1187, sy: 219,
		wx: -1096, wy: 354,
	})
	if ok {
		t.Fatal("held cursor moved the window")
	}

	dx, dy, ok := d.prepareDrag(dragSample{
		rawX: 3115, rawY: 64,
		sx: 2151, sy: 64,
		wx: -1096, wy: 354,
	})
	if ok {
		t.Fatalf("cursor jump applied %d,%d", dx, dy)
	}

	if !d.jumpDrop || d.jumpSnap || d.lrx != 1187 || d.fx != 1187 {
		t.Fatalf("first jump snapped, drop=%v snap=%v cursor=%d fx=%v", d.jumpDrop, d.jumpSnap, d.lrx, d.fx)
	}

	dx, dy, ok = d.prepareDrag(dragSample{
		rawX: 3115, rawY: 64,
		sx: 2151, sy: 64,
		wx: -1096, wy: 354,
	})
	if ok {
		t.Fatalf("stable coordinate moved the window by %d,%d", dx, dy)
	}

	if !d.jumpSnap || d.lrx != 3115 || d.lry != 64 {
		t.Fatalf("stable coordinate not adopted, snap=%v cursor=%d,%d", d.jumpSnap, d.lrx, d.lry)
	}

	dx, dy, ok = d.prepareDrag(dragSample{
		rawX: 3145, rawY: 64,
		sx: 3145, sy: 64,
		wx: -1096, wy: 354,
	})
	if !ok || dx != 30 || dy != 0 {
		t.Fatalf("motion after adopt = %d,%d ok=%v, want 30,0", dx, dy, ok)
	}
}

func TestDragKeepsRealHand(t *testing.T) {
	t.Parallel()

	// Motion under the jump threshold moves the window by the full hand
	// delta, including a step that crosses a monitor.
	var d windowDrag
	d.tracked = true
	d.lrx, d.lry = 1640, 138
	d.lsx, d.lsy = 1640, 138
	d.lwx, d.lwy = -452, 244
	d.haveReq = true
	d.reqx, d.reqy = -452, 244

	dx, dy, ok := d.prepareDrag(dragSample{
		rawX: 1450, rawY: 174,
		sx: 1450, sy: 174,
		wx: -452, wy: 244,
	})
	if !ok || dx != -190 || dy != 36 {
		t.Fatalf("hand = %d,%d ok=%v, want -190,36", dx, dy, ok)
	}
}

func TestDragSeamWindowRebaseStaysWithCursor(t *testing.T) {
	t.Parallel()

	// drag.log lines 188-189. The presented origin jumps by the monitor
	// rebase. The cursor only moved a few pixels in the titlebar.
	var d windowDrag
	d.tracked = true
	d.lrx, d.lry = 630, 281
	d.lsx, d.lsy = 630, 281
	d.lwx, d.lwy = -1290, 382
	d.reqx, d.reqy = -1290, 382
	d.haveReq = true

	dx, dy, ok := d.prepareDrag(dragSample{
		rawX: 605, rawY: 275,
		sx: 605, sy: 275,
		wx: 630, wy: 62,
	})
	if !ok || dx != -25 || dy != -6 {
		t.Fatalf("rebased hand = %d,%d ok=%v, want -25,-6", dx, dy, ok)
	}

	if !d.rebased || d.rebaseX != -1290 || d.rebaseY != 382 {
		t.Fatalf("rebase from %d,%d flagged %v", d.rebaseX, d.rebaseY, d.rebased)
	}
}

func TestDragDoesNotOscillateAfterApply(t *testing.T) {
	t.Parallel()

	var d windowDrag
	_, _, ok := d.prepareDrag(dragSample{
		rawX: 100, rawY: 50,
		sx: 100, sy: 50,
		wx: 0, wy: 0,
	})
	if ok {
		t.Fatal("anchor frame moved the window")
	}

	dx, dy, ok := d.prepareDrag(dragSample{
		rawX: 200, rawY: 50,
		sx: 200, sy: 50,
		wx: 0, wy: 0,
	})
	if !ok || dx != 100 || dy != 0 {
		t.Fatalf("hand = %d,%d ok=%v, want 100,0", dx, dy, ok)
	}

	d.commitDrag(0, 0, dx, dy, false)

	// The window caught the hand and the window-relative cursor mirrored
	// back. A steady hand must read zero, not the inverse of the move.
	dx, dy, ok = d.prepareDrag(dragSample{
		rawX: 100, rawY: 50,
		sx: 100, sy: 50,
		wx: 100, wy: 0,
	})
	if ok {
		t.Fatalf("settled hand moved %d,%d", dx, dy)
	}
}

func TestDragCursorSpikeRewindsWhenTheHandReturns(t *testing.T) {
	t.Parallel()

	// drag.log line 781. The smoothed cursor was -414 and the raw sample
	// reconstructed from the low-pass is 3414. The next sample is back
	// on the titlebar. Applying the spike is what threw the window from
	// x=-1328 to about 1551.
	var d windowDrag
	d.tracked = true
	d.fx, d.fy = -414, 15
	d.lrx, d.lry = -414, 15
	d.lsx, d.lsy = -414, 15
	d.lwx, d.lwy = -1328, 425
	d.haveReq = true
	d.reqx, d.reqy = -1328, 425

	dx, dy, ok := d.prepareDrag(dragSample{
		rawX: 3414, rawY: -697,
		sx: 1500, sy: -341,
		wx: -1328, wy: 425,
	})
	if ok || dx != 0 || dy != 0 {
		t.Fatalf("spike applied %d,%d ok=%v", dx, dy, ok)
	}

	if d.lrx != -414 || d.lwx != -1328 || d.fx != -414 {
		t.Fatalf("spike moved anchors cursor=%d window=%d fx=%v", d.lrx, d.lwx, d.fx)
	}

	dx, dy, ok = d.prepareDrag(dragSample{
		rawX: -411, rawY: 16,
		sx: -414, sy: 15,
		wx: -1328, wy: 425,
	})
	if ok || dx != 0 || dy != 0 {
		t.Fatalf("return moved the window by %d,%d ok=%v", dx, dy, ok)
	}
}

func TestDragCursorSpikeTwoFramesThenReturn(t *testing.T) {
	t.Parallel()

	// drag.log lines 1141-1143. Two raw samples sit in the bogus
	// coordinate space and are too far apart to be the same rebase.
	// The third sample is back near the titlebar. None of the bogus
	// steps may move the window.
	var d windowDrag
	d.tracked = true
	d.fx, d.fy = 1001, 39
	d.lrx, d.lry = 1001, 39
	d.lsx, d.lsy = 1001, 39
	d.lwx, d.lwy = -1219, 427

	dx, dy, ok := d.prepareDrag(dragSample{
		rawX: 2923, rawY: -279,
		sx: 1962, sy: -120,
		wx: -1219, wy: 427,
	})
	if ok || dx != 0 || dy != 0 {
		t.Fatalf("first spike applied %d,%d", dx, dy)
	}

	dx, dy, ok = d.prepareDrag(dragSample{
		rawX: 1900, rawY: -100,
		sx: 1450, sy: -30,
		wx: -1219, wy: 427,
	})
	if ok || dx != 0 || dy != 0 || d.jumpSnap {
		t.Fatalf("second spike applied %d,%d snap=%v", dx, dy, d.jumpSnap)
	}

	dx, dy, ok = d.prepareDrag(dragSample{
		rawX: 984, rawY: 45,
		sx: 992, sy: 42,
		wx: -1219, wy: 427,
	})
	if !ok || dx != -9 || dy != 3 {
		t.Fatalf("return = %d,%d ok=%v, want -9,3", dx, dy, ok)
	}
}

func TestDragCursorJumpAdoptsStableCoordinates(t *testing.T) {
	t.Parallel()

	// drag.log lines 892-893. The cursor's coordinate space changes and
	// stays there. The window must not take the jump, and a later small
	// move in the new space must apply.
	var d windowDrag
	d.tracked = true
	d.fx, d.fy = 765, 3
	d.lrx, d.lry = 765, 3
	d.lsx, d.lsy = 765, 3
	d.lwx, d.lwy = -135, -7

	dx, dy, ok := d.prepareDrag(dragSample{
		rawX: -735, rawY: 350,
		sx: 15, sy: 176,
		wx: -135, wy: -7,
	})
	if ok || dx != 0 || dy != 0 || d.lrx != 765 {
		t.Fatalf("first jump applied %d,%d cursor=%d", dx, dy, d.lrx)
	}

	dx, dy, ok = d.prepareDrag(dragSample{
		rawX: -700, rawY: 360,
		sx: -700, sy: 360,
		wx: -135, wy: -7,
	})
	if ok || dx != 0 || dy != 0 || !d.jumpSnap || d.lrx != -700 {
		t.Fatalf("adopt = %d,%d ok=%v snap=%v cursor=%d", dx, dy, ok, d.jumpSnap, d.lrx)
	}

	dx, dy, ok = d.prepareDrag(dragSample{
		rawX: -660, rawY: 360,
		sx: -660, sy: 360,
		wx: -135, wy: -7,
	})
	if !ok || dx != 40 || dy != 0 {
		t.Fatalf("motion after adopt = %d,%d ok=%v, want 40,0", dx, dy, ok)
	}
}

func TestDragAppliesThreeHundredPixelStep(t *testing.T) {
	t.Parallel()

	var d windowDrag
	d.tracked = true
	d.lrx, d.lry = 1000, 100
	d.lsx, d.lsy = 1000, 100
	d.lwx, d.lwy = 0, 0

	dx, dy, ok := d.prepareDrag(dragSample{
		rawX: 1300, rawY: 80,
		sx: 1300, sy: 80,
		wx: 0, wy: 0,
	})
	if !ok || dx != 300 || dy != -20 {
		t.Fatalf("step = %d,%d ok=%v, want 300,-20", dx, dy, ok)
	}
}

func TestDragSpikeLeavesInFlightRequestToCancel(t *testing.T) {
	t.Parallel()

	// A position request is in flight. The next sample is a cursor spike
	// while the window is still at the old origin. The sample after that
	// is the request landing and the cursor back on the titlebar. The
	// landing cancels through the cursor mirror.
	var d windowDrag
	d.tracked = true
	d.fx, d.fy = 500, 40
	d.lrx, d.lry = 500, 40
	d.lsx, d.lsy = 500, 40
	d.lwx, d.lwy = 0, 0

	dx, _, ok := d.prepareDrag(dragSample{
		rawX: 600, rawY: 40,
		sx: 600, sy: 40,
		wx: 0, wy: 0,
	})
	if !ok || dx != 100 {
		t.Fatalf("hand = %d ok=%v, want 100", dx, ok)
	}

	d.commitDrag(0, 0, 100, 0, false)
	_, _, ok = d.prepareDrag(dragSample{
		rawX: 2500, rawY: 40,
		sx: 1550, sy: 40,
		wx: 0, wy: 0,
	})
	if ok || d.lwx != 0 || d.lrx != 600 {
		t.Fatalf("spike moved anchors window=%d cursor=%d ok=%v", d.lwx, d.lrx, ok)
	}

	dx, _, ok = d.prepareDrag(dragSample{
		rawX: 500, rawY: 40,
		sx: 500, sy: 40,
		wx: 100, wy: 0,
	})
	if ok {
		t.Fatalf("landed request applied again, dx=%d", dx)
	}
}

func TestNegligibleDragWithholdsMicroJitter(t *testing.T) {
	t.Parallel()

	if !negligibleDrag(1, -1) {
		t.Fatal("1px jitter not withheld")
	}

	if negligibleDrag(2, 0) {
		t.Fatal("2px move withheld")
	}
}

func TestSmoothCursorHalvesDistance(t *testing.T) {
	t.Parallel()

	if got := smoothCursor(0, 100); got != 50 {
		t.Fatalf("smoothed = %v, want 50", got)
	}

	if got := smoothCursor(100, 100); got != 100 {
		t.Fatalf("settled cursor moved to %v", got)
	}

	prev := 0.0
	for range 20 {
		prev = smoothCursor(prev, 100)
	}

	if prev < 99.9 {
		t.Fatalf("cursor did not converge, stuck at %v", prev)
	}
}
