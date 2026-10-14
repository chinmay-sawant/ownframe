package dump

// Rect is a JSON-friendly CSS-pixel rectangle.
type Rect struct{ X0, Y0, X1, Y1 int }

// Frame is one measured frame after a Tick.
type Frame struct {
	Kind        string // "full-replay", "full-buffer", "rect", or "blit"
	TickNS      int64
	Ops         int
	ReplayedOps int  // ops meeting any dirty rect, approximate for kind "rect"
	Rect        Rect // bounding box of the damage list
	Rects       int  // damage-list length; 0 on a blit frame
	Area        int  // summed damage-list area in CSS pixels
	FullArea    int  // content area in CSS pixels
}

// Report is one scenario's measurement.
type Report struct {
	Scenario     string
	Ops          int
	Boxes        int
	RedrawNS     int64
	Frames       []Frame
	Kinds        map[string]int
	TickP50NS    int64
	TickP95NS    int64
	TickMaxNS    int64
	RectMaxArea  int // largest bounding box, for old and new dumps to compare
	RectMeanArea int
	RectsMean    float64 // mean damage-list length on a reporting frame
	AreaMean     int     // mean summed damage area on a reporting frame
	AreaMax      int     // largest summed damage area in one frame
	AllocsFrame  float64
}

// rectArea is the pixel area of r, or 0 when r is empty or unset.
func rectArea(r Rect) int {
	w, h := r.X1-r.X0, r.Y1-r.Y0
	if w <= 0 || h <= 0 {
		return 0
	}
	return w * h
}
