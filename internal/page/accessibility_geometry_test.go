package page

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

func TestAccessibilityAnonymousNodesGetUnambiguousBounds(t *testing.T) {
	p := accessibilityTestPage(t, `<main><p>Anonymous paragraph</p><button>Save</button><button>Save</button></main>`)
	tree, err := p.AccessibilityTree()
	if err != nil {
		t.Fatal(err)
	}
	var buttons int
	for _, node := range tree.Nodes {
		if node.Role != host.AccessibilityButton && node.Role != host.AccessibilityParagraph {
			continue
		}
		if !node.HasBounds {
			t.Fatalf("anonymous node lacks bounds: %+v", node)
		}
		if node.Role == host.AccessibilityButton {
			buttons++
		}
	}
	if buttons != 2 {
		t.Fatalf("anonymous buttons = %d, want 2", buttons)
	}
}

func TestAccessibilityAnonymousBoundsOmitAmbiguousCSSHiddenMatch(t *testing.T) {
	p := accessibilityTestPage(t, `<button style="display:none">Save</button><button>Save</button>`)
	tree, err := p.AccessibilityTree()
	if err != nil {
		t.Fatal(err)
	}
	for _, node := range tree.Nodes {
		if node.Role == host.AccessibilityButton && node.HasBounds {
			t.Fatalf("ambiguous anonymous node got bounds: %+v", node)
		}
	}
}
