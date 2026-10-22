// Package main runs the OwnFrame CSS property catalog example.
package main

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

func run(perf bool, width, height int) error {
	ctx := context.Background()
	v := &view{Started: map[string]bool{}}
	p, err := openCatalog(ctx, catalogOpen{
		Title: "OwnFrame CSS property catalog",
		Data:  v, Width: width, Height: height, RetainState: true,
	})
	if err != nil {
		return err
	}
	if perf {
		p.SetPerf(true)
		p.SetDevTools(true)
	}
	v.Catalog = catalogRows(nil)
	v.Rows = v.Catalog
	if err := redrawCatalog(ctx, p, v); err != nil {
		return err
	}
	v.Catalog = catalogRows(p)
	v.Rows = v.Catalog
	v.Readout = describe(p)
	if err := redrawCatalog(ctx, p, v); err != nil {
		return err
	}
	metrics, ok := measureCatalog(p)
	v.Readout = "Select a Start button.\nIts value and action appear here."
	d := newDemo(p, v)
	p.Handle(ownframe.Handlers{Click: d.click, ClickPaint: d.clickPaint})
	p.Probe().SetAfterRedraw(func() { d.afterRedraw(context.Background()) })
	if ok {
		if err := installCatalogWindow(ctx, p, v, d, metrics); err != nil {
			return err
		}
	}
	d.gen = p.Probe().Generation()
	startWalk(p, d, v, metrics)
	return ownframe.Run(ctx, p)
}
