package atspi

import (
	"context"
	"errors"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

type fakeAcc struct {
	tree host.AccessibilityTree
	req  host.AccessibilityRequest
}

func (f *fakeAcc) AccessibilityTree() (host.AccessibilityTree, error) {
	return f.tree, nil
}

func (f *fakeAcc) AccessibilityPerform(_ context.Context, req host.AccessibilityRequest) error {
	f.req = req
	return nil
}

func TestPerformMapsClickAndRejectsPasswordEdit(t *testing.T) {
	p := NewPublisher(NewFakeBus())
	tree := sampleTree()
	if err := p.Publish(tree); err != nil {
		t.Fatal(err)
	}
	acc := &fakeAcc{tree: tree}
	goPath := byID(t, p.app, "id:go").Path
	if err := p.Perform(context.Background(), acc, goPath, ActionClick, ""); err != nil {
		t.Fatal(err)
	}
	if acc.req.Kind != host.AccessibilityActivate || acc.req.Node != "id:go" {
		t.Fatalf("req %+v", acc.req)
	}
	pwPath := byID(t, p.app, "id:pw").Path
	before := acc.req
	if err := p.Perform(context.Background(), acc, pwPath, ActionSetText, "secret"); !errors.Is(err, host.ErrAccessibilityAction) {
		t.Fatalf("password edit = %v", err)
	}
	if acc.req != before {
		t.Fatal("rejected edit reached the page")
	}
}
