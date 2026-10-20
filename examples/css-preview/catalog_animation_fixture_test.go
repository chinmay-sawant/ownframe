package main

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/cssanimation"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

type animationRuntimeCase struct {
	property, before, after string
	at                      time.Duration
	compareActive           bool
}

func animationCSS(property, value string) string {
	base := "animation:route 1000ms linear 1 normal both running;"
	if property == "animation" || property == "-webkit-animation" {
		base = ""
	}
	return `#target{width:40px;height:20px;background-color:rgb(20,40,60);` + base + property + `:` + value + `;}` + keyframeCSS
}

const keyframeCSS = `@keyframes route{from{transform:translateX(0px)}to{transform:translateX(100px)}}
@keyframes returnRoute{from{transform:translateX(0px)}to{transform:translateX(-100px)}}
@-webkit-keyframes route{from{transform:translateX(0px)}to{transform:translateX(100px)}}
@-webkit-keyframes returnRoute{from{transform:translateX(0px)}to{transform:translateX(-100px)}}`

func sampleAnimation(t *testing.T, c animationRuntimeCase, value string) (cssmotion.Paint, bool) {
	t.Helper()
	d, b, indices, initial := motionFixture(t, `<div id="target"></div>`, animationCSS(c.property, value))
	start := time.Unix(100, 0)
	p, err := cssanimation.Bind(d, b, indices, initial, start)
	if err != nil {
		t.Fatal(err)
	}
	active := p.Step(start.Add(c.at))
	return cssmotion.Read(&d.Ops[indices[0]]), active
}
