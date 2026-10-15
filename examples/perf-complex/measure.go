package main

import (
	"runtime"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/page"
)

type sample struct {
	ElapsedNS          int64
	AllocBytes, Allocs uint64
	Stats              host.Stats
}
type dump struct {
	Mode, GoVersion, GOOS, GOARCH string
	Samples                       []sample
	ProfileIterations             int
}

func measure(mode string, p *page.Page, v view) (dump, *page.Page, error) {
	d := dump{Mode: mode, GoVersion: runtime.Version(), GOOS: runtime.GOOS, GOARCH: runtime.GOARCH}
	for i := 0; i < 7; i++ {
		if mode == "initial" {
			var err error
			var fresh *page.Page
			fresh, v, err = fixture(false)
			if err == nil {
				p = fresh
			}
			if err != nil {
				return d, p, err
			}
		}
		prepare(mode, p, &v, i)
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		start := time.Now()
		if err := redraw(p); err != nil {
			return d, p, err
		}
		elapsed := time.Since(start).Nanoseconds()
		runtime.ReadMemStats(&after)
		d.Samples = append(d.Samples, sample{elapsed, after.TotalAlloc - before.TotalAlloc, after.Mallocs - before.Mallocs, p.Stats()})
	}
	start := time.Now()
	for time.Since(start) < 2*time.Second {
		prepare(mode, p, &v, d.ProfileIterations)
		if mode == "initial" {
			var err error
			var fresh *page.Page
			fresh, v, err = fixture(false)
			if err == nil {
				p = fresh
			}
			if err != nil {
				return d, p, err
			}
		}
		if err := redraw(p); err != nil {
			return d, p, err
		}
		d.ProfileIterations++
	}
	return d, p, nil
}

func prepare(mode string, p *page.Page, v *view, i int) {
	if mode == "data" {
		v.Revision++
		p.SetData(*v)
	}
	if mode == "resize" {
		p.SetSize(1440+(i%2)*80, 1000)
	}
}
