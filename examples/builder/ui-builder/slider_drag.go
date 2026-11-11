package main

import (
	"context"
	"fmt"
	"math"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
)

const (
	catalogSliderID = "catalog-slider"
	widgetsSliderID = "widgets-slider"
)

func (a *builderApp) catalogSliderOptions() chadcn.SliderOptions {
	return chadcn.SliderOptions{ID: catalogSliderID, Interactions: a.interactions, Label: "Volume", Value: a.catalogSlider,
		OnValueChange: func(ctx context.Context, value float64) error { return a.setSliderValue(ctx, catalogSliderID, value) }}
}

func (a *builderApp) widgetsSliderOptions() chadcn.SliderOptions {
	return chadcn.SliderOptions{
		ID: widgetsSliderID, Interactions: a.interactions,
		OnValueChange: func(ctx context.Context, value float64) error { return a.setSliderValue(ctx, widgetsSliderID, value) },
		Label:         "Upload progress",
		Value:         float64(a.widgets.progress),
		Max:           float64(widgetsProgressMax),
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
