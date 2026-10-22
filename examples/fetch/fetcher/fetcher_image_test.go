package fetcher_test

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/fetch/fetcher"
)

// redImage serves a 2x2 opaque red PNG.
func redImage(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "image/png")

	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i] = 255
		img.Pix[i+3] = 255
	}

	_ = png.Encode(w, img)
}

func TestFetchedImageBecomesBackground(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(redImage))
	t.Cleanup(srv.Close)

	ctx := ownframe.AllowPrivate(context.Background())

	app, err := fetcher.New()
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Get(ctx, srv.URL); err != nil {
		t.Fatal(err)
	}

	if got := app.View().Status; !strings.Contains(got, "image=2x2") {
		t.Fatalf("Status = %q, want image=2x2", got)
	}

	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	img, err := png.Decode(bytes.NewReader(app.Page().PNG()))
	if err != nil {
		t.Fatal(err)
	}

	got, ok := color.NRGBAModel.Convert(img.At(5, 5)).(color.NRGBA)
	if !ok {
		t.Fatal("pixel model changed")
	}
	if got.R < 200 || got.G > 60 || got.B > 60 {
		t.Fatalf("corner pixel %+v, want the fetched red", got)
	}
}
