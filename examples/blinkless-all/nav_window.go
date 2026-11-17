//go:build !js

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// ebitenLiveSource reads the real window: the cursor, the left button, typed
// characters, and the editing keys that went down this frame.
func ebitenLiveSource(page *ownframe.Page) liveSource {
	return liveSource{
		cursor: func() (float64, float64) {
			x, y := page.Window().CursorPosition()

			return float64(x), float64(y)
		},
		click: func() bool { return inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) },
		down:  func() bool { return ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) },
		keys: func() []string {
			var out []string

			for _, k := range inpututil.AppendJustPressedKeys(nil) {
				switch k {
				case ebiten.KeyArrowUp, ebiten.KeyArrowDown, ebiten.KeyArrowLeft, ebiten.KeyArrowRight,
					ebiten.KeyBackspace, ebiten.KeyEnter:
					out = append(out, k.String())
				}
			}

			return out
		},
		text: func() string { return string(ebiten.AppendInputChars(nil)) },
	}
}

// attachLive installs the live sessions on a window page. A non-nil hub also
// lets the snapshot endpoint ask the window for its current picture.
func attachLive(page *ownframe.Page, hub *snapshotHub) {
	sessions := liveSessions()
	if hub != nil {
		sessions = append(sessions, &snapshotSession{hub: hub})
	}

	if hub != nil {
		hub.wake = func() { page.WantFrameAt(time.Now()) }
	}

	r := &liveRunner{page: page, sessions: sessions, clock: wallClock{}, src: ebitenLiveSource(page)}
	r.reportErr = func(err error) {
		fmt.Fprintln(os.Stderr, "blinkless-all: patch refused, the page keeps running:", err)
	}
	r.attach()
}
