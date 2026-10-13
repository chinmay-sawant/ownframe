package replay

import (
	"context"
	"fmt"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

// groupModes are the separable mix-blend-modes the compositor accepts.
var groupModes = []string{"normal", "multiply", "screen", "overlay", "darken", "lighten",
	"difference", "exclusion", "hard-light", "soft-light", "color-dodge", "color-burn"}

// checkGroupComposite compares each accepted mode's composited group with the
// bitmap painter inside and outside the blend overlap, then checks nesting.
func checkGroupComposite() error {
	for _, mode := range groupModes {
		if err := checkGroupMode(mode); err != nil {
			return err
		}
	}

	return checkGroupNested()
}

func checkGroupMode(mode string) error {
	source := `<body style="margin:0;height:140px;background:#d2d2d2">` +
		`<div style="position:absolute;left:20px;top:20px;width:120px;height:80px;background:#3060c0">` +
		`<div style="position:absolute;left:30px;top:20px;width:60px;height:40px;background:#c05030;` +
		`mix-blend-mode:` + mode + `"></div></div></body>`

	d, err := render.DisplayList(context.Background(), source, 200, 140)
	if err != nil {
		return err
	}

	if !render.Replayable(d) {
		return fmt.Errorf("mode %s should replay", mode)
	}

	return compareGroupPixels(d, mode, [][2]int{{60, 50}, {25, 25}, {5, 5}})
}
