package ownframe

import (
	"context"
	"errors"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	ebitenmobile "github.com/hajimehoshi/ebiten/v2/mobile"

	pagepkg "github.com/chinmay-sawant/ownframe/internal/page"
	"github.com/chinmay-sawant/ownframe/internal/window"
)

// Run opens a window and blocks until it closes.
// On a WebAssembly build the window is the browser canvas.
// Run draws the page first when it has not been drawn yet.
func Run(ctx context.Context, page *Page) (err error) {
	return RunWithOptions(ctx, page, WindowOptions{})
}

// BindMobile registers the page with Ebitengine's mobile view.
// Call it from the package that ebitenmobile bind compiles.
// Do not call Run from that package. BindMobile draws the page first
// when it has not been drawn yet. Responsive pages follow the native
// viewport instead of the desktop minimum; LockView retains its canvas.
func BindMobile(ctx context.Context, page *Page) (err error) {
	defer savePanic(page, &err)

	if err := pagepkg.PrepareMobile(ctx, innerPage(page)); err != nil {
		return err
	}

	ensureAudio()

	ebiten.SetTPS(ebiten.SyncWithFPS)
	ebitenmobile.SetGame(window.NewGame(ctx, innerPage(page)))

	return nil
}

func prepare(ctx context.Context, p *Page) error {
	return pagepkg.Prepare(ctx, innerPage(p))
}

func savePanic(page *Page, errp *error) {
	rec := recover()
	if rec == nil {
		return
	}
	var recErr error
	if err, ok := rec.(error); ok {
		recErr = err
	} else {
		recErr = fmt.Errorf("%v", rec)
	}
	*errp = reportError(page, errors.Join(ErrPanicked, recErr))
}
