package window

import "testing"

func TestFramePointScalesOntoThePicture(t *testing.T) {
	t.Parallel()

	x, y := framePoint(50, 25, 200, 100, 100, 50)
	if x != 100 || y != 50 {
		t.Fatalf("point = %v, %v", x, y)
	}
}

func TestFramePointMatchesWhenSizesMatch(t *testing.T) {
	t.Parallel()

	x, y := framePoint(12, 34, 480, 640, 480, 640)
	if x != 12 || y != 34 {
		t.Fatalf("point = %v, %v", x, y)
	}
}

func TestPanScrollClampsToThePage(t *testing.T) {
	t.Parallel()

	x, y := panScroll(0, 0, 0, 1, 200, 800, 200, 400)
	if x != 0 || y != 0 {
		t.Fatalf("wheel up at the top = %d, %d", x, y)
	}

	x, y = panScroll(0, 0, 0, -1, 200, 800, 200, 400)
	if x != 0 || y != scrollStep {
		t.Fatalf("wheel down = %d, %d", x, y)
	}

	x, y = panScroll(0, 400, 0, -1, 200, 800, 200, 400)
	if x != 0 || y != 400 {
		t.Fatalf("past the end = %d, %d", x, y)
	}
}

func TestContentPointAddsTheScroll(t *testing.T) {
	t.Parallel()

	x, y := contentPoint(10, 20, 0, 100, false, 200, 800, 200, 400)
	if x != 10 || y != 120 {
		t.Fatalf("scrolled point = %v, %v", x, y)
	}
}

func TestFramePointRejectsAnEmptyScreen(t *testing.T) {
	t.Parallel()

	x, y := framePoint(12, 34, 480, 640, 0, 640)
	if x != 0 || y != 0 {
		t.Fatalf("point = %v, %v", x, y)
	}
}
