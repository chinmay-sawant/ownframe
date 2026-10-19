package main

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
	"github.com/chinmay-sawant/ownframe/internal/csstransition"
)

type transitionRuntimeCase struct {
	property, selectedProperty, before, after string
	at                                        time.Duration
}

func transitionCSS(property, value string) string {
	base := "transition:left 1000ms linear 0ms;"
	if property == "transition" || property == "-webkit-transition" {
		base = ""
	}
	return `#target{position:absolute;top:5px;left:0px;width:20px;height:20px;background-color:rgb(20,40,60);` + base + property + `:` + value + `;}`
}

func sampleTransition(t *testing.T, c transitionRuntimeCase, value string) (cssmotion.Paint, bool) {
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
