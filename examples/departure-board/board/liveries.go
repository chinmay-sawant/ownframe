package board

// Livery is one board skin: a day sheet, a dimmer night sheet, and the
// accent both sheets paint with.
type Livery struct {
	Name   string
	CSS    string
	Night  string
	Accent [3]float64
}

// DefaultLivery is the skin the board opens on.
const DefaultLivery = "amber"

// liveryList holds the four skins in button order. Each colors row lists
// the eight custom properties in sheet order, day then night, and the
// accent literals must match --accent in both sheets of that skin.
var liveryList = []Livery{
	livery("amber",
		"#0b0906 #171310 #ffb000 #ffcc66 #ffb000 #3a2c10 #8a6c2a 6px",
		"#050403 #0c0a07 #c9881f #a87f39 #ffb000 #241b08 #57451b 6px",
		[3]float64{1, 176.0 / 255, 0}),
	livery("flap",
		"#0d0c0a #1a1815 #f2ecdd #fffaf0 #e8dcc4 #38332b #8a8375 4px",
		"#070605 #100f0d #b8b1a2 #c9c2b2 #e8dcc4 #241f19 #5c564c 4px",
		[3]float64{232.0 / 255, 220.0 / 255, 196.0 / 255}),
	livery("lcd",
		"#06101f #0b1a30 #8ecbff #c9e6ff #4fa3ff #1d3a63 #5b7ea6 3px",
		"#03070f #060f1c #5b8cb3 #7fa8cc #4fa3ff #112240 #33506b 3px",
		[3]float64{79.0 / 255, 163.0 / 255, 1}),
	livery("paper",
		"#f2ead8 #fbf6ea #2a2318 #1c1810 #b4552a #d8c9a6 #7a6a4d 2px",
		"#1a1712 #221d16 #d8cdb8 #c0b49c #b4552a #3a3328 #6b604f 2px",
		[3]float64{180.0 / 255, 85.0 / 255, 42.0 / 255}),
}

// LiveryNames returns the livery names in button order.
func LiveryNames() []string {
	names := make([]string, len(liveryList))
	for i, l := range liveryList {
		names[i] = l.Name
	}

	return names
}

// LiveryAccent returns the accent of one livery.
func LiveryAccent(name string) ([3]float64, bool) {
	l, ok := liveryByName(name)
	if !ok {
		return [3]float64{}, false
	}

	return l.Accent, true
}

// liveryByName looks a livery up by name.
func liveryByName(name string) (Livery, bool) {
	for _, l := range liveryList {
		if l.Name == name {
			return l, true
		}
	}

	return Livery{}, false
}
