package main

// patchCostCard names the catalog card the harness measures: the card row,
// the state span inside it, and the class strings both carry. The card row
// holds an absolutely positioned preview, so the engine refuses to splice it.
// The state span holds none, so the engine splices it.
type patchCostCard struct {
	id       string
	state    string
	name     string
	cardBase string
	cardOn   string
	spanBase string
	spanOn   string
}

// patchCostCardOf returns the first catalog card of the view.
func patchCostCardOf(v *view) patchCostCard {
	for _, row := range v.Catalog {
		if row.Name == "trigger-scope" {
			continue
		}
		state := row.StateID
		if state == "" {
			state = "catalog-state-" + row.Number
		}
		return patchCostCard{id: row.CardID, state: state, name: row.Name,
			cardBase: row.BaseClass, cardOn: row.BaseClass + " is-selected",
			spanBase: "catalog-state", spanOn: "catalog-state is-selected"}
	}
	return patchCostCard{}
}
