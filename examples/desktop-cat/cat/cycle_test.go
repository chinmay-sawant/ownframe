package cat

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

func TestExpressionsCycleAndWrap(t *testing.T) {
	a := testAnimation(t, true)
	ctx := context.Background()
	if len(a.files) < 2 {
		t.Fatal("expression collection needs multiple images")
	}
	if err := a.paint(ctx, 0); err != nil {
		t.Fatal(err)
	}
	first, _, _ := a.image.ImageBytes()
	before := a.page.Testing().Stats()
	if err := a.paint(ctx, 8); err != nil {
		t.Fatal(err)
	}
	second, _, _ := a.image.ImageBytes()
	if bytes.Equal(first, second) || a.current != 1 {
		t.Fatal("expression did not change after eight seconds")
	}
	if a.page.Testing().Stats().Parses != before.Parses {
		t.Fatal("expression switch reparsed HTML")
	}
	if err := a.paint(ctx, float64(len(a.files))*8); err != nil {
		t.Fatal(err)
	}
	wrapped, _, _ := a.image.ImageBytes()
	if !bytes.Equal(first, wrapped) || a.current != 0 {
		t.Fatal("collection did not wrap to the first expression")
	}
}

func TestFixedExpressionAndCancellation(t *testing.T) {
	a := testAnimation(t, false)
	ctx := context.Background()
	if err := a.paint(ctx, 24); err != nil {
		t.Fatal(err)
	}
	if a.current != 0 {
		t.Fatal("fixed expression changed")
	}
	ctx, cancel := context.WithCancel(ctx)
	cancel()
	if err := a.paint(ctx, 32); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled tick returned %v", err)
	}
	if _, err := NewVariant(-1); err == nil {
		t.Fatal("negative expression number accepted")
	}
	if _, err := NewVariant(len(a.files) + 1); err == nil {
		t.Fatal("out-of-range expression accepted")
	}
}
