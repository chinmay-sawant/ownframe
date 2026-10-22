package window

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
	"github.com/hajimehoshi/ebiten/v2"
)

const repaintWidth, repaintHeight, repaintStageCount = 600, 300, 6

// TestWindowRepaintPixels compares the retained window framebuffer against a
// fresh full replay. Both painting and readback run inside Ebiten's Draw loop.
func TestWindowRepaintPixels(t *testing.T) {
	if os.Getenv("OWNFRAME_FRAME_GPU_TEST") != "1" {
		t.Skip("run make graphics-check for an isolated graphics session")
	}
	ctx := context.Background()
	p, err := page.New(page.Config{
		Width: repaintWidth, Height: repaintHeight,
		HTML: `<body style="margin:0;background:white"><div style="height:900px">
<div style="position:absolute;left:240px;top:70px;width:80px;height:80px;background:{{.Color}};border-radius:8px"></div>
{{if .Visible}}<div style="position:absolute;left:40px;top:160px;width:60px;height:60px;background:blue"></div>{{end}}
<div style="position:fixed;z-index:10;left:0;top:0;width:600px;height:24px;background:black"></div>
</div></body>`,
	})
	if err != nil {
		t.Fatal(err)
	}
	view := repaintView{Color: "red", Visible: true}
	p.SetData(view)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	g := &repaintPixelGame{page: p, shell: newShell(ctx, p), view: view}
	defer g.shell.cancel()
	ebiten.SetWindowSize(repaintWidth, repaintHeight)
	ebiten.SetScreenClearedEveryFrame(false)
	if err := ebiten.RunGame(g); err != nil {
		t.Fatal(err)
	}
	if g.err != nil {
		t.Fatal(g.err)
	}
	if g.step != repaintStageCount {
		t.Fatalf("only completed %d of %d stages", g.step, repaintStageCount)
	}
}

type repaintView struct {
	Color   string
	Visible bool
}

// Repeating a clean frame catches accidental clearing of untouched pixels.
func repaintStageName(i int) string {
	return [...]string{"initial", "color across tiles", "remove element", "scroll with fixed header", "clean frame", "change while scrolled"}[i]
}

type repaintPixelGame struct {
	page     *page.Page
	shell    *shell
	view     repaintView
	expected *ebiten.Image
	step     int
	err      error
}

func (g *repaintPixelGame) Layout(int, int) (int, int) { return repaintWidth, repaintHeight }
func (g *repaintPixelGame) Update() error {
	if g.err != nil || g.step == repaintStageCount {
		return ebiten.Termination
	}
	return nil
}

func (g *repaintPixelGame) Draw(dst *ebiten.Image) {
	if g.err != nil || g.step == repaintStageCount {
		return
	}
	if err := g.shell.syncImage(); err != nil {
		g.err = err
		return
	}
	g.shell.partial.focus = ebiten.IsFocused()
	g.shell.Draw(dst)
	if g.shell.partial.tiles == nil {
		w, h := g.shell.contentSize()
		g.err = fmt.Errorf("%s bypassed tiles: content=%dx%d screen=%dx%d fallback=%v stretched=%v display=%v safe=%v", repaintStageName(g.step), w, h, g.shell.screenW, g.shell.screenH, g.shell.fallback, g.shell.stretched(), g.shell.display != nil, g.shell.bufferOrderSafe())
		return
	}
	if g.step == 1 && g.shell.partial.mode != repaintRect {
		g.err = fmt.Errorf("color change did not take partial repaint path")
		return
	}
	if g.expected == nil {
		g.expected = ebiten.NewImage(repaintWidth, repaintHeight)
	}
	g.shell.paintFullReplay(g.expected, g.shell.display)
	g.shell.drawScrollbars(g.expected)
	actual, expected := make([]byte, repaintWidth*repaintHeight*4), make([]byte, repaintWidth*repaintHeight*4)
	dst.ReadPixels(actual)
	g.expected.ReadPixels(expected)
	if !bytes.Equal(actual, expected) {
		g.err = fmt.Errorf("%s: retained window differs from full replay", repaintStageName(g.step))
		return
	}
	g.prepareNext()
	g.step++
}

func (g *repaintPixelGame) prepareNext() {
	const removeStage, scrollStage, cleanStage = 1, 2, 4
	switch g.step {
	case 0:
		g.view.Color = "green"
	case removeStage:
		g.view.Visible = false
	case scrollStage:
		const offset = 90
		g.shell.scrollY = offset
		g.page.SetScrollOffset(0, offset)
		return
	case cleanStage:
		g.view.Color = "purple"
	default:
		return
	}
	g.page.SetData(g.view)
	g.err = g.page.Redraw(context.Background())
}
