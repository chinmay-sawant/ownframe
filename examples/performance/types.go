package main

import (
	"time"
)

// Scenario represents one performance benchmark scenario.
type Scenario struct {
	ID            string
	Name          string
	Level         string // Low, Medium-Low, Medium, Medium-High, High, Complex, Extreme
	Description   string
	ItemCount     int
	CSSProperties []string
	HTML          string
	Data          any
	Width         int
	Height        int
	Mutate        func(any) any
}

// Result captures the measured cost of rendering a scenario.
type Result struct {
	ID             string
	Name           string
	Level          string
	ItemCount      int
	Boxes          int
	Ops            int
	ColdRender     time.Duration // Time for New + SetData + first Redraw
	WarmRedraw     time.Duration // Time for warm Redraw (cached DOM)
	MutationRedraw time.Duration // Time for SetData(mutated) + Redraw
	ResizeRelayout time.Duration // Time for SetSize(w+1, h) + Redraw
	TemplateTime   time.Duration // html/template execution stage
	StyleTime      time.Duration // CSS cascade & style computation stage
	LayoutTime     time.Duration // Box layout & text shaping stage
	PaintTime      time.Duration // Display list / drawing emission stage
	AllocFrame     uint64        // Memory allocated during sampled redraw (bytes)
}
