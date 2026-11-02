package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestCatalogBackdropFilterReachesOwnFrameDisplay(t *testing.T) {
	p, err := ownframe.New(ownframe.Config{HTML: `<body style="margin:0;background:#203040">` +
		`<div style="position:absolute;left:20px;top:20px;width:100px;height:60px;` +
		`backdrop-filter:blur(12px) saturate(1.25);background:rgba(255,255,255,0.2)"></div></body>`, Width: 160, Height: 120})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	display := p.Probe().Display()
	if display == nil {
		t.Fatal("OwnFrame did not produce a display list")
	}
	found := false
	for i := range display.Ops {
		group := display.Ops[i].Group()
		if group == nil || group.BackdropFilter == "" {
			continue
		}
		found = true
		if group.BackdropFilter == "blur(12px) saturate(1.25)" {
			return
		}
	}
	if !found {
		t.Fatal("OwnFrame did not mark the display for full backdrop replay")
	}

	t.Fatal("backdrop-filter did not reach OwnFrame's display list")
}
