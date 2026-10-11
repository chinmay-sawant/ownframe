package demo

import (
	"context"
	"fmt"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

func (a *App) onClick(ctx context.Context, box ownframe.Box) error {
	status := "Tree refreshed; snapshot reflects current page state."
	var err error
	switch box.Action {
	case "activation-target":
		a.view.ActivationCount++
		status = "Activation target clicked."
	case "api-focus":
		err = a.performElement(ctx, "display-name", ownframe.AccessibilityFocus, "")
	case "api-set-value":
		err = a.performElement(ctx, "display-name", ownframe.AccessibilitySetValue, a.page.FormValue("action-value"))
	case "api-activate-checkbox":
		err = a.performElement(ctx, "updates", ownframe.AccessibilityActivate, "")
	case "api-activate-radio":
		err = a.performElement(ctx, "contact-phone", ownframe.AccessibilityActivate, "")
	case "api-activate-select":
		err = a.performElement(ctx, "plan", ownframe.AccessibilityActivate, "")
	case "api-activate-button":
		err = a.performElement(ctx, "activation-target", ownframe.AccessibilityActivate, "")
	case "probe-stale", "probe-node", "probe-action", "probe-disabled", "probe-range":
		err = a.runErrorProbe(ctx, box.Action)
		status = "Error probe returned " + errorLabel(err) + "."
	case "refresh-tree":
	default:
		return nil
	}
	if err != nil && !strings.HasPrefix(box.Action, "probe-") {
		status = fmt.Sprintf("Action failed: %v", err)
	}
	return a.refreshTree(status)
}

func (a *App) onChange(context.Context, ownframe.Box) error {
	return a.refreshTree("Control changed; snapshot refreshed.")
}

func (a *App) refreshTree(status string) error {
	tree, err := a.accessible().AccessibilityTree()
	if err != nil {
		return err
	}
	a.view.Name = a.page.FormValue("display-name")
	a.view.ActionValue = a.page.FormValue("action-value")
	a.view.Tree = treeText(tree)
	a.view.Status = status
	a.page.SetData(a.view)
	return nil
}
