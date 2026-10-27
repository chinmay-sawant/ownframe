package main

import (
	"bytes"
	"hash/fnv"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"strconv"
	"strings"
	"sync"

	mermaid "github.com/bvolpato/mermaid-go-renderer"
)

const (
	diagramStripCap = 40
	diagramMinPx    = 120
	diagramBgR      = 0x0a
	diagramBgG      = 0x0d
	diagramBgB      = 0x13
	diagramBgA      = 0xff
)

// diagramPictures caches one rendered diagram by its source and column budget.
//
//nolint:gochecknoglobals // preview cache, same role as previewWidths
var diagramPictures sync.Map

type diagramPicture struct {
	strips [][]byte
	width  int
}

// appendMermaidImage paints a Mermaid fence as row-high image strips. A
// renderer error leaves the text diagram in place.
func appendMermaidImage(rows []previewRow, src []string, cols int) ([]previewRow, []int, bool) {
	source := strings.TrimSpace(strings.Join(src, "\n"))
	if source == "" {
		return nil, nil, false
	}

	maxW := int(float64(cols) * codeCellW)
	if maxW < diagramMinPx {
		maxW = diagramMinPx
	}

	picture, ok := cachedDiagram(source, maxW)
	if !ok {
		return nil, nil, false
	}

	base := len(rows)
	id := diagramID(source, maxW)
	lineRows := make([]int, len(src))

	for i, strip := range picture.strips {
		alt := ""
		if i == 0 {
			alt = "Mermaid diagram"
		}

		rows = append(rows, previewRow{
			Kind:   "diagram",
			Marker: "◇",
			Image:  "mermaid-" + id + "-" + strconv.Itoa(i) + ".png",
			ImageW: picture.width,
			Alt:    alt,
			png:    strip,
		})
	}

	return rows, lineRows, base < len(rows)
}

// cachedDiagram returns the strips for one diagram, rendering them once.
func cachedDiagram(source string, maxW int) (diagramPicture, bool) {
	key := source + "\n" + strconv.Itoa(maxW)
	if hit, ok := diagramPictures.Load(key); ok {
		picture, ok := hit.(diagramPicture)
		return picture, ok && len(picture.strips) > 0
	}

	picture, ok := renderDiagram(source, maxW)
	if !ok {
		return diagramPicture{}, false
	}

	diagramPictures.Store(key, picture)

	return picture, true
}

// renderDiagram rasterizes one Mermaid source and cuts it into editor rows.
func renderDiagram(source string, maxW int) (diagramPicture, bool) {
	svg, err := mermaid.RenderWithOptions(source, diagramOptions())
	if err != nil || strings.TrimSpace(svg) == "" {
		return diagramPicture{}, false
	}

	file, err := os.CreateTemp("", "ownframe-mermaid-*.png")
	if err != nil {
		return diagramPicture{}, false
	}

	name := file.Name()
	_ = file.Close()
	defer func() {
		_ = os.Remove(name)
	}()

	if err := mermaid.WriteOutputPNG(svg, name); err != nil {
		return diagramPicture{}, false
	}

	raw, err := os.ReadFile(name) //nolint:gosec // temp path from CreateTemp, not user input
	if err != nil {
		return diagramPicture{}, false
	}

	img, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		return diagramPicture{}, false
	}

	fitted := fitDiagramWidth(img, maxW)
	strips := diagramStrips(fitted)
	if len(strips) == 0 {
		return diagramPicture{}, false
	}

	return diagramPicture{strips: strips, width: fitted.Bounds().Dx()}, true
}

// diagramOptions uses the editor's dark page behind the diagram.
func diagramOptions() mermaid.RenderOptions {
	theme := mermaid.ModernTheme()
	theme.Background = "#0a0d13"
	theme.PrimaryColor = "#1a2333"
	theme.PrimaryBorderColor = "#7aa2f7"
	theme.PrimaryTextColor = "#c4cddd"
	theme.SecondaryColor = "#121826"
	theme.TertiaryColor = "#182033"
	theme.LineColor = "#7aa2f7"
	theme.TextColor = "#c4cddd"
	theme.EdgeLabelBackground = "#121826"
	theme.ClusterBorder = "#394253"

	opts := mermaid.DefaultRenderOptions()
	opts.Theme = theme

	return opts
}

// diagramID is a stable file name for one source at one width.
func diagramID(source string, maxW int) string {
	sum := fnv.New64a()
	_, _ = sum.Write([]byte(source))
	_, _ = sum.Write([]byte(strconv.Itoa(maxW)))

	return strconv.FormatUint(sum.Sum64(), 16)
}

// fitDiagramWidth scales a diagram down to the preview width.
func fitDiagramWidth(src image.Image, maxW int) image.Image {
	bounds := src.Bounds()
	if maxW < 1 || bounds.Dx() <= maxW {
		return src
	}

	height := bounds.Dy() * maxW / bounds.Dx()
	if height < 1 {
		height = 1
	}

	dst := image.NewNRGBA(image.Rect(0, 0, maxW, height))
	for y := 0; y < height; y++ {
		sy := bounds.Min.Y + y*bounds.Dy()/height
		for x := 0; x < maxW; x++ {
			sx := bounds.Min.X + x*bounds.Dx()/maxW
			dst.Set(x, y, src.At(sx, sy))
		}
	}

	return dst
}

// diagramStrips cuts a diagram into row-high PNG strips.
func diagramStrips(img image.Image) [][]byte {
	bounds := img.Bounds()
	fill := color.NRGBA{R: diagramBgR, G: diagramBgG, B: diagramBgB, A: diagramBgA}
	var strips [][]byte

	for y := bounds.Min.Y; y < bounds.Max.Y && len(strips) < diagramStripCap; y += rowH {
		height := rowH
		if y+height > bounds.Max.Y {
			height = bounds.Max.Y - y
		}

		strip := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), rowH))
		draw.Draw(strip, strip.Bounds(), &image.Uniform{C: fill}, image.Point{}, draw.Src)
		draw.Draw(strip, image.Rect(0, 0, bounds.Dx(), height), img, image.Pt(bounds.Min.X, y), draw.Src)

		var buf bytes.Buffer
		if err := png.Encode(&buf, strip); err != nil {
			return nil
		}

		strips = append(strips, buf.Bytes())
	}

	return strips
}
