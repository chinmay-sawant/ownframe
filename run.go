package ownframe

import (
	"context"
	"errors"
	"fmt"
	"github.com/chinmay-sawant/ownframe/internal/filepick"
	pagepkg "github.com/chinmay-sawant/ownframe/internal/page"
	"github.com/chinmay-sawant/ownframe/internal/window"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	ebitenmobile "github.com/hajimehoshi/ebiten/v2/mobile"
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

type desktopWindowHost struct{}

func (desktopWindowHost) cursorPosition() (int, int) { return window.ContentCursorPosition() }
func (desktopWindowHost) minimize()                  { window.Minimize() }
func (desktopWindowHost) toggleMaximize()            { window.ToggleMaximize() }
func (desktopWindowHost) maximized() bool            { return window.Maximized() }
func (desktopWindowHost) close()                     { window.RequestClose() }

// NewGame returns the Ebiten game Run and BindMobile use.
// Tests that need the window loop call this instead of reaching
// into internal/window. It does not export host.Screen.
func NewGame(ctx context.Context, page *Page) ebiten.Game {
	page.bindHost(desktopWindowHost{})
	return window.NewGame(ctx, innerPage(page))
}

// WindowOptions controls desktop window appearance and input.
// The zero value opens the same window as Run. Risky surfaces stay off
// unless opted in: MousePassthrough passes pointer input through painted
// pixels, so keep it false unless the app is a click-through overlay.
type WindowOptions = window.Options

// RunWithOptions opens a page with optional desktop window settings.
// A startup or window failure is stored in a crash report file first; the
// returned error keeps the original without the file path. A canceled
// context or a deadline returns that error and writes no file. A panic
// still goes through the recover in savePanic. Use LastReportPath for the file.
func RunWithOptions(ctx context.Context, page *Page, options WindowOptions) (err error) {
	defer savePanic(page, &err)

	if err := prepare(ctx, page); err != nil {
		return reportRunErr(page, err)
	}

	pagepkg.InstallPicker(innerPage(page), filepick.Pick)
	if !options.NoAudio {
		ensureAudio()
	}

	page.bindHost(desktopWindowHost{})
	defer page.bindHost(nil)

	if err := window.RunWithOptions(ctx, innerPage(page), options); err != nil {
		return reportRunErr(page, err)
	}

	return nil
}

// reportRunErr writes a crash report for a startup or window failure.
// Stopping the window by canceling its context, or by a deadline, is a
// normal return. Those errors are handed back as they are.
func reportRunErr(page *Page, err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}

	return reportError(page, err)
}

// audioSampleRate is the sample rate of the context Run creates.
const audioSampleRate = 48000

// ensureAudio creates the Ebiten audio context once. Ebiten v2.10 does not
// create one itself, and the audio package needs a context before it can
// make a player. Run, BindMobile, and Serve call this, so a page that plays
// sound through the host finds a context. In -web the sound comes out of
// this machine, the same way the native window plays it.
func ensureAudio() {
	if audio.CurrentContext() == nil {
		audio.NewContext(audioSampleRate)
	}
}
