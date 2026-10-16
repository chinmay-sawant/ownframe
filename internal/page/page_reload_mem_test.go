package page

import (
	"context"
	"testing"
)

func TestPollReloadMemorySource(t *testing.T) {
	ctx := context.Background()
	fake := newMemSource()
	path := "mem://one.html"
	fake.write(path, []byte(`<p id="p">one</p>`))
	installSourceReader(fake)
	defer installSourceReader(nil)
	p, err := New(Config{File: path, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	gen := p.Generation()
	if changed, err := p.PollReload(ctx); changed || err != nil {
		t.Fatalf("clean poll = %v, %v", changed, err)
	}
	fake.write(path, []byte(`<p id="p">two words longer</p>`))
	changed, err := p.PollReload(ctx)
	if !changed || err != nil {
		t.Fatalf("dirty poll = %v, %v", changed, err)
	}
	if p.HTML() != `<p id="p">two words longer</p>` {
		t.Fatalf("HTML = %q", p.HTML())
	}
	if p.Generation() != gen+1 {
		t.Fatalf("gen = %d, want %d", p.Generation(), gen+1)
	}
}
