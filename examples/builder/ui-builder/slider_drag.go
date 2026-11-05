package main

import (
	"context"
	"fmt"
	"math"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/components/chadcn"
)

const (
	catalogSliderID = "catalog-slider"
	widgetsSliderID = "widgets-slider"
)

// sliderGesture is one horizontal drag on a catalog or widgets slider.
// An empty id means no slider drag is active.
type sliderGesture struct {
	id      string
	width   float64
	origin  float64
	dragged float64
}

func (a *builderApp) sliderDragBox(box ownframe.Box) bool {
	_, ok := a.sliderOptions(box.ID)
	return ok && box.W > 0
}

// claimSliderDrag takes a horizontal drag when box is a slider.
// It returns false for every other box so the scroll-thumb drag still runs.
func (a *builderApp) claimSliderDrag(_ context.Context, box ownframe.Box) (bool, error) {
	opts, ok := a.sliderOptions(box.ID)
	if !ok || opts.Disabled || box.W <= 0 || a.page == nil {
		return false, nil
	}
	a.widgets.gesture = sliderGesture{
		id:     box.ID,
		width:  box.W,
		origin: sliderOrigin(opts, box.W),
	}
	a.page.AddHandlers(ownframe.Handlers{
		DragMove: a.moveSliderDrag,
		DragEnd:  a.endSliderDrag,
	})
	return true, nil
}

// moveSliderDrag is the DragMove callback. The vertical delta is ignored.
func (a *builderApp) moveSliderDrag(ctx context.Context, dx, _ float64) error {
	if a.widgets.gesture.id == "" || dx == 0 {
		return nil
	}
	opts, ok := a.sliderOptions(a.widgets.gesture.id)
	if !ok {
		return nil
	}
	a.widgets.gesture.dragged += dx
	next := chadcn.SliderValueAt(a.widgets.gesture.origin+a.widgets.gesture.dragged, a.widgets.gesture.width, opts)
	return a.setSliderValue(ctx, a.widgets.gesture.id, next)
}

func (a *builderApp) endSliderDrag(context.Context) error {
	a.widgets.gesture = sliderGesture{}
	if a.page == nil {
		return nil
	}
	a.page.AddHandlers(ownframe.Handlers{
		DragMove: a.moveCenterScrollDrag,
		DragEnd:  a.endCenterScrollDrag,
	})
	return nil
}

func (a *builderApp) sliderOptions(id string) (chadcn.SliderOptions, bool) {
	switch id {
	case catalogSliderID:
		return a.catalogSliderOptions(), true
	case widgetsSliderID:
		return a.widgetsSliderOptions(), true
	default:
		return chadcn.SliderOptions{}, false
	}
}

func (a *builderApp) catalogSliderOptions() chadcn.SliderOptions {
	return chadcn.SliderOptions{Label: "Volume", Value: a.catalogSlider}
}

func (a *builderApp) widgetsSliderOptions() chadcn.SliderOptions {
	return chadcn.SliderOptions{
		Label: "Upload progress",
		Value: float64(a.widgets.progress),
		Max:   float64(widgetsProgressMax),
	}
}

func (a *builderApp) setSliderValue(ctx context.Context, id string, value float64) error {
	switch id {
	case catalogSliderID:
		if a.catalogSlider == value {
			return nil
		}
		a.catalogSlider = value
		a.notice = fmt.Sprintf("Catalog slider %.0f.", value)
	case widgetsSliderID:
		next := int(math.Round(value))
		if next == a.widgets.progress {
			return nil
		}
		a.widgets.progress = next
		a.notice = fmt.Sprintf("Progress set to %d%%.", next)
	default:
		return nil
	}
	return a.refresh(ctx)
}

// sliderOrigin is the track-local x of the current value. SliderValueAt
// supplies the same ends the control paints, including the default range.
func sliderOrigin(opts chadcn.SliderOptions, width float64) float64 {
	if width <= 0 {
		return 0
	}
	low := chadcn.SliderValueAt(0, width, opts)
	high := chadcn.SliderValueAt(width, width, opts)
	span := high - low
	if span <= 0 {
		return 0
	}
	value := opts.Value
	if math.IsNaN(value) || value < low {
		value = low
	}
	if value > high {
		value = high
	}
	return (value - low) / span * width
}
