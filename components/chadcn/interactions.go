package chadcn

import (
	"context"
	"fmt"
	"slices"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/ui"
)

// Interactions owns input mechanics for a builder view. The application still
// owns component values and open state. Use Mount, Handle, and Update on the
// page loop; rebuilding through ui.View.Update directly bypasses registration.
type Interactions struct {
	view         *ui.View
	handlers     ownframe.Handlers
	active       *interactionFrame
	staging      *interactionFrame
	modal        DialogFocus
	modalID      string
	pressOrigin  *sliderPress
	drag         *sliderDrag
	fallbackDrag bool
	closed       bool
}

type interactionFrame struct {
	tabs    map[string]tabBinding
	sliders map[string]SliderOptions
	dialog  *DialogOptions
	err     error
}

type tabBinding struct {
	options TabsOptions
	index   int
}

type sliderPress struct {
	id            string
	origin, width float64
}

type sliderDrag struct {
	options SliderOptions
	width   float64
	origin  float64
	delta   float64
}

func NewInteractions() *Interactions { return &Interactions{} }

func (i *Interactions) build(build func() *ui.Element) (*ui.Element, *interactionFrame, error) {
	if build == nil {
		return nil, nil, fmt.Errorf("chadcn: component builder is required")
	}
	if i.staging != nil {
		return nil, nil, fmt.Errorf("chadcn: interaction build is already active")
	}
	frame := &interactionFrame{tabs: make(map[string]tabBinding), sliders: make(map[string]SliderOptions)}
	i.staging = frame
	defer func() { i.staging = nil }()
	root := build()
	return root, frame, frame.err
}

// Mount registers the initial component snapshot. Initially open dialogs need
// a first redraw so their rendered controls can supply the focus scope.
func (i *Interactions) Mount(ctx context.Context, cfg ownframe.Config, build func() *ui.Element) (*ui.View, error) {
	if ctx == nil {
		return nil, fmt.Errorf("chadcn: context is required")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if i.view != nil || i.closed {
		return nil, fmt.Errorf("chadcn: interactions already mounted or closed")
	}
	root, frame, err := i.build(build)
	if err != nil {
		return nil, err
	}
	view, err := ui.Mount(cfg, root)
	if err != nil {
		return nil, err
	}
	i.view, i.active = view, frame
	i.Handle(ownframe.Handlers{})
	if frame.dialog != nil {
		if err := view.Page.Redraw(ctx); err != nil {
			return nil, err
		}
		if err := i.syncDialog(ctx, ""); err != nil {
			return nil, err
		}
	}
	return view, nil
}

// Update adopts registrations only after the builder update succeeds. Removed
// or disabled sliders cancel an owned drag, and removed modals restore focus.
func (i *Interactions) Update(ctx context.Context, build func() *ui.Element) error {
	if ctx == nil {
		return fmt.Errorf("chadcn: context is required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if i.view == nil || i.closed {
		return fmt.Errorf("chadcn: interactions are not mounted")
	}
	root, frame, err := i.build(build)
	if err != nil {
		return err
	}
	previous := i.view.Page.Form().FocusID()
	oldTab, hadTab := i.active.tabs[previous]
	if err := i.view.Update(ctx, root); err != nil {
		return err
	}
	i.active = frame
	if i.drag != nil {
		options, ok := frame.sliders[i.drag.options.ID]
		if !ok || options.Disabled {
			i.drag = nil
		}
	}
	if err := i.syncDialog(ctx, previous); err != nil {
		return err
	}
	if hadTab && i.active.dialog == nil {
		if _, exists := frame.tabs[previous]; !exists {
			for _, item := range oldTab.options.Items {
				if binding, ok := frame.tabs[partID(item.ID, "-tab")]; ok {
					selected := activeTabIndex(binding.options)
					if selected >= 0 {
						return i.view.Page.Form().Focus(ctx, partID(binding.options.Items[selected].ID, "-tab"))
					}
				}
			}
			return i.view.Page.Form().FocusNext(ctx)
		}
	}
	return nil
}

// Handle installs application callbacks once while preserving builder OnClick
// dispatch. Components consume only the inputs they own; other inputs reach h.
func (i *Interactions) Handle(h ownframe.Handlers) {
	i.handlers = h
	wrapped := h
	wrapped.KeyDown = i.keyDown
	wrapped.Press = i.press
	wrapped.DragStart = i.dragStart
	wrapped.DragMove = i.dragMove
	wrapped.DragEnd = i.dragEnd
	i.view.Handle(wrapped)
}

// Close detaches component mechanics and restores an active modal's focus.
// The page and its application callbacks remain usable.
func (i *Interactions) Close(ctx context.Context) error {
	if i.closed {
		return nil
	}
	if ctx == nil {
		return fmt.Errorf("chadcn: context is required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	i.closed = true
	i.drag = nil
	i.active = &interactionFrame{}
	if i.modal.active {
		i.modal.Deactivate(i.view.Page)
		i.modalID = ""
		return i.view.Page.RedrawIfChanged(ctx)
	}
	return nil
}

func (i *Interactions) addTabs(options TabsOptions) {
	if i.staging == nil {
		return
	}
	options.Items = slices.Clone(options.Items)
	for index, item := range options.Items {
		id := partID(item.ID, "-tab")
		if item.ID == "" || options.OnSelect == nil {
			i.staging.err = fmt.Errorf("chadcn: managed tabs require item IDs and OnSelect")
			return
		}
		if _, exists := i.staging.tabs[id]; exists {
			i.staging.err = fmt.Errorf("chadcn: duplicate tab ID %q", id)
		}
		i.staging.tabs[id] = tabBinding{options: options, index: index}
	}
}

func (i *Interactions) addSlider(options SliderOptions) {
	if i.staging == nil {
		return
	}
	if options.ID == "" || (!options.Disabled && options.OnValueChange == nil) {
		i.staging.err = fmt.Errorf("chadcn: managed sliders require an ID and OnValueChange")
		return
	}
	if _, exists := i.staging.sliders[options.ID]; exists {
		i.staging.err = fmt.Errorf("chadcn: duplicate slider ID %q", options.ID)
	}
	i.staging.sliders[options.ID] = options
}

func (i *Interactions) addDialog(options DialogOptions) {
	if i.staging == nil || (!options.Open && !options.legacyOpen()) {
		return
	}
	if i.staging.dialog != nil {
		i.staging.err = fmt.Errorf("chadcn: managed nested or simultaneous dialogs are unsupported")
		return
	}
	if options.OnOpenChange == nil && options.CloseAction == "" {
		i.staging.err = fmt.Errorf("chadcn: managed dialog requires OnOpenChange or CloseAction")
	}
	i.staging.dialog = &options
}

func (i *Interactions) syncDialog(ctx context.Context, previous string) error {
	page := i.view.Page
	if i.active.dialog == nil {
		if i.modal.active {
			i.modal.Deactivate(page)
			i.modalID = ""
			if err := page.RedrawIfChanged(ctx); err != nil {
				return err
			}
			if page.Form().FocusID() == "" {
				return page.Form().FocusNext(ctx)
			}
			return nil
		}
		return nil
	}
	panel := i.active.dialog.panelID()
	ids, err := dialogFocusIDs(page, panel)
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		ids = []string{panel}
	}
	if i.modalID != panel {
		if i.modal.active {
			i.modal.Deactivate(page)
		}
		i.modal.Activate(page, ids...)
		if previous != "" {
			i.modal.previous = previous
		}
		i.modalID = panel
		return page.RedrawIfChanged(ctx)
	}
	page.Form().SetFocusScope(ids...)
	if !slices.Contains(ids, page.Form().FocusID()) {
		page.Form().RequestFocus(ids[0])
		return page.RedrawIfChanged(ctx)
	}
	return nil
}

func dialogFocusIDs(page *ownframe.Page, panel string) ([]string, error) {
	tree, err := page.Access().AccessibilityTree()
	if err != nil {
		return nil, err
	}
	inside := make(map[ownframe.AccessibilityNodeID]bool)
	ids := []string{}
	for _, node := range tree.Nodes {
		if node.ElementID == panel {
			inside[node.ID] = true
			continue
		}
		if inside[node.Parent] {
			inside[node.ID] = true
			if node.Focusable && !node.Disabled && node.ElementID != "" {
				ids = append(ids, node.ElementID)
			}
		}
	}
	return ids, nil
}
