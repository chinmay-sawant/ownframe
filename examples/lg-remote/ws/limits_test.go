package ws

import (
	"bufio"
	"errors"
	"strings"
	"testing"
)

func TestRejectBadScheme(t *testing.T) {
	t.Parallel()

	if _, err := Dial(t.Context(), "http://tv/"); !errors.Is(err, ErrHandshake) {
		t.Fatalf("err = %v", err)
	}
}

func TestRejectHugeMessage(t *testing.T) {
	t.Parallel()

	c := &Conn{wmu: make(chan struct{}, 1)}
	if err := c.WriteText(make([]byte, maxMsgLen+1)); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("err = %v", err)
	}
}

func TestRejectHugeHandshake(t *testing.T) {
	t.Parallel()

	line := strings.Repeat("a", maxHeadBytes+2) + "\n"
	if err := accept(bufio.NewReader(strings.NewReader(line)), "k"); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("err = %v", err)
	}
}

func TestRejectHugeCtrl(t *testing.T) {
	t.Parallel()

	c := &Conn{wmu: make(chan struct{}, 1)}
	if err := c.writeCtrl(opcodePing, make([]byte, maxCtrlLen+1)); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("err = %v", err)
	}
}
