package atspi

import (
	"errors"
	"fmt"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/host"
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

func TestPublisherPublishesContentWhenRevisionIsZero(t *testing.T) {
	bus := NewFakeBus()
	p := NewPublisher(bus)
	tree := sampleTree()
	tree.Revision = 0
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

func TestPublisherSkipsUnchangedRevision(t *testing.T) {
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
	if len(bus.Apps()) != 1 || len(bus.Events()) != 1 {
		t.Fatalf("same revision republished: apps=%d events=%d", len(bus.Apps()), len(bus.Events()))
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

type readyFakeBus struct {
	*FakeBus
	ready bool
}

func (b *readyFakeBus) Ready() bool { return b.ready }

func TestPublisherReadyUsesBusBackoffWhileLive(t *testing.T) {
	bus := &readyFakeBus{FakeBus: NewFakeBus()}
	p := NewPublisher(bus)
	if !p.Live() {
		t.Fatal("fake bus must stay live")
	}
	if p.Ready() {
		t.Fatal("live bus with backoff reported ready")
	}
	bus.ready = true
	if !p.Ready() {
		t.Fatal("bus readiness change was ignored")
	}
}

func TestPublisherLargeReplacementEmitsStandardRootChildChanges(t *testing.T) {
	bus := NewFakeBus()
	p := NewPublisher(bus)
	if err := p.Publish(sampleTree()); err != nil {
		t.Fatal(err)
	}
	before := len(bus.Events())

	next := host.AccessibilityTree{
		Document: testDoc + 1,
		Revision: testRev + 1,
		Root:     "root",
		Focused:  "n1",
	}
	children := make([]host.AccessibilityNodeID, 0, 2000)
	for i := 1; i <= 2000; i++ {
		children = append(children, host.AccessibilityNodeID(fmt.Sprintf("n%d", i)))
	}
	next.Nodes = append(next.Nodes, host.AccessibilityNode{
		ID: "root", Role: host.AccessibilityDocument, Children: children,
	})
	for i := 1; i <= 2000; i++ {
		next.Nodes = append(next.Nodes, host.AccessibilityNode{
			ID: host.AccessibilityNodeID(fmt.Sprintf("n%d", i)), Parent: "root",
			Role: host.AccessibilityTextbox, Name: "field",
		})
	}
	if err := p.Publish(next); err != nil {
		t.Fatal(err)
	}
	events := bus.Events()[before:]
	if len(events) > maxPublishEvents {
		t.Fatalf("large replacement emitted %d events", len(events))
	}
	childrenChanged := 0
	for _, ev := range events {
		if ev.Kind == "children-changed" {
			childrenChanged++
			if ev.Detail != "add" && ev.Detail != "remove" {
				t.Fatalf("non-standard child change detail %q", ev.Detail)
			}
		}
	}
	if childrenChanged != 2 {
		t.Fatalf("large replacement emitted %d child changes", childrenChanged)
	}
}

func TestPublisherCoalescesChildChangesPerParent(t *testing.T) {
	bus := NewFakeBus()
	p := NewPublisher(bus)
	if err := p.Publish(sampleTree()); err != nil {
		t.Fatal(err)
	}
	before := len(bus.Events())

	next := sampleTree()
	next.Revision++
	next.Focused = ""
	children := make([]host.AccessibilityNodeID, 0, 300)
	for i := 0; i < 300; i++ {
		id := host.AccessibilityNodeID(fmt.Sprintf("new%d", i))
		children = append(children, id)
		next.Nodes = append(next.Nodes, host.AccessibilityNode{
			ID: id, Parent: "document", Role: host.AccessibilityButton, Name: "added",
		})
	}
	for i := range next.Nodes {
		if next.Nodes[i].ID == "document" {
			next.Nodes[i].Children = children
		}
	}
	if err := p.Publish(next); err != nil {
		t.Fatal(err)
	}
	childrenChanged := 0
	for _, ev := range bus.Events()[before:] {
		if ev.Kind == "children-changed" {
			childrenChanged++
		}
	}
	if childrenChanged != 1 {
		t.Fatalf("one parent emitted %d child changes", childrenChanged)
	}
}

func TestPublisherCapsStateEvents(t *testing.T) {
	bus := NewFakeBus()
	p := NewPublisher(bus)
	tree := host.AccessibilityTree{
		Document: testDoc,
		Revision: testRev,
		Root:     "root",
		Nodes: []host.AccessibilityNode{{
			ID: "root", Role: host.AccessibilityDocument,
		}},
	}
	for i := 0; i < 300; i++ {
		tree.Nodes = append(tree.Nodes, host.AccessibilityNode{
			ID: host.AccessibilityNodeID(fmt.Sprintf("n%d", i)), Parent: "root",
			Role: host.AccessibilityButton, Name: "button",
			Actions: []host.AccessibilityActionKind{host.AccessibilityActivate},
		})
	}
	if err := p.Publish(tree); err != nil {
		t.Fatal(err)
	}
	before := len(bus.Events())

	tree.Revision++
	for i := range tree.Nodes {
		tree.Nodes[i].Disabled = true
	}
	if err := p.Publish(tree); err != nil {
		t.Fatal(err)
	}
	events := bus.Events()[before:]
	if len(events) != maxPublishEvents {
		t.Fatalf("state burst = %d events, want %d", len(events), maxPublishEvents)
	}
	for _, ev := range events {
		if ev.Kind != "state-changed" {
			t.Fatalf("state burst carried %q", ev.Kind)
		}
	}
}
