package main

import (
	"context"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/page"
)

// channelMax scales 0..255 hover tints down to the 0..1 op channels.
const channelMax = 255.0

// Hover tints in 0..255 channels: even rows, odd rows, and the hovered row.
var (
	hoverBase = [3]float64{20, 24, 32}
	hoverAlt  = [3]float64{29, 37, 50}
	hoverNext = [3]float64{52, 65, 89}
)

func installHover(p *page.Page) error {
	if err := p.SetTheme(`.row{background:#141820}`); err != nil {
		return err
	}
	cache := rowPaintCache{rows: map[string]*ownframe.DisplayOp{}}
	p.Handle(ownframe.Handlers{Hover: func(_ context.Context, old, next ownframe.Box) (bool, error) {
		allowed := func(id string) bool {
			return id == "" || id == "grid" || id == "top-space" || id == "bottom-space" || strings.HasPrefix(id, "row-")
		}
		if !allowed(old.ID) || !allowed(next.ID) {
			return false, nil
		}
		cache.bind(p)
		if op := cache.rows[old.ID]; op != nil {
			n, _ := strconv.Atoi(strings.TrimPrefix(old.ID, "row-"))
			color := hoverBase
			if n%2 == 0 {
				color = hoverAlt
			}
			op.R, op.G, op.B = color[0]/channelMax, color[1]/channelMax, color[2]/channelMax
		}
		if op := cache.rows[next.ID]; op != nil {
			op.R, op.G, op.B = hoverNext[0]/channelMax, hoverNext[1]/channelMax, hoverNext[2]/channelMax
		}
		return true, nil
	}})
	return nil
}
