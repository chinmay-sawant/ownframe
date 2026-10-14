package runtimebehavior

import (
	"testing"
)

func TestVisualCompareNilAndEmpty(t *testing.T) {
	var nilC *VisualComparer
	if !nilC.Compare() {
		t.Fatal("nil comparer should be equal")
	}
	if !nilC.IsEqual() || !nilC.Matches() {
		t.Fatal("nil comparer flags wrong")
	}
	v := NewVisualComparer()
	v.SetImages(nil, nil)
	if !v.Compare() {
		t.Fatal("nil bytes should match")
	}
	if !v.Equal || v.Mismatch {
		t.Fatal("equal flags wrong")
	}
	v.SetImages([]byte("full"), []byte("partial"))
	if v.Compare() {
		t.Fatal("different bytes should mismatch")
	}
	if v.Equal || !v.Mismatch {
		t.Fatal("mismatch flags wrong")
	}
}
