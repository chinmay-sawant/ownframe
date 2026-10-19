package main

import (
	"fmt"

	"github.com/chinmay-sawant/ownframe"
)

type catalogRow struct {
	Name, Number, Button, Value, Action, SampleKey string
	TargetID, PreviewClass, PreviewLabel           string
}

func catalogRows(p *ownframe.Page) []catalogRow {
	rows := make([]catalogRow, 0, len(catalogProperties))
	for i, name := range catalogProperties {
		sample, action := catalogAction(name)
		row := catalogRow{Name: name, Number: fmt.Sprintf("%02d", i+1), Button: catalogButton(name),
			Action: action, SampleKey: sample, TargetID: catalogTargetID(name),
			PreviewClass: catalogPreviewClass(name), PreviewLabel: catalogPreviewLabel(name)}
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

func catalogTargetID(name string) string { return "catalog-target-" + catalogSlug(name) }
