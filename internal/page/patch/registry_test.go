package patch

import "testing"

func TestElementRegistryNode(t *testing.T) {
	if Root(nil) != nil || len(All(nil)) != 0 {
		t.Fatal("nil registry was not empty")
	}

	r := Ensure(nil)
	if r == nil {
		t.Fatal("Ensure(nil) returned nil")
	}
	if Ensure(r) != r {
		t.Fatal("Ensure replaced a live registry")
	}

	n := Node(r, "box")
	if n == nil || n.ID != "box" {
		t.Fatalf("node = %+v", n)
	}
	if again := Node(r, "box"); again != n {
		t.Fatal("Node minted a second handle")
	}
	if Root(r) != n {
		t.Fatal("root is not the first handle")
	}
	if got := All(r); len(got) != 1 || got[0] != n {
		t.Fatalf("all = %v", got)
	}

	other := Node(r, "other")
	if Root(r) != n || other.ID != "other" {
		t.Fatal("root moved")
	}
}
