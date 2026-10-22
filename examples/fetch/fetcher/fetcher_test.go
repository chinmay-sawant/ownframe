package fetcher_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/fetch/fetcher"
)

func TestFetchExample(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(echo))
	t.Cleanup(srv.Close)

	ctx := ownframe.AllowPrivate(context.Background())
	app, err := fetcher.New()
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if len(app.Page().Testing().PNG()) == 0 {
		t.Fatal("PNG is empty after Redraw")
	}

	if err := app.Get(ctx, srv.URL); err != nil {
		t.Fatal(err)
	}

	if got := app.View().Status; !strings.Contains(got, "200") || !strings.Contains(got, "hello") {
		t.Fatalf("GET Status = %q, want 200 and hello", got)
	}

	if err := app.Post(ctx, srv.URL); err != nil {
		t.Fatal(err)
	}

	if got := app.View().Status; !strings.Contains(got, "200") || !strings.Contains(got, "echo:note=hello") {
		t.Fatalf("POST Status = %q, want the echoed body", got)
	}

	err = app.Get(ctx, "ftp://x")
	if !errors.Is(err, ownframe.ErrScheme) {
		t.Fatalf("Get(ftp) err = %v, want ErrScheme", err)
	}

	if got := app.View().Status; !strings.Contains(got, "scheme") {
		t.Fatalf("bad Status = %q, want a scheme error", got)
	}
}

// echo returns hello on GET and echoes the POST body.
func echo(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		body, _ := io.ReadAll(r.Body)
		_, _ = w.Write([]byte("echo:" + string(body)))

		return
	}

	_, _ = w.Write([]byte("hello"))
}
