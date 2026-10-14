package window

import (
	"sync"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// devMeasure memoizes panel text widths and the badge face line height.
// Panel labels repeat on every frame and text.Measure shapes its run per
// call, so the memo turns a frame of label measures into map hits.
var devMeasure = struct {
	mu    sync.Mutex
	width map[string]float64
	faceH float64
}{}

// devTextWidth returns the badge-face width of s, measuring it once.
func devTextWidth(s string) float64 {
	devMeasure.mu.Lock()
	defer devMeasure.mu.Unlock()

	if w, ok := devMeasure.width[s]; ok {
		return w
	}

	w, _ := text.Measure(s, badgeFace, 0)
	if devMeasure.width == nil {
		devMeasure.width = map[string]float64{}
	}
	devMeasure.width[s] = w

	return w
}

// devFaceHeight returns the badge face line height, measured once.
func devFaceHeight() float64 {
	devMeasure.mu.Lock()
	defer devMeasure.mu.Unlock()

	if devMeasure.faceH == 0 {
		_, devMeasure.faceH = text.Measure("Ag", badgeFace, 0)
	}

	return devMeasure.faceH
}
