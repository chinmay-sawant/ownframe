package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// TestWalkBoxRouteDuringRedraw reads the /box route while the game loop
// redraws the page. The route serves the snapshot the loop published, so
// it never reads the page while the loop swaps the boxes. Run it with the
// race detector on.
func TestWalkBoxRouteDuringRedraw(t *testing.T) {
	p, d, v := fixturePage(t)
	w := &walker{p: p, d: d, v: v}
	w.mu.Lock()
	w.publishBoxes()
	w.mu.Unlock()
	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/box?id=catalog-row-01", nil)
		for {
			select {
			case <-stop:
				return
			default:
			}
			rec.Body.Reset()
			w.handleBox(rec, req)
		}
	}()
	for i := 0; i < 5; i++ {
		if err := p.Redraw(context.Background()); err != nil {
			t.Fatal(err)
		}
		w.mu.Lock()
		w.publishBoxes()
		w.mu.Unlock()
	}
	close(stop)
	wg.Wait()
}
