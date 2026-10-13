package replay

import (
	"context"
	"fmt"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

// checkGroupNested compares a multiply element holding a screen child, so the
// child's layer buffers inside the parent's and both blend once.
func checkGroupNested() error {
	source := `<body style="margin:0;height:140px;background:#d2d2d2">` +
		`<div style="position:absolute;left:20px;top:20px;width:120px;height:80px;background:#3060c0;` +
		`mix-blend-mode:multiply"><div style="position:absolute;left:30px;top:20px;width:60px;height:40px;` +
		`background:#c05030;mix-blend-mode:screen"></div></div></body>`

	d, err := render.DisplayList(context.Background(), source, 200, 140)
	if err != nil {
		return err
	}

	if !render.Replayable(d) {
		return fmt.Errorf("nested groups should replay")
	}

	return compareGroupPixels(d, "nested multiply/screen", [][2]int{{60, 50}, {25, 25}, {5, 5}})
}
