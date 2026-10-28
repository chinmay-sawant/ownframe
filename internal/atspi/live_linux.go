//go:build linux && !android && !ios && !js

package atspi

import (
	"bufio"
	"context"
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
	"time"
)

type liveBus struct {
	mu         sync.Mutex
	conn       net.Conn
	reader     *bufio.Reader
	addr       string
	name       string
	uniqueName string
	claimed    bool
	closed     bool
	reason     string
	serial     uint32
	app        Application
}

const (
	liveDialTimeout = 2 * time.Second
	liveCallTimeout = 2 * time.Second
	liveEmitTimeout = 500 * time.Millisecond
)

func newLiveBus() Bus {
	netw, addr, err := discoverATSPISocket()
	if err != nil {
		return &liveBus{reason: err.Error()}
	}

	conn, err := (&net.Dialer{Timeout: liveDialTimeout}).DialContext(context.Background(), netw, addr)
	if err != nil {
		return &liveBus{reason: fmt.Sprintf("connect to %s failed: %v", addr, err)}
	}

	b := &liveBus{
		conn: conn,
		addr: addr,
	}

	if err := b.handshake(conn); err != nil {
		_ = conn.Close()
		return &liveBus{reason: fmt.Sprintf("handshake failed: %v", err)}
	}

	if err := b.callHello(); err != nil {
		_ = conn.Close()
		return &liveBus{reason: fmt.Sprintf("hello failed: %v", err)}
	}

	if err := b.requestName(DefaultName); err != nil {
		_ = conn.Close()
		return &liveBus{reason: fmt.Sprintf("claim failed: %v", err)}
	}

	b.reason = "connected to " + addr
	return b
}

func discoverATSPISocket() (string, string, error) {
	addr := os.Getenv("AT_SPI_BUS_ADDRESS")
	if addr == "disabled" || addr == "none" {
		return "", "", errors.New("at-spi bus disabled by environment")
	}
	if addr == "" {
		addr = os.Getenv("DBUS_SESSION_BUS_ADDRESS")
	}
	if addr == "" {
		uid := os.Getuid()
		matches, _ := filepath.Glob(fmt.Sprintf("/run/user/%d/at-spi/bus*", uid))
		if len(matches) > 0 {
			addr = "unix:path=" + matches[0]
		}
	}
	if addr == "" {
		return "", "", errors.New(liveReason())
	}
	return parseDBusAddress(addr)
}

func parseDBusAddress(addr string) (string, string, error) {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return "", "", errors.New("empty bus address")
	}
	if strings.HasPrefix(addr, "unix:") {
		rest := addr[len("unix:"):]
		parts := strings.Split(rest, ",")
		for _, part := range parts {
			if strings.HasPrefix(part, "path=") {
				return "unix", part[len("path="):], nil
			}
			if strings.HasPrefix(part, "abstract=") {
				return "unix", "@" + part[len("abstract="):], nil
			}
		}
		return "", "", fmt.Errorf("no path or abstract in unix address: %s", addr)
	}
	if strings.HasPrefix(addr, "/") {
		return "unix", addr, nil
	}
	return "", "", fmt.Errorf("unsupported bus address: %s", addr)
}

func (b *liveBus) handshake(conn net.Conn) error {
	_ = conn.SetDeadline(time.Now().Add(liveCallTimeout))
	defer func() { _ = conn.SetDeadline(time.Time{}) }()

	if _, err := conn.Write([]byte{0}); err != nil {
		return fmt.Errorf("write null byte: %w", err)
	}

	uidStr := strconv.Itoa(os.Getuid())
	hexUID := hex.EncodeToString([]byte(uidStr))
	authMsg := fmt.Sprintf("AUTH EXTERNAL %s\r\n", hexUID)
	if _, err := conn.Write([]byte(authMsg)); err != nil {
		return fmt.Errorf("write auth: %w", err)
	}

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("read auth response: %w", err)
	}
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "OK") {
		return fmt.Errorf("auth rejected: %s", line)
	}

	if _, err := conn.Write([]byte("BEGIN\r\n")); err != nil {
		return fmt.Errorf("write begin: %w", err)
	}

	b.reader = reader
	return nil
}

func (b *liveBus) callHello() error {
	b.serial++
	msg := &dbusMsg{
		msgType: dbusTypeMethodCall,
		serial:  b.serial,
		path:    "/org/freedesktop/DBus",
		dest:    "org.freedesktop.DBus",
		iface:   "org.freedesktop.DBus",
		member:  "Hello",
	}
	raw := encodeMessage(msg)
	_ = b.conn.SetDeadline(time.Now().Add(liveCallTimeout))
	defer func() { _ = b.conn.SetDeadline(time.Time{}) }()

	if _, err := b.conn.Write(raw); err != nil {
		return fmt.Errorf("write hello: %w", err)
	}

	reply, err := readMessage(b.reader)
	if err != nil {
		return fmt.Errorf("read hello reply: %w", err)
	}
	if reply.msgType == dbusTypeError {
		return fmt.Errorf("hello error: %s", reply.errorName)
	}
	if reply.msgType != dbusTypeMethodReturn {
		return fmt.Errorf("unexpected hello reply type: %d", reply.msgType)
	}
	name, _ := decodeStringBody(reply.body)
	b.uniqueName = name
	return nil
}

func (b *liveBus) requestName(name string) error {
	b.serial++
	msg := &dbusMsg{
		msgType: dbusTypeMethodCall,
		serial:  b.serial,
		path:    "/org/freedesktop/DBus",
		dest:    "org.freedesktop.DBus",
		iface:   "org.freedesktop.DBus",
		member:  "RequestName",
		sig:     "su",
		body:    encodeRequestNameBody(name, 0),
	}
	raw := encodeMessage(msg)
	_ = b.conn.SetDeadline(time.Now().Add(liveCallTimeout))
	defer func() { _ = b.conn.SetDeadline(time.Time{}) }()

	if _, err := b.conn.Write(raw); err != nil {
		return fmt.Errorf("write request name: %w", err)
	}

	reply, err := readMessage(b.reader)
	if err != nil {
		return fmt.Errorf("read request name reply: %w", err)
	}
	if reply.msgType == dbusTypeError {
		return fmt.Errorf("request name error: %s", reply.errorName)
	}
	if reply.msgType != dbusTypeMethodReturn {
		return fmt.Errorf("unexpected request name reply type: %d", reply.msgType)
	}
	retCode, err := decodeUint32Body(reply.body)
	if err != nil {
		return fmt.Errorf("decode request name body: %w", err)
	}
	if retCode != 1 && retCode != 4 {
		return fmt.Errorf("request name returned code %d", retCode)
	}
	b.name = name
	b.claimed = true
	return nil
}

func (b *liveBus) Live() bool {
	if b == nil {
		return false
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.conn != nil && !b.closed && b.claimed
}

func (b *liveBus) Status() Status {
	if b == nil {
		return Status{Platform: runtime.GOOS, Live: false, Reason: "no bus"}
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	live := b.conn != nil && !b.closed && b.claimed
	reason := b.reason
	if b.closed {
		live = false
		reason = "bus closed"
	}
	return Status{Platform: runtime.GOOS, Live: live, Reason: reason}
}

func (b *liveBus) Claim(name string) error {
	if err := CheckName(name); err != nil {
		return err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.conn == nil || b.closed {
		return ErrLiveUnavailable
	}
	if b.claimed && b.name == name {
		return nil
	}
	return b.requestName(name)
}

func (b *liveBus) Publish(app Application) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.conn == nil || b.closed || !b.claimed {
		return ErrLiveUnavailable
	}
	b.app = app

	appName := app.Name
	if appName == "" {
		appName = DefaultName
	}
	rootPath := app.RootPath
	if rootPath == "" {
		rootPath = "/org/a11y/atspi/accessible/root"
	}

	b.serial++
	msg := &dbusMsg{
		msgType: dbusTypeMethodCall,
		flags:   dbusFlagNoReplyExpected,
		serial:  b.serial,
		path:    "/org/a11y/atspi/registry",
		dest:    "org.a11y.atspi.Registry",
		iface:   "org.a11y.atspi.Socket",
		member:  "Embed",
		sig:     "(so)",
		body:    encodeEmbedBody(appName, rootPath),
	}
	raw := encodeMessage(msg)
	_ = b.conn.SetDeadline(time.Now().Add(liveEmitTimeout))
	defer func() { _ = b.conn.SetDeadline(time.Time{}) }()
	if _, err := b.conn.Write(raw); err != nil {
		b.closed = true
		b.claimed = false
		b.reason = "connection error: " + err.Error()
		_ = b.conn.Close()
		return err
	}
	return nil
}

func (b *liveBus) Emit(ev Event) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.conn == nil || b.closed || !b.claimed {
		return ErrLiveUnavailable
	}
	member := "ChildrenChanged"
	if ev.Kind != "" && ev.Kind != "children-changed" {
		member = kindToMember(ev.Kind)
	}
	path := ev.Path
	if path == "" {
		path = "/org/a11y/atspi/accessible/root"
	}

	b.serial++
	msg := &dbusMsg{
		msgType: dbusTypeSignal,
		flags:   dbusFlagNoReplyExpected,
		serial:  b.serial,
		path:    path,
		iface:   "org.a11y.atspi.Event.Object",
		member:  member,
		sig:     "s",
		body:    encodeStringBody(ev.Kind),
	}
	raw := encodeMessage(msg)
	_ = b.conn.SetDeadline(time.Now().Add(liveEmitTimeout))
	defer func() { _ = b.conn.SetDeadline(time.Time{}) }()
	if _, err := b.conn.Write(raw); err != nil {
		b.closed = true
		b.claimed = false
		b.reason = "connection error: " + err.Error()
		_ = b.conn.Close()
		return err
	}
	return nil
}

func (b *liveBus) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return nil
	}
	b.closed = true
	b.claimed = false
	if b.conn != nil {
		err := b.conn.Close()
		b.conn = nil
		return err
	}
	return nil
}

func kindToMember(kind string) string {
	parts := strings.Split(kind, "-")
	var sb strings.Builder
	for _, p := range parts {
		if len(p) > 0 {
			sb.WriteString(strings.ToUpper(p[:1]) + p[1:])
		}
	}
	return sb.String()
}
