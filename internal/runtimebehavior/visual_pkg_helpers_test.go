package runtimebehavior

import (
	"testing"
)

func TestVisualPackageHelpers(t *testing.T) {
	if !CompareVisual(nil, nil) {
		t.Fatal("CompareVisual broken")
	}
	if !VisualEqual([]byte("z"), []byte("z")) {
		t.Fatal("VisualEqual broken")
	}
	if VisualEqual([]byte("z"), []byte("y")) {
		t.Fatal("VisualEqual should fail")
	}
}
