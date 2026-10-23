package player

import (
	"fmt"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func StableID(p string, i int) string { return fmt.Sprintf("%s-%d", p, i) }

func StableIDs(p string, n int) []string {
	o := make([]string, 0, n)
	for i := 0; i < n; i++ {
		o = append(o, StableID(p, i))
	}
	return o
}

func (a *App) RetainedElement(id string) *ownframe.Element {
	if a == nil || a.page == nil {
		return nil
	}
	return a.page.Probe().Element(id)
}

func (a *App) RetainedElements(ids ...string) []*ownframe.Element {
	o := make([]*ownframe.Element, 0, len(ids))
	for _, id := range ids {
		o = append(o, a.RetainedElement(id))
	}
	return o
}

func (a *App) RetainElement(id, kind string) *ownframe.Element {
	if a == nil || a.page == nil {
		return nil
	}
	n := a.page.Probe().ElementRegistry().GetOrCreate(id, ownframe.ElementKind(kind))
	if n != nil {
		n.ID = id
	}
	return n
}

func RetainedRegistry(p *ownframe.Page) *ownframe.ElementRegistry {
	if p == nil {
		return nil
	}
	return p.Probe().ElementRegistry()
}

func RetainFromRegistry(r *ownframe.ElementRegistry, id, kind string) *ownframe.Element {
	if r == nil {
		r = ownframe.NewElementRegistry()
	}
	return r.GetOrCreate(id, ownframe.ElementKind(kind))
}

func AssertRetainedIdentity(t testing.TB, b, a *ownframe.Element) {
	t.Helper()
	if b == nil || a == nil || b != a {
		t.Fatalf("retained identity changed: %p != %p", b, a)
	}
}

func AssertRetainedLifecycle(t testing.TB, n *ownframe.Element, m, u, x bool) {
	t.Helper()
	if n == nil {
		t.Fatal("nil retained node")
	}
	if n.Mounted != m || n.Lifecycle.Mounted != m {
		t.Fatalf("mounted=%v/%v want=%v", n.Mounted, n.Lifecycle.Mounted, m)
	}
	if n.Updated != u || n.Lifecycle.Updated != u {
		t.Fatalf("updated=%v/%v want=%v", n.Updated, n.Lifecycle.Updated, u)
	}
	if n.Lifecycle.Unmounted != x {
		t.Fatalf("unmounted=%v want=%v", n.Lifecycle.Unmounted, x)
	}
}
