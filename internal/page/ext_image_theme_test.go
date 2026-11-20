package page_test

import (
	"bytes"
	"context"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/page"
	"image"
	"image/color"
	"image/png"
	"reflect"
	"testing"
)

// TestPNGReturnsCopy mutates the returned bytes and requires the next PNG
// to still decode the same picture.
func TestPNGReturnsCopy(t *testing.T) {
	t.Parallel()

	screen, err := page.New(page.Config{HTML: `<div id="a">copy</div>`, Width: 120, Height: 80})
	if err != nil {
		t.Fatal(err)
	}

	if err := screen.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	first := screen.PNG()
	if first == nil {
		t.Fatal("no PNG to copy")
	}

	want := append([]byte(nil), first...)
	for i := range first {
		first[i] ^= 0xff
	}

	if got := screen.PNG(); !bytes.Equal(got, want) {
		t.Fatal("PNG cache changed after caller edit")
	}
}

// TestBoxesReturnsCopy mutates the returned slice and requires the page
// boxes to stay unchanged.
func TestBoxesReturnsCopy(t *testing.T) {
	t.Parallel()

	screen, err := page.New(page.Config{HTML: `<div id="a">copy</div>`, Width: 120, Height: 80})
	if err != nil {
		t.Fatal(err)
	}

	if err := screen.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	boxes := screen.Boxes()
	if len(boxes) == 0 {
		t.Fatal("no boxes to copy")
	}

	want := boxes[0]
	boxes[0].X = want.X + 10000
	boxes[0].W = want.W + 10000

	if got := screen.Boxes(); !reflect.DeepEqual(got[0], want) {
		t.Fatalf("boxes changed after caller edit: %+v, want %+v", got[0], want)
	}
}

func TestImageBeforeRedrawIsNil(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen, err := page.New(page.Config{
		HTML:   "<p>Hi</p>",
		Width:  320,
		Height: 400,
	})
	if err != nil {
		t.Fatal(err)
	}

	if screen.Image() != nil || screen.Display() != nil {
		t.Fatal("page drawn before redraw")
	}

	if screen.PNG() != nil {
		t.Fatal("png before redraw")
	}

	if err := screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	display := screen.Display()
	if display == nil {
		t.Fatal("display is nil")
	}

	if display.Width < 320 || display.Height < 400 {
		t.Fatalf("display = %d x %d", display.Width, display.Height)
	}

	if screen.Image() != nil {
		t.Fatal("replayable page kept a bitmap")
	}

	cfg, err := png.DecodeConfig(bytes.NewReader(screen.PNG()))
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Width != display.Width || cfg.Height < 400 {
		t.Fatalf("png = %d x %d, display = %d x %d", cfg.Width, cfg.Height, display.Width, display.Height)
	}
}

func TestPNGRejectsUnsupportedRetainedPaint(t *testing.T) {
	p, err := page.New(page.Config{HTML: `<body style="background:red">text</body>`, Width: 100, Height: 100})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	d := p.Display()
	if d == nil || len(d.Ops) == 0 {
		t.Fatal("fixture has no replay operations")
	}
	d.Ops[0].Kind = layout.DisplayKind(100)
	if png := p.PNG(); png != nil {
		t.Fatal("unsupported retained paint was replaced by source repaint")
	}
}

// TestSetImageKeepsCache checks that a new image resolver does not force a
// parse or a cascade: only the layout that reads the resolver changes.
func TestSetImageKeepsCache(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newCachePage(t, backgroundHTML)

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	p.SetImage("bg", redPNG(t))
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	stats := p.Stats()
	if stats.Parses != 1 || stats.Cascades != 1 {
		t.Fatalf("parses = %d, cascades = %d, want 1 and 1", stats.Parses, stats.Cascades)
	}
}

// TestSetImageClonesCallerBytes corrupts the caller's slice after SetImage
// and requires the painted background to stay red.
func TestSetImageClonesCallerBytes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen, err := page.New(page.Config{HTML: backgroundHTML, Width: 40, Height: 40})
	if err != nil {
		t.Fatal(err)
	}

	data := redPNG(t)
	screen.SetImage("bg", data)
	for i := range data {
		data[i] = 0
	}

	if err := screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if got := centerPixel(t, screen.PNG()); got.R < 200 || got.G > 60 || got.B > 60 {
		t.Fatalf("center pixel %+v after caller edit, want red", got)
	}
}

const backgroundHTML = `<body style="margin:0">` +
	`<div style="width:40px;height:40px;background-image:url('bg');` +
	`background-size:100% 100%"></div></body>`

// redPNG builds a 2x2 opaque red PNG.
func redPNG(t *testing.T) []byte {
	t.Helper()

	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i] = 255
		img.Pix[i+3] = 255
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}

	return buf.Bytes()
}

// centerPixel decodes a PNG and returns its center pixel.
func centerPixel(t *testing.T, data []byte) color.NRGBA {
	t.Helper()

	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	bounds := img.Bounds()

	c, ok := color.NRGBAModel.Convert(img.At(bounds.Dx()/2, bounds.Dy()/2)).(color.NRGBA)
	if !ok {
		t.Fatal("pixel model changed")
	}

	return c
}

func TestSetImagePaintsAndRemovesBackground(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	screen, err := page.New(page.Config{HTML: backgroundHTML, Width: 40, Height: 40})
	if err != nil {
		t.Fatal(err)
	}

	paint := func() color.NRGBA {
		if err := screen.Redraw(ctx); err != nil {
			t.Fatal(err)
		}

		return centerPixel(t, screen.PNGWithContext(ctx))
	}

	screen.SetImage("bg", redPNG(t))
	if got := paint(); got.R < 200 || got.G > 60 || got.B > 60 {
		t.Fatalf("center pixel %+v, want red", got)
	}

	screen.SetImage("bg", nil)
	if got := paint(); got.R < 200 || got.G < 200 || got.B < 200 {
		t.Fatalf("center pixel %+v after removal, want white", got)
	}
}

func TestSetThemeRejectsBrokenCSS(t *testing.T) {
	t.Parallel()

	p := newThemePage(t, "")
	if err := p.SetTheme(`#swatch{background:#ff0000`); err == nil {
		t.Fatal("unbalanced theme parsed without error")
	}
}

func TestThemeReachesPNG(t *testing.T) {
	t.Parallel()

	p := newThemePage(t, `#swatch{background:#ff0000}`)
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	img, err := png.Decode(bytes.NewReader(p.PNG()))
	if err != nil {
		t.Fatal(err)
	}

	r, g, b, _ := img.At(20, 20).RGBA()
	if r < 0xf000 || g > 0x0f00 || b > 0x0f00 {
		t.Fatalf("pixel = %x,%x,%x, want red", r, g, b)
	}
}

const themePage = `<html><head><style>` +
	`#swatch{width:100px;height:40px;background:#ffffff}` +
	`</style></head><body><div id="swatch"></div></body></html>`

// swatchRGB returns the swatch fill channels, 0..1. The swatch is 100x40 CSS
// pixels, 75x30 points in the display list.
func swatchRGB(t *testing.T, p *page.Page) (float64, float64, float64) {
	t.Helper()

	d := p.Display()
	if d == nil {
		t.Fatal("theme page fell back to a bitmap")
	}

	for i := range d.Ops {
		op := &d.Ops[i]
		if op.Kind == layout.DisplayOpFillRect && op.W == 75 && op.H == 30 {
			return op.R, op.G, op.B
		}
	}

	t.Fatal("no swatch fill")

	return 0, 0, 0
}

// newThemePage builds the swatch page with theme CSS.
func newThemePage(t *testing.T, theme string) *page.Page {
	t.Helper()

	p, err := page.New(page.Config{HTML: themePage, Theme: theme, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}

	return p
}

func TestConfigThemeWinsTie(t *testing.T) {
	t.Parallel()

	p := newThemePage(t, `#swatch{background:#ff0000}`)
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	if r, g, b := swatchRGB(t, p); r < 1 || g > 0 || b > 0 {
		t.Fatalf("swatch = %v,%v,%v, want red", r, g, b)
	}
}

func TestSetThemeSwitchesAndClears(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newThemePage(t, "")
	if err := p.SetTheme(`#swatch{background:#00ff00}`); err != nil {
		t.Fatal(err)
	}

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if r, g, b := swatchRGB(t, p); g < 1 || r > 0 || b > 0 {
		t.Fatalf("green swatch = %v,%v,%v", r, g, b)
	}

	if err := p.SetTheme(""); err != nil {
		t.Fatal(err)
	}

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if r, g, b := swatchRGB(t, p); r < 1 || g < 1 || b < 1 {
		t.Fatalf("cleared swatch = %v,%v,%v, want template white", r, g, b)
	}
}
