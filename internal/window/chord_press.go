package window

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (s *shell) submitIfEnter() error {
	enter := inpututil.IsKeyJustPressed(ebiten.KeyEnter)
	numpad := inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter)
	if enter || numpad {
		return s.kbd().Submit(s.ctx)
	}

	return nil
}

func backspaceDue() bool {
	held := inpututil.KeyPressDuration(ebiten.KeyBackspace)
	if held == 1 {
		return true
	}

	if held <= backspaceRepeatAt {
		return false
	}

	return (held-backspaceRepeatAt)%backspaceRepeatEvery == 0
}
