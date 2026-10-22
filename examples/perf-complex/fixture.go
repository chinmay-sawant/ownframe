package main

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/chinmay-sawant/ownframe"
)

//go:embed layout.html
var source string

type row struct {
	ID, Region int
	Name       string
}
type view struct {
	Revision, Top, Bottom int
	Nav                   []string
	Rows                  []row
}

// Fixture geometry in CSS pixels: row counts, row height, page size, and the
// resize-mode width step.
const (
	fixtureRows       = 480
	fixtureWindowRows = 48
	fixtureRowH       = 76
	fixtureWidth      = 1440
	fixtureHeight     = 1000
	resizeStep        = 80
	resizeAlt         = 2
)

func fixture(window bool) (*ownframe.Page, view, error) {
	v := view{Nav: []string{"Overview", "Services", "Regions", "Deployments", "Incidents", "Capacity"}}
	count := fixtureRows
	if window {
		count = fixtureWindowRows
		v.Bottom = (fixtureRows - count) * fixtureRowH
	}
	for i := 0; i < count; i++ {
		v.Rows = append(v.Rows, row{i, i % 8, fmt.Sprintf("service-%04d", i)})
	}
	p, err := ownframe.New(ownframe.Config{HTML: source, Width: fixtureWidth, Height: fixtureHeight, Perf: true})
	if err == nil {
		p.SetData(v)
	}
	return p, v, err
}

func redraw(p *ownframe.Page) error {
	if err := p.Redraw(context.Background()); err != nil {
		return err
	}
	if p.Testing().Display() == nil {
		return fmt.Errorf("fixture triggered raster fallback")
	}
	p.Testing().TakeDirty()
	return nil
}
