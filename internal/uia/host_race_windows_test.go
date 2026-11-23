//go:build windows && (amd64 || arm64)

package uia

import (
	"context"
	"sync"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// bareHost builds a Host without a window so the object lifetime rules can be
// stressed without subclassing anything.
func bareHost() *Host {
	h := &Host{
		prov:  NewProvider(func(context.Context, host.AccessibilityRequest) error { return nil }),
		elems: map[host.AccessibilityNodeID]*comObj{},
	}
	h.prov.Update(sampleTree())
	h.rootObj = newObj(h, "", true)
	return h
}

// TestObjForSurvivesConcurrentPublish asserts that an object returned by
// objFor stays registered until its caller releases it, even while Publish
// swaps documents and drops the cache's own references. Run with -race on a
// Windows host.
func TestObjForSurvivesConcurrentPublish(t *testing.T) {
	h := bareHost()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := range 2000 {
			tree := sampleTree()
			tree.Document = uint64(100 + i%2)
			h.Publish(tree)
		}
	}()
	go func() {
		defer wg.Done()
		for range 2000 {
			for _, id := range []host.AccessibilityNodeID{"go", "email", "root"} {
				o := h.objFor(id)
				if o == nil {
					continue
				}
				if _, _, ok := resolve(o.slotAddr(slotFragment)); !ok {
					t.Errorf("object %q unregistered while owned", id)
				}
				o.release()
			}
		}
	}()
	wg.Wait()
}

// TestObjForRefusesAfterClose asserts a released root cannot be revived.
func TestObjForRefusesAfterClose(t *testing.T) {
	h := bareHost()
	h.mu.Lock()
	h.closed = true
	h.mu.Unlock()
	if h.objFor("go") != nil || h.acquireRoot() != nil {
		t.Fatal("closed host handed out an object")
	}

	o := newObj(h, "x", false)
	o.release()
	if o.tryAddRef() {
		t.Fatal("tryAddRef resurrected a released object")
	}
}
