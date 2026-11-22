package replay

import (
	"context"
	"fmt"
	"image"

	"github.com/chinmay-sawant/ownframe/internal/bitmap"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

// shapedPageInk draws a one-line page through replay and reports its ink box
// next to the software painter's.
func shapedPageInk(style, content string) (gpu, soft image.Rectangle, err error) {
	source := `<body style="margin:0;background:#fff"><div style="position:absolute;left:10px;top:10px;color:#000;font-size:18px;` +
		style + `">` + content + `</div></body>`
	d, err := render.DisplayList(context.Background(), source, 300, 60)
	if err != nil {
		return gpu, soft, err
	}
	if !render.Replayable(d) {
		return gpu, soft, fmt.Errorf("%q should replay", style)
	}
	softImg, err := bitmap.Paint(d)
	if err != nil {
		return gpu, soft, fmt.Errorf("%q software painter: %w", style, err)
	}
	img := newParityImage(300, 60)
	defer img.Deallocate()
	DrawVisible(img, d, 0, 0)
	pixels := make([]byte, 300*60*4)
	img.ReadPixels(pixels)
	for y := 0; y < 60; y++ {
		for x := 0; x < 300; x++ {
			at := (y*300 + x) * 4
			if pixels[at+3] >= 128 && (pixels[at] < 200 || pixels[at+1] < 200 || pixels[at+2] < 200) {
				gpu = gpu.Union(image.Rect(x, y, x+1, y+1))
			}
			r, g, b, alpha := softImg.At(x, y).RGBA()
			if alpha >= 0x8000 && (r < 0xc800 || g < 0xc800 || b < 0xc800) {
				soft = soft.Union(image.Rect(x, y, x+1, y+1))
			}
		}
	}

	return gpu, soft, nil
}

// checkShapedText holds the two text shaping options replay applies natively
// to what layout and the software painter agree on:
//   - text-autospace opens the gap layout measured: the replay ink grows by
//     about one gap per ideograph and Latin boundary, and its box matches the
//     software painter's, which adds the same gap;
//   - OpenType features reach the Ebiten face: kerning off makes the replay
//     run wider, which the software painter (no shaping) cannot show.
func checkShapedText() error {
	const content = "日本語abc 日本語"
	plainGPU, _, err := shapedPageInk("text-autospace:no-autospace", content)
	if err != nil {
		return err
	}
	spacedGPU, spacedSoft, err := shapedPageInk("text-autospace:normal", content)
	if err != nil {
		return err
	}
	// The run has one ideograph and Latin boundary (語|a), a 1.5 pt gap, so
	// the ink grows by about 2 px; allow a few for glyph rounding.
	grow := spacedGPU.Dx() - plainGPU.Dx()
	if grow < 1 || grow > 8 {
		return fmt.Errorf("autospace grew the replay ink by %d px (plain %v, spaced %v)", grow, plainGPU, spacedGPU)
	}
	for _, e := range [4][2]int{
		{spacedGPU.Min.X, spacedSoft.Min.X}, {spacedGPU.Min.Y, spacedSoft.Min.Y},
		{spacedGPU.Max.X, spacedSoft.Max.X}, {spacedGPU.Max.Y, spacedSoft.Max.Y},
	} {
		if d := e[0] - e[1]; d < -3 || d > 3 {
			return fmt.Errorf("autospace ink box: replay %v, software %v", spacedGPU, spacedSoft)
		}
	}

	kerned, _, err := shapedPageInk("", "AVAVAVAV")
	if err != nil {
		return err
	}
	unkerned, _, err := shapedPageInk("font-kerning:none", "AVAVAVAV")
	if err != nil {
		return err
	}
	if unkerned.Dx() <= kerned.Dx() {
		return fmt.Errorf("kerning off did not widen the replay run: kerned %v, unkerned %v", kerned, unkerned)
	}

	return nil
}
