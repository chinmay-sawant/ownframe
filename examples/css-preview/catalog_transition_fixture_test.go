package main

// These transition cases exercise csstransition directly. The automatic
// transition driver covers the colour and opacity cards (see
// catalog_transition_automatic_test.go); the left and width cases stay here
// until the transition track's geometry phase lands.

import (
	"context"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
	"github.com/chinmay-sawant/ownframe/internal/csstransition"
)

type transitionRuntimeCase struct {
	property, selectedProperty, before, after string
	at                                        time.Duration
}

// motionFixture builds one page and returns the display, the target box, its
// operation indices, and the declared paints. The transition tests compare two
// pages because a transition starts between them; the transition runtime does
// not capture a style change yet.
func motionFixture(t *testing.T, html, css string) (*ownframe.Display, ownframe.Box, []int, []paint) {
	t.Helper()
	p, err := ownframe.New(ownframe.Config{HTML: html, Theme: css, Width: 300, Height: 120})
	if err != nil {
		t.Fatal(err)
	}
	if err = p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	d := p.Probe().Display()
	b, ok := findBox(p.Boxes(), "target")
	if d == nil || !ok {
		t.Fatal("fixture did not produce target box and display")
	}
	indices := cssmotion.OperationIndices(d, b)
	initial := cssmotion.Capture(d, b)
	if len(indices) == 0 || len(initial) != len(indices) {
		t.Fatal("fixture did not produce target paint operations")
	}
	return d, b, indices, initial
}

func transitionCSS(property, value string) string {
	base := "transition:left 1000ms linear 0ms;"
	if property == "transition" || property == "-webkit-transition" {
		base = ""
	}
	return `#target{position:absolute;top:5px;left:0px;width:20px;height:20px;background-color:rgb(20,40,60);` + base + property + `:` + value + `;}`
}

func sampleTransition(t *testing.T, c transitionRuntimeCase, value string) (paint, bool) {
	t.Helper()
	style := transitionCSS(c.property, value)
	beforeHTML := `<div id="target" style="left:0px;width:20px"></div>`
	afterHTML := `<div id="target" style="left:100px;width:100px"></div>`
	if (c.property == "transition-property" || c.property == "-webkit-transition-property") && value == "width" {
		afterHTML = `<div id="target" style="left:0px;width:100px"></div>`
	}
	_, _, _, initial := motionFixture(t, beforeHTML, style)
	d, b, indices, _ := motionFixture(t, afterHTML, style)
	property := c.selectedProperty
	if c.property == "transition-property" || c.property == "-webkit-transition-property" {
		property = value
	}
	start := time.Unix(200, 0)
	p, err := csstransition.Bind(d, b, property, indices, initial, start)
	if err != nil {
		t.Fatal(err)
	}
	active := p.Step(start.Add(c.at))
	return cssmotion.Read(&d.Ops[indices[0]]), active
}
