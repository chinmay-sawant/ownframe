//go:build linux && !android && !ios && !js

package atspi

import (
	"bufio"
	"context"
	"net"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

func TestLiveReaderServesQueriesWhileActionWaits(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	bus, ok := NewLiveBus(func(ctx context.Context, _ host.AccessibilityRequest) error {
		close(entered)
		select {
		case <-release:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}).(*liveBus)
	if !ok {
		t.Fatal("expected Linux bus")
	}
	view, _ := testServerView(t)
	client, peer := net.Pipe()
	bus.conn = client
	bus.connDone = make(chan struct{})
	bus.view = view
	bus.actionCtx, bus.cancelActions = context.WithCancel(t.Context())
	t.Cleanup(func() { _ = peer.Close(); _ = bus.Close() })
	bus.wg.Add(1)
	go bus.readLoop(client, bufio.NewReader(client))
	if err := peer.SetDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	path := objectPathFor(t, view, "id:name")
	for _, call := range []*dbusMsg{
		{msgType: dbusTypeMethodCall, serial: 10, sender: ":1.9", path: path, iface: atspiComponent, member: "GrabFocus"},
		{msgType: dbusTypeMethodCall, serial: 11, sender: ":1.9", path: path, iface: atspiAccessible, member: "GetName"},
	} {
		if _, err := peer.Write(encodeMessage(call)); err != nil {
			t.Fatal(err)
		}
	}
	reader := bufio.NewReader(peer)
	query, err := readMessage(reader)
	if err != nil || query.replySerial != 11 {
		t.Fatal("waiting action blocked the reader's query reply")
	}
	<-entered
	close(release)
	action, err := readMessage(reader)
	if err != nil || action.replySerial != 10 {
		t.Fatal("action reply was lost after execution")
	}
}
