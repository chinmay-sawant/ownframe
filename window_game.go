package ownframe

import (
	"context"

	"github.com/chinmay-sawant/ownframe/internal/window"
	"github.com/hajimehoshi/ebiten/v2"
)

// NewGame returns the Ebiten game Run and BindMobile use.
// Tests that need the window loop call this instead of reaching
// into internal/window. It does not export host.Screen.
func NewGame(ctx context.Context, page *Page) ebiten.Game {
	return window.NewGame(ctx, innerPage(page))
}
