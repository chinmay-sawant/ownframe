package atspi

import (
	"errors"
	"testing"
)

func TestPublisherFakeBusReplacesTree(t *testing.T) {
	bus := NewFakeBus()
	p := NewPublisher(bus)
	if !p.Live() {
		t.Fatal("fake must be live as a seam")
	}
	tree := sampleTree()
	if err := p.Publish(tree); err != nil {
		t.Fatal(err)
	}
	if err := p.Publish(tree); err != nil {
		t.Fatal(err)
	}
	if len(bus.Apps()) != 1 {
		t.Fatalf("skip same revision: %d", len(bus.Apps()))
	}
	tree.Revision = testRev + 1
	tree.Focused = "id:go"
	if err := p.Publish(tree); err != nil {
		t.Fatal(err)
	}
	if len(bus.Apps()) != 2 || len(bus.Events()) != 2 {
		t.Fatal("replace")
	}
	if p.Mapped()[0].Role != RoleDocumentWeb {
		t.Fatal("mapped")
	}
}

func TestPublisherLiveBusDoesNotClaimSuccess(t *testing.T) {
	p := NewPublisher(NewLiveBus())
	if p.Live() {
		t.Fatal("live client is not implemented")
	}
	err := p.Publish(sampleTree())
	if !errors.Is(err, ErrLiveUnavailable) {
		t.Fatalf("got %v", err)
	}
	if len(p.Mapped()) == 0 {
		t.Fatal("map still stored")
	}
	if p.Status().Live {
		t.Fatal("status")
	}
}

func TestPublisherPublishesContentWithoutRevisionChange(t *testing.T) {
	bus := NewFakeBus()
	p := NewPublisher(bus)
	tree := sampleTree()
	if err := p.Publish(tree); err != nil {
		t.Fatal(err)
	}
	tree.Nodes[1].Value = "Grace"
	if err := p.Publish(tree); err != nil {
		t.Fatal(err)
	}
	apps := bus.Apps()
	if len(apps) != 2 || byID(t, apps[1], "id:name").Value != "Grace" {
		t.Fatal("value-only edit was not published")
	}
}

type retryBus struct {
	*FakeBus
	failed bool
}

func (b *retryBus) Publish(app Application) error {
	if !b.failed {
		b.failed = true
		return ErrLiveUnavailable
	}
	return b.FakeBus.Publish(app)
}

func TestPublisherRetriesFailedPublication(t *testing.T) {
	bus := &retryBus{FakeBus: NewFakeBus()}
	p := NewPublisher(bus)
	tree := sampleTree()
	if err := p.Publish(tree); !errors.Is(err, ErrLiveUnavailable) {
		t.Fatalf("first publish = %v", err)
	}
	if err := p.Publish(tree); err != nil {
		t.Fatal(err)
	}
	if len(bus.Apps()) != 1 {
		t.Fatal("unchanged tree was not retried after failure")
	}
}
