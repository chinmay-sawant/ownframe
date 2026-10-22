package ui

import (
	"fmt"
	"strings"

	"github.com/chinmay-sawant/ownframe/examples/system-monitor/domain"
)

// percentScale turns a 0..100 percent reading back into a 0..1 fraction.
const percentScale = 100

// toSummary converts one merged sample into the four overview readings.
func toSummary(s domain.Sample) Summary {
	return Summary{
		At:  s.Stamp.At,
		Gap: s.Gap,
		Readings: []Reading{
			cpuReading(s),
			memReading(s),
			diskReading(s),
			netReading(s),
		},
	}
}

// cpuReading shows percent of all cores plus the load averages.
func cpuReading(s domain.Sample) Reading {
	pct, ok := s.CPUPercent.Get()
	frac := pct / percentScale
	r := Reading{ID: "cpu", Value: frac, OK: ok, Scale: 1, Text: formatPercent(frac, ok)}

	load := make([]string, 0, len(s.Load))

	for _, v := range s.Load {
		load = append(load, v.String())
	}

	sub := "load " + strings.Join(load, " ")
	if len(s.CPUCores) > 0 {
		sub = fmt.Sprintf("%d cores \u00b7 %s", len(s.CPUCores), sub)
	}

	r.Sub = sub

	return r
}

// memReading shows used percent with used and total bytes.
func memReading(s domain.Sample) Reading {
	pct, ok := s.Mem.UsedPercent.Get()
	frac := pct / percentScale
	r := Reading{ID: "mem", Value: frac, OK: ok, Scale: 1, Text: formatPercent(frac, ok)}

	used, uok := s.Mem.Used.Get()
	total, tok := s.Mem.Total.Get()

	switch {
	case uok && tok:
		r.Sub = formatBytes(used, true) + " / " + formatBytes(total, true)
	case tok:
		r.Sub = formatBytes(total, true) + " total"
	}

	return r
}
