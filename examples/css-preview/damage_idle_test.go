package main

import (
	"context"
	"testing"
)

// TestIdleTickOptsIntoFrameDirty keeps the walker from forcing a full replay
// every frame. The same opt-in must survive finished samples that remain in
// the map.
func TestIdleTickOptsIntoFrameDirty(t *testing.T) {
	ctx := context.Background()
	p, d, _ := fixturePage(t)
	p.Testing().TakeDirtyRects()

	if err := d.tick(ctx); err != nil {
		t.Fatal(err)
	}
	if !p.FrameDirty() {
		t.Fatal("empty idle tick did not opt into the dirty path")
	}

	d.samples["finished"] = &sample{active: false}
	if err := d.tick(ctx); err != nil {
		t.Fatal(err)
	}
	if !p.FrameDirty() {
		t.Fatal("finished-only tick did not opt into the dirty path")
	}
}
