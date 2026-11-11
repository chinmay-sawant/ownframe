package chadcn

import (
	"context"
	"math"
	"strconv"

	"github.com/chinmay-sawant/ownframe/ui"
)

// SliderOptions describes a controlled horizontal slider.
type SliderOptions struct {
	Interactions  *Interactions
	OnValueChange func(context.Context, float64) error
	// ID is the stable host ID. The filled range derives its own ID from
	// it, so a value rebuild patches instead of reparsing.
	ID string
	// Label is the control's accessible name.
	Label string
	// Value is the current value. The host updates it as the thumb moves.
	Value float64
	// Min is the low end of the range. The zero value is 0.
	Min float64
	// Max is the high end of the range. The zero value defaults to 100.
	Max float64
	// Step is the value increment. The zero value defaults to 1.
	Step float64
	// Disabled marks the control as non-interactive.
	Disabled bool
}

const (
	defaultSliderMax  = 100
	defaultSliderStep = 1
	sliderPercentMax  = 100
)

// Slider renders a range track, the filled portion, and a thumb. The thumb
// follows Value. A click on the element carries no point, so there is no
// pointer callback without Interactions. With Interactions, OnValueChange
// receives pointer and keyboard changes; the host updates Value and rebuilds.
// A disabled slider does not move.
func Slider(options SliderOptions) *ui.Element {
	if options.Interactions != nil {
		options.Interactions.addSlider(options)
	}
	minimum, maximum, step := sliderRange(options)
	value := clampSliderValue(options.Value, minimum, maximum, step)
	fill := (value - minimum) / (maximum - minimum) * sliderPercentMax
	fillBar := ui.Div().Class("of-slider-range").Width(ui.Percent(fill))
	if fillID := partID(options.ID, "-range"); fillID != "" {
		fillBar.ID(fillID)
	}
	rail := ui.Div().Class("of-slider-rail")
	if options.Interactions != nil && options.ID != "" {
		rail.ID(options.ID + "-rail")
	}
	track := ui.Div().Class("of-slider-track").Children(
		rail.Children(
			fillBar,
			ui.Div().Class("of-slider-thumb").Attr("aria-hidden", "true"),
		),
	)
	slider := ui.Div().Class(classes("of-slider")).Attr("role", "slider").
		Attr("aria-valuemin", sliderValueText(minimum)).
		Attr("aria-valuemax", sliderValueText(maximum)).
		Attr("aria-valuenow", sliderValueText(value)).
		Attr("aria-orientation", "horizontal").
		Children(track)
	if options.ID != "" {
		slider.ID(options.ID)
	}
	if options.Label != "" {
		slider.Attr("aria-label", options.Label)
	}
	if !options.Disabled {
		slider.Attr("tabindex", "0")
	}
	slider = setDisabled(slider, options.Disabled)
	if options.Disabled {
		slider.Attr("aria-disabled", "true")
	}
	return slider
}

// sliderRange returns the effective range and step. A max that is not above
// min falls back to a 100-unit range, and a non-positive step becomes 1.
func sliderRange(options SliderOptions) (float64, float64, float64) {
	minimum, maximum := options.Min, options.Max
	if maximum <= minimum {
		maximum = minimum + defaultSliderMax
	}
	step := options.Step
	if step <= 0 {
		step = defaultSliderStep
	}
	return minimum, maximum, step
}

// SliderValueAt maps localX, measured from the left edge of a track
// trackWidth wide, onto the clamped step grid Slider paints. Min, Max, and
// Step are read from options; Value and Disabled are not. A non-finite
// input or a non-positive width returns the clamped minimum.
func SliderValueAt(localX, trackWidth float64, options SliderOptions) float64 {
	minimum, maximum, step := sliderRange(options)
	if !finite(localX) || !finite(trackWidth) || trackWidth <= 0 {
		return clampSliderValue(minimum, minimum, maximum, step)
	}
	if localX < 0 {
		localX = 0
	}
	if localX > trackWidth {
		localX = trackWidth
	}
	value := minimum + (localX/trackWidth)*(maximum-minimum)
	return clampSliderValue(value, minimum, maximum, step)
}

// clampSliderValue keeps value inside the range and on a step boundary.
func clampSliderValue(value, minimum, maximum, step float64) float64 {
	if math.IsNaN(value) {
		value = minimum
	}
	value = math.Min(math.Max(value, minimum), maximum)
	value = minimum + math.Round((value-minimum)/step)*step
	return math.Min(math.Max(value, minimum), maximum)
}

func sliderValueText(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
