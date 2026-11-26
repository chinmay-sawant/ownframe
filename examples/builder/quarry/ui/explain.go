package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/chinmay-sawant/ownframe/ui"

	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/driver"
)

// explainView renders an ExplainPlan as an indented tree of cards or, when raw
// is set, as the engine's raw text.
func (a *App) explainView(plan *driver.ExplainPlan, raw bool, onToggle func(context.Context) error) *ui.Element {
	if plan == nil || plan.Root == nil {
		return emptyPlain("explain-empty", "Run Explain to see the query plan.")
	}
	head := ui.Div().Class("qr-explain-head").Children(
		ui.Text(fmt.Sprintf("%s plan", plan.Engine)).Class("qr-muted"),
		planPill(plan.FullScans, "full scan", "qr-badge-warn"),
		planPill(plan.IndexSearches, "index search", "qr-badge-ok"),
	)
	if plan.Analyzed {
		head.Children(ui.Text(fmt.Sprintf("execution %.2f ms", plan.ExecutionMs)).Class("qr-muted"))
	}
	head.Children(ui.Div().Class("qr-spacer"),
		ui.Button(map[bool]string{false: "Raw", true: "Tree"}[raw]).ID("explain-raw-toggle").Class("qr-btn").OnClick(onToggle))
	body := ui.Div().ID("explain-body").Class("qr-explain-body")
	if raw {
		body.Children(ui.Tag("pre").ID("explain-raw").Class("qr-mono qr-raw").Text(plan.Raw))
	} else {
		maxMs := planMaxMs(plan.Root)
		body.Children(planCard(plan.Root, 0, maxMs))
	}
	return ui.Div().ID("explain").Class("qr-explain").Children(head, a.pane("explain", "Plan scroll", "", body))
}

func planPill(n int, label, class string) *ui.Element {
	if n == 0 {
		return nil
	}
	return ui.Text(fmt.Sprintf("%d %s", n, label)).Class("qr-badge " + class)
}

func planMaxMs(n *driver.PlanNode) float64 {
	m := n.ActualMs
	for _, c := range n.Children {
		m = max(m, planMaxMs(c))
	}
	return m
}

// planCard renders one node card followed by its indented children.
func planCard(n *driver.PlanNode, depth int, maxMs float64) *ui.Element {
	title := n.NodeType
	target := n.Relation
	if n.Index != "" {
		target += " using " + n.Index
	}
	card := ui.Div().ID(fmt.Sprintf("plan-node-%d", n.ID)).Class("qr-plan-card").Children(
		ui.Div().Class("qr-plan-line").Children(
			ui.Text(title).Class("qr-plan-op"),
			scanBadge(n),
			ui.Text(target).Class("qr-muted"),
		),
	)
	if n.Detail != "" && n.Relation == "" {
		card.Children(ui.Div().Class("qr-muted qr-plan-detail").Text(n.Detail))
	}
	meta := []string{}
	if n.TotalCost > 0 {
		meta = append(meta, fmt.Sprintf("cost %.2f..%.2f", n.StartupCost, n.TotalCost))
	}
	if n.PlanRows > 0 {
		meta = append(meta, fmt.Sprintf("rows ~%.0f", n.PlanRows))
	}
	if n.ActualMs >= 0 && (n.ActualRows > 0 || n.ActualMs > 0) {
		meta = append(meta, fmt.Sprintf("actual %.0f rows in %.2f ms", n.ActualRows, n.ActualMs))
	}
	if len(meta) > 0 {
		card.Children(ui.Div().Class("qr-muted qr-plan-meta").Text(strings.Join(meta, "  ")))
	}
	if n.ActualMs > 0 && maxMs > 0 {
		card.Children(ui.Div().Class("qr-plan-bar-track").Children(
			ui.Div().Class("qr-plan-bar").Width(ui.Percent(max(planBarMinPct, n.ActualMs/maxMs*chartHundred)))))
	}
	wrap := ui.Div().Class("qr-plan-node").Children(card)
	if len(n.Children) > 0 {
		kids := ui.Div().Class("qr-plan-children")
		for _, c := range n.Children {
			kids.Children(planCard(c, depth+1, maxMs))
		}
		wrap.Children(kids)
	}
	return wrap
}

func scanBadge(n *driver.PlanNode) *ui.Element {
	switch n.Scan {
	case driver.ScanFull:
		return ui.Text("Seq Scan").Class("qr-badge qr-badge-warn")
	case driver.ScanSearch:
		return ui.Text("Index Scan").Class("qr-badge qr-badge-ok")
	}
	return nil
}

// emptyPlain is the one-line empty state used where no action applies.
func emptyPlain(id, text string) *ui.Element {
	return ui.Div().ID(id).Class("qr-empty").Children(ui.Div().Class("qr-empty-hint").Text(text))
}
