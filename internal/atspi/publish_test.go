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
	if len(bus.Apps()) != 2 || len(bus.Events()) != 3 {
		t.Fatalf("replace: apps=%d events=%d", len(bus.Apps()), len(bus.Events()))
	}
	if p.Mapped()[0].Role != RoleApplication {
		t.Fatal("mapped")
	}
}

func TestPublisherLiveBusRetriesWithoutABus(t *testing.T) {
	t.Setenv("AT_SPI_BUS_ADDRESS", "disabled")
	p := NewPublisher(NewLiveBus())
	if p.Live() {
		t.Fatal("live client claimed a disabled bus")
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

func TestPublisherEmitsStateChanges(t *testing.T) {
	bus := NewFakeBus()
	p := NewPublisher(bus)
	tree := sampleTree()
	if err := p.Publish(tree); err != nil {
		t.Fatal(err)
	}
	boxPath := ""
	for _, obj := range p.Mapped() {
		if obj.ID == "id:on" {
			boxPath = obj.Path
		}
	}
	if boxPath == "" {
		t.Fatal("missing checkbox")
	}
	for i := range tree.Nodes {
		if tree.Nodes[i].ID != "id:on" {
			continue
		}
		tree.Nodes[i].Checked = false
		tree.Nodes[i].Disabled = true
	}
	tree.Revision++
	if err := p.Publish(tree); err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, ev := range bus.Events() {
		if ev.Path == boxPath && ev.Kind == "state-changed" {
			seen[ev.Detail] = ev.Enabled
		}
	}
	if _, ok := seen["checked"]; !ok || seen["checked"] {
		t.Fatalf("checked event %v", seen)
	}
	if _, ok := seen["enabled"]; !ok || seen["enabled"] {
		t.Fatalf("enabled event %v", seen)
	}
	if _, ok := seen["sensitive"]; !ok || seen["sensitive"] {
		t.Fatalf("sensitive event %v", seen)
	}
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
