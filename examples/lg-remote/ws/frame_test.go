package ws

import (
	"bufio"
	"net"
	"testing"
)

func TestTextRoundTrip(t *testing.T) {
	left, right := net.Pipe()
	t.Cleanup(func() {
		_ = left.Close()
		_ = right.Close()
	})

	writer := &Conn{c: left, r: bufio.NewReader(left), wmu: make(chan struct{}, 1)}
	reader := &Conn{c: right, r: bufio.NewReader(right), wmu: make(chan struct{}, 1)}

	go func() { _ = writer.WriteText([]byte("HOME")) }()

	got, err := reader.ReadText()
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != "HOME" {
		t.Fatalf("got %q", got)
	}
}
