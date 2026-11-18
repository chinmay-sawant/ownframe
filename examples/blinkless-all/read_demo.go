package main

import (
	"fmt"
	"html"
	"strings"
)

const (
	readSnipFlow = `<div id="r-flex" style="display:flex;flex-direction:row-reverse;reading-flow:flex-visual;width:300px;background:#f1efe7">
  <div id="r-a" style="width:60px;height:24px">A</div><div id="r-b" style="width:60px;height:24px">B</div><div id="r-c" style="width:60px;height:24px">C</div>
 </div>
 <div id="r-src" style="reading-flow:source-order;position:relative;height:30px;width:300px;background:#eef2f7;margin-top:4px">
  <div id="s-a" style="position:absolute;left:100px;top:0;width:60px;height:24px">s-a</div>
  <div id="s-b" style="position:absolute;left:0;top:0;width:60px;height:24px">s-b</div>
 </div>
 <div id="r-rows" style="reading-flow:grid-rows;position:relative;height:70px;width:300px;background:#eef7ee;margin-top:4px">
  <div id="p1" style="position:absolute;left:100px;top:0;width:60px;height:24px">p1</div>
  <div id="p2" style="position:absolute;left:0;top:0;width:60px;height:24px">p2</div>
  <div id="p3" style="position:absolute;left:0;top:40px;width:60px;height:24px">p3</div>
 </div>
 <div id="r-cols" style="reading-flow:grid-columns;position:relative;height:70px;width:300px;background:#f7eeee;margin-top:4px">
  <div id="q1" style="position:absolute;left:100px;top:0;width:60px;height:24px">q1</div>
  <div id="q2" style="position:absolute;left:0;top:0;width:60px;height:24px">q2</div>
  <div id="q3" style="position:absolute;left:0;top:40px;width:60px;height:24px">q3</div>
 </div>`
	readSnipOrder = `<div id="r-ord" style="reading-flow:source-order;width:300px;background:#f7f3e3">
  <div id="o1" style="reading-order:3">first in source, reading-order 3</div>
  <div id="o2" style="reading-order:1">second in source, reading-order 1</div>
  <div id="o3" style="reading-order:2">third in source, reading-order 2</div>
 </div>`
	readSnipOverlay = `<div id="ov-auto" style="overlay:auto">overlay auto</div>
 <div id="ov-none" style="overlay:none">overlay none</div>`
	readSnipAttach = `<div id="bg-local" style="background-attachment:local, fixed">attachment local, fixed</div>`
	readSnipTbd    = `<div id="bg-tbd" style="background-tbd:url(a.png) top left / 10px 10px">background-tbd</div>`
)

const readDemoHTML = readSnipFlow + readSnipOrder + readSnipOverlay + readSnipAttach + readSnipTbd

const readHeadCSS = navHeadCSS

// readIDs lists the direct children of parent in reading order.
func readIDs(m navModel, parent string) []string {
	root := m.ByID(parent)

	var ids []string

	for _, i := range m.readingSequence() {
		if m.Items[i].Parent == root {
			ids = append(ids, m.Items[i].ID)
		}
	}

	return ids
}

// readChips draws numbered chips for one container's reading order.
func readChips(m navModel, parent string) string {
	var b strings.Builder

	b.WriteString(`<div>` + html.EscapeString(parent+" ("+m.Items[m.ByID(parent)].Declared["reading-flow"]+") "))

	for n, id := range readIDs(m, parent) {
		b.WriteString(chip("hit", fmt.Sprintf("%d %s", n+1, id)))
	}

	return b.String() + `</div>`
}

func readDemoFlow() string {
	return navShow(readSnipFlow, func(m navModel) string {
		var b strings.Builder
		for _, g := range []string{"r-flex", "r-src", "r-rows", "r-cols"} {
			b.WriteString(readChips(m, g))
		}

		return b.String() + resLine("flex items arrive in layout order already; rows and columns reorder the positioned boxes.")
	})
}

func readDemoOrder() string {
	return navShow(readSnipOrder, func(m navModel) string {
		return readChips(m, "r-ord")
	})
}

func readDemoOverlay() string {
	return navShow(readSnipOverlay, func(m navModel) string {
		return chip("hit", fmt.Sprintf("ov-auto closing: top layer %v", readTopLayer(false, m.Items[m.ByID("ov-auto")].Declared["overlay"]))) +
			chip("no", fmt.Sprintf("ov-none closing: top layer %v", readTopLayer(false, m.Items[m.ByID("ov-none")].Declared["overlay"])))
	})
}

func readDemoAttach() string {
	return navShow(readSnipAttach, func(m navModel) string {
		return resLine("stored %q. Page scrolled 100px, box scrolled 30px: local shifts %g, fixed shifts %g.",
			m.Items[m.ByID("bg-local")].Declared["background-attachment"],
			readBackgroundShift("local", readPageScroll, readBoxScroll), readBackgroundShift("fixed", readPageScroll, readBoxScroll)) +
			resLine("the engine still paints every layer as scroll")
	})
}

func readDemoTbd() string {
	return navShow(readSnipTbd, func(m navModel) string {
		return resLine("stored as %q; no defined behavior", m.Items[m.ByID("bg-tbd")].Declared["background-tbd"])
	})
}

// readPageScroll is how far the demo page is scrolled, in CSS pixels.
const readPageScroll = 100

// readBoxScroll is how far the demo box is scrolled, in CSS pixels.
const readBoxScroll = 30
