package atspi

import (
	"context"
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

func TestPerformMapsClickAndRedactsPasswordValue(t *testing.T) {
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
	if err := p.Perform(context.Background(), acc, pwPath, ActionSetText, "secret"); err != nil {
		t.Fatal(err)
	}
	if acc.req.Value != "" || acc.req.Kind != host.AccessibilitySetValue {
		t.Fatalf("password value leaked: %+v", acc.req)
	}
}
