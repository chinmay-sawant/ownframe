package web

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

func TestWebScrollMapsClickAndHoverToDocument(t *testing.T) {
	p, err := page.New(page.Config{Width: 300, Height: 200, RetainState: true, HTML: `<style>body{margin:0}#target{position:absolute;top:600px;width:150px;height:40px}#target:hover{background:red}</style><button id="target">Choose date</button>`})
	if err != nil {
		t.Fatal(err)
	}
	clicked := ""
	p.Handle(page.Handlers{Click: func(_ context.Context, b page.Box) error { clicked = b.ID; p.ScrollTo(0, 0); return nil }})
	if err = p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	srv := &server{app: p}
	rec := formRec(t, srv.wheel, "/wheel", "x=100&y=100&dx=0&dy=-10", "http://example.com")
	if rec.Code != http.StatusNoContent || srv.scrollY == 0 {
		t.Fatalf("scroll failed %d offset=%d", rec.Code, srv.scrollY)
	}
	y := 620 - srv.scrollY
	body := fmt.Sprintf("x=50&y=%d", y)
	rec = formRec(t, srv.hover, "/hover", body, "http://example.com")
	if rec.Code != http.StatusNoContent || p.HoverID() != "target" {
		t.Fatalf("hover failed %d id=%s", rec.Code, p.HoverID())
	}
	rec = formRec(t, srv.click, "/click", body, "http://example.com")
	if srv.scrollY != 0 {
		t.Fatal("navigation scroll request was not consumed")
	}
	if rec.Code != http.StatusSeeOther || clicked != "target" {
		t.Fatalf("click failed %d id=%s", rec.Code, clicked)
	}
}
