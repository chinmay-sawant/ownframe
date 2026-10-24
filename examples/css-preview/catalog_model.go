package main

import (
	"fmt"

	"github.com/chinmay-sawant/ownframe"
)

// catalogRow is one property card. Class, Started, and Selected are rendered
// from the view's click state on every draw, so a Start click is a template
// data change.
type catalogRow struct {
	Name, Number, Button, Value, Action, SampleKey string
	TargetID, PreviewClass, PreviewLabel           string
	CardID, BaseClass, Class, StateID              string
	Started, Selected                              bool
}

func catalogRows(p *ownframe.Page) []catalogRow {
	rows := make([]catalogRow, 0, len(catalogProperties))
	for i, name := range catalogProperties {
		sample, action := catalogAction(name)
		base := catalogBaseClass(name)
		row := catalogRow{Name: name, Number: fmt.Sprintf("%02d", i+1), Button: catalogButton(name),
			Action: action, SampleKey: sample, TargetID: catalogTargetID(name),
			PreviewClass: catalogPreviewClass(name), PreviewLabel: catalogPreviewLabel(name),
			CardID: fmt.Sprintf("catalog-row-%02d", i+1), BaseClass: base, Class: base,
			StateID: fmt.Sprintf("catalog-state-%02d", i+1)}
		if p != nil {
			row.Value = catalogValue(p, name)
		}
		rows = append(rows, row)
	}
	return rows
}

func catalogButton(name string) string {
	return "start-property-" + catalogSlug(name)
}

func catalogCardID(name string) string {
	for i, property := range catalogProperties {
		if property == name {
			return fmt.Sprintf("catalog-row-%02d", i+1)
		}
	}
	return ""
}

func catalogTargetID(name string) string { return "catalog-target-" + catalogSlug(name) }

// catalogBaseClass returns the class list a card carries without the click
// state. A card whose preview waits for an animation-trigger also scopes the
// name on the card, so the Start button's activation reaches the target.
func catalogBaseClass(name string) string {
	switch name {
	case "trigger-scope":
		return "catalog-row has-trigger-scope"
	case "animation-trigger":
		return "catalog-row has-trigger-start"
	case "animation":
		return "catalog-row has-standard-trigger"
	}
	return "catalog-row"
}
