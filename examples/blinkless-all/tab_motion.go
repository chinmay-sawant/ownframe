package main

// motionTab is the "Motion path" tab. The engine turns offset-path, distance,
// rotate, anchor, position and the shorthand into one matrix on the box. The
// host reads the matrix back and prints the numbers below the lanes.
func motionTab() Tab {
	return Tab{
		ID:    "motion",
		Title: "Motion path",
		Props: []Prop{
			{
				Name: "offset-path",
				Demo: func() string {
					return motionDemoHTML("mp-dist-0", "mp-dist-50", "mp-dist-100", "mp-rot-auto", "mp-rot-reverse", "mp-rot-fixed")
				},
				Covers: []string{"offset-distance", "offset-rotate"},
				Note: "A path() line 200px long; the box rides it by offset-distance (0%, 50%, 100%) " +
					"and turns with the path unless offset-rotate says otherwise (auto, reverse, 45deg).",
			},
			{
				Name: "offset-anchor",
				Demo: func() string { return motionDemoHTML("mp-anchor-center", "mp-anchor-corner") },
				Note: "Which point of the box sits on the path: the center by default, or the top-left corner.",
			},
			{
				Name: "offset-position",
				Demo: func() string { return motionDemoHTML("mp-ray-origin") },
				Note: "Where a ray() starts, here 60px 30px from the box corner.",
			},
			{
				Name: "offset",
				Demo: func() string { return motionDemoHTML("mp-short") },
				Note: "The shorthand: path, distance, rotate and anchor in one declaration.",
			},
		},
		Head: motionHead,
	}
}

func motionHead() string { return `<style>` + motionCSS + `</style>` }
