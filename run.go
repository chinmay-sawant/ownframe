package ownframe

import (
	"context"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	ebitenmobile "github.com/hajimehoshi/ebiten/v2/mobile"

	"github.com/chinmay-sawant/ownframe/internal/crash"
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

	if err := pagepkg.PrepareMobile(ctx, page); err != nil {
		return err
	}

	ensureAudio()

	ebiten.SetTPS(ebiten.SyncWithFPS)
	ebitenmobile.SetGame(window.NewGame(ctx, page))

	return nil
}

func prepare(ctx context.Context, p *Page) error {
	return pagepkg.Prepare(ctx, p)
}

func savePanic(page *Page, errp *error) {
	rec := recover()
	if rec == nil {
		return
	}

	title := ""
	if page != nil {
		title = page.Title()
	}

	path, werr := crash.Write(title, fmt.Sprint(rec))
	if werr != nil {
		*errp = fmt.Errorf("ownframe: panic: %v: %s: %w", rec, path, werr)

		return
	}

	*errp = fmt.Errorf("ownframe: panic: %v: %s", rec, path)
}
