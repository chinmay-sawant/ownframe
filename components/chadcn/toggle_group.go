package chadcn

import (
	"context"

	"github.com/chinmay-sawant/ownframe/ui"
)

// ToggleItem is one value in a ToggleGroup.
type ToggleItem struct {
	Icon     *ui.Element
	ID       string
	Value    string
	Label    string
	Disabled bool
}

// ToggleGroupOptions describes the values and selected set of a toggle group.
// Multiple tells the host handler that it may keep more than one value selected;
// the group itself only paints the Selected set.
type ToggleGroupOptions struct {
	OnSelect func(ctx context.Context, value string) error
	Items    []ToggleItem
	Selected []string
	Label    string
	Multiple bool
	Vertical bool
	Size     ButtonSize
}

// ToggleGroup builds a set of toggle buttons that share a selection. The host
// sets Selected after OnSelect, then rebuilds the view; the pressed face comes
// from each item's aria-pressed state. A horizontal group wraps its chips when
// the container is narrow, so a long filter bar never overlaps its neighbours.
func ToggleGroup(options ToggleGroupOptions) *ui.Element {
	selected := make(map[string]struct{}, len(options.Selected))
	for _, value := range options.Selected {
		selected[value] = struct{}{}
	}
	groupClass := "of-toggle-group"
	if options.Vertical {
		groupClass += " of-toggle-group--vertical"
	} else {
		groupClass += " of-toggle-group--wrap"
	}
	group := ui.Div().Class(groupClass).Attr("role", "group").Attr("data-multiple", ariaBool(options.Multiple))
	if options.Label != "" {
		group.Attr("aria-label", options.Label)
	}
	for _, item := range options.Items {
		_, pressed := selected[item.Value]
		toggle := Toggle(item.Label, pressed, ToggleOptions{
			Icon:     item.Icon,
			Size:     options.Size,
			Disabled: item.Disabled,
		}).Attr("data-value", item.Value)
		if item.ID != "" {
			toggle.ID(item.ID)
		}
		if options.OnSelect != nil {
			value := item.Value
			toggle.OnClick(func(ctx context.Context) error {
				return options.OnSelect(ctx, value)
			})
		}
		group.Children(toggle)
	}
	return group
}
