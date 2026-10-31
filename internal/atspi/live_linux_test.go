//go:build linux && !android && !ios && !js

package atspi

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

type mockDBusServer struct {
	listener net.Listener
	addr     string
	mu       sync.Mutex
	authCmd  string
	helloGot bool
	claims   []string
	embeds   []string
	signals  []dbusMsg
	err      error

	rejectAuth bool
	helloErr   bool
	claimCode  uint32
	embedErr   bool
	noEmbed    bool
	injectAck  bool

	// Security-test knobs.
	serverGUID       string
	authLine         string
	getAddress       string
	busOwner         string
	registryOwner    string
	claimCodes       map[string]uint32
	forgeHello       bool
	forgeRequestName bool
	forgeGetAddress  bool
	forgeEmbed       bool
	forgedSender     string
	forgedAddress    string
	setIDBeforeEmbed bool
}

func startMockDBusServer(t *testing.T, opts ...func(*mockDBusServer)) *mockDBusServer {
	t.Helper()
	dir := t.TempDir()
	sockPath := filepath.Join(dir, "mock-atspi.sock")
	l, err := (&net.ListenConfig{}).Listen(t.Context(), "unix", sockPath)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	srv := &mockDBusServer{
		listener:      l,
		addr:          "unix:path=" + sockPath,
		claimCode:     1, // default PrimaryOwner
		serverGUID:    mockBusGUID,
		busOwner:      ":1.9",
		registryOwner: ":1.7",
		forgedSender:  ":1.66",
	}
	for _, opt := range opts {
		opt(srv)
	}
	if srv.getAddress == "" {
		srv.getAddress = srv.addr
	}
	if srv.forgedAddress == "" {
		srv.forgedAddress = "unix:path=/tmp/forged-atspi"
	}

	go srv.serve()

	t.Cleanup(func() {
		_ = l.Close()
	})

	return srv
}

const mockBusGUID = "mock-guid-for-tests"

func (s *mockDBusServer) serve() {
	conn, err := s.listener.Accept()
	if err != nil {
		return
	}
	defer func() { _ = conn.Close() }()

	var nullByte [1]byte
	if _, err := conn.Read(nullByte[:]); err != nil {
		s.recordErr(err)
		return
	}

	r := bufio.NewReader(conn)
	authLine, err := r.ReadString('\n')
	if err != nil {
		s.recordErr(err)
		return
	}
	s.mu.Lock()
	s.authCmd = strings.TrimSpace(authLine)
	s.mu.Unlock()

	if s.rejectAuth {
		_, _ = conn.Write([]byte("REJECTED EXTERNAL\r\n"))
		return
	}

	s.mu.Lock()
	responseLine := s.authLine
	if responseLine == "" {
		responseLine = "OK " + s.serverGUID + "\r\n"
	}
	s.mu.Unlock()
	if _, err := conn.Write([]byte(responseLine)); err != nil {
		s.recordErr(err)
		return
	}

	beginLine, err := r.ReadString('\n')
	if err != nil {
		s.recordErr(err)
		return
	}
	if strings.TrimSpace(beginLine) != "BEGIN" {
		s.recordErr(fmt.Errorf("unexpected begin line: %s", beginLine))
		return
	}

	for {
		msg, err := readMessage(r)
		if err != nil {
			return
		}
		switch msg.msgType {
		case dbusTypeMethodCall:
			switch {
			case msg.iface == dbusDaemonName && msg.member == "Hello":
				s.mu.Lock()
				s.helloGot = true
				s.mu.Unlock()

				if s.helloErr {
					reply := &dbusMsg{
						msgType:     dbusTypeError,
						replySerial: msg.serial,
						serial:      100,
						dest:        ":1.42",
						sender:      dbusDaemonName,
						errorName:   "org.freedesktop.DBus.Error.Failed",
						sig:         "s",
						body:        encodeStringBody("hello rejected"),
					}
					_, _ = conn.Write(encodeMessage(reply))
					return
				}

				if s.forgeHello {
					s.writeForgedReply(conn, msg.serial, 98, "s", encodeStringBody(":1.99"))
				}
				reply := &dbusMsg{
					msgType:     dbusTypeMethodReturn,
					replySerial: msg.serial,
					serial:      100,
					dest:        ":1.42",
					sender:      dbusDaemonName,
					sig:         "s",
					body:        encodeStringBody(":1.42"),
				}
				_, _ = conn.Write(encodeMessage(reply))
				if s.injectAck {
					// The real bus queues NameAcquired after Hello.
					ack := &dbusMsg{
						msgType: dbusTypeSignal,
						serial:  99,
						path:    "/org/freedesktop/DBus",
						iface:   "org.freedesktop.DBus",
						member:  "NameAcquired",
						sig:     "s",
						body:    encodeStringValue(":1.42"),
					}
					_, _ = conn.Write(encodeMessage(ack))
				}
			case msg.iface == dbusDaemonName && msg.member == "RequestName":
				name, _, _ := decodeRequestNameBody(msg.body)
				s.mu.Lock()
				s.claims = append(s.claims, name)
				code := s.claimCode
				if override, ok := s.claimCodes[name]; ok {
					code = override
				}
				s.mu.Unlock()

				if s.forgeRequestName {
					s.writeForgedReply(conn, msg.serial, 97, "u", encodeUint32Body(3))
				}
				reply := &dbusMsg{
					msgType:     dbusTypeMethodReturn,
					replySerial: msg.serial,
					serial:      101,
					dest:        ":1.42",
					sender:      dbusDaemonName,
					sig:         "u",
					body:        encodeUint32Body(code),
				}
				_, _ = conn.Write(encodeMessage(reply))
			case msg.iface == dbusDaemonName && msg.member == "StartServiceByName":
				reply := &dbusMsg{
					msgType:     dbusTypeMethodReturn,
					replySerial: msg.serial,
					serial:      106,
					dest:        ":1.42",
					sender:      dbusDaemonName,
					sig:         "u",
					body:        encodeUint32Body(1),
				}
				_, _ = conn.Write(encodeMessage(reply))
			case msg.iface == dbusDaemonName && msg.member == "GetNameOwner":
				name, _ := decodeStringBody(msg.body)
				owner := ""
				switch name {
				case a11yBusName:
					owner = s.busOwner
				case a11yRegistryName:
					owner = s.registryOwner
				}
				if owner == "" {
					reply := &dbusMsg{
						msgType:     dbusTypeError,
						replySerial: msg.serial,
						serial:      104,
						dest:        ":1.42",
						sender:      dbusDaemonName,
						errorName:   "org.freedesktop.DBus.Error.NameHasNoOwner",
						sig:         "s",
						body:        encodeStringBody("name has no owner"),
					}
					_, _ = conn.Write(encodeMessage(reply))
					continue
				}
				reply := &dbusMsg{
					msgType:     dbusTypeMethodReturn,
					replySerial: msg.serial,
					serial:      105,
					dest:        ":1.42",
					sender:      dbusDaemonName,
					sig:         "s",
					body:        encodeStringValue(owner),
				}
				_, _ = conn.Write(encodeMessage(reply))
			case msg.iface == "org.a11y.atspi.Socket" && msg.member == "Embed":
				name, root, _ := decodeEmbedBody(msg.body)
				s.mu.Lock()
				s.embeds = append(s.embeds, name+":"+root)
				noEmbed, embedErr := s.noEmbed, s.embedErr
				s.mu.Unlock()
				if noEmbed {
					continue
				}
				if embedErr {
					reply := &dbusMsg{
						msgType:     dbusTypeError,
						replySerial: msg.serial,
						serial:      102,
						dest:        ":1.42",
						sender:      s.registryOwner,
						errorName:   "org.freedesktop.DBus.Error.Failed",
						sig:         "s",
						body:        encodeStringBody("embed rejected"),
					}
					_, _ = conn.Write(encodeMessage(reply))
					continue
				}
				if s.forgeEmbed {
					s.writeForgedReply(conn, msg.serial, 96, "(so)", encodeStructSO(":1.2", "/org/a11y/atspi/accessible/forged"))
				}
				if s.setIDBeforeEmbed {
					// registryd writes Application.Id before it sends
					// the Embed reply, and sends it one-way.
					var e dbusEncoder
					e.stringVal(atspiApplication)
					e.stringVal("Id")
					e.variant("i", func(e *dbusEncoder) { e.int32Val(7) })
					setCall := &dbusMsg{
						msgType: dbusTypeMethodCall,
						flags:   dbusFlagNoReplyExpected,
						serial:  90,
						path:    root,
						dest:    name,
						iface:   dbusProperties,
						member:  "Set",
						sig:     "ssv",
						body:    e.bytes(),
						sender:  s.registryOwner,
					}
					_, _ = conn.Write(encodeMessage(setCall))
				}
				reply := &dbusMsg{
					msgType:     dbusTypeMethodReturn,
					replySerial: msg.serial,
					serial:      102,
					dest:        ":1.42",
					sender:      s.registryOwner,
					sig:         "(so)",
					body:        encodeStructSO(s.registryOwner, "/org/a11y/atspi/accessible/root"),
				}
				_, _ = conn.Write(encodeMessage(reply))
			case msg.iface == a11yBusName && msg.member == "GetAddress":
				if s.forgeGetAddress {
					s.writeForgedReply(conn, msg.serial, 95, "s", encodeStringValue(s.forgedAddress))
				}
				reply := &dbusMsg{
					msgType:     dbusTypeMethodReturn,
					replySerial: msg.serial,
					serial:      103,
					dest:        ":1.42",
					sender:      s.busOwner,
					sig:         "s",
					body:        encodeStringValue(s.getAddress),
				}
				_, _ = conn.Write(encodeMessage(reply))
			}
		case dbusTypeSignal:
			s.mu.Lock()
			s.signals = append(s.signals, *msg)
			s.mu.Unlock()
		}
	}
}

func (s *mockDBusServer) writeForgedReply(conn net.Conn, replySerial, serial uint32, sig string, body []byte) {
	reply := &dbusMsg{
		msgType:     dbusTypeMethodReturn,
		replySerial: replySerial,
		serial:      serial,
		dest:        ":1.42",
		sender:      s.forgedSender,
		sig:         sig,
		body:        body,
	}
	_, _ = conn.Write(encodeMessage(reply))
}

func (s *mockDBusServer) recordErr(err error) {
	s.mu.Lock()
	s.err = err
	s.mu.Unlock()
}

func TestMockDBusLifecycle(t *testing.T) {
	srv := startMockDBusServer(t)
	t.Setenv("AT_SPI_BUS_ADDRESS", srv.addr)

	bus := NewLiveBus()
	if err := bus.Claim(DefaultName); err != nil {
		t.Fatalf("claim: %v", err)
	}
	if !bus.Live() {
		t.Fatalf("expected bus to be live, status: %+v", bus.Status())
	}
	st := bus.Status()
	if !st.Live || st.Platform != runtime.GOOS || !strings.Contains(st.Reason, "connected to") {
		t.Fatalf("unexpected status: %+v", st)
	}

	// Verify handshake was recorded
	expectedUIDHex := hex.EncodeToString([]byte(strconv.Itoa(os.Getuid())))
	expectedAuth := "AUTH EXTERNAL " + expectedUIDHex
	srv.mu.Lock()
	if srv.authCmd != expectedAuth {
		srv.mu.Unlock()
		t.Fatalf("expected auth %q, got %q", expectedAuth, srv.authCmd)
	}
	if !srv.helloGot {
		srv.mu.Unlock()
		t.Fatal("expected Hello method call")
	}
	if len(srv.claims) != 1 || srv.claims[0] != DefaultName {
		srv.mu.Unlock()
		t.Fatalf("expected initial claim %s, got %v", DefaultName, srv.claims)
	}
	srv.mu.Unlock()

	// Claim existing name (fast path, already claimed)
	if err := bus.Claim(DefaultName); err != nil {
		t.Fatalf("claim default name: %v", err)
	}

	// Claim new valid name
	customName := "org.a11y.atspi.ownframe.custom"
	if err := bus.Claim(customName); err != nil {
		t.Fatalf("claim custom name: %v", err)
	}
	srv.mu.Lock()
	if len(srv.claims) != 2 || srv.claims[1] != customName {
		srv.mu.Unlock()
		t.Fatalf("expected claims to include %s, got %v", customName, srv.claims)
	}
	srv.mu.Unlock()

	// Claim invalid name
	if err := bus.Claim("org.freedesktop.DBus"); !errors.Is(err, ErrName) {
		t.Fatalf("expected ErrName, got %v", err)
	}

	// Emit signal
	ev := Event{Path: "/org/a11y/atspi/accessible/n1", Kind: "children-changed", Detail: "add", Child: "/org/a11y/atspi/accessible/n1"}
	if err := bus.Emit(ev); err != nil {
		t.Fatalf("emit: %v", err)
	}

	// Publish app
	app := Application{
		Name:     ApplicationName,
		RootPath: "/org/a11y/atspi/accessible/nroot",
	}
	if err := bus.Publish(app); err != nil {
		t.Fatalf("publish: %v", err)
	}

	// Wait briefly for server to process messages
	time.Sleep(50 * time.Millisecond)

	srv.mu.Lock()
	if len(srv.signals) != 1 {
		srv.mu.Unlock()
		t.Fatalf("expected 1 signal, got %d", len(srv.signals))
	}
	sig := srv.signals[0]
	if sig.path != "/org/a11y/atspi/accessible/n1" || sig.iface != "org.a11y.atspi.Event.Object" || sig.member != "ChildrenChanged" {
		srv.mu.Unlock()
		t.Fatalf("unexpected signal: %+v", sig)
	}
	if sig.sig != "siiva{sv}" {
		srv.mu.Unlock()
		t.Fatalf("unexpected signal signature: %s", sig.sig)
	}
	if len(srv.embeds) != 1 || srv.embeds[0] != ":1.42:/org/a11y/atspi/accessible/nroot" {
		srv.mu.Unlock()
		t.Fatalf("unexpected embeds: %v", srv.embeds)
	}
	srv.mu.Unlock()

	// Close bus
	if err := bus.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	if bus.Live() {
		t.Fatal("expected bus to not be live after Close")
	}
	if err := bus.Claim(DefaultName); !errors.Is(err, ErrLiveUnavailable) {
		t.Fatalf("expected ErrLiveUnavailable after close, got %v", err)
	}
	if err := bus.Publish(app); !errors.Is(err, ErrLiveUnavailable) {
		t.Fatalf("expected ErrLiveUnavailable after close, got %v", err)
	}
	if err := bus.Emit(ev); !errors.Is(err, ErrLiveUnavailable) {
		t.Fatalf("expected ErrLiveUnavailable after close, got %v", err)
	}
}

func TestMockDBusPublisherIntegration(t *testing.T) {
	srv := startMockDBusServer(t)
	t.Setenv("AT_SPI_BUS_ADDRESS", srv.addr)

	bus := NewLiveBus()
	p := NewPublisher(bus)
	tree := sampleTree()

	if err := p.Publish(tree); err != nil {
		t.Fatalf("publish: %v", err)
	}
	if !p.Live() {
		t.Fatal("expected publisher to be live")
	}

	time.Sleep(50 * time.Millisecond)

	srv.mu.Lock()
	defer srv.mu.Unlock()
	if len(srv.signals) == 0 {
		t.Fatal("expected children-changed signal")
	}
	if len(srv.embeds) == 0 {
		t.Fatal("expected embed registration")
	}
}

func TestMockDBusIgnoresQueuedNameAcquired(t *testing.T) {
	srv := startMockDBusServer(t, func(s *mockDBusServer) {
		s.injectAck = true
	})
	t.Setenv("AT_SPI_BUS_ADDRESS", srv.addr)

	bus := NewLiveBus()
	if err := bus.Claim(DefaultName); err != nil {
		t.Fatalf("claim with queued signal: %v", err)
	}
	if !bus.Live() {
		t.Fatalf("not live: %+v", bus.Status())
	}
}

func TestMockDBusEmbedError(t *testing.T) {
	srv := startMockDBusServer(t, func(s *mockDBusServer) {
		s.embedErr = true
	})
	t.Setenv("AT_SPI_BUS_ADDRESS", srv.addr)

	bus := NewLiveBus()
	if err := bus.Claim(DefaultName); err != nil {
		t.Fatalf("claim: %v", err)
	}
	err := bus.Publish(Application{Name: ApplicationName, RootPath: accessibleRootPath})
	if err == nil || !strings.Contains(err.Error(), "embed") {
		t.Fatalf("publish error: %v", err)
	}
}

func TestMockDBusAuthRejected(t *testing.T) {
	srv := startMockDBusServer(t, func(s *mockDBusServer) {
		s.rejectAuth = true
	})
	t.Setenv("AT_SPI_BUS_ADDRESS", srv.addr)

	bus := NewLiveBus()
	if err := bus.Claim(DefaultName); err == nil {
		t.Fatal("expected claim to fail when auth is rejected")
	}
	st := bus.Status()
	if st.Live || !strings.Contains(st.Reason, "auth rejected") {
		t.Fatalf("unexpected status: %+v", st)
	}
}

func TestMockDBusHelloError(t *testing.T) {
	srv := startMockDBusServer(t, func(s *mockDBusServer) {
		s.helloErr = true
	})
	t.Setenv("AT_SPI_BUS_ADDRESS", srv.addr)

	bus := NewLiveBus()
	if err := bus.Claim(DefaultName); err == nil {
		t.Fatal("expected claim to fail when Hello fails")
	}
	st := bus.Status()
	if st.Live || !strings.Contains(st.Reason, "hello failed") {
		t.Fatalf("unexpected status: %+v", st)
	}
}

func TestMockDBusClaimFailed(t *testing.T) {
	srv := startMockDBusServer(t, func(s *mockDBusServer) {
		s.claimCode = 3 // Exists
	})
	t.Setenv("AT_SPI_BUS_ADDRESS", srv.addr)

	bus := NewLiveBus()
	if err := bus.Claim(DefaultName); err == nil {
		t.Fatal("expected claim to fail when the name exists")
	}
	st := bus.Status()
	if st.Live || !strings.Contains(st.Reason, "claim failed") {
		t.Fatalf("unexpected status: %+v", st)
	}
}

func TestDiscoverATSPISocketFormats(t *testing.T) {
	t.Setenv("NO_AT_BRIDGE", "")

	// 1. unix:path
	netw, addr, guid, err := parseDBusAddress("unix:path=/run/user/1000/at-spi/bus_0,guid=123")
	if err != nil || netw != "unix" || addr != "/run/user/1000/at-spi/bus_0" || guid != "123" {
		t.Fatalf("got net=%s, addr=%s, guid=%s, err=%v", netw, addr, guid, err)
	}

	// 2. unix:abstract
	netw, addr, guid, err = parseDBusAddress("unix:abstract=/tmp/dbus-test,guid=abc")
	if err != nil || netw != "unix" || addr != "@/tmp/dbus-test" || guid != "abc" {
		t.Fatalf("got net=%s, addr=%s, guid=%s, err=%v", netw, addr, guid, err)
	}

	// 3. raw unix path
	netw, addr, guid, err = parseDBusAddress("/run/user/1000/at-spi/bus")
	if err != nil || netw != "unix" || addr != "/run/user/1000/at-spi/bus" || guid != "" {
		t.Fatalf("got net=%s, addr=%s, guid=%s, err=%v", netw, addr, guid, err)
	}

	// 4. explicit accessibility bus address (trusted, even outside the runtime dir)
	t.Setenv("AT_SPI_BUS_ADDRESS", "unix:path=/tmp/atspi.sock,guid=abc")
	netw, addr, guid, err = discoverATSPISocket()
	if err != nil || netw != "unix" || addr != "/tmp/atspi.sock" || guid != "abc" {
		t.Fatalf("got net=%s, addr=%s, guid=%s, err=%v", netw, addr, guid, err)
	}

	// 5. disabled
	t.Setenv("AT_SPI_BUS_ADDRESS", "disabled")
	_, _, _, err = discoverATSPISocket()
	if err == nil || !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("expected disabled error, got %v", err)
	}

	// 6. none
	t.Setenv("AT_SPI_BUS_ADDRESS", "none")
	_, _, _, err = discoverATSPISocket()
	if err == nil || !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("expected disabled error, got %v", err)
	}

	// 7. no AT_SPI_BUS_ADDRESS: ask org.a11y.Bus on the session bus.
	runtimeDir := t.TempDir()
	wantAddr := filepath.Join(runtimeDir, "at-spi", "bus")
	wantGUID := "mock-guid-second"
	srv := startMockDBusServer(t, func(s *mockDBusServer) {
		s.getAddress = "unix:path=" + wantAddr + ",guid=" + wantGUID
	})
	t.Setenv("AT_SPI_BUS_ADDRESS", "")
	t.Setenv("DBUS_SESSION_BUS_ADDRESS", srv.addr)
	t.Setenv("XDG_RUNTIME_DIR", runtimeDir)
	netw, addr, guid, err = discoverATSPISocket()
	if err != nil || netw != "unix" || addr != wantAddr || guid != wantGUID {
		t.Fatalf("got net=%s, addr=%s, guid=%s, err=%v; want %s", netw, addr, guid, err, wantAddr)
	}

	// 8. invalid address
	_, _, _, err = parseDBusAddress("tcp:host=localhost,port=1234")
	if err == nil {
		t.Fatal("expected error for unsupported tcp address")
	}
}

func TestDBusWireRoundTrip(t *testing.T) {
	msg := &dbusMsg{
		msgType:     dbusTypeMethodCall,
		flags:       dbusFlagNoReplyExpected,
		serial:      42,
		replySerial: 21,
		path:        "/org/test/path",
		iface:       "org.test.Interface",
		member:      "TestMethod",
		dest:        "org.test.Dest",
		sender:      ":1.99",
		sig:         "s",
		body:        encodeStringBody("hello world"),
	}

	raw := encodeMessage(msg)
	decoded, err := readMessage(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("readMessage: %v", err)
	}

	if decoded.msgType != msg.msgType {
		t.Fatalf("msgType: got %d, want %d", decoded.msgType, msg.msgType)
	}
	if decoded.flags != msg.flags {
		t.Fatalf("flags: got %d, want %d", decoded.flags, msg.flags)
	}
	if decoded.serial != msg.serial {
		t.Fatalf("serial: got %d, want %d", decoded.serial, msg.serial)
	}
	if decoded.replySerial != msg.replySerial {
		t.Fatalf("replySerial: got %d, want %d", decoded.replySerial, msg.replySerial)
	}
	if decoded.path != msg.path {
		t.Fatalf("path: got %s, want %s", decoded.path, msg.path)
	}
	if decoded.iface != msg.iface {
		t.Fatalf("iface: got %s, want %s", decoded.iface, msg.iface)
	}
	if decoded.member != msg.member {
		t.Fatalf("member: got %s, want %s", decoded.member, msg.member)
	}
	if decoded.dest != msg.dest {
		t.Fatalf("dest: got %s, want %s", decoded.dest, msg.dest)
	}
	if decoded.sender != msg.sender {
		t.Fatalf("sender: got %s, want %s", decoded.sender, msg.sender)
	}
	if decoded.sig != msg.sig {
		t.Fatalf("sig: got %s, want %s", decoded.sig, msg.sig)
	}

	bodyStr, err := decodeStringBody(decoded.body)
	if err != nil || bodyStr != "hello world" {
		t.Fatalf("body: got %s, err %v", bodyStr, err)
	}
}
