package csstransition

import (
	"context"
	"fmt"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func positionFrame(t *testing.T, behavior, position string) (*ownframe.Display, ownframe.Box) {
	t.Helper()
	html := fmt.Sprintf(`<style>
html,body{margin:0}#target{width:50px;height:20px;background:#f00;
position:%s;left:80px;top:70px;transition-property:position;
transition-duration:1000ms;transition-timing-function:linear;
transition-behavior:%s}
</style><div id="target"></div>`, position, behavior)
	p, err := ownframe.New(ownframe.Config{HTML: html, Width: 200, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	for _, box := range p.Boxes() {
		if box.ID == "target" {
			return p.Testing().Display(), box
		}
	}
	t.Fatal("Blinkless display has no target box")
	return nil, ownframe.Box{}
}

func colorFrame(t *testing.T, color string) (*ownframe.Display, ownframe.Box) {
	t.Helper()
	html := `<style>#target{width:50px;height:20px;background-color:` + color + `;
transition-property:background-color;transition-duration:1000ms;
transition-timing-function:linear;transition-behavior:normal}</style><div id="target"></div>`
	p, err := ownframe.New(ownframe.Config{HTML: html, Width: 200, Height: 100})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	for _, box := range p.Boxes() {
		if box.ID == "target" {
			return p.Testing().Display(), box
		}
	}
	t.Fatal("Blinkless display has no target box")
	return nil, ownframe.Box{}
}
