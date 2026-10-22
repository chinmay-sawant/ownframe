package login_test

import (
	"bytes"
	"context"
	"image/png"
	"math"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/login/login"
)

func TestLoginFillsAndCentersTheFrame(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)

	assertFilledFrame(t, app)

	app.SetSize(720, 480)
	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	assertFilledFrame(t, app)
}

func assertFilledFrame(t *testing.T, app *login.App) {
	t.Helper()

	frameW, frameH := app.Size()
	body := boxByTag(t, app, "body")
	if math.Abs(body.H-float64(frameH)) > 1 || math.Abs(body.W-float64(frameW)) > 1 {
		t.Fatalf("body = %.1f x %.1f, frame = %d x %d", body.W, body.H, frameW, frameH)
	}

	card := boxByID(t, app, "card")
	midX := card.X + card.W/2
	midY := card.Y + card.H/2
	if math.Abs(midX-float64(frameW)/2) > 2 || math.Abs(midY-float64(frameH)/2) > 2 {
		t.Fatalf("card center = %.1f, %.1f, frame center = %.1f, %.1f",
			midX, midY, float64(frameW)/2, float64(frameH)/2)
	}

	img := decodePNG(t, app)
	if img.Bounds().Dx() != frameW || img.Bounds().Dy() != frameH {
		t.Fatalf("png = %d x %d, frame = %d x %d", img.Bounds().Dx(), img.Bounds().Dy(), frameW, frameH)
	}

	if got := rgbaAt(img, 4, 4); got != pageBackground {
		t.Fatalf("page corner = #%02x%02x%02x, want #%02x%02x%02x",
			got.R, got.G, got.B, pageBackground.R, pageBackground.G, pageBackground.B)
	}

	if got := rgbaAt(img, int(card.X)+8, int(card.Y)+8); got != cardBackground {
		t.Fatalf("card padding = #%02x%02x%02x, want #%02x%02x%02x",
			got.R, got.G, got.B, cardBackground.R, cardBackground.G, cardBackground.B)
	}
}

func TestSetSizeClampsAndRedraws(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)
	before := app.Generation()

	app.SetSize(10, 10)
	width, height := app.Size()
	if width != login.MinWidth || height != login.MinHeight {
		t.Fatalf("clamped size = %d x %d", width, height)
	}

	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	cfg, err := png.DecodeConfig(bytes.NewReader(app.PNG()))
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Width != login.MinWidth || cfg.Height != login.MinHeight {
		t.Fatalf("png = %d x %d", cfg.Width, cfg.Height)
	}

	if app.Generation() <= before {
		t.Fatalf("generation = %d, before = %d", app.Generation(), before)
	}
}

func TestFieldHitBoxStaysInsideTheCard(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)
	card := boxByID(t, app, "card")
	email := boxByID(t, app, "email")

	if email.X < card.X || email.Y < card.Y {
		t.Fatalf("email origin = %.1f, %.1f, card origin = %.1f, %.1f", email.X, email.Y, card.X, card.Y)
	}

	if email.X+email.W > card.X+card.W+1 || email.Y+email.H > card.Y+card.H+1 {
		t.Fatalf("email extends outside the card")
	}

	click(t, ctx, app, "email", "")
	if got := app.Page().FocusedField(); got != "email" {
		t.Fatalf("focus = %q", got)
	}
}
