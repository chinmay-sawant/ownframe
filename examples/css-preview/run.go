package main

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

func run(perf bool, width, height int) error {
	ctx := context.Background()
	v := newView()
	p, err := openCatalog(ctx, catalogOpen{
		Title: "OwnFrame CSS property catalog",
		Data:  v, Width: width, Height: height, RetainState: true,
	})
	if err != nil {
		return err
	}
	if perf {
		p.Window().SetPerf(true)
		p.Window().SetDevTools(true)
	}
	v.Catalog = catalogRows(nil)
	v.Rows = v.Catalog
	v.applyRows()
	if err := redrawCatalog(ctx, p, v); err != nil {
		return err
	}
	v.Catalog = catalogRows(p)
	v.Rows = v.Catalog
	v.Readout = describe(p)
	v.applyRows()
	if err := redrawCatalog(ctx, p, v); err != nil {
		return err
	}
	metrics, ok := measureCatalog(p)
	v.Readout = "Select a Start button.\nIts value and action appear here."
	d := newDemo(p, v)
	p.Handle(ownframe.Handlers{Click: d.click})
	p.Probe().SetAfterRedraw(func() { d.afterRedraw() })
	if ok {
		if err := installCatalogWindow(ctx, p, v, d, metrics); err != nil {
			return err
		}
	}
	startWalk(p, d, v, metrics)
	return ownframe.Run(ctx, p)
}
