package chadcn

import (
	"context"
	"strconv"

	"github.com/chinmay-sawant/ownframe/ui"
)

// CarouselOptions describes a carousel that shows one slide at a time. Index is
// the visible slide and is clamped to the slide range. Labels default to
// "Carousel" and per-slide "N of M". The host updates Index and rebuilds from
// OnPrev and OnNext.
type CarouselOptions struct {
	ID     string
	Label  string
	Index  int
	Slides []*ui.Element
	OnPrev func(context.Context) error
	OnNext func(context.Context) error
}

// Carousel builds a labeled region with previous and next controls, one visible
// slide, and decorative dot indicators. Both controls disable at the ends of
// the list, so the host does not need to loop.
func Carousel(options CarouselOptions) *ui.Element {
	count := len(options.Slides)
	index := options.Index
	if index < 0 {
		index = 0
	}
	if index >= count {
		index = count - 1
	}
	label := options.Label
	if label == "" {
		label = "Carousel"
	}
	prev := ui.Button("‹").Attr("type", "button").
		Class(classes("of-carousel-control", "of-carousel-prev")).Attr("aria-label", "Previous slide")
	prev = setDisabled(prev, index <= 0)
	if options.OnPrev != nil {
		prev.OnClick(options.OnPrev)
	}
	next := ui.Button("›").Attr("type", "button").
		Class(classes("of-carousel-control", "of-carousel-next")).Attr("aria-label", "Next slide")
	next = setDisabled(next, index < 0 || index >= count-1)
	if options.OnNext != nil {
		next.OnClick(options.OnNext)
	}
	stage := ui.Div().Class("of-carousel-stage")
	if stageID := partID(options.ID, "-stage"); stageID != "" {
		stage.ID(stageID)
		prev.Attr("aria-controls", stageID)
		next.Attr("aria-controls", stageID)
	}
	if index >= 0 {
		stage.Children(ui.Div().Class("of-carousel-slide").Attr("role", "group").
			Attr("aria-roledescription", "slide").
			Attr("aria-label", strconv.Itoa(index+1)+" of "+strconv.Itoa(count)).
			Children(options.Slides[index]))
	}
	root := ui.Section().Class(classes("of-scope", "of-carousel")).Attr("role", "region").
		Attr("aria-roledescription", "carousel").Attr("aria-label", label)
	if options.ID != "" {
		root.ID(options.ID)
	}
	root.Children(ui.Div().Class("of-carousel-viewport").Children(prev, stage, next))
	if count > 0 {
		root.Children(carouselIndicators(index, count))
	}
	return root
}

// carouselIndicators builds decorative position dots. The slide's own label
// announces the position to assistive technology, so the dots stay hidden.
func carouselIndicators(index, count int) *ui.Element {
	dots := ui.Div().Class("of-carousel-indicators").Attr("aria-hidden", "true")
	for dot := range count {
		dots.Children(ui.Span().Class(classes("of-carousel-dot", carouselDotClass(dot == index))))
	}
	return dots
}

func carouselDotClass(active bool) string {
	if active {
		return "of-carousel-dot--active"
	}
	return ""
}
