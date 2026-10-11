package cat

import (
	"context"
	"fmt"
	"math"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/desktop-cat/assets"
)

type animation struct {
	page    *ownframe.Page
	files   []string
	current int
	cycle   bool
	random  bool
	step    int
	shape   *silhouette
	display *ownframe.Display
	image   *ownframe.DisplayOp
	y       float64
}

func (a *animation) load(index int) error {
	data, err := assets.Cats.ReadFile(a.files[index])
	if err != nil {
		return err
	}

	shape, err := makeSilhouette(data)
	if err != nil {
		return err
	}
	a.shape = shape
	a.page.SetImage("companion", data)
	a.current = index

	return nil
}

func (a *animation) paint(ctx context.Context, seconds float64) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if a.cycle {
		index := a.nextIndex(seconds)
		if index != a.current {
			if err := a.load(index); err != nil {
				return err
			}

			if err := a.page.Redraw(ctx); err != nil {
				return err
			}
		}
	}

	d := a.page.Display()
	if d == nil {
		return fmt.Errorf("desktop cat: the PNG template requires display-list replay")
	}

	if a.display != d {
		if err := a.bind(d); err != nil {
			return err
		}
	}

	a.image.Y = a.y + math.Sin(seconds*2.4)*2*d.PointsPerPixel

	return nil
}
