package masking

var blendModes = map[string]struct {
	sep bool
	id  int
}{
	"": {true, 0}, "normal": {true, 0},
	"multiply": {true, 1}, "screen": {true, 2},
	"overlay": {true, 3}, "darken": {true, 4},
	"lighten": {true, 5}, "color-dodge": {true, 6},
	"color-burn": {true, 7}, "hard-light": {true, 8},
	"soft-light": {true, 9}, "difference": {true, 10},
	"exclusion": {true, 11}, "hue": {}, "saturation": {},
	"color": {}, "luminosity": {},
}

func BlendSupported(mode string) bool {
	_, ok := blendModes[mode]
	return ok
}

func BlendSeparable(mode string) bool {
	return blendModes[mode].sep
}

func BlendModeID(mode string) int {
	return blendModes[mode].id
}

func BlendNormal(mode string) bool {
	return mode == "" || mode == "normal"
}
