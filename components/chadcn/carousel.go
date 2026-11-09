package chadcn

import (
	"context"
	"strconv"

	"github.com/chinmay-sawant/ownframe/ui"
)

// CarouselOptions describes a carousel that shows one slide at a time. Index is
// the visible slide and is clamped to the slide range. Labels default to
// "Carousel" and per-slide "N of M". The host updates Index and rebuilds from
// OnPrev, OnNext, and OnSelect. OnSelect receives the dot index the host should
// show.
type CarouselOptions struct {
	ID       string
	Label    string
	Index    int
	Slides   []*ui.Element
	OnPrev   func(context.Context) error
	OnNext   func(context.Context) error
	OnSelect func(ctx context.Context, index int) error
}

// Carousel builds a labeled region with previous and next controls, one visible
// slide, and one button per slide. Both controls disable at the ends of the
// list and drop their callbacks there, so the host does not need to loop.
// The controls use ID+"-prev" and ID+"-next" when the host supplies an ID.
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
	prevDisabled := index <= 0
	nextDisabled := index < 0 || index >= count-1
	prev := ui.Button("‹").Attr("type", "button").
		Class(classes("of-carousel-control", "of-carousel-prev")).Attr("aria-label", "Previous slide")
	if prevID := partID(options.ID, "-prev"); prevID != "" {
		prev.ID(prevID)
	}
	prev = setDisabled(prev, prevDisabled)
	if options.OnPrev != nil && !prevDisabled {
		prev.OnClick(options.OnPrev)
	}
	next := ui.Button("›").Attr("type", "button").
		Class(classes("of-carousel-control", "of-carousel-next")).Attr("aria-label", "Next slide")
	if nextID := partID(options.ID, "-next"); nextID != "" {
		next.ID(nextID)
	}
	next = setDisabled(next, nextDisabled)
	if options.OnNext != nil && !nextDisabled {
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
		root.Children(carouselIndicators(options.ID, index, count, options.OnSelect))
	}
	return root
}

// carouselIndicators builds one button per slide. OnSelect chooses that index.
// Without a callback the dots stay out of the tab order; the slide label still
// announces the position.
func carouselIndicators(id string, index, count int, onSelect func(context.Context, int) error) *ui.Element {
	dots := ui.Div().Class("of-carousel-indicators").Attr("role", "group").Attr("aria-label", "Slides")
	if onSelect == nil {
		dots.Attr("aria-hidden", "true")
	}
	for dot := range count {
		button := ui.Button("").Attr("type", "button").
			Class(classes("of-carousel-dot", carouselDotClass(dot == index))).
			Attr("aria-label", "Go to slide "+strconv.Itoa(dot+1))
		if dotID := partID(id, "-dot-"+strconv.Itoa(dot)); dotID != "" {
			button.ID(dotID)
		}
		if dot == index {
			button.Attr("aria-current", "true")
		}
		if onSelect == nil {
			button.Attr("tabindex", "-1")
		} else {
			selected := dot
			button.OnClick(func(ctx context.Context) error { return onSelect(ctx, selected) })
		}
		dots.Children(button)
	}
	return dots
}

func carouselDotClass(active bool) string {
	if active {
		return "of-carousel-dot--active"
	}
	return ""
}
