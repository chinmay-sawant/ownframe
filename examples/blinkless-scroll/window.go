//go:build !js

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/chinmay-sawant/ownframe"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// ebitenSource reads the real window: the cursor, the left button, and the
// keys that went down this frame.
func ebitenSource() inputSource {
	return inputSource{
		cursor: func() (float64, float64) {
			x, y := ebiten.CursorPosition()

			return float64(x), float64(y)
		},
		click: func() bool { return inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) },
		keys: func() []string {
			var out []string
			for _, k := range inpututil.AppendJustPressedKeys(nil) {
				out = append(out, k.String())
			}

			return out
		},
	}
}

// attachWindow installs the per-frame callback with the real window input.
func (s *session) attachWindow() {
	s.reportErr = func(err error) {
		fmt.Fprintln(os.Stderr, "blinkless-scroll: style patch refused, host keeps running:", err)
	}
	// The window offers each wheel event here before it scrolls the page. The
	// coordinates are page coordinates, the same ones the host works in.
	s.page.AddHandlers(ownframe.Handlers{
		Wheel: func(_ context.Context, x, y, _, dy float64) (bool, error) { return s.wheel(x, y, dy), nil },
	})
	s.attach(ebitenSource())
}
