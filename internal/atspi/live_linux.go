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

	"github.com/chinmay-sawant/ownframe/internal/safego"
)

type liveBus struct {
	mu         sync.Mutex
	conn       net.Conn
	reader     *bufio.Reader
	writeMu    sync.Mutex
	connectMu  sync.Mutex
	pending    map[uint32]chan *dbusMsg
	done       chan struct{}
	connDone   chan struct{}
	closeOnce  sync.Once
	wg         sync.WaitGroup
	name       string
	uniqueName string
	addr       string
	claimed    bool
	closed     bool
	reason     string
	serial     uint32
	app        Application
	view       *serverView
	originX    int
	originY    int
	parent     dbusRef
	appID      int32
	lastTry    time.Time
	failAt     time.Time
	embedded   bool
	width      int
	height     int
}

const (
	liveDialTimeout   = 2 * time.Second
	liveCallTimeout   = 2 * time.Second
	liveEmitTimeout   = 500 * time.Millisecond
	liveRetryInterval = 2 * time.Second
	serialGetAddress  = 2
)

func newLiveBus() Bus {
	return &liveBus{
		pending: make(map[uint32]chan *dbusMsg),
		done:    make(chan struct{}),
		reason:  "not connected",
	}
}

// Ready reports whether Publish should attempt to connect or publish.
// A dead bus is retried at most once per liveRetryInterval.
func (b *liveBus) Ready() bool {
	if b == nil {
		return false
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return false
	}
	if !b.failAt.IsZero() && time.Since(b.failAt) < liveRetryInterval {
		return false
	}
	if b.conn != nil && b.claimed {
		return true
	}
	if b.lastTry.IsZero() {
		return true
	}
	return time.Since(b.lastTry) >= liveRetryInterval
}

// refNameLocked is the bus name used in object references. Callers hold b.mu.
func (b *liveBus) refNameLocked() string {
	if b.uniqueName != "" {
		return b.uniqueName
	}
	return b.name
}

func (b *liveBus) ensure() error {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return ErrLiveUnavailable
	}
	if b.conn != nil {
		b.mu.Unlock()
		return nil
	}
	if time.Since(b.lastTry) < liveRetryInterval {
		reason := b.reason
		b.mu.Unlock()
		return fmt.Errorf("%w: %s", ErrLiveUnavailable, reason)
	}
	b.lastTry = time.Now()
	b.mu.Unlock()
	return b.connect()
}

func (b *liveBus) connect() error {
	b.connectMu.Lock()
	defer b.connectMu.Unlock()

	b.mu.Lock()
	if b.conn != nil {
		b.mu.Unlock()
		return nil
	}
	if b.closed {
		b.mu.Unlock()
		return ErrLiveUnavailable
	}
	b.mu.Unlock()

	netw, addr, err := discoverATSPISocket()
	if err != nil {
		b.setReason(err.Error())
		return fmt.Errorf("%w: %w", ErrLiveUnavailable, err)
	}

	conn, err := (&net.Dialer{Timeout: liveDialTimeout}).DialContext(context.Background(), netw, addr)
	if err != nil {
		b.setReason(fmt.Sprintf("connect to %s failed: %v", addr, err))
		return fmt.Errorf("%w: connect %s: %w", ErrLiveUnavailable, addr, err)
	}

	reader, unique, err := greet(conn)
	if err != nil {
		_ = conn.Close()
		b.setReason(err.Error())
		return fmt.Errorf("%w: %w", ErrLiveUnavailable, err)
	}

	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		_ = conn.Close()
		return ErrLiveUnavailable
	}
	b.conn = conn
	b.connDone = make(chan struct{})
	b.reader = reader
	b.uniqueName = unique
	b.addr = addr
	b.serial = 1
	b.reason = "connected to " + addr
	b.embedded = false
	b.wg.Add(1)
	b.mu.Unlock()

	safego.Go("atspi reader", func() { b.readLoop(conn, reader) })
	return nil
}

func (b *liveBus) setReason(reason string) {
	b.mu.Lock()
	b.reason = reason
	b.mu.Unlock()
}

func handshake(conn net.Conn) (*bufio.Reader, error) {
	_ = conn.SetDeadline(time.Now().Add(liveCallTimeout))
	defer func() { _ = conn.SetDeadline(time.Time{}) }()

	if _, err := conn.Write([]byte{0}); err != nil {
		return nil, fmt.Errorf("write null byte: %w", err)
	}

	uidStr := strconv.Itoa(os.Getuid())
	hexUID := hex.EncodeToString([]byte(uidStr))
	authMsg := fmt.Sprintf("AUTH EXTERNAL %s\r\n", hexUID)
	if _, err := conn.Write([]byte(authMsg)); err != nil {
		return nil, fmt.Errorf("write auth: %w", err)
	}

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("read auth response: %w", err)
	}
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "OK") {
		return nil, fmt.Errorf("auth rejected: %s", line)
	}

	if _, err := conn.Write([]byte("BEGIN\r\n")); err != nil {
		return nil, fmt.Errorf("write begin: %w", err)
	}
	return reader, nil
}

// greet performs the auth handshake and Hello, returning the buffered
// reader and the unique bus name. It reads only the matching Hello
// reply and leaves any queued signal for the read loop.
func greet(conn net.Conn) (*bufio.Reader, string, error) {
	reader, err := handshake(conn)
	if err != nil {
		return nil, "", err
	}
	hello := &dbusMsg{
		msgType: dbusTypeMethodCall,
		serial:  1,
		path:    "/org/freedesktop/DBus",
		dest:    "org.freedesktop.DBus",
		iface:   "org.freedesktop.DBus",
		member:  "Hello",
	}
	reply, err := syncCall(conn, reader, hello)
	if err != nil {
		return nil, "", fmt.Errorf("hello failed: %w", err)
	}
	name, err := decodeStringBody(reply.body)
	if err != nil {
		return nil, "", fmt.Errorf("decode hello: %w", err)
	}
	return reader, name, nil
}

func syncCall(conn net.Conn, reader *bufio.Reader, call *dbusMsg) (*dbusMsg, error) {
	_ = conn.SetDeadline(time.Now().Add(liveCallTimeout))
	defer func() { _ = conn.SetDeadline(time.Time{}) }()
	if _, err := conn.Write(encodeMessage(call)); err != nil {
		return nil, err
	}
	for {
		reply, err := readMessage(reader)
		if err != nil {
			return nil, err
		}
		if reply.msgType == dbusTypeSignal {
			continue
		}
		if reply.replySerial != call.serial {
			continue
		}
		if reply.msgType == dbusTypeError {
			return nil, fmt.Errorf("%s: %s", reply.errorName, decodeStringOrEmpty(reply.body))
		}
		if reply.msgType != dbusTypeMethodReturn {
			return nil, fmt.Errorf("unexpected reply type %d", reply.msgType)
		}
		return reply, nil
	}
}

func decodeStringOrEmpty(body []byte) string {
	s, _ := decodeStringBody(body)
	return s
}

func (b *liveBus) readLoop(conn net.Conn, reader *bufio.Reader) {
	defer b.wg.Done()
	for {
		msg, err := readMessage(reader)
		if err != nil {
			b.disconnect(conn, fmt.Sprintf("bus read failed: %v", err))
			return
		}
		switch msg.msgType {
		case dbusTypeMethodReturn, dbusTypeError:
			b.deliver(msg)
		case dbusTypeMethodCall:
			b.handleCall(msg)
		}
	}
}

func (b *liveBus) deliver(msg *dbusMsg) {
	b.mu.Lock()
	ch := b.pending[msg.replySerial]
	delete(b.pending, msg.replySerial)
	b.mu.Unlock()
	if ch == nil {
		return
	}
	select {
	case ch <- msg:
	default:
	}
}

func (b *liveBus) handleCall(msg *dbusMsg) {
	b.mu.Lock()
	view := b.view
	b.mu.Unlock()
	var res callResult
	if view == nil {
		res = errResult("org.freedesktop.DBus.Error.UnknownObject", "application has no published tree")
	} else {
		res = view.handle(msg)
	}
	if msg.flags&dbusFlagNoReplyExpected != 0 {
		return
	}
	b.reply(msg, res)
}

func (b *liveBus) reply(call *dbusMsg, res callResult) {
	msg := &dbusMsg{
		msgType:     dbusTypeMethodReturn,
		replySerial: call.serial,
		dest:        call.sender,
		path:        call.path,
		sig:         res.sig,
		body:        res.body,
	}
	if res.errName != "" {
		msg.msgType = dbusTypeError
		msg.errorName = res.errName
		msg.body = encodeErrorBody(res.errText)
		msg.sig = "s"
	}
	if err := b.send(msg, liveCallTimeout); err != nil {
		b.disconnect(b.connSnapshot(), "reply failed: "+err.Error())
	}
}

func (b *liveBus) call(msg *dbusMsg, timeout time.Duration) (*dbusMsg, error) {
	b.mu.Lock()
	if b.conn == nil || b.closed {
		b.mu.Unlock()
		return nil, ErrLiveUnavailable
	}
	b.serial++
	msg.serial = b.serial
	ch := make(chan *dbusMsg, 1)
	b.pending[msg.serial] = ch
	connDone := b.connDone
	b.mu.Unlock()

	defer func() {
		b.mu.Lock()
		delete(b.pending, msg.serial)
		b.mu.Unlock()
	}()

	if err := b.writeMessage(msg, timeout); err != nil {
		return nil, err
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case reply := <-ch:
		if reply.msgType == dbusTypeError {
			return nil, fmt.Errorf("%s: %s", reply.errorName, decodeStringOrEmpty(reply.body))
		}
		return reply, nil
	case <-timer.C:
		return nil, fmt.Errorf("timed out waiting for %s", msg.member)
	case <-connDone:
		return nil, ErrLiveUnavailable
	case <-b.done:
		return nil, ErrLiveUnavailable
	}
}

func (b *liveBus) send(msg *dbusMsg, timeout time.Duration) error {
	b.mu.Lock()
	if b.conn == nil || b.closed {
		b.mu.Unlock()
		return ErrLiveUnavailable
	}
	b.serial++
	msg.serial = b.serial
	b.mu.Unlock()
	return b.writeMessage(msg, timeout)
}

func (b *liveBus) writeMessage(msg *dbusMsg, timeout time.Duration) error {
	b.mu.Lock()
	conn := b.conn
	b.mu.Unlock()
	if conn == nil {
		return ErrLiveUnavailable
	}
	b.writeMu.Lock()
	defer b.writeMu.Unlock()
	_ = conn.SetWriteDeadline(time.Now().Add(timeout))
	_, err := conn.Write(encodeMessage(msg))
	_ = conn.SetWriteDeadline(time.Time{})
	if err != nil {
		return err
	}
	return nil
}

func (b *liveBus) connSnapshot() net.Conn {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.conn
}

func (b *liveBus) disconnect(conn net.Conn, reason string) {
	b.mu.Lock()
	if b.closed || b.conn != conn {
		b.mu.Unlock()
		return
	}
	done := b.connDone
	b.conn = nil
	b.reader = nil
	b.uniqueName = ""
	b.claimed = false
	b.embedded = false
	b.parent = dbusRef{}
	b.reason = reason
	b.connDone = nil
	b.mu.Unlock()
	if done != nil {
		close(done)
	}
	if conn != nil {
		_ = conn.Close()
	}
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
	}
	return Status{Platform: runtime.GOOS, Live: live, Reason: reason}
}

func (b *liveBus) Claim(name string) error {
	if err := CheckName(name); err != nil {
		return err
	}
	if err := b.ensure(); err != nil {
		return err
	}
	b.mu.Lock()
	if b.closed || b.conn == nil {
		b.mu.Unlock()
		return ErrLiveUnavailable
	}
	if b.claimed && b.name == name {
		b.mu.Unlock()
		return nil
	}
	b.mu.Unlock()
	return b.requestName(name)
}

func (b *liveBus) requestName(name string) error {
	code, err := b.requestNameCode(name)
	if err != nil {
		b.setReason("claim failed: " + err.Error())
		return fmt.Errorf("claim failed: %w", err)
	}
	if code == 1 || code == 4 {
		b.claimedAs(name)
		b.mu.Lock()
		b.failAt = time.Time{}
		b.mu.Unlock()
		return nil
	}
	if name == DefaultName {
		fallback := fmt.Sprintf("%s.p%d", DefaultName, os.Getpid())
		if fcode, ferr := b.requestNameCode(fallback); ferr == nil && (fcode == 1 || fcode == 4) {
			b.claimedAs(fallback)
			b.mu.Lock()
			b.failAt = time.Time{}
			b.mu.Unlock()
			return nil
		}
	}
	b.mu.Lock()
	b.failAt = time.Now()
	b.mu.Unlock()
	b.setReason(fmt.Sprintf("claim failed: name %s returned code %d", name, code))
	return fmt.Errorf("claim failed: name %s returned code %d", name, code)
}

func (b *liveBus) requestNameCode(name string) (uint32, error) {
	msg := &dbusMsg{
		msgType: dbusTypeMethodCall,
		path:    "/org/freedesktop/DBus",
		dest:    "org.freedesktop.DBus",
		iface:   "org.freedesktop.DBus",
		member:  "RequestName",
		sig:     "su",
		body:    encodeRequestNameBody(name, 0),
	}
	reply, err := b.call(msg, liveCallTimeout)
	if err != nil {
		return 0, err
	}
	return decodeUint32Body(reply.body)
}

func (b *liveBus) claimedAs(name string) {
	b.mu.Lock()
	b.name = name
	b.claimed = true
	b.reason = "connected to " + b.addr + " as " + name
	b.mu.Unlock()
}

func (b *liveBus) ensureClaimed() error {
	if err := b.ensure(); err != nil {
		return err
	}
	b.mu.Lock()
	claimed := b.claimed
	name := b.name
	b.mu.Unlock()
	if claimed && name != "" {
		return nil
	}
	return b.Claim(DefaultName)
}

// SetOrigin records the window content origin used for screen coordinates.
func (b *liveBus) SetOrigin(x, y int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed || (b.originX == x && b.originY == y) {
		return
	}
	b.originX, b.originY = x, y
	if b.app.RootPath != "" {
		b.view = newServerView(b.app, b.refNameLocked(), b.parent, b.originX, b.originY, b.width, b.height, &b.appID)
	}
}

// SetSize records the window content size used for the root extents.
func (b *liveBus) SetSize(width, height int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed || (b.width == width && b.height == height) {
		return
	}
	b.width, b.height = width, height
	if b.app.RootPath != "" {
		b.view = newServerView(b.app, b.refNameLocked(), b.parent, b.originX, b.originY, b.width, b.height, &b.appID)
	}
}

func (b *liveBus) Publish(app Application) error {
	if err := b.ensureClaimed(); err != nil {
		return err
	}
	b.mu.Lock()
	if b.conn == nil || b.closed || !b.claimed {
		b.mu.Unlock()
		return ErrLiveUnavailable
	}
	refName := b.refNameLocked()
	b.app = app
	b.view = newServerView(app, refName, b.parent, b.originX, b.originY, b.width, b.height, &b.appID)
	embedded := b.embedded && b.parent.Path != ""
	b.mu.Unlock()
	if embedded {
		return nil
	}

	plug := encodeStructSO(refName, app.RootPath)
	msg := &dbusMsg{
		msgType: dbusTypeMethodCall,
		path:    accessibleRootPath,
		dest:    "org.a11y.atspi.Registry",
		iface:   "org.a11y.atspi.Socket",
		member:  "Embed",
		sig:     "(so)",
		body:    plug,
	}
	reply, err := b.call(msg, liveCallTimeout)
	if err != nil {
		b.mu.Lock()
		b.failAt = time.Now()
		b.mu.Unlock()
		b.setReason("embed failed: " + err.Error())
		return fmt.Errorf("embed failed: %w", err)
	}
	name, path, err := decodeSO(reply.body)
	if err != nil {
		b.setReason("embed reply: " + err.Error())
		return fmt.Errorf("embed reply: %w", err)
	}
	b.mu.Lock()
	b.parent = dbusRef{Name: name, Path: path}
	b.embedded = true
	b.failAt = time.Time{}
	b.view = newServerView(b.app, refName, b.parent, b.originX, b.originY, b.width, b.height, &b.appID)
	b.reason = "connected to " + b.addr + " as " + b.name
	b.mu.Unlock()
	return nil
}

func (b *liveBus) Emit(ev Event) error {
	b.mu.Lock()
	live := b.conn != nil && !b.closed && b.claimed
	name := b.refNameLocked()
	root := b.app.RootPath
	b.mu.Unlock()
	if !live {
		return ErrLiveUnavailable
	}
	sig, body, member, err := encodeEvent(ev, name, root)
	if err != nil {
		return err
	}
	path := ev.Path
	if path == "" {
		path = root
	}
	if path == "" {
		path = accessibleRootPath
	}
	msg := &dbusMsg{
		msgType: dbusTypeSignal,
		path:    path,
		iface:   "org.a11y.atspi.Event.Object",
		member:  member,
		sig:     sig,
		body:    body,
	}
	return b.send(msg, liveEmitTimeout)
}

func encodeEvent(ev Event, busName, root string) (string, []byte, string, error) {
	switch ev.Kind {
	case "children-changed":
		detail := ev.Detail
		if detail == "" {
			detail = "add"
		}
		child := ev.Child
		if child == "" {
			child = ev.Path
		}
		if child == "" {
			child = root
		}
		var e dbusEncoder
		e.stringVal(detail)
		e.int32Val(ev.Index)
		e.int32Val(0)
		e.variant("(so)", func(e *dbusEncoder) {
			e.structValue(func(e *dbusEncoder) {
				e.stringVal(busName)
				e.stringVal(child)
			})
		})
		e.array("{sv}", func(*dbusEncoder) {})
		return "siiva{sv}", e.bytes(), "ChildrenChanged", nil
	case "state-changed":
		detail := ev.Detail
		if detail == "" {
			detail = "focused"
		}
		enabled := int32(0)
		if ev.Enabled {
			enabled = 1
		}
		var e dbusEncoder
		e.stringVal(detail)
		e.int32Val(enabled)
		e.int32Val(0)
		e.variant("i", func(e *dbusEncoder) { e.int32Val(0) })
		e.array("{sv}", func(*dbusEncoder) {})
		return "siiva{sv}", e.bytes(), "StateChanged", nil
	}
	return "", nil, "", fmt.Errorf("atspi: unsupported event %q", ev.Kind)
}

func (b *liveBus) Close() error {
	b.connectMu.Lock()
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		b.connectMu.Unlock()
		b.wg.Wait()
		return nil
	}
	b.closed = true
	b.claimed = false
	b.embedded = false
	b.reason = "bus closed"
	conn := b.conn
	b.conn = nil
	b.reader = nil
	b.uniqueName = ""
	b.parent = dbusRef{}
	connDone := b.connDone
	b.connDone = nil
	b.closeOnce.Do(func() { close(b.done) })
	if connDone != nil {
		close(connDone)
	}
	b.mu.Unlock()
	b.connectMu.Unlock()
	if conn != nil {
		_ = conn.Close()
	}
	b.wg.Wait()
	return nil
}

func discoverATSPISocket() (string, string, error) {
	addr := os.Getenv("AT_SPI_BUS_ADDRESS")
	if addr == "disabled" || addr == "none" {
		return "", "", errors.New("at-spi bus disabled by environment")
	}
	if addr != "" {
		return parseDBusAddress(addr)
	}
	if a11y, err := sessionA11yAddress(); err == nil && a11y != "" {
		return parseDBusAddress(a11y)
	}
	uid := os.Getuid()
	if runtime := os.Getenv("XDG_RUNTIME_DIR"); runtime != "" {
		matches, _ := filepath.Glob(filepath.Join(runtime, "at-spi", "bus*"))
		if len(matches) > 0 {
			return parseDBusAddress("unix:path=" + matches[0])
		}
	}
	matches, _ := filepath.Glob(fmt.Sprintf("/run/user/%d/at-spi/bus*", uid))
	if len(matches) > 0 {
		return parseDBusAddress("unix:path=" + matches[0])
	}
	return "", "", errors.New(liveReason())
}

// sessionA11yAddress asks the session bus for the accessibility bus
// address. The two buses are separate; DBUS_SESSION_BUS_ADDRESS alone
// is not an AT-SPI bus.
func sessionA11yAddress() (string, error) {
	session := os.Getenv("DBUS_SESSION_BUS_ADDRESS")
	if session == "" {
		return "", errors.New("no session bus address")
	}
	netw, addr, err := parseDBusAddress(session)
	if err != nil {
		return "", err
	}
	conn, err := (&net.Dialer{Timeout: liveDialTimeout}).DialContext(context.Background(), netw, addr)
	if err != nil {
		return "", err
	}
	defer func() { _ = conn.Close() }()

	reader, _, err := greet(conn)
	if err != nil {
		return "", err
	}
	msg := &dbusMsg{
		msgType: dbusTypeMethodCall,
		serial:  serialGetAddress,
		path:    "/org/a11y/bus",
		dest:    "org.a11y.Bus",
		iface:   "org.a11y.Bus",
		member:  "GetAddress",
	}
	reply, err := syncCall(conn, reader, msg)
	if err != nil {
		return "", err
	}
	return decodeStringBody(reply.body)
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
