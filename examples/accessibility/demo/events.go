package demo

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

func (a *App) onClick(ctx context.Context, box ownframe.Box) error {
	if box.Action == "set-via-tree" {
		if err := a.setNameViaTree(ctx); err != nil {
			return err
		}
	}
	return a.refreshTree()
}

func (a *App) onChange(context.Context, ownframe.Box) error {
	return a.refreshTree()
}

func (a *App) refreshTree() error {
	tree, err := a.page.AccessibilityTree()
	if err != nil {
		return err
	}
	a.view.Tree = treeText(tree)
	a.view.Status = "Tree refreshed; values reflect committed edits."
	a.page.SetData(a.view)
	return nil
}
