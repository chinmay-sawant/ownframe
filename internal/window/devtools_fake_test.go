package window

import (
	"context"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// devScreen is a screen double with an inspector and a click count.
type devScreen struct {
	*fakeScreen
	on      bool
	clicked int
	hovers  int
	draw    time.Duration
}

func (d *devScreen) DevTools() bool              { return d.on }
func (d *devScreen) SetDevTools(on bool)         { d.on = on }
func (d *devScreen) Stats() host.Stats           { return host.Stats{LastDraw: d.draw} }
func (d *devScreen) SetDrawTime(v time.Duration) { d.draw = v }

func (d *devScreen) Click(context.Context, float64, float64) error {
	d.clicked++

	return nil
}

func (d *devScreen) Hover(context.Context, float64, float64) error {
	d.hovers++

	return nil
}

func newDevScreen() *devScreen {
	return &devScreen{
		fakeScreen: &fakeScreen{
			width:  800,
			height: 600,
			boxes:  []layout.Box{{ID: "known", Tag: "div", X: 10, Y: 10, W: 100, H: 40}},
		},
		on: true,
	}
}

func newDevShell(d *devScreen) *shell {
	return &shell{
		app:     d,
		ctx:     context.Background(),
		screenW: 1920,
		screenH: 1080,
		dev:     devState{on: d.on, available: d.on, watch: newKeyWatch()},
	}
}
