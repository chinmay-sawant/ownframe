package androidaccess_test

import (
	"encoding/json"
	"testing"

	"github.com/chinmay-sawant/ownframe/androidaccess"
	"github.com/chinmay-sawant/ownframe/internal/a11ybridge"
)

// The behavior of the snapshot and of the actions is tested against a real
// page in internal/a11ybridge. These tests pin that the public wrappers read
// what the window loop publishes and refuse until a page is attached.

func TestTreeIsParsableBeforeAnyPage(t *testing.T) {
	a11ybridge.Reset()

	var tree struct {
		Nodes []json.RawMessage `json:"nodes"`
	}

	if err := json.Unmarshal([]byte(androidaccess.Tree()), &tree); err != nil {
		t.Fatalf("Tree before a page is not valid JSON: %v", err)
	}

	if len(tree.Nodes) != 0 {
		t.Fatalf("Tree before a page has %d nodes", len(tree.Nodes))
	}
}

func TestTreeReturnsWhatTheLoopPublished(t *testing.T) {
	t.Cleanup(a11ybridge.Reset)

	const published = `{"doc":3,"rev":9,"root":"r","nodes":[]}`
	a11ybridge.SetTree(published)

	if got := androidaccess.Tree(); got != published {
		t.Fatalf("Tree = %q, want %q", got, published)
	}
}

func TestPerformIsRefusedWithoutAPage(t *testing.T) {
	a11ybridge.Reset()

	if msg := androidaccess.Perform(1, "node", androidaccess.ActionActivate, ""); msg == "" {
		t.Fatal("Perform with no page attached must refuse")
	}
}

func TestActionNumbersMatchTheHost(t *testing.T) {
	if androidaccess.ActionFocus != 1 || androidaccess.ActionActivate != 2 || androidaccess.ActionSetValue != 3 {
		t.Fatalf("action numbers changed: focus=%d activate=%d setValue=%d; the Java host hard-codes 1, 2, 3",
			androidaccess.ActionFocus, androidaccess.ActionActivate, androidaccess.ActionSetValue)
	}
}
