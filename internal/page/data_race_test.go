package page

import (
	"context"
	"sync"
	"testing"
)

// TestBindWriteRacesSetData runs the bound-field write against SetData; under
// -race it fails when bindWrite touches p.data or formDirty without dataMu.
func TestBindWriteRacesSetData(t *testing.T) {
	t.Parallel()

	c := bindCtl("input", "text")
	c.Bind, c.Value = "Text", "v"
	p := &Page{data: &bindData{}}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for range 500 {
			p.SetData(&bindData{})
		}
	}()
	go func() {
		defer wg.Done()
		for range 500 {
			bindWrite(p, c)
		}
	}()
	wg.Wait()
}

// TestScrollAndReplaceHTMLRaceSetData covers formDirty and lastSyncBody.
func TestScrollAndReplaceHTMLRaceSetData(t *testing.T) {
	t.Parallel()

	p, err := New(Config{HTML: `<body><p>a</p></body>`, Width: 200, Height: 100})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for range 300 {
			p.SetData(nil)
		}
	}()
	go func() {
		defer wg.Done()
		for i := range 300 {
			p.SetScrollOffset(0, i%5)
			if err := ReplaceHTML(context.Background(), p, `<body><p>b</p></body>`, true); err != nil {
				t.Error(err)

				return
			}
		}
	}()
	wg.Wait()
}
