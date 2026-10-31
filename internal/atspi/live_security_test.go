//go:build linux && !android && !ios && !js

package atspi

import (
	"bufio"
	"encoding/hex"
	"net"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSyncCallRejectsForgedSender(t *testing.T) {
	call := &dbusMsg{
		msgType: dbusTypeMethodCall,
		serial:  7,
		path:    "/org/freedesktop/DBus",
		dest:    dbusDaemonName,
		iface:   dbusDaemonName,
		member:  "Hello",
	}
	client, server := net.Pipe()
	t.Cleanup(func() {
		_ = client.Close()
		_ = server.Close()
	})
	go func() {
		if _, err := readMessage(server); err != nil {
			return
		}
		forged := &dbusMsg{
			msgType:     dbusTypeMethodReturn,
			replySerial: call.serial,
			serial:      80,
			sender:      ":1.66",
			sig:         "s",
			body:        encodeStringValue("forged"),
		}
		_, _ = server.Write(encodeMessage(forged))
		real := &dbusMsg{
			msgType:     dbusTypeMethodReturn,
			replySerial: call.serial,
			serial:      81,
			sender:      dbusDaemonName,
			sig:         "s",
			body:        encodeStringValue("real"),
		}
		_, _ = server.Write(encodeMessage(real))
	}()

	reply, err := syncCall(client, bufio.NewReader(client), call, dbusDaemonName)
	if err != nil {
		t.Fatalf("syncCall: %v", err)
	}
	got, err := decodeStringBody(reply.body)
	if err != nil {
		t.Fatalf("decode reply: %v", err)
	}
	if got != "real" {
		t.Fatalf("accepted forged reply %q, want real", got)
	}
}

func TestSyncCallRejectsForgedSenderOnly(t *testing.T) {
	call := &dbusMsg{
		msgType: dbusTypeMethodCall,
		serial:  7,
		path:    "/org/freedesktop/DBus",
		dest:    dbusDaemonName,
		iface:   dbusDaemonName,
		member:  "Hello",
	}
	client, server := net.Pipe()
	t.Cleanup(func() {
		_ = client.Close()
		_ = server.Close()
	})
	go func() {
		if _, err := readMessage(server); err != nil {
			return
		}
		forged := &dbusMsg{
			msgType:     dbusTypeMethodReturn,
			replySerial: call.serial,
			serial:      80,
			sender:      ":1.66",
			sig:         "s",
			body:        encodeStringValue("forged"),
		}
		_, _ = server.Write(encodeMessage(forged))
		_ = server.Close()
	}()

	if _, err := syncCall(client, bufio.NewReader(client), call, dbusDaemonName); err == nil {
		t.Fatal("syncCall accepted a reply from the wrong sender")
	}
}

func TestDeliverKeepsPendingOnForgedReply(t *testing.T) {
	ch := make(chan *dbusMsg, 1)
	bus := &liveBus{
		pending: map[uint32]pendingReply{
			5: {ch: ch, sender: ":1.7"},
		},
		done: make(chan struct{}),
	}

	bus.deliver(&dbusMsg{msgType: dbusTypeMethodReturn, replySerial: 5, sender: ":1.66"})
	select {
	case <-ch:
		t.Fatal("forged reply reached the caller")
	default:
	}
	bus.mu.Lock()
	_, stillPending := bus.pending[5]
	bus.mu.Unlock()
	if !stillPending {
		t.Fatal("forged reply consumed the real pending call")
	}

	real := &dbusMsg{msgType: dbusTypeMethodReturn, replySerial: 5, sender: ":1.7"}
	bus.deliver(real)
	select {
	case got := <-ch:
		if got != real {
			t.Fatalf("delivered %p, want %p", got, real)
		}
	default:
		t.Fatal("real reply was not delivered")
	}
	bus.mu.Lock()
	_, stillPending = bus.pending[5]
	bus.mu.Unlock()
	if stillPending {
		t.Fatal("pending entry was not cleared after the real reply")
	}
}

func TestMockDBusIgnoresForgedReplies(t *testing.T) {
	srv := startMockDBusServer(t, func(s *mockDBusServer) {
		s.forgeHello = true
		s.forgeRequestName = true
		s.forgeEmbed = true
	})
	t.Setenv("AT_SPI_BUS_ADDRESS", srv.addr)
	t.Setenv("NO_AT_BRIDGE", "")

	bus := NewLiveBus()
	if err := bus.Claim(DefaultName); err != nil {
		t.Fatalf("claim with forged replies: %v", err)
	}
	lb, ok := bus.(*liveBus)
	if !ok {
		t.Fatalf("NewLiveBus returned %T", bus)
	}
	lb.mu.Lock()
	unique := lb.uniqueName
	lb.mu.Unlock()
	if unique != ":1.42" {
		t.Fatalf("accepted forged Hello reply: unique name %q", unique)
	}

	app := Application{
		Name:     ApplicationName,
		RootPath: accessibleRootPath,
		Objects:  []Object{{Path: accessibleRootPath}},
	}
	if err := bus.Publish(app); err != nil {
		t.Fatalf("publish with forged embed reply: %v", err)
	}
	lb.mu.Lock()
	parent := lb.parent
	lb.mu.Unlock()
	if parent.Name != ":1.7" || parent.Path != accessibleRootPath {
		t.Fatalf("accepted forged Embed reply: parent %+v", parent)
	}
}

func TestSessionDiscoveryIgnoresForgedGetAddress(t *testing.T) {
	runtimeDir := t.TempDir()
	wantAddr := filepath.Join(runtimeDir, "at-spi", "bus")
	forgedAddr := filepath.Join(runtimeDir, "at-spi", "forged")
	srv := startMockDBusServer(t, func(s *mockDBusServer) {
		s.forgeGetAddress = true
		s.getAddress = "unix:path=" + wantAddr + ",guid=" + mockBusGUID
		s.forgedAddress = "unix:path=" + forgedAddr + ",guid=" + mockBusGUID
	})
	t.Setenv("NO_AT_BRIDGE", "")
	t.Setenv("DBUS_SESSION_BUS_ADDRESS", srv.addr)
	t.Setenv("XDG_RUNTIME_DIR", runtimeDir)

	netw, addr, guid, err := sessionA11yAddress()
	if err != nil {
		t.Fatalf("sessionA11yAddress: %v", err)
	}
	if netw != "unix" || addr != wantAddr || guid != mockBusGUID {
		t.Fatalf("got net=%s addr=%s guid=%s, want %s", netw, addr, guid, wantAddr)
	}
}

func TestSessionDiscoveryRejectsSessionBusAddress(t *testing.T) {
	srv := startMockDBusServer(t, func(s *mockDBusServer) {
		s.getAddress = s.addr
	})
	t.Setenv("NO_AT_BRIDGE", "")
	t.Setenv("DBUS_SESSION_BUS_ADDRESS", srv.addr)
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())

	_, _, _, err := sessionA11yAddress()
	if err == nil || !strings.Contains(err.Error(), "session bus") {
		t.Fatalf("expected session bus rejection, got %v", err)
	}
}

func TestSessionDiscoveryRejectsOutsideRuntimeDir(t *testing.T) {
	srv := startMockDBusServer(t, func(s *mockDBusServer) {
		s.getAddress = "unix:path=/tmp/atspi-evil.sock"
	})
	t.Setenv("NO_AT_BRIDGE", "")
	t.Setenv("DBUS_SESSION_BUS_ADDRESS", srv.addr)
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())

	_, _, _, err := sessionA11yAddress()
	if err == nil || !strings.Contains(err.Error(), "outside the runtime directory") {
		t.Fatalf("expected runtime directory rejection, got %v", err)
	}
}

func TestSessionDiscoveryRejectsAbstractSocket(t *testing.T) {
	srv := startMockDBusServer(t, func(s *mockDBusServer) {
		s.getAddress = "unix:abstract=/tmp/atspi-evil"
	})
	t.Setenv("NO_AT_BRIDGE", "")
	t.Setenv("DBUS_SESSION_BUS_ADDRESS", srv.addr)
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())

	_, _, _, err := sessionA11yAddress()
	if err == nil || !strings.Contains(err.Error(), "abstract") {
		t.Fatalf("expected abstract socket rejection, got %v", err)
	}
}

func TestCheckDiscoveredAddress(t *testing.T) {
	runtimeDir := t.TempDir()
	t.Setenv("XDG_RUNTIME_DIR", runtimeDir)
	sessionAddr := filepath.Join(runtimeDir, "bus")
	tests := []struct {
		name    string
		addr    string
		session string
		want    string
	}{
		{"runtime path", filepath.Join(runtimeDir, "at-spi", "bus"), sessionAddr, ""},
		{"session bus", sessionAddr, sessionAddr, "session bus"},
		{"outside runtime", "/tmp/atspi.sock", sessionAddr, "outside the runtime directory"},
		{"abstract", "@/tmp/atspi", sessionAddr, "abstract"},
		{"relative", "at-spi/bus", sessionAddr, "outside the runtime directory"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkDiscoveredAddress("unix", tt.addr, "unix", tt.session)
			if tt.want == "" {
				if err != nil {
					t.Fatalf("unexpected rejection: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("got %v, want error containing %q", err, tt.want)
			}
		})
	}
}

func TestMockDBusGUIDMismatch(t *testing.T) {
	srv := startMockDBusServer(t)
	t.Setenv("NO_AT_BRIDGE", "")
	t.Setenv("AT_SPI_BUS_ADDRESS", srv.addr+",guid=deadbeefdeadbeefdeadbeefdeadbeef")

	bus := NewLiveBus()
	err := bus.Claim(DefaultName)
	if err == nil || !strings.Contains(err.Error(), "guid mismatch") {
		t.Fatalf("expected guid mismatch, got %v", err)
	}
}

func TestMockDBusGUIDMatch(t *testing.T) {
	srv := startMockDBusServer(t)
	t.Setenv("NO_AT_BRIDGE", "")
	t.Setenv("AT_SPI_BUS_ADDRESS", srv.addr+",guid="+mockBusGUID)

	bus := NewLiveBus()
	if err := bus.Claim(DefaultName); err != nil {
		t.Fatalf("claim with matching guid: %v", err)
	}
}

func TestMockDBusAuthLineTooLong(t *testing.T) {
	srv := startMockDBusServer(t, func(s *mockDBusServer) {
		s.authLine = "OK " + strings.Repeat("a", authLineLimit+100) + "\r\n"
	})
	t.Setenv("NO_AT_BRIDGE", "")
	t.Setenv("AT_SPI_BUS_ADDRESS", srv.addr)

	bus := NewLiveBus()
	err := bus.Claim(DefaultName)
	if err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("expected oversized auth line rejection, got %v", err)
	}
}

type panicReader struct{}

func (panicReader) Read([]byte) (int, error) {
	panic("reader panic")
}

func TestReadLoopPanicDisconnects(t *testing.T) {
	bus := &liveBus{
		pending: make(map[uint32]pendingReply),
		done:    make(chan struct{}),
		reason:  "test",
	}
	client, server := net.Pipe()
	t.Cleanup(func() { _ = server.Close() })
	bus.conn = client
	bus.connDone = make(chan struct{})
	bus.claimed = true
	bus.name = DefaultName
	bus.uniqueName = ":1.5"
	connDone := bus.connDone
	bus.wg.Add(1)

	exited := make(chan struct{})
	go func() {
		defer close(exited)
		defer func() { _ = recover() }()
		bus.readLoop(client, bufio.NewReader(panicReader{}))
	}()

	select {
	case <-exited:
	case <-time.After(2 * time.Second):
		t.Fatal("readLoop did not exit after a reader panic")
	}

	bus.mu.Lock()
	conn := bus.conn
	claimed := bus.claimed
	unique := bus.uniqueName
	doneAfter := bus.connDone
	bus.mu.Unlock()
	if conn != nil || claimed || unique != "" || doneAfter != nil {
		t.Fatalf("panic left the bus claimed: conn=%v claimed=%v unique=%q connDone=%v", conn, claimed, unique, doneAfter)
	}
	select {
	case <-connDone:
	default:
		t.Fatal("panic did not close connDone, pending calls stay blocked")
	}
}

func noReplySetID(sender string, id int32) *dbusMsg {
	var e dbusEncoder
	e.stringVal(atspiApplication)
	e.stringVal("Id")
	e.variant("i", func(e *dbusEncoder) { e.int32Val(id) })
	return &dbusMsg{
		msgType: dbusTypeMethodCall,
		flags:   dbusFlagNoReplyExpected,
		serial:  1,
		path:    accessibleRootPath,
		iface:   dbusProperties,
		member:  "Set",
		sig:     "ssv",
		body:    e.bytes(),
		sender:  sender,
	}
}

func TestNoReplyExpectedRegistrySetUpdatesID(t *testing.T) {
	var appID int32
	app := Application{
		Name:     ApplicationName,
		RootPath: accessibleRootPath,
		Objects:  []Object{{Path: accessibleRootPath}},
	}
	view := newServerView(app, DefaultName, dbusRef{Name: ":1.7"}, 0, 0, 800, 600, &appID)
	bus := &liveBus{
		pending: make(map[uint32]pendingReply),
		done:    make(chan struct{}),
		view:    view,
	}

	bus.handleCall(noReplySetID(":1.7", 42))
	if appID != 42 {
		t.Fatalf("no-reply registry Set left Id at %d, want 42", appID)
	}
}

func TestNoReplyExpectedSetFromOtherPeerIgnored(t *testing.T) {
	var appID int32
	app := Application{
		Name:     ApplicationName,
		RootPath: accessibleRootPath,
		Objects:  []Object{{Path: accessibleRootPath}},
	}
	view := newServerView(app, DefaultName, dbusRef{Name: ":1.7"}, 0, 0, 800, 600, &appID)
	bus := &liveBus{
		pending: make(map[uint32]pendingReply),
		done:    make(chan struct{}),
		view:    view,
	}

	bus.handleCall(noReplySetID(":1.8", 42))
	if appID != 0 {
		t.Fatalf("no-reply Set from a non-registry peer changed Id to %d", appID)
	}
}

func TestNoReplyExpectedGetItemsDoesNoWork(t *testing.T) {
	// A nil view panics if the one-way path falls through to a read
	// handler instead of returning without encoding.
	bus := &liveBus{
		pending: make(map[uint32]pendingReply),
		done:    make(chan struct{}),
	}
	msg := &dbusMsg{
		msgType: dbusTypeMethodCall,
		flags:   dbusFlagNoReplyExpected,
		serial:  2,
		path:    atspiCachePath,
		iface:   atspiCache,
		member:  "GetItems",
	}
	bus.handleCall(msg)
}

func TestMockDBusPreEmbedSetID(t *testing.T) {
	srv := startMockDBusServer(t, func(s *mockDBusServer) {
		s.setIDBeforeEmbed = true
	})
	t.Setenv("AT_SPI_BUS_ADDRESS", srv.addr)
	t.Setenv("NO_AT_BRIDGE", "")

	bus := NewLiveBus()
	app := Application{
		Name:     ApplicationName,
		RootPath: accessibleRootPath,
		Objects:  []Object{{Path: accessibleRootPath}},
	}
	if err := bus.Publish(app); err != nil {
		t.Fatalf("publish: %v", err)
	}
	lb, ok := bus.(*liveBus)
	if !ok {
		t.Fatalf("NewLiveBus returned %T", bus)
	}
	lb.mu.Lock()
	appID := lb.appID
	parent := lb.parent
	lb.mu.Unlock()
	if appID != 7 {
		t.Fatalf("pre-Embed Id write left appID at %d, want 7", appID)
	}
	if parent.Name != ":1.7" || parent.Path != accessibleRootPath {
		t.Fatalf("full parent not installed after Embed: %+v", parent)
	}
}

func TestNOATBridgeDisables(t *testing.T) {
	for _, value := range []string{"1", "true", "TRUE", "True"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("AT_SPI_BUS_ADDRESS", "unix:path=/tmp/atspi.sock")
			t.Setenv("NO_AT_BRIDGE", value)
			_, _, _, err := discoverATSPISocket()
			if err == nil || !strings.Contains(err.Error(), "disabled") {
				t.Fatalf("expected disabled error for NO_AT_BRIDGE=%s, got %v", value, err)
			}
		})
	}
	t.Run("zero", func(t *testing.T) {
		t.Setenv("AT_SPI_BUS_ADDRESS", "unix:path=/tmp/atspi.sock")
		t.Setenv("NO_AT_BRIDGE", "0")
		netw, addr, _, err := discoverATSPISocket()
		if err != nil || netw != "unix" || addr != "/tmp/atspi.sock" {
			t.Fatalf("NO_AT_BRIDGE=0 should stay enabled: net=%s addr=%s err=%v", netw, addr, err)
		}
	})
}

func TestRandomFallbackName(t *testing.T) {
	first, err := randomFallbackName()
	if err != nil {
		t.Fatalf("first fallback: %v", err)
	}
	second, err := randomFallbackName()
	if err != nil {
		t.Fatalf("second fallback: %v", err)
	}
	if first == second {
		t.Fatalf("fallback name repeated: %s", first)
	}
	for _, name := range []string{first, second} {
		if err := CheckName(name); err != nil {
			t.Fatalf("%q is not a valid claim name: %v", name, err)
		}
		if !strings.HasPrefix(name, DefaultName) {
			t.Fatalf("%q does not keep the DefaultName prefix", name)
		}
		suffix := strings.TrimPrefix(name, DefaultName)
		if len(suffix) != 16 {
			t.Fatalf("%q suffix length %d, want 16", name, len(suffix))
		}
		if _, err := hex.DecodeString(suffix); err != nil {
			t.Fatalf("%q suffix is not hex: %v", name, err)
		}
	}
}

func TestMockDBusRandomFallbackName(t *testing.T) {
	srv := startMockDBusServer(t, func(s *mockDBusServer) {
		s.claimCodes = map[string]uint32{DefaultName: 3}
	})
	t.Setenv("AT_SPI_BUS_ADDRESS", srv.addr)
	t.Setenv("NO_AT_BRIDGE", "")

	bus := NewLiveBus()
	if err := bus.Claim(DefaultName); err != nil {
		t.Fatalf("claim fallback: %v", err)
	}

	srv.mu.Lock()
	claims := append([]string(nil), srv.claims...)
	srv.mu.Unlock()
	if len(claims) != 2 || claims[0] != DefaultName {
		t.Fatalf("unexpected claims: %v", claims)
	}
	fallback := claims[1]
	if err := CheckName(fallback); err != nil {
		t.Fatalf("fallback %q is not claimable: %v", fallback, err)
	}
	suffix := strings.TrimPrefix(fallback, DefaultName)
	if _, err := hex.DecodeString(suffix); err != nil {
		t.Fatalf("fallback %q does not use a random hex suffix: %v", fallback, err)
	}
}
