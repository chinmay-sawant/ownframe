package player

import (
	"fmt"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/element"
)

func StableID(p string, i int) string { return fmt.Sprintf("%s-%d", p, i) }

func StableIDs(p string, n int) []string {
	o := make([]string, 0, n)
	for i := 0; i < n; i++ {
		o = append(o, StableID(p, i))
	}
	return o
}

func (a *App) RetainedElement(id string) *element.Node {
	if a == nil || a.page == nil {
		return nil
	}
	return a.page.RetainedElement(id)
}

func (a *App) RetainedElements(ids ...string) []*element.Node {
	o := make([]*element.Node, 0, len(ids))
	for _, id := range ids {
		o = append(o, a.RetainedElement(id))
	}
	return o
}

func (a *App) RetainElement(id, kind string) *element.Node {
	if a == nil || a.page == nil {
		return nil
	}
	n := a.page.ElementRegistry().GetOrCreate(id, element.Kind(kind))
	if n != nil {
		n.ID = id
	}
	return n
}

func RetainedRegistry(p *ownframe.Page) *element.Registry {
	if p == nil {
		return nil
	}
	return p.ElementRegistry()
}

func RetainFromRegistry(r *element.Registry, id, kind string) *element.Node {
	if r == nil {
		r = element.NewRegistry()
	}
	return r.GetOrCreate(id, element.Kind(kind))
}

func AssertRetainedIdentity(t testing.TB, b, a *element.Node) {
	t.Helper()
	if b == nil || a == nil || b != a {
		t.Fatalf("retained identity changed: %p != %p", b, a)
	}
}

func AssertRetainedLifecycle(t testing.TB, n *element.Node, m, u, x bool) {
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
