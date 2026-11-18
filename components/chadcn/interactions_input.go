package chadcn

import (
	"context"
	"fmt"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

func (i *Interactions) keyDown(ctx context.Context, key string) error {
	if !i.closed {
		if key == "escape" && i.active.dialog != nil {
			dialog := i.active.dialog
			if dialog.OnOpenChange != nil {
				return dialog.OnOpenChange(ctx, false)
			}
			if i.handlers.Click != nil {
				return i.handlers.Click(ctx, ownframe.Box{Action: dialog.CloseAction})
			}
			return fmt.Errorf("chadcn: dialog CloseAction has no application click handler")
		}
		focus := i.view.Page.Form().FocusID()
		if tab, ok := i.active.tabs[focus]; ok {
			if handled, err := i.tabKey(ctx, key, tab); handled || err != nil {
				return err
			}
		}
		if options, ok := i.active.sliders[focus]; ok && !options.Disabled {
			low, high, step := sliderRange(options)
			value := options.Value
			switch key {
			case "arrowright", "arrowup":
				value += step
			case "arrowleft", "arrowdown":
				value -= step
			case "home":
				value = low
			case "end":
				value = high
			default:
				return i.fallbackKey(ctx, key)
			}
			return i.setSlider(ctx, options, clampSliderValue(value, low, high, step))
		}
	}
	return i.fallbackKey(ctx, key)
}

func (i *Interactions) fallbackKey(ctx context.Context, key string) error {
	if i.handlers.KeyDown != nil {
		return i.handlers.KeyDown(ctx, key)
	}
	return nil
}

func (i *Interactions) tabKey(ctx context.Context, key string, tab tabBinding) (bool, error) {
	items := tab.options.Items
	next := tab.index
	switch key {
	case "arrowright":
		next = (next + 1) % len(items)
	case "arrowleft":
		next = (next + len(items) - 1) % len(items)
	case "home":
		next = 0
	case "end":
		next = len(items) - 1
	default:
		return false, nil
	}
	if err := tab.options.OnSelect(ctx, items[next].ID); err != nil {
		return true, err
	}
	return true, i.view.Page.Form().Focus(ctx, partID(items[next].ID, "-tab"))
}

func (i *Interactions) slider(id string) (SliderOptions, bool) {
	if i.closed {
		return SliderOptions{}, false
	}
	if options, ok := i.active.sliders[id]; ok {
		return options, true
	}
	for _, suffix := range []string{"-range", "-rail"} {
		if root, ok := strings.CutSuffix(id, suffix); ok {
			options, found := i.active.sliders[root]
			return options, found
		}
	}
	return SliderOptions{}, false
}

func (i *Interactions) box(id string) (ownframe.Box, bool) {
	for _, box := range i.view.Page.Boxes() {
		if box.ID == id {
			return box, true
		}
	}
	return ownframe.Box{}, false
}

func (i *Interactions) press(ctx context.Context, box ownframe.Box, x, y float64) (bool, error) {
	options, ok := i.slider(box.ID)
	if ok {
		if options.Disabled {
			return true, nil
		}
		rail, exists := i.box(options.ID + "-rail")
		if !exists || rail.W <= 0 {
			return true, fmt.Errorf("chadcn: slider rail is unavailable")
		}
		i.pressOrigin = &sliderPress{id: options.ID, origin: x - rail.X, width: rail.W}
		if i.drag != nil && i.drag.options.ID == options.ID {
			i.drag.origin, i.drag.width = x-rail.X, rail.W
		}
		if err := i.view.Page.Form().Focus(ctx, options.ID); err != nil {
			return true, err
		}
		return true, i.setSlider(ctx, options, SliderValueAt(x-rail.X, rail.W, options))
	}
	if i.handlers.Press != nil {
		return i.handlers.Press(ctx, box, x, y)
	}
	return false, nil
}

func (i *Interactions) setSlider(ctx context.Context, options SliderOptions, value float64) error {
	if value == options.Value {
		return nil
	}
	if err := options.OnValueChange(ctx, value); err != nil {
		return err
	}
	if current, ok := i.active.sliders[options.ID]; ok && current.Value == options.Value {
		current.Value = value
		i.active.sliders[options.ID] = current
	}
	return nil
}

func (i *Interactions) dragStart(ctx context.Context, box ownframe.Box) (bool, error) {
	options, ok := i.slider(box.ID)
	if ok && !options.Disabled {
		rail, exists := i.box(options.ID + "-rail")
		if !exists || rail.W <= 0 {
			return false, nil
		}
		low, high, _ := sliderRange(options)
		origin := (options.Value - low) / (high - low) * rail.W
		if pressed, exists := i.slider(i.view.Page.Probe().PressedID()); exists && pressed.ID == options.ID && i.pressOrigin != nil && i.pressOrigin.id == options.ID && i.pressOrigin.width == rail.W {
			origin = i.pressOrigin.origin
		}
		i.drag = &sliderDrag{options: options, width: rail.W, origin: origin}
		i.fallbackDrag = false
		return true, nil
	}
	if i.handlers.DragStart != nil {
		claimed, err := i.handlers.DragStart(ctx, box)
		i.fallbackDrag = claimed && err == nil
		return claimed, err
	}
	return false, nil
}

func (i *Interactions) dragMove(ctx context.Context, dx, dy float64) error {
	if i.drag != nil {
		if dx == 0 {
			return nil
		}
		i.drag.delta += dx
		options := i.drag.options
		value := SliderValueAt(i.drag.origin+i.drag.delta, i.drag.width, options)
		current, ok := i.active.sliders[options.ID]
		if ok && !current.Disabled {
			return i.setSlider(ctx, current, value)
		}
		return nil
	}
	if i.fallbackDrag && i.handlers.DragMove != nil {
		return i.handlers.DragMove(ctx, dx, dy)
	}
	return nil
}

func (i *Interactions) dragEnd(ctx context.Context) error {
	i.pressOrigin = nil
	if i.drag != nil {
		i.drag = nil
		return nil
	}
	if i.fallbackDrag {
		i.fallbackDrag = false
		if i.handlers.DragEnd != nil {
			return i.handlers.DragEnd(ctx)
		}
	}
	return nil
}
