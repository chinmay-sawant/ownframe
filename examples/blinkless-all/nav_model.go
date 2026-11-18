package main

import (
	"context"

	"github.com/chinmay-sawant/blinkless/css"
	"github.com/chinmay-sawant/blinkless/html"
	"github.com/chinmay-sawant/blinkless/layout"
)

// navItem is one element as the navigation and reading code sees it: its id,
// its place in the tree, its border box in CSS pixels, and the declared text
// the engine exported for it.
type navItem struct {
	ID       string
	Tag      string
	Parent   int
	X, Y     float64
	W, H     float64
	Declared map[string]string
}

// navModel is a display list reduced to navItems, in document order.
type navModel struct {
	Items []navItem
}

// navModelOf reads the boxes of one display list.
func navModelOf(d *layout.Display) navModel {
	m := navModel{Items: make([]navItem, len(d.Boxes))}

	for i, b := range d.Boxes {
		m.Items[i] = navItem{ID: b.ID, Tag: b.Tag, Parent: b.ParentIndex, X: b.X, Y: b.Y, W: b.W, H: b.H, Declared: b.Declared}
	}

	return m
}

// navLayout lays out one HTML document at the given width and returns its model.
func navLayout(ctx context.Context, page string, width int) (navModel, error) {
	doc, err := html.Parse([]byte(page))
	if err != nil {
		return navModel{}, err
	}

	styled, err := css.Apply(ctx, doc, css.Options{WidthPx: width, HeightPx: navPageHeight, Media: "screen"})
	if err != nil {
		return navModel{}, err
	}

	d, err := layout.DisplayList(ctx, styled)
	if err != nil {
		return navModel{}, err
	}

	return navModelOf(d), nil
}

// ByID returns the index of the first item with this id, or -1.
func (m navModel) ByID(id string) int {
	if id == "" {
		return -1
	}

	for i, it := range m.Items {
		if it.ID == id {
			return i
		}
	}

	return -1
}

// declared returns the nearest value of prop on item i or an ancestor, and
// the index that carries it. Only the properties that inherit should use it
// without thinking; the caller chooses.
func (m navModel) declaredUp(i int, prop string) (string, int) {
	for n := 0; i >= 0 && i < len(m.Items) && n <= len(m.Items); n++ {
		if v := m.Items[i].Declared[prop]; v != "" {
			return v, i
		}

		i = m.Items[i].Parent
	}

	return "", -1
}

// isUnder reports whether item i is container or sits inside it.
func (m navModel) isUnder(i, container int) bool {
	for n := 0; i >= 0 && n <= len(m.Items); n++ {
		if i == container {
			return true
		}

		i = m.Items[i].Parent
	}

	return false
}

// navPageHeight is the height of the nav demo page, in CSS pixels.
const navPageHeight = 600
