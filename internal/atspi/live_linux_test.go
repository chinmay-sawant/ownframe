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
	closed   bool
	err      error

	rejectAuth bool
	helloErr   bool
	claimCode  uint32
}

func startMockDBusServer(t *testing.T, opts ...func(*mockDBusServer)) *mockDBusServer {
	t.Helper()
	dir := t.TempDir()
	sockPath := filepath.Join(dir, "mock-atspi.sock")
	l, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	srv := &mockDBusServer{
		listener:  l,
		addr:      "unix:path=" + sockPath,
		claimCode: 1, // default PrimaryOwner
	}
	for _, opt := range opts {
		opt(srv)
	}

	go srv.serve()

	t.Cleanup(func() {
		_ = l.Close()
	})

	return srv
}

func (s *mockDBusServer) serve() {
	conn, err := s.listener.Accept()
	if err != nil {
		return
	}
	defer conn.Close()

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

	if _, err := conn.Write([]byte("OK 0123456789abcdef0123456789abcdef\r\n")); err != nil {
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
			if msg.iface == "org.freedesktop.DBus" && msg.member == "Hello" {
				s.mu.Lock()
				s.helloGot = true
				s.mu.Unlock()

				if s.helloErr {
					reply := &dbusMsg{
						msgType:     dbusTypeError,
						replySerial: msg.serial,
						serial:      100,
						errorName:   "org.freedesktop.DBus.Error.Failed",
						dest:        ":1.42",
						sig:         "s",
						body:        encodeStringBody("hello rejected"),
					}
					_, _ = conn.Write(encodeMessage(reply))
					return
				}

				reply := &dbusMsg{
					msgType:     dbusTypeMethodReturn,
					replySerial: msg.serial,
					serial:      100,
					dest:        ":1.42",
					sig:         "s",
					body:        encodeStringBody(":1.42"),
				}
				_, _ = conn.Write(encodeMessage(reply))
			} else if msg.iface == "org.freedesktop.DBus" && msg.member == "RequestName" {
				name, _, _ := decodeRequestNameBody(msg.body)
				s.mu.Lock()
				s.claims = append(s.claims, name)
				code := s.claimCode
				s.mu.Unlock()

				reply := &dbusMsg{
					msgType:     dbusTypeMethodReturn,
					replySerial: msg.serial,
					serial:      101,
					dest:        ":1.42",
					sig:         "u",
					body:        encodeUint32Body(code),
				}
				_, _ = conn.Write(encodeMessage(reply))
			} else if msg.iface == "org.a11y.atspi.Socket" && msg.member == "Embed" {
				name, root, _ := decodeEmbedBody(msg.body)
				s.mu.Lock()
				s.embeds = append(s.embeds, name+":"+root)
				s.mu.Unlock()
			}
		case dbusTypeSignal:
			s.mu.Lock()
			s.signals = append(s.signals, *msg)
			s.mu.Unlock()
		}
	}
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
	ev := Event{Path: "/org/a11y/atspi/accessible/n1", Kind: "children-changed"}
	if err := bus.Emit(ev); err != nil {
		t.Fatalf("emit: %v", err)
	}

	// Publish app
	app := Application{
		Name:     DefaultName,
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
	if len(srv.embeds) != 1 || srv.embeds[0] != DefaultName+":/org/a11y/atspi/accessible/nroot" {
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
	if !p.Live() {
		t.Fatal("expected publisher to be live")
	}

	tree := sampleTree()

	if err := p.Publish(tree); err != nil {
		t.Fatalf("publish: %v", err)
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

func TestMockDBusAuthRejected(t *testing.T) {
	srv := startMockDBusServer(t, func(s *mockDBusServer) {
		s.rejectAuth = true
	})
	t.Setenv("AT_SPI_BUS_ADDRESS", srv.addr)

	bus := NewLiveBus()
	if bus.Live() {
		t.Fatal("expected bus to not be live when auth is rejected")
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
	if bus.Live() {
		t.Fatal("expected bus to not be live when Hello fails")
	}
	st := bus.Status()
	if st.Live || !strings.Contains(st.Reason, "hello error") {
		t.Fatalf("unexpected status: %+v", st)
	}
}

func TestMockDBusClaimFailed(t *testing.T) {
	srv := startMockDBusServer(t, func(s *mockDBusServer) {
		s.claimCode = 3 // Exists
	})
	t.Setenv("AT_SPI_BUS_ADDRESS", srv.addr)

	bus := NewLiveBus()
	if bus.Live() {
		t.Fatal("expected bus to not be live when Claim fails")
	}
	st := bus.Status()
	if st.Live || !strings.Contains(st.Reason, "claim failed") {
		t.Fatalf("unexpected status: %+v", st)
	}
}

func TestDiscoverATSPISocketFormats(t *testing.T) {
	// 1. unix:path
	netw, addr, err := parseDBusAddress("unix:path=/run/user/1000/at-spi/bus_0,guid=123")
	if err != nil || netw != "unix" || addr != "/run/user/1000/at-spi/bus_0" {
		t.Fatalf("got net=%s, addr=%s, err=%v", netw, addr, err)
	}

	// 2. unix:abstract
	netw, addr, err = parseDBusAddress("unix:abstract=/tmp/dbus-test,guid=abc")
	if err != nil || netw != "unix" || addr != "@/tmp/dbus-test" {
		t.Fatalf("got net=%s, addr=%s, err=%v", netw, addr, err)
	}

	// 3. raw unix path
	netw, addr, err = parseDBusAddress("/run/user/1000/at-spi/bus")
	if err != nil || netw != "unix" || addr != "/run/user/1000/at-spi/bus" {
		t.Fatalf("got net=%s, addr=%s, err=%v", netw, addr, err)
	}

	// 4. disabled
	t.Setenv("AT_SPI_BUS_ADDRESS", "disabled")
	_, _, err = discoverATSPISocket()
	if err == nil || !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("expected disabled error, got %v", err)
	}

	// 5. none
	t.Setenv("AT_SPI_BUS_ADDRESS", "none")
	_, _, err = discoverATSPISocket()
	if err == nil || !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("expected disabled error, got %v", err)
	}

	// 6. DBUS_SESSION_BUS_ADDRESS fallback
	t.Setenv("AT_SPI_BUS_ADDRESS", "")
	t.Setenv("DBUS_SESSION_BUS_ADDRESS", "unix:path=/tmp/session.sock")
	netw, addr, err = discoverATSPISocket()
	if err != nil || netw != "unix" || addr != "/tmp/session.sock" {
		t.Fatalf("got net=%s, addr=%s, err=%v", netw, addr, err)
	}

	// 7. invalid address
	_, _, err = parseDBusAddress("tcp:host=localhost,port=1234")
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
