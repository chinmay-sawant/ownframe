package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/render"
	"github.com/chinmay-sawant/ownframe/internal/replay"
)

const backdropFilterCatalogCase = "backdrop-filter"

func TestCatalogBackdropFilterReachesOwnFrameDisplay(t *testing.T) {
	display, err := render.DisplayList(context.Background(), `<body style="margin:0;background:#203040">`+
		`<div style="position:absolute;left:20px;top:20px;width:100px;height:60px;`+
		`backdrop-filter:blur(12px) saturate(1.25);background:rgba(255,255,255,0.2)"></div></body>`, 160, 120)
	if err != nil {
		t.Fatal(err)
	}
	renderer := &replay.Renderer{}
	if !renderer.HasBackdropFilter(display) {
		t.Fatal("OwnFrame did not mark the display for full backdrop replay")
	}

	for i := range display.Ops {
		group := display.Ops[i].Group()
		if group != nil && group.BackdropFilter == "blur(12px) saturate(1.25)" {
			return
		}
	}

	t.Fatal("backdrop-filter did not reach OwnFrame's display list")
}
