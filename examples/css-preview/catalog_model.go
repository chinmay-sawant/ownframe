package main

import (
	"fmt"

	"github.com/chinmay-sawant/ownframe"
)

type catalogRow struct {
	Name, Number, Button, Value, Action, SampleKey string
	TargetID, PreviewClass, PreviewLabel           string
	CardID, BaseClass, StateID                     string
}

func catalogRows(p *ownframe.Page) []catalogRow {
	rows := make([]catalogRow, 0, len(catalogProperties))
	for i, name := range catalogProperties {
		sample, action := catalogAction(name)
		row := catalogRow{Name: name, Number: fmt.Sprintf("%02d", i+1), Button: catalogButton(name),
			Action: action, SampleKey: sample, TargetID: catalogTargetID(name),
			PreviewClass: catalogPreviewClass(name), PreviewLabel: catalogPreviewLabel(name),
			CardID: fmt.Sprintf("catalog-row-%02d", i+1), BaseClass: catalogBaseClass(name),
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
// state. The trigger-scope row keeps the class its scope rule needs.
func catalogBaseClass(name string) string {
	if name == "trigger-scope" {
		return "catalog-row has-trigger-scope"
	}
	return "catalog-row"
}
