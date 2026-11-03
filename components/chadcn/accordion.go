package chadcn

import (
	"context"

	"github.com/chinmay-sawant/ownframe/ui"
)

// AccordionItem is one disclosure trigger and the content it reveals.
type AccordionItem struct {
	Content *ui.Element
	ID      string
	Title   string
}

// AccordionOptions describes a controlled accordion. Open lists the item IDs
// whose panels are expanded; the host decides whether one item or many may be
// open at once. OnToggle receives the clicked item ID.
type AccordionOptions struct {
	OnToggle func(ctx context.Context, id string) error
	Items    []AccordionItem
	Open     []string
}

// Accordion builds a stack of disclosure items. Each trigger is a button with
// aria-expanded and aria-controls, and each open panel is a labelled region.
// Item IDs derive the trigger and panel part IDs.
func Accordion(options AccordionOptions) *ui.Element {
	accordion := ui.Div().Class("of-accordion")
	for _, item := range options.Items {
		open := false
		for _, id := range options.Open {
			if id == item.ID {
				open = true
				break
			}
		}
		triggerID, panelID := partID(item.ID, "-trigger"), partID(item.ID, "-panel")
		trigger := ui.Button("").Attr("type", "button").Class("of-accordion-trigger").
			Attr("aria-expanded", ariaBool(open))
		if triggerID != "" {
			trigger.ID(triggerID)
		}
		if open && panelID != "" {
			trigger.Attr("aria-controls", panelID)
		}
		chevron := "▾"
		if open {
			chevron = "▴"
		}
		trigger.Children(
			ui.Span().Class("of-accordion-title").Text(item.Title),
			ui.Span().Class("of-accordion-chevron").Attr("aria-hidden", "true").Text(chevron),
		)
		if options.OnToggle != nil {
			id := item.ID
			trigger.OnClick(func(ctx context.Context) error { return options.OnToggle(ctx, id) })
		}
		children := []*ui.Element{trigger}
		if open {
			panel := ui.Div().Class("of-accordion-content").Attr("role", "region")
			if panelID != "" {
				panel.ID(panelID)
			}
			if triggerID != "" {
				panel.Attr("aria-labelledby", triggerID)
			} else if item.Title != "" {
				panel.Attr("aria-label", item.Title)
			}
			panel.Children(item.Content)
			children = append(children, panel)
		}
		accordion.Children(ui.Div().Class("of-accordion-item").Children(children...))
	}
	return accordion
}
