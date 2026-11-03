//go:build !js

package main

import (
	"context"
	"fmt"
	"image"
	"image/png"
	"os"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	scrollbarSettleFrames = 2
	scrollbarDownFrames   = 8
	scrollbarUpFrames     = 8
	scrollbarFrameCount   = scrollbarSettleFrames + scrollbarDownFrames + scrollbarUpFrames
	scrollbarExtraDrafts  = 24
	scrollbarHitInsetY    = 20
	scrollbarMinRun       = 8
	scrollbarRGBABytes    = 4
	// Sage thumb from the custom theme, with slack for the rounded-cap blend.
	scrollbarSageR        = 117
	scrollbarSageG        = 134
	scrollbarSageB        = 106
	scrollbarChannelSlack = 28
	scrollbarBlueSlack    = 32
	scrollbarMinGreen     = 90
	scrollbarShotMode     = 0o600
)

// TestScrollbarThumbStaysOnePiece scrolls the drafts list through the desktop
// window and requires the thumb to stay one vertical run. A fresh picture can
// show one thumb while the retained framebuffer keeps a piece from each step.
func TestScrollbarThumbStaysOnePiece(t *testing.T) {
	if os.Getenv("OWNFRAME_FRAME_GPU_TEST") != "1" {
		t.Skip("run make graphics-check for an isolated graphics session")
	}

	ctx := context.Background()
	app, err := newScrollbarApp(ctx)
	if err != nil {
		t.Fatal(err)
	}

	game := &scrollbarPixelGame{Game: ownframe.NewGame(ctx, app.Page()), app: app}
	ebiten.SetWindowSize(mailWindowWidth, mailWindowHeight)
	ebiten.SetScreenClearedEveryFrame(false)
	if err = ebiten.RunGame(game); err != nil {
		t.Fatal(err)
	}
	if game.err != nil {
		t.Fatal(game.err)
	}
	if game.step != scrollbarFrameCount {
		t.Fatalf("completed %d frames, want %d", game.step, scrollbarFrameCount)
	}
}

func newScrollbarApp(ctx context.Context) (*app, error) {
	a, err := newApp()
	if err != nil {
		return nil, err
	}
	a.appearance = mailThemeCustom
	a.themeDirty = true
	for i := range scrollbarExtraDrafts {
		a.messages = append(a.messages, mail{
			ID: fmt.Sprintf("extra-%d", i), ThreadID: "extra",
			From:       contact{Name: "Alex Morgan", Email: "alex@example.com"},
			To:         contact{Name: "Nora", Email: "nora@example.org"},
			Subject:    "New draft · OwnFrame Mail",
			Preview:    "This is a local draft, not a sent email.",
			Body:       "Hi Nora,\n\nA first iteration.\n\nAlex",
			ReceivedAt: time.Now(),
			MailboxIDs: map[string]bool{"drafts": true},
			Keywords:   map[string]bool{"$draft": true, "$seen": true},
		})
	}
	a.folder = "drafts"
	a.ensureSelection()
	if err = a.refresh(ctx); err != nil {
		return nil, err
	}
	return a, nil
}

type scrollbarPixelGame struct {
	ebiten.Game
	app  *app
	step int
	err  error
}

func (g *scrollbarPixelGame) Update() error {
	if g.err != nil || g.step >= scrollbarFrameCount {
		return ebiten.Termination
	}
	// The window syncs the display, then handles the wheel. A redraw from
	// that wheel has to be on screen this frame. Scrolling after Update
	// matches that order.
	if err := g.Game.Update(); err != nil {
		return err
	}
	if dy, ok := scrollbarWheel(g.step); ok {
		g.scroll(dy)
	}
	if g.err != nil {
		return ebiten.Termination
	}
	return nil
}

func (g *scrollbarPixelGame) Draw(screen *ebiten.Image) {
	if g.err != nil || g.step >= scrollbarFrameCount {
		return
	}
	g.Game.Draw(screen)
	if _, scrolled := scrollbarWheel(g.step); scrolled {
		g.checkThumb(screen)
	}
	g.step++
}

func (g *scrollbarPixelGame) Layout(int, int) (int, int) {
	return mailWindowWidth, mailWindowHeight
}

func scrollbarWheel(step int) (float64, bool) {
	if step >= scrollbarSettleFrames && step < scrollbarSettleFrames+scrollbarDownFrames {
		return -1, true
	}
	if step >= scrollbarSettleFrames+scrollbarDownFrames && step < scrollbarFrameCount {
		return 1, true
	}
	return 0, false
}

func (g *scrollbarPixelGame) scroll(dy float64) {
	box, ok := boxByID(g.app.page, "mail-list-viewport")
	if !ok {
		g.err = fmt.Errorf("frame %d: message list viewport is missing", g.step)
		return
	}
	if _, err := g.app.handleWheel(context.Background(), box.X+box.W/2, box.Y+scrollbarHitInsetY, 0, dy); err != nil {
		g.err = err
	}
}

func (g *scrollbarPixelGame) checkThumb(screen *ebiten.Image) {
	box, ok := boxByID(g.app.page, "mail-list-thumb")
	if !ok {
		g.err = fmt.Errorf("frame %d: scrollbar thumb is missing", g.step)
		return
	}
	track, ok := boxByID(g.app.page, "mail-list-viewport")
	if !ok {
		g.err = fmt.Errorf("frame %d: message list viewport is missing", g.step)
		return
	}
	bounds := screen.Bounds()
	pixels := make([]byte, scrollbarRGBABytes*bounds.Dx()*bounds.Dy())
	screen.ReadPixels(pixels)
	x0, x1 := int(box.X), int(box.X+box.W)
	// Only rows inside the list track can hold the thumb. The toolbar
	// search field sits in the same x column, and its muted placeholder
	// ink falls inside the sage thumb tolerance.
	y0, y1 := int(track.Y), int(track.Y+track.H)
	runs := tallRuns(pixels, bounds.Dx(), bounds.Dy(), x0, x1, y0, y1, scrollbarMinRun)
	if len(runs) <= 1 {
		return
	}
	_ = writeScrollbarShot(screen)
	g.err = fmt.Errorf(
		"frame %d offset %d: thumb box (%.0f,%.0f) %.0fx%.0f screen %dx%d split into %d runs %v",
		g.step,
		g.app.listScroll.Offset(),
		box.X,
		box.Y,
		box.W,
		box.H,
		bounds.Dx(),
		bounds.Dy(),
		len(runs),
		runs,
	)
}

// tallRuns are the vertical sage runs inside [x0, x1) and [yLo, yHi) that are
// at least min pixels tall. A run is kept when any column in the span sees
// it, so a rounded cap that leaves the center still counts.
func tallRuns(pixels []byte, width, height, x0, x1, yLo, yHi, min int) [][2]int {
	if x0 < 0 {
		x0 = 0
	}
	if x1 > width {
		x1 = width
	}
	if yLo < 0 {
		yLo = 0
	}
	if yHi > height {
		yHi = height
	}
	var runs [][2]int
	y := yLo
	for y < yHi {
		if !spanThumb(pixels, width, x0, x1, y) {
			y++
			continue
		}
		start := y
		for y < yHi && spanThumb(pixels, width, x0, x1, y) {
			y++
		}
		if y-start >= min {
			runs = append(runs, [2]int{start, y - 1})
		}
	}
	return runs
}

func spanThumb(pixels []byte, width, x0, x1, y int) bool {
	for x := x0; x < x1; x++ {
		if thumbPixel(pixels, width, x, y) {
			return true
		}
	}
	return false
}

func boxByID(page *ownframe.Page, id string) (ownframe.Box, bool) {
	for _, box := range page.Boxes() {
		if box.ID == id {
			return box, true
		}
	}
	return ownframe.Box{}, false
}

func thumbPixel(pixels []byte, width, x, y int) bool {
	i := (y*width + x) * scrollbarRGBABytes
	r, g, b := int(pixels[i]), int(pixels[i+1]), int(pixels[i+2])
	nearRed := absInt(r-scrollbarSageR) < scrollbarChannelSlack
	nearGreen := absInt(g-scrollbarSageG) < scrollbarChannelSlack
	nearBlue := absInt(b-scrollbarSageB) < scrollbarBlueSlack
	greenLeads := g > r && g > b && g > scrollbarMinGreen
	return nearRed && nearGreen && nearBlue && greenLeads
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func writeScrollbarShot(screen *ebiten.Image) error {
	bounds := screen.Bounds()
	img := image.NewRGBA(bounds)
	screen.ReadPixels(img.Pix)
	f, err := os.OpenFile("/tmp/scrollbar-repro.png", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, scrollbarShotMode)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return png.Encode(f, img)
}
